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
