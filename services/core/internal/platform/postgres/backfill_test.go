package postgres_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/testdb"
	"github.com/AndriiOpryshko/spilno/services/core/migrations"
)

// lastBeforeBackfill is the live version before the three backfills (drop of accounts.association_id).
const lastBeforeBackfill = 37

// TestBackfillsFillNullColumns inserts rows with NULLs in the three expanded columns at the version
// before the backfills, applies the rest, and expects no NULL left (AC-19c, AC-29).
func TestBackfillsFillNullColumns(t *testing.T) {
	ctx := context.Background()
	pool := testdb.New(t)
	db := stdlib.OpenDBFromPool(pool)
	t.Cleanup(func() { _ = db.Close() })
	provider, err := goose.NewProvider(goose.DialectPostgres, db, migrations.FS)
	if err != nil {
		t.Fatalf("goose provider: %v", err)
	}
	if _, err := provider.DownTo(ctx, 0); err != nil {
		t.Fatalf("down to 0: %v", err)
	}
	if _, err := provider.UpTo(ctx, lastBeforeBackfill); err != nil {
		t.Fatalf("up to %d: %v", lastBeforeBackfill, err)
	}

	const seeded = "00000000-0000-7000-8000-000000000001"
	stmts := []string{
		`INSERT INTO associations (id, name) VALUES ('00000000-0000-7000-8000-0000000000c1', 'Hand-made') ON CONFLICT DO NOTHING`,
		`INSERT INTO board_members (id, association_id) VALUES ('00000000-0000-7000-8000-0000000000d1', '` + seeded + `')`,
		`INSERT INTO board_member_revisions (id, association_id, board_member_id, full_name, term_start)
			VALUES ('00000000-0000-7000-8000-0000000000e1', '` + seeded + `', '00000000-0000-7000-8000-0000000000d1', 'Test User', '2026-01-01')`,
		`INSERT INTO contracts (counterparty, amount_cents, currency, signed_at) VALUES ('Test Counterparty', 100, 'UAH', '2026-01-01')`,
	}
	for _, s := range stmts {
		if _, err := pool.Exec(ctx, s); err != nil {
			t.Fatalf("seed %q: %v", s, err)
		}
	}

	if _, err := provider.Up(ctx); err != nil {
		t.Fatalf("up: %v", err)
	}

	checks := []struct{ name, query, want string }{
		{"associations.state", `SELECT count(*) FILTER (WHERE state IS NULL), count(*) FILTER (WHERE state <> 'connected') FROM associations`, ""},
		{"board_member_revisions.position", `SELECT count(*) FILTER (WHERE position IS NULL), count(*) FILTER (WHERE position <> 'member') FROM board_member_revisions`, ""},
		{"contracts.association_id", `SELECT count(*) FILTER (WHERE association_id IS NULL), count(*) FILTER (WHERE association_id <> '` + seeded + `') FROM contracts`, ""},
	}
	for _, c := range checks {
		var nulls, wrong int
		if err := pool.QueryRow(ctx, c.query).Scan(&nulls, &wrong); err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if nulls != 0 || wrong != 0 {
			t.Errorf("%s after backfill: %d NULL, %d wrong value; want 0 and 0", c.name, nulls, wrong)
		}
	}

	// Down is a no-op and the chain re-applies cleanly.
	if _, err := provider.DownTo(ctx, lastBeforeBackfill); err != nil {
		t.Fatalf("down to %d: %v", lastBeforeBackfill, err)
	}
	if _, err := provider.Up(ctx); err != nil {
		t.Fatalf("re-apply: %v", err)
	}
}
