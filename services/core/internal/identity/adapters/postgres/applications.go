package postgres

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/adapters/postgres/sqlcgen"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/ports"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/db"
)

// Applications implements ports.ApplicationRepository.
type Applications struct {
	pool *pgxpool.Pool
}

var _ ports.ApplicationRepository = (*Applications)(nil)

func NewApplications(pool *pgxpool.Pool) *Applications { return &Applications{pool: pool} }

func pgUUID(id uuid.UUID) pgtype.UUID { return pgtype.UUID{Bytes: id, Valid: true} }

func (a *Applications) File(ctx context.Context, f ports.NewFiling) (ports.FiledApplication, error) {
	tx, err := a.pool.Begin(ctx)
	if err != nil {
		return ports.FiledApplication{}, err
	}
	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()
	q := sqlcgen.New(tx)

	if _, err := q.GetAssociationByEdrpou(ctx, pgtype.Text{String: f.Edrpou, Valid: true}); err == nil {
		return ports.FiledApplication{}, ports.ErrAssociationConnected
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return ports.FiledApplication{}, err
	}
	if _, err := q.FindOpenApplicationOfAccount(ctx, sqlcgen.FindOpenApplicationOfAccountParams{AccountID: pgUUID(f.AccountID), Edrpou: f.Edrpou}); err == nil {
		return ports.FiledApplication{}, ports.ErrApplicationOpen
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return ports.FiledApplication{}, err
	}

	out, err := store(ctx, q, f, nil)
	if err != nil {
		return out, err
	}
	if err := tx.Commit(ctx); err != nil {
		return ports.FiledApplication{}, fmt.Errorf("commit: %w", err)
	}
	return out, nil
}

// FileTransfer implements ports.ApplicationRepository.
func (a *Applications) FileTransfer(ctx context.Context, f ports.NewTransfer) (ports.FiledApplication, error) {
	if f.AssociationID != nil {
		return a.fileFromCard(ctx, f)
	}
	tx, err := a.pool.Begin(ctx)
	if err != nil {
		return ports.FiledApplication{}, err
	}
	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()
	q := sqlcgen.New(tx)
	id, err := q.GetAssociationByEdrpou(ctx, pgtype.Text{String: f.Edrpou, Valid: true})
	if errors.Is(err, pgx.ErrNoRows) {
		return ports.FiledApplication{}, ports.ErrAssociationNotConnected
	}
	if err != nil {
		return ports.FiledApplication{}, err
	}
	assoc := uuid.UUID(id.Bytes)
	out, err := storeChecked(ctx, q, f.NewFiling, &assoc)
	if err != nil {
		return out, err
	}
	if err := tx.Commit(ctx); err != nil {
		return ports.FiledApplication{}, fmt.Errorf("commit: %w", err)
	}
	return out, nil
}

// fileFromCard files for the association of the card; only its current member may, and anyone else
// learns nothing but ErrAssociationNotConnected.
func (a *Applications) fileFromCard(ctx context.Context, f ports.NewTransfer) (ports.FiledApplication, error) {
	var out ports.FiledApplication
	if f.Roster == nil {
		return out, ports.ErrAssociationNotConnected
	}
	assoc := *f.AssociationID
	err := db.InAssociation(ctx, a.pool, assoc, f.ApplicantEmail, func(ctx context.Context) error {
		ok, err := f.Roster.CurrentMember(ctx, assoc, f.ApplicantEmail, time.Now().UTC().Truncate(24*time.Hour))
		if err != nil {
			return err
		}
		if !ok {
			return ports.ErrAssociationNotConnected
		}
		tx, err := db.From(ctx)
		if err != nil {
			return err
		}
		out, err = storeChecked(ctx, sqlcgen.New(tx), f.NewFiling, &assoc)
		return err
	})
	return out, err
}

func storeChecked(ctx context.Context, q *sqlcgen.Queries, f ports.NewFiling, assoc *uuid.UUID) (ports.FiledApplication, error) {
	if _, err := q.FindOpenApplicationOfAccount(ctx, sqlcgen.FindOpenApplicationOfAccountParams{AccountID: pgUUID(f.AccountID), Edrpou: f.Edrpou}); err == nil {
		return ports.FiledApplication{}, ports.ErrApplicationOpen
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return ports.FiledApplication{}, err
	}
	return store(ctx, q, f, assoc)
}

