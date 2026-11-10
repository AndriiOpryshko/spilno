package postgres

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/adapters/postgres/sqlcgen"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/ports"
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
	if err := tx.Commit(ctx); err != nil {
		return ports.FiledApplication{}, fmt.Errorf("commit: %w", err)
	}
	return ports.FiledApplication{ID: appID, CreatedAt: appCreated.Time, FiledAt: filed.Time}, nil
}

// mapOpen turns the unique-index race of two concurrent filings into ErrApplicationOpen.
func mapOpen(err error) error {
	var pg *pgconn.PgError
	if errors.As(err, &pg) && pg.Code == "23505" {
		return ports.ErrApplicationOpen
	}
	return err
}
