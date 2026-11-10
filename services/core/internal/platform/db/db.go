// Package db carries the association and the actor on a transaction.
package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

// DBTX is the query runner shape sqlc generates against.
type DBTX interface {
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
}

const setScopeSQL = "SELECT set_config('app.association_id', $1, true), set_config('app.actor_email', $2, true)"

type scope struct {
	tx          pgx.Tx
	association uuid.UUID
}

type scopeKey struct{}

// ErrNoTransaction is returned by From when ctx carries no InAssociation transaction.
var ErrNoTransaction = errors.New("db: no association transaction in context")

// ErrOtherAssociation is returned when InAssociation nests for another association.
var ErrOtherAssociation = errors.New("db: nested transaction for another association")

// InAssociation runs fn in a transaction with app.association_id and
// app.actor_email set. A nested call for the same association reuses the
// transaction; for another association it fails.
func InAssociation(ctx context.Context, pool *pgxpool.Pool, associationID uuid.UUID, actorEmail string, fn func(ctx context.Context) error) (err error) {
	if s, ok := ctx.Value(scopeKey{}).(*scope); ok {
		if s.association != associationID {
			return ErrOtherAssociation
		}
		return fn(ctx)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(context.WithoutCancel(ctx))
		}
	}()

	if _, err = tx.Exec(ctx, setScopeSQL, associationID.String(), actorEmail); err != nil {
		return fmt.Errorf("set association: %w", err)
	}
	if err = fn(context.WithValue(ctx, scopeKey{}, &scope{tx: tx, association: associationID})); err != nil {
		return err
	}
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	committed = true
	return nil
}

// From returns the transaction InAssociation put in ctx. Without one it
// counts platform.db.association_unset and returns an error.
func From(ctx context.Context) (DBTX, error) {
	if s, ok := ctx.Value(scopeKey{}).(*scope); ok {
		return s.tx, nil
	}
	c, err := otel.Meter("platform/db").Int64Counter("platform.db.association_unset",
		metric.WithDescription("Scoped runner requested outside InAssociation"))
	if err == nil {
		c.Add(ctx, 1)
	}
	return nil, ErrNoTransaction
}
