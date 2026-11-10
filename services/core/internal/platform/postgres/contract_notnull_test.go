package postgres_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/testdb"
	"github.com/AndriiOpryshko/spilno/services/core/migrations"
)

// lastBeforeContractSteps is the live version of the last backfill; the contract steps follow it.
const lastBeforeContractSteps = 40

// TestContractStepsSetNotNullAndIndex proves the promoted contract steps: the three expanded columns are
// NOT NULL, idx_contracts_association_created exists, and the steps revert and re-apply (AC-08, AC-21).
func TestContractStepsSetNotNullAndIndex(t *testing.T) {
	ctx := context.Background()
	pool := testdb.New(t)
	db := stdlib.OpenDBFromPool(pool)
	t.Cleanup(func() { _ = db.Close() })
	provider, err := goose.NewProvider(goose.DialectPostgres, db, migrations.FS)
	if err != nil {
		t.Fatalf("goose provider: %v", err)
	}

	notNull := func() map[string]string {
		rows, err := pool.Query(ctx, `SELECT table_name || '.' || column_name, is_nullable FROM information_schema.columns
			WHERE table_schema = 'public' AND (table_name, column_name) IN
			(('associations','state'),('board_member_revisions','position'),('contracts','association_id'))`)
		if err != nil {
			t.Fatalf("columns: %v", err)
		}
		defer rows.Close()
		got := map[string]string{}
		for rows.Next() {
			var k, v string
			if err := rows.Scan(&k, &v); err != nil {
				t.Fatal(err)
			}
			got[k] = v
		}
		return got
	}
	indexExists := func() bool {
		var n int
		if err := pool.QueryRow(ctx, `SELECT count(*) FROM pg_indexes WHERE schemaname='public' AND indexname='idx_contracts_association_created'`).Scan(&n); err != nil {
			t.Fatalf("index: %v", err)
		}
		return n == 1
	}
	assertState := func(label, nullable string, index bool) {
		t.Helper()
		got := notNull()
		for _, c := range []string{"associations.state", "board_member_revisions.position", "contracts.association_id"} {
			if got[c] != nullable {
				t.Errorf("%s: %s is_nullable = %q, want %q", label, c, got[c], nullable)
			}
		}
		if got := indexExists(); got != index {
			t.Errorf("%s: idx_contracts_association_created exists = %v, want %v", label, got, index)
		}
	}

	if _, err := provider.Up(ctx); err != nil {
		t.Fatalf("up: %v", err)
	}
	assertState("after up", "NO", true)

	if _, err := provider.DownTo(ctx, lastBeforeContractSteps); err != nil {
		t.Fatalf("down to %d: %v", lastBeforeContractSteps, err)
	}
	assertState("after down", "YES", false)

	if _, err := provider.Up(ctx); err != nil {
		t.Fatalf("re-apply: %v", err)
	}
	assertState("after re-apply", "NO", true)
}