func store(ctx context.Context, q *sqlcgen.Queries, f ports.NewFiling, assoc *uuid.UUID) (ports.FiledApplication, error) {
	var err error
	ids := [3]uuid.UUID{}
	for i := range ids {
		if ids[i], err = uuid.NewV7(); err != nil {
			return ports.FiledApplication{}, err
		}
	}
	extractID, appID, filingID := ids[0], ids[1], ids[2]
	sum := sha256.Sum256(f.ExtractContent)
	if err := q.CreateExtract(ctx, sqlcgen.CreateExtractParams{
		ID: pgUUID(extractID), AccountID: pgUUID(f.AccountID), Content: f.ExtractContent, Sha256: sum[:],
		SizeBytes: int64(len(f.ExtractContent)), MediaType: f.ExtractMediaType, OriginalName: f.ExtractName,
	}); err != nil {
		return ports.FiledApplication{}, err
	}
	appCreated, err := q.CreateApplication(ctx, sqlcgen.CreateApplicationParams{
		ID: pgUUID(appID), Kind: f.Kind, AccountID: pgUUID(f.AccountID), Edrpou: f.Edrpou, State: "pending",
		AssociationID: assocOf(assoc),
	})
	if err != nil {
		return ports.FiledApplication{}, mapOpen(err)
	}
	termEnd := pgtype.Date{}
	if f.TermEnd != nil {
		termEnd = pgtype.Date{Time: *f.TermEnd, Valid: true}
	}
	filed, err := q.CreateFiling(ctx, sqlcgen.CreateFilingParams{
		ID: pgUUID(filingID), ApplicationID: pgUUID(appID), ExtractID: pgUUID(extractID),
		AssociationName: f.AssociationName, AssociationAddress: f.AssociationAddress,
		ApplicantFullName: f.ApplicantFullName, Position: f.Position,
		TermStart: pgtype.Date{Time: f.TermStart, Valid: true}, TermEnd: termEnd,
	})
	if err != nil {
		return ports.FiledApplication{}, err
	}
	return ports.FiledApplication{ID: appID, CreatedAt: appCreated.Time, FiledAt: filed.Time}, nil
}

func assocOf(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{}
	}
	return pgUUID(*id)
}

// mapOpen turns the unique-index race of two concurrent filings into ErrApplicationOpen.
func mapOpen(err error) error {
	var pg *pgconn.PgError
	if errors.As(err, &pg) && pg.Code == "23505" {
		return ports.ErrApplicationOpen
	}
	return err
}

func optDate(d pgtype.Date) *time.Time {
	if !d.Valid {
		return nil
	}
	return &d.Time
}

// viewOf builds an application view; the association an application names is shown only once approved.
func viewOf(id pgtype.UUID, kind, edrpou string, assoc pgtype.UUID, state string, created pgtype.Timestamptz, name string, filed pgtype.Timestamptz, reason pgtype.Text) ports.ApplicationView {
	v := ports.ApplicationView{
		ID: uuid.UUID(id.Bytes), Kind: kind, State: state, Edrpou: edrpou, AssociationName: name,
		CreatedAt: created.Time, FiledAt: filed.Time, RejectionReason: optText(reason),
	}
	if state == "approved" {
		v.AssociationID = optUUID(assoc)
	}
	return v
}

// ListOfAccount implements ports.ApplicationRepository.
func (a *Applications) ListOfAccount(ctx context.Context, account uuid.UUID, after *ports.ApplicationCursor, limit int) ([]ports.ApplicationView, error) {
	p := sqlcgen.ListApplicationsOfAccountParams{AccountID: pgUUID(account), PageSize: int32(limit)}
	if after != nil {
		p.HasCursor = true
		p.CursorCreatedAt = pgtype.Timestamptz{Time: after.CreatedAt, Valid: true}
		p.CursorID = pgUUID(after.ID)
	}
	rows, err := sqlcgen.New(a.pool).ListApplicationsOfAccount(ctx, p)
	if err != nil {
		return nil, err
	}
	out := make([]ports.ApplicationView, len(rows))
	for i, r := range rows {
		out[i] = viewOf(r.ID, r.Kind, r.Edrpou, r.AssociationID, r.State, r.CreatedAt, r.AssociationName, r.FiledAt, r.RejectionReason)
	}
	return out, nil
}

// GetOfAccount implements ports.ApplicationRepository.
func (a *Applications) GetOfAccount(ctx context.Context, account, id uuid.UUID) (ports.ApplicationDetail, error) {
	q := sqlcgen.New(a.pool)
	r, err := q.GetApplicationOfAccount(ctx, sqlcgen.GetApplicationOfAccountParams{ID: pgUUID(id), AccountID: pgUUID(account)})
	if errors.Is(err, pgx.ErrNoRows) {
		return ports.ApplicationDetail{}, ports.ErrApplicationNotFound
	}
	if err != nil {
		return ports.ApplicationDetail{}, err
	}
	d := ports.ApplicationDetail{Application: viewOf(r.ID, r.Kind, r.Edrpou, r.AssociationID, r.State, r.CreatedAt, r.AssociationName, r.FiledAt, r.RejectionReason)}
	filings, err := q.ListFilings(ctx, pgUUID(id))
	if err != nil {
		return d, err
	}
	d.Filings = make([]ports.FilingView, len(filings))
	for i, f := range filings {
		d.Filings[i] = ports.FilingView{
			ID: uuid.UUID(f.ID.Bytes), AssociationName: f.AssociationName, AssociationAddress: f.AssociationAddress,
			ApplicantFullName: f.ApplicantFullName, Position: f.Position, TermStart: f.TermStart.Time,
			TermEnd: optDate(f.TermEnd), CreatedAt: f.CreatedAt.Time, ExtractID: uuid.UUID(f.ExtractID.Bytes),
			MediaType: f.MediaType, OriginalName: f.OriginalName, Sha256: f.Sha256, SizeBytes: f.SizeBytes,
		}
	}
	decisions, err := q.ListDecisions(ctx, pgUUID(id))
	if err != nil {
		return d, err
	}
	d.Decisions = make([]ports.DecisionView, len(decisions))
	for i, x := range decisions {
		d.Decisions[i] = ports.DecisionView{
			ID: uuid.UUID(x.ID.Bytes), FilingID: uuid.UUID(x.FilingID.Bytes), AssociationID: optUUID(x.AssociationID),
			Outcome: x.Outcome, Reason: optText(x.Reason), ReliedOnDocument: optText(x.ReliedOnDocument), CreatedAt: x.CreatedAt.Time,
		}
	}
	return d, nil
}

