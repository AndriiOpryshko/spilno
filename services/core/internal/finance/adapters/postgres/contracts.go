// Package postgres implements the finance context's ports on Postgres.
package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/AndriiOpryshko/spilno/services/core/internal/finance/adapters/postgres/sqlcgen"
	"github.com/AndriiOpryshko/spilno/services/core/internal/finance/domain"
	"github.com/AndriiOpryshko/spilno/services/core/internal/finance/ports"
)

type Contracts struct {
	q *sqlcgen.Queries
}

var _ ports.ContractRepository = (*Contracts)(nil)

func NewContracts(pool *pgxpool.Pool) *Contracts {
	return &Contracts{q: sqlcgen.New(pool)}
}

func (c *Contracts) Create(ctx context.Context, t domain.ContractTerms) (domain.Contract, error) {
	row, err := c.q.CreateContract(ctx, sqlcgen.CreateContractParams{
		Counterparty: t.Counterparty,
		AmountCents:  t.AmountCents,
		Currency:     string(t.Currency),
		SignedAt:     pgtype.Date{Time: t.SignedAt, Valid: true},
		EndsAt:       toDate(t.EndsAt),
	})
	if err != nil {
		return domain.Contract{}, err
	}
	return toDomain(row), nil
}

func (c *Contracts) Get(ctx context.Context, id uuid.UUID) (domain.Contract, error) {
	row, err := c.q.GetContract(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Contract{}, domain.ErrContractNotFound
	}
	if err != nil {
		return domain.Contract{}, err
	}
	return toDomain(row), nil
}

func (c *Contracts) List(ctx context.Context, after *ports.ContractKey, limit int) ([]domain.Contract, error) {
	// limit is bounded by app.MaxPageSize, so it fits int32.
	arg := sqlcgen.ListContractsParams{RowLimit: int32(limit)}
	if after != nil {
		arg.AfterCreatedAt = pgtype.Timestamptz{Time: after.CreatedAt, Valid: true}
		arg.AfterID = pgtype.UUID{Bytes: after.ID, Valid: true}
	}
	rows, err := c.q.ListContracts(ctx, arg)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Contract, len(rows))
	for i, r := range rows {
		out[i] = toDomain(r)
	}
	return out, nil
}

func toDate(t *time.Time) pgtype.Date {
	if t == nil {
		return pgtype.Date{}
	}
	return pgtype.Date{Time: *t, Valid: true}
}

func toDomain(r sqlcgen.Contract) domain.Contract {
	c := domain.Contract{
		ID: r.ID.Bytes,
		ContractTerms: domain.ContractTerms{
			Counterparty: r.Counterparty,
			AmountCents:  r.AmountCents,
			Currency:     domain.Currency(r.Currency),
			SignedAt:     r.SignedAt.Time,
		},
		CreatedAt: r.CreatedAt.Time,
		UpdatedAt: r.UpdatedAt.Time,
	}
	if r.EndsAt.Valid {
		end := r.EndsAt.Time
		c.EndsAt = &end
	}
	return c
}
