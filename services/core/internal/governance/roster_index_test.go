package governance_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/AndriiOpryshko/spilno/services/core/internal/governance"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/db"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/testdb"
)

const lowerEmailIndex = "idx_board_member_revisions_lower_sign_in_email"

// membershipsOfSQL returns the MembershipsOf statement from roster.sql with $1 = email, $2 = on_date.
func membershipsOfSQL(t *testing.T) string {
	t.Helper()
	raw, err := os.ReadFile("adapters/postgres/queries/roster.sql")
	if err != nil {
		t.Fatal(err)
	}
	_, rest, ok := strings.Cut(string(raw), "-- name: MembershipsOf :many\n")
	if !ok {
		t.Fatal("roster.sql has no MembershipsOf query")
	}
	body, _, _ := strings.Cut(rest, "-- name:")
	body = strings.ReplaceAll(body, "sqlc.arg(email)", "$1")
	body = strings.ReplaceAll(body, "sqlc.arg(on_date)", "$2")
	return strings.TrimSuffix(strings.TrimSpace(body), ";")
}

// T59 / F20: on a large roster MembershipsOf, run as the app role under RLS, is served by the index
// on lower(sign_in_email), not by a sequential scan of board_member_revisions.
func TestMembershipsOfUsesTheLoweredSignInEmailIndex(t *testing.T) {
	pool, owner, as, _ := testdb.NewTwoAssociationsWithOwner(t)
	ctx := context.Background()
	a := uuid.MustParse(as)

	// 5000 entries, one revision each, distinct addresses (owner seeds in bulk, then refreshes statistics).
	if _, err := owner.Exec(ctx, `
		WITH m AS (
		  INSERT INTO board_members (id, association_id)
		  SELECT gen_random_uuid(), $1 FROM generate_series(1, 5000) RETURNING id)
		INSERT INTO board_member_revisions (id, association_id, board_member_id, full_name, sign_in_email, term_start, position)
		SELECT gen_random_uuid(), $1, id, 'Test User', 'user-' || row_number() OVER () || '@example.test', current_date - 365, 'member' FROM m`, a); err != nil {
		t.Fatalf("seed roster: %v", err)
	}
	if _, err := owner.Exec(ctx, `ANALYZE board_member_revisions`); err != nil {
		t.Fatalf("analyze: %v", err)
	}

	const actor = "User-2500@Example.test"
	day := time.Now().UTC().Truncate(24 * time.Hour)

	var plan []string
	err := db.InAssociation(ctx, pool, uuid.Nil, actor, func(ctx context.Context) error {
		tx, err := db.From(ctx)
		if err != nil {
			return err
		}
		rows, err := tx.Query(ctx, "EXPLAIN "+membershipsOfSQL(t), actor, day)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var line string
			if err := rows.Scan(&line); err != nil {
				return err
			}
			plan = append(plan, line)
		}
		return rows.Err()
	})
	if err != nil {
		t.Fatal(err)
	}
	text := strings.Join(plan, "\n")
	if !strings.Contains(text, lowerEmailIndex) {
		t.Errorf("plan does not use %s, want an index scan on it:\n%s", lowerEmailIndex, text)
	}
	if strings.Contains(text, "Seq Scan on board_member_revisions") {
		t.Errorf("plan has Seq Scan on board_member_revisions:\n%s", text)
	}

	// The call itself still answers: the one entry carrying the actor's address.
	err = db.InAssociation(ctx, pool, uuid.Nil, actor, func(ctx context.Context) error {
		got, err := governance.New(pool, nil).MembershipsOf(ctx, actor, day)
		if err != nil {
			return err
		}
		if len(got) != 1 || got[0].AssociationID != a {
			t.Errorf("MembershipsOf = %+v, want one membership in %s", got, a)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// Pin (T59): the latest revision of each entry decides, and the address compares case-insensitively.
func TestMembershipsOfLatestRevisionDecides(t *testing.T) {
	pool, as, _ := testdb.NewTwoAssociations(t)
	ctx := context.Background()
	a := uuid.MustParse(as)
	const actor = "me@example.test"
	day := time.Now().UTC().Truncate(24 * time.Hour)

	revision := func(ctx context.Context, tx db.DBTX, member uuid.UUID, email string, age time.Duration) {
		t.Helper()
		if _, err := tx.Exec(ctx, `INSERT INTO board_member_revisions (id, association_id, board_member_id, full_name, sign_in_email, term_start, position, created_at)
			VALUES ($1, $2, $3, 'Test User', $4, $5, 'member', now() - $6::interval)`,
			uuid.Must(uuid.NewV7()), a, member, email, day.AddDate(-1, 0, 0), age.String()); err != nil {
			t.Fatalf("insert revision: %v", err)
		}
	}
	newMember := func(ctx context.Context, tx db.DBTX) uuid.UUID {
		t.Helper()
		id := uuid.Must(uuid.NewV7())
		if _, err := tx.Exec(ctx, `INSERT INTO board_members (id, association_id) VALUES ($1, $2)`, id, a); err != nil {
			t.Fatal(err)
		}
		return id
	}

	var inA int
	err := db.InAssociation(ctx, pool, a, actor, func(ctx context.Context) error {
		tx, err := db.From(ctx)
		if err != nil {
			return err
		}
		// Moved away: old revision has the actor's address, the latest another one -> no membership.
		moved := newMember(ctx, tx)
		revision(ctx, tx, moved, actor, 2*time.Hour)
		revision(ctx, tx, moved, "someone-else@example.test", time.Hour)
		got, err := governance.New(pool, nil).MembershipsOf(ctx, actor, day)
		if err != nil {
			return err
		}
		if len(got) != 0 {
			t.Errorf("entry whose latest revision has another address: MembershipsOf = %+v, want none", got)
		}
		// Moved in: old revision another address, latest the actor's in other letter case -> membership.
		in := newMember(ctx, tx)
		revision(ctx, tx, in, "old@example.test", 2*time.Hour)
		revision(ctx, tx, in, "ME@Example.TEST", time.Hour)
		got, err = governance.New(pool, nil).MembershipsOf(ctx, actor, day)
		if err != nil {
			return err
		}
		for _, m := range got {
			if m.AssociationID == a {
				inA++
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if inA != 1 {
		t.Errorf("latest revision with the actor's address in another case: %d memberships in A, want 1", inA)
	}
}
