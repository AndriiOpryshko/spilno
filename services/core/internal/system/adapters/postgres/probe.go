// Package postgres implements the system context's ports on Postgres.
package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/AndriiOpryshko/spilno/services/core/internal/system/adapters/postgres/sqlcgen"
	"github.com/AndriiOpryshko/spilno/services/core/internal/system/ports"
)

type Probe struct {
	q *sqlcgen.Queries
}

var _ ports.DatabaseProbe = (*Probe)(nil)

func NewProbe(pool *pgxpool.Pool) *Probe {
	return &Probe{q: sqlcgen.New(pool)}
}

func (p *Probe) ServerVersion(ctx context.Context) (string, error) {
	return p.q.ServerVersion(ctx)
}
