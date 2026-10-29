package postgres_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/testdb"
	"github.com/AndriiOpryshko/spilno/services/core/migrations"
)

// liveTables are the tables the migrations must create; every Down must drop them again.
var liveTables = []string{
	"contracts", "meetings", "associations",
	"accounts", "sessions", "google_connections",
}

func TestMigrationsRoundtrip(t *testing.T) {
	ctx := context.Background()
	pool := testdb.New(t)
	db := stdlib.OpenDBFromPool(pool)
	t.Cleanup(func() { _ = db.Close() })
	provider, err := goose.NewProvider(goose.DialectPostgres, db, migrations.FS)
	if err != nil {
		t.Fatalf("goose provider: %v", err)
	}

	assertTables(t, pool, liveTables)

	if _, err := provider.DownTo(ctx, 0); err != nil {
		t.Fatalf("down to 0: %v", err)
	}
	var left []string
	rows, err := pool.Query(ctx, `SELECT table_name FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name <> 'goose_db_version' ORDER BY 1`)
	if err != nil {
		t.Fatalf("list tables: %v", err)
	}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("scan: %v", err)
		}
		left = append(left, name)
	}
	if len(left) > 0 {
		t.Errorf("tables left after down to 0: %v", left)
	}

	if _, err := provider.Up(ctx); err != nil {
		t.Fatalf("re-apply: %v", err)
	}
	assertTables(t, pool, liveTables)
}

func assertTables(t *testing.T, pool *pgxpool.Pool, names []string) {
	t.Helper()
	for _, name := range names {
		var found *string
		if err := pool.QueryRow(context.Background(), `SELECT to_regclass($1)::text`, "public."+name).Scan(&found); err != nil {
			t.Fatalf("to_regclass %s: %v", name, err)
		}
		if found == nil {
			t.Errorf("table %s does not exist", name)
		}
	}
}
