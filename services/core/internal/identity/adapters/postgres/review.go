package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/adapters/postgres/sqlcgen"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/domain"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/ports"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/db"
)

// Reviews implements ports.ReviewRepository. The application tables carry no policy (ADR-0003).
type Reviews struct {
	pool *pgxpool.Pool
}

var _ ports.ReviewRepository = (*Reviews)(nil)

func NewReviews(pool *pgxpool.Pool) *Reviews { return &Reviews{pool: pool} }

func optUUID(u pgtype.UUID) *uuid.UUID {
	if !u.Valid {
		return nil
	}
	id := uuid.UUID(u.Bytes)
	return &id
}

func optText(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}

func reviewApplication(id pgtype.UUID, kind, edrpou string, assoc pgtype.UUID, state string, created pgtype.Timestamptz, name string, filed pgtype.Timestamptz, reason pgtype.Text) ports.ReviewApplication {
	return ports.ReviewApplication{
		ID: id.Bytes, Kind: kind, Edrpou: edrpou, AssociationID: optUUID(assoc), State: state,
		CreatedAt: created.Time, AssociationName: name, FiledAt: filed.Time, RejectionReason: optText(reason),
	}
}

func (s *Reviews) List(ctx context.Context, state string, after *ports.AssociationCursor, limit int) ([]ports.ReviewApplication, error) {
	q := sqlcgen.New(s.pool)
	var cAt pgtype.Timestamptz
	var cID pgtype.UUID
	if after != nil {
		cAt, cID = pgtype.Timestamptz{Time: after.CreatedAt, Valid: true}, pgUUID(after.ID)
	}
	size := int32(limit) //nolint:gosec // limit is 1..101
	var out []ports.ReviewApplication
	if state == string(domain.StatePending) {
		rows, err := q.ListPendingApplications(ctx, sqlcgen.ListPendingApplicationsParams{State: state, CursorCreatedAt: cAt, CursorID: cID, PageSize: size})
		if err != nil {
			return nil, err
		}
		for _, r := range rows {
			out = append(out, reviewApplication(r.ID, r.Kind, r.Edrpou, r.AssociationID, r.State, r.CreatedAt, r.AssociationName, r.FiledAt, r.RejectionReason))
		}
		return out, nil
	}
	rows, err := q.ListDecidedApplications(ctx, sqlcgen.ListDecidedApplicationsParams{State: state, CursorCreatedAt: cAt, CursorID: cID, PageSize: size})
	if err != nil {
		return nil, err
	}
	for _, r := range rows {
		out = append(out, reviewApplication(r.ID, r.Kind, r.Edrpou, r.AssociationID, r.State, r.CreatedAt, r.AssociationName, r.FiledAt, r.RejectionReason))
	}
	return out, nil
}

func (s *Reviews) Get(ctx context.Context, id uuid.UUID) (ports.ReviewDetail, error) {
	q := sqlcgen.New(s.pool)
	pid := pgUUID(id)
	a, err := q.GetApplication(ctx, pid)
	if errors.Is(err, pgx.ErrNoRows) {
		return ports.ReviewDetail{}, ports.ErrApplicationNotFound
	}
	if err != nil {
		return ports.ReviewDetail{}, err
	}
	d := ports.ReviewDetail{
		Application:    reviewApplication(a.ID, a.Kind, a.Edrpou, a.AssociationID, a.State, a.CreatedAt, a.AssociationName, a.FiledAt, a.RejectionReason),
		ApplicantEmail: a.ApplicantEmail,
		Filings:        []ports.ReviewFiling{}, Checks: []ports.ReviewCheck{}, Decisions: []ports.ReviewDecision{}, Competing: []uuid.UUID{},
	}
	filings, err := q.ListFilingsOfApplication(ctx, pid)
	if err != nil {
		return d, err
	}
	for _, f := range filings {
		rf := ports.ReviewFiling{
			ID: f.ID.Bytes, AssociationName: f.AssociationName, AssociationAddress: f.AssociationAddress,
			ApplicantFullName: f.ApplicantFullName, Position: f.Position, TermStart: f.TermStart.Time,
			CreatedAt: f.CreatedAt.Time, ExtractID: f.ExtractID.Bytes, ExtractMediaType: f.MediaType,
			ExtractName: f.OriginalName, ExtractSha256: f.Sha256, ExtractSizeBytes: f.SizeBytes,
		}
		if f.TermEnd.Valid {
			t := f.TermEnd.Time
			rf.TermEnd = &t
		}
		d.Filings = append(d.Filings, rf)
	}
	checks, err := q.ListChecksOfApplication(ctx, pid)
	if err != nil {
		return d, err
	}
	for _, c := range checks {
		d.Checks = append(d.Checks, ports.ReviewCheck{
			ID: c.ID.Bytes, FilingID: c.FilingID.Bytes, SignatureValid: c.SignatureValid, SignerFullName: c.SignerFullName,
			SignedOn: c.SignedOn.Time, HeadFullName: c.HeadFullName, ExtractEdrpou: c.ExtractEdrpou,
			ExtractDatedOn: c.ExtractDatedOn.Time, CheckedAt: c.CreatedAt.Time,
		})
	}
	decisions, err := q.ListDecisionsOfApplication(ctx, pid)
	if err != nil {
		return d, err
	}
	for _, x := range decisions {
		d.Decisions = append(d.Decisions, ports.ReviewDecision{
			ID: x.ID.Bytes, FilingID: x.FilingID.Bytes, AssociationID: optUUID(x.AssociationID), Outcome: x.Outcome,
			Reason: optText(x.Reason), ReliedOnDocument: optText(x.ReliedOnDocument), DecidedAt: x.CreatedAt.Time,
		})
	}
	competing, err := q.ListCompetingApplications(ctx, pid)
	if err != nil {
		return d, err
	}
	for _, c := range competing {
		d.Competing = append(d.Competing, c.Bytes)
	}
	if a.Kind == string(domain.KindTransfer) && a.AssociationID.Valid {
		if d.StorageAccountEmail, err = s.storageAccountEmail(ctx, a.AssociationID); err != nil {
			return d, err
		}
	}
	return d, nil
}

