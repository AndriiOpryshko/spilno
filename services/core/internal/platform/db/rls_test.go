package db_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/db"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/testdb"
)

// Every table with an association_id column must have forced row-level security and a policy;
// the catalogue walk means a table added later is checked without editing this test.
// noPolicyByDesign names the tables that carry an association_id and still have no policy: the
// association is NULL until approval and the rows are reached only by account or by the оператор
// (association-onboarding ADR-0003). Every other table with the column must be isolated.
var noPolicyByDesign = map[string]bool{
	"applications":          true,
	"application_decisions": true,
}

func TestRLS_EveryAssociationTableIsIsolated(t *testing.T) {
	pool, _, _ := testdb.NewTwoAssociations(t)
	ctx := context.Background()

	rows, err := pool.Query(ctx, `
		SELECT c.relname, c.relrowsecurity, c.relforcerowsecurity,
		       (SELECT count(*) FROM pg_policy p WHERE p.polrelid = c.oid)
		FROM pg_class c
		JOIN pg_namespace n ON n.oid = c.relnamespace AND n.nspname = 'public'
		JOIN pg_attribute a ON a.attrelid = c.oid AND a.attname = 'association_id' AND NOT a.attisdropped
		WHERE c.relkind = 'r'
		ORDER BY c.relname`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	seen := 0
	for rows.Next() {
		var name string
		var enabled, forced bool
		var policies int
		if err := rows.Scan(&name, &enabled, &forced, &policies); err != nil {
			t.Fatal(err)
		}
		seen++
		if noPolicyByDesign[name] {
			if enabled || policies != 0 {
				t.Errorf("table %s: rls enabled=%v policies=%d, want none (ADR-0003)", name, enabled, policies)
			}
			continue
		}
		if !enabled || !forced || policies == 0 {
			t.Errorf("table %s: rls enabled=%v forced=%v policies=%d, want enabled, forced and >=1 policy", name, enabled, forced, policies)
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if seen == 0 {
		t.Fatal("no table with an association_id column found")
	}
}

func TestRLS_AppRoleSeesOnlyTheSetAssociation(t *testing.T) {
	pool, a, b := testdb.NewTwoAssociations(t)
	ctx := context.Background()
	aID, bID := uuid.MustParse(a), uuid.MustParse(b)

	insert := func(assoc uuid.UUID) error {
		return db.InAssociation(ctx, pool, assoc, "", func(ctx context.Context) error {
			q, err := db.From(ctx)
			if err != nil {
				return err
			}
			_, err = q.Exec(ctx, `INSERT INTO protocol_registers (id, association_id, next_number) VALUES ($1, $2, 1)`, uuid.New(), assoc)
			return err
		})
	}
	if err := insert(aID); err != nil {
		t.Fatalf("insert for A: %v", err)
	}
	if err := insert(bID); err != nil {
		t.Fatalf("insert for B: %v", err)
	}

	var n int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM protocol_registers`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Errorf("rows with no association set = %d, want 0", n)
	}

	err := db.InAssociation(ctx, pool, aID, "", func(ctx context.Context) error {
		q, err := db.From(ctx)
		if err != nil {
			return err
		}
		var other int
		if err := q.QueryRow(ctx, `SELECT count(*) FROM protocol_registers WHERE association_id <> $1`, aID).Scan(&other); err != nil {
			return err
		}
		if other != 0 {
			t.Errorf("rows of other associations visible with A set = %d, want 0", other)
		}
		var own int
		if err := q.QueryRow(ctx, `SELECT count(*) FROM protocol_registers`).Scan(&own); err != nil {
			return err
		}
		if own != 1 {
			t.Errorf("rows visible with A set = %d, want 1", own)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	// With A set, an insert for B is refused by the policy (a fresh row, so not a unique clash first).
	err = db.InAssociation(ctx, pool, aID, "", func(ctx context.Context) error {
		q, err := db.From(ctx)
		if err != nil {
			return err
		}
		_, err = q.Exec(ctx, `INSERT INTO protocol_registers (id, association_id, next_number) VALUES ($1, $2, 1)`, uuid.New(), bID)
		return err
	})
	if err == nil {
		t.Error("insert for B with A set succeeded, want refusal")
	}
}