// Refile implements ports.ApplicationRepository.
func (a *Applications) Refile(ctx context.Context, account, id uuid.UUID, prepare func(ports.StoredApplication) (ports.NewFiling, error)) (ports.ApplicationView, error) {
	tx, err := a.pool.Begin(ctx)
	if err != nil {
		return ports.ApplicationView{}, err
	}
	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()
	q := sqlcgen.New(tx)
	row, err := q.LockApplicationOfAccount(ctx, sqlcgen.LockApplicationOfAccountParams{ID: pgUUID(id), AccountID: pgUUID(account)})
	if errors.Is(err, pgx.ErrNoRows) {
		return ports.ApplicationView{}, ports.ErrApplicationNotFound
	}
	if err != nil {
		return ports.ApplicationView{}, err
	}
	f, err := prepare(ports.StoredApplication{Kind: row.Kind, State: row.State, Edrpou: row.Edrpou})
	if err != nil {
		return ports.ApplicationView{}, err
	}
	if row.Kind == "connection" {
		if _, err := q.GetAssociationByEdrpou(ctx, pgtype.Text{String: row.Edrpou, Valid: true}); err == nil {
			return ports.ApplicationView{}, ports.ErrAssociationConnected
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return ports.ApplicationView{}, err
		}
	}
	ids := [2]uuid.UUID{}
	for i := range ids {
		if ids[i], err = uuid.NewV7(); err != nil {
			return ports.ApplicationView{}, err
		}
	}
	sum := sha256.Sum256(f.ExtractContent)
	if err := q.CreateExtract(ctx, sqlcgen.CreateExtractParams{
		ID: pgUUID(ids[0]), AccountID: pgUUID(account), Content: f.ExtractContent, Sha256: sum[:],
		SizeBytes: int64(len(f.ExtractContent)), MediaType: f.ExtractMediaType, OriginalName: f.ExtractName,
	}); err != nil {
		return ports.ApplicationView{}, err
	}
	termEnd := pgtype.Date{}
	if f.TermEnd != nil {
		termEnd = pgtype.Date{Time: *f.TermEnd, Valid: true}
	}
	if _, err := q.CreateFiling(ctx, sqlcgen.CreateFilingParams{
		ID: pgUUID(ids[1]), ApplicationID: pgUUID(id), ExtractID: pgUUID(ids[0]),
		AssociationName: f.AssociationName, AssociationAddress: f.AssociationAddress,
		ApplicantFullName: f.ApplicantFullName, Position: f.Position,
		TermStart: pgtype.Date{Time: f.TermStart, Valid: true}, TermEnd: termEnd,
	}); err != nil {
		return ports.ApplicationView{}, err
	}
	if err := q.UpdateApplicationState(ctx, sqlcgen.UpdateApplicationStateParams{ID: pgUUID(id), State: "pending"}); err != nil {
		return ports.ApplicationView{}, mapOpen(err)
	}
	r, err := q.GetApplicationOfAccount(ctx, sqlcgen.GetApplicationOfAccountParams{ID: pgUUID(id), AccountID: pgUUID(account)})
	if err != nil {
		return ports.ApplicationView{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return ports.ApplicationView{}, fmt.Errorf("commit: %w", err)
	}
	return viewOf(r.ID, r.Kind, r.Edrpou, r.AssociationID, r.State, r.CreatedAt, r.AssociationName, r.FiledAt, r.RejectionReason), nil
}

// GetExtract implements ports.ApplicationRepository.
func (a *Applications) GetExtract(ctx context.Context, account, application, filing uuid.UUID) (ports.StoredExtract, error) {
	r, err := sqlcgen.New(a.pool).GetExtract(ctx, sqlcgen.GetExtractParams{FilingID: pgUUID(filing), ApplicationID: pgUUID(application), AccountID: pgUUID(account)})
	if errors.Is(err, pgx.ErrNoRows) {
		return ports.StoredExtract{}, ports.ErrApplicationNotFound
	}
	if err != nil {
		return ports.StoredExtract{}, err
	}
	return ports.StoredExtract{Name: r.OriginalName, MediaType: r.MediaType, Content: r.Content, Sha256: r.Sha256}, nil
}
