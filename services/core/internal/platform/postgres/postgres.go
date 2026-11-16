// Package postgres opens the shared pgx pool and applies goose migrations.
package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/AndriiOpryshko/spilno/services/core/migrations"
)

// Open connects a pool with OpenTelemetry query tracing and pool metrics.
func Open(ctx context.Context, url string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}
	cfg.ConnConfig.Tracer = otelpgx.NewTracer()

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("open pool: %w", err)
	}
	if err := otelpgx.RecordStats(pool); err != nil {
		pool.Close()
		return nil, fmt.Errorf("record pool stats: %w", err)
	}
	return pool, nil
}

// Migrate applies every pending migration embedded in the migrations package.
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	db := stdlib.OpenDBFromPool(pool)
	defer db.Close()

	provider, err := goose.NewProvider(goose.DialectPostgres, db, migrations.FS)
	if err != nil {
		return fmt.Errorf("goose provider: %w", err)
	}
	if _, err := provider.Up(ctx); err != nil {
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}

// RequireRLSEnforced fails when the pool's role is a superuser or has BYPASSRLS,
// because row-level security (ADR-0003) would not apply to it.
func RequireRLSEnforced(ctx context.Context, pool *pgxpool.Pool) error {
	var bypass bool
	err := pool.QueryRow(ctx,
		`SELECT rolsuper OR rolbypassrls FROM pg_roles WHERE rolname = current_user`).Scan(&bypass)
	if err != nil {
		return fmt.Errorf("check serving role: %w", err)
	}
	if bypass {
		return errors.New("DATABASE_URL names a role that bypasses row-level security (superuser or BYPASSRLS); serve as spilno_app_login and keep the owner in MIGRATIONS_DATABASE_URL")
	}
	return nil
}
