package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/adapters/postgres/sqlcgen"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/ports"
)

// Sessions implements ports.AccountRepository and ports.SessionRepository.
type Sessions struct {
	q *sqlcgen.Queries
}

var (
	_ ports.AccountRepository = (*Sessions)(nil)
	_ ports.SessionRepository = (*Sessions)(nil)
)

func NewSessions(pool *pgxpool.Pool) *Sessions { return &Sessions{q: sqlcgen.New(pool)} }

func (s *Sessions) Upsert(ctx context.Context, id uuid.UUID, issuer, subject, email string) (ports.Account, error) {
	row, err := s.q.UpsertAccount(ctx, sqlcgen.UpsertAccountParams{
		ID:          pgtype.UUID{Bytes: id, Valid: true},
		OidcIssuer:  issuer,
		OidcSubject: subject,
		Email:       email,
	})
	if err != nil {
		return ports.Account{}, err
	}
	return ports.Account{ID: row.ID.Bytes, Email: row.Email}, nil
}

func (s *Sessions) Create(ctx context.Context, id, accountID uuid.UUID, tokenHash []byte, expiresAt time.Time) error {
	return s.q.InsertSession(ctx, sqlcgen.InsertSessionParams{
		ID:        pgtype.UUID{Bytes: id, Valid: true},
		AccountID: pgtype.UUID{Bytes: accountID, Valid: true},
		TokenHash: tokenHash,
		ExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true},
	})
}

func (s *Sessions) Account(ctx context.Context, tokenHash []byte) (ports.Account, error) {
	row, err := s.q.GetSessionPrincipal(ctx, tokenHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return ports.Account{}, ports.ErrSessionNotFound
	}
	if err != nil {
		return ports.Account{}, err
	}
	return ports.Account{ID: row.ID.Bytes, Email: row.Email}, nil
}

func (s *Sessions) Delete(ctx context.Context, tokenHash []byte) error {
	return s.q.DeleteSession(ctx, tokenHash)
}