// storageAccountEmail returns only the address of the association's connection, read inside the
// association (ADR-0003); nil when the association has no connection.
func (s *Reviews) storageAccountEmail(ctx context.Context, associationID pgtype.UUID) (*string, error) {
	var email *string
	err := db.InAssociation(ctx, s.pool, associationID.Bytes, "", func(ctx context.Context) error {
		tx, err := db.From(ctx)
		if err != nil {
			return err
		}
		row, err := sqlcgen.New(tx).GetGoogleConnection(ctx, associationID)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		email = &row.GoogleAccountEmail
		return nil
	})
	return email, err
}

func (s *Reviews) Extract(ctx context.Context, applicationID, filingID uuid.UUID) (ports.ReviewExtract, error) {
	e, err := sqlcgen.New(s.pool).GetExtractOfFiling(ctx, sqlcgen.GetExtractOfFilingParams{FilingID: pgUUID(filingID), ApplicationID: pgUUID(applicationID)})
	if errors.Is(err, pgx.ErrNoRows) {
		return ports.ReviewExtract{}, ports.ErrApplicationNotFound
	}
	return ports.ReviewExtract{Content: e.Content, MediaType: e.MediaType, OriginalName: e.OriginalName}, err
}

func (s *Reviews) RecordCheck(ctx context.Context, c ports.NewCheck) (ports.ReviewCheck, error) {
	q := sqlcgen.New(s.pool)
	if _, err := q.FilingBelongsToApplication(ctx, sqlcgen.FilingBelongsToApplicationParams{FilingID: pgUUID(c.FilingID), ApplicationID: pgUUID(c.ApplicationID)}); errors.Is(err, pgx.ErrNoRows) {
		return ports.ReviewCheck{}, ports.ErrApplicationNotFound
	} else if err != nil {
		return ports.ReviewCheck{}, err
	}
	id, err := uuid.NewV7()
	if err != nil {
		return ports.ReviewCheck{}, err
	}
	at, err := q.CreateCheck(ctx, sqlcgen.CreateCheckParams{
		ID: pgUUID(id), FilingID: pgUUID(c.FilingID), SignatureValid: c.SignatureValid, SignerFullName: c.SignerFullName,
		SignedOn: pgtype.Date{Time: c.SignedOn, Valid: true}, HeadFullName: c.HeadFullName, ExtractEdrpou: c.ExtractEdrpou,
		ExtractDatedOn: pgtype.Date{Time: c.ExtractDatedOn, Valid: true}, CheckedByEmail: c.CheckedByEmail,
	})
	if err != nil {
		return ports.ReviewCheck{}, err
	}
	return ports.ReviewCheck{
		ID: id, FilingID: c.FilingID, SignatureValid: c.SignatureValid, SignerFullName: c.SignerFullName, SignedOn: c.SignedOn,
		HeadFullName: c.HeadFullName, ExtractEdrpou: c.ExtractEdrpou, ExtractDatedOn: c.ExtractDatedOn, CheckedAt: at.Time,
	}, nil
}
