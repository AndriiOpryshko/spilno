package db_test

import (
	"context"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/db"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/testdb"
)

const revisionsInsertSQL = `INSERT INTO board_member_revisions (id, association_id, board_member_id, full_name, sign_in_email, term_start, position)
	VALUES ($1, $2, $3, 'Test User', $4, $5, 'member')`

// seedRevisionsAB seeds A with ann and carol and B with ann and dave through the app role, each
// inside its own association.
func seedRevisionsAB(t *testing.T, run func(assoc uuid.UUID, fn func(ctx context.Context, q db.DBTX) error) error, a, b uuid.UUID) {
	t.Helper()
	for assoc, emails := range map[uuid.UUID][]string{a: {"ann@example.test", "carol@example.test"}, b: {"Ann@Example.test", "dave@example.test"}} {
		err := run(assoc, func(ctx context.Context, q db.DBTX) error {
			for _, email := range emails {
				member := uuid.Must(uuid.NewV7())
				if _, err := q.Exec(ctx, `INSERT INTO board_members (id, association_id) VALUES ($1, $2)`, member, assoc); err != nil {
					return err
				}
				if _, err := q.Exec(ctx, revisionsInsertSQL, uuid.Must(uuid.NewV7()), assoc, member, email, time.Now().AddDate(-1, 0, 0)); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			t.Fatalf("seed %s: %v", assoc, err)
		}
	}
}

// DoD (T20): with association A set and the actor's address, the actor's own board_member_revisions
// entries in B are visible and nobody else's (governance MembershipsOf works without a bypass).
func TestRLS_RevisionsShowTheActorsOwnEntriesInOtherAssociations(t *testing.T) {
	pool, as, bs := testdb.NewTwoAssociations(t)
	ctx := context.Background()
	a, b := uuid.MustParse(as), uuid.MustParse(bs)
	seedRevisionsAB(t, func(assoc uuid.UUID, fn func(ctx context.Context, q db.DBTX) error) error {
		return db.InAssociation(ctx, pool, assoc, "seed@example.test", func(ctx context.Context) error {
			q, err := db.From(ctx)
			if err != nil {
				return err
			}
			return fn(ctx, q)
		})
	}, a, b)

	visible := func(actor string) []string {
		t.Helper()
		var got []string
		err := db.InAssociation(ctx, pool, a, actor, func(ctx context.Context) error {
			q, err := db.From(ctx)
			if err != nil {
				return err
			}
			rows, err := q.Query(ctx, `SELECT lower(sign_in_email) || '@' || (CASE WHEN association_id = $1 THEN 'A' ELSE 'B' END) FROM board_member_revisions`, a)
			if err != nil {
				return err
			}
			defer rows.Close()
			for rows.Next() {
				var s string
				if err := rows.Scan(&s); err != nil {
					return err
				}
				got = append(got, s)
			}
			return rows.Err()
		})
		if err != nil {
			t.Fatal(err)
		}
		slices.Sort(got)
		return got
	}

	want := []string{"ann@example.test@A", "ann@example.test@B", "carol@example.test@A"}
	if got := visible("ANN@example.test"); !slices.Equal(got, want) {
		t.Errorf("A set, actor ann: visible = %v, want %v (own entry in B, nobody else's)", got, want)
	}
	if got := visible("carol@example.test"); !slices.Equal(got, []string{"ann@example.test@A", "carol@example.test@A"}) {
		t.Errorf("A set, actor carol (only on A): visible = %v, want the two entries of A", got)
	}
	if got := visible(""); strings.Contains(strings.Join(got, ","), "@B") {
		t.Errorf("A set, no actor: visible = %v, want entries of A only", got)
	}

	// Reading across associations does not extend to writing: with A set, an entry for B is refused.
	err := db.InAssociation(ctx, pool, a, "ann@example.test", func(ctx context.Context) error {
		q, err := db.From(ctx)
		if err != nil {
			return err
		}
		member := uuid.Must(uuid.NewV7())
		if _, err := q.Exec(ctx, `INSERT INTO board_members (id, association_id) VALUES ($1, $2)`, member, a); err != nil {
			return err
		}
		_, err = q.Exec(ctx, revisionsInsertSQL, uuid.Must(uuid.NewV7()), b, member, "ann@example.test", time.Now())
		return err
	})
	if err == nil {
		t.Error("insert of a revision for B with A set succeeded, want refusal")
	}
}

// DoD (T20): the owner role still reads every row of board_member_revisions.
func TestRLS_OwnerStillReadsEveryRevision(t *testing.T) {
	pool := testdb.New(t)
	ctx := context.Background()
	a, b := uuid.MustParse("00000000-0000-7000-8000-0000000000a1"), uuid.MustParse("00000000-0000-7000-8000-0000000000b1")
	if _, err := pool.Exec(ctx, `INSERT INTO associations (id, name, state) VALUES ($1, 'A', 'connected'), ($2, 'B', 'connected')`, a, b); err != nil {
		t.Fatal(err)
	}
	for _, assoc := range []uuid.UUID{a, b} {
		member := uuid.Must(uuid.NewV7())
		if _, err := pool.Exec(ctx, `INSERT INTO board_members (id, association_id) VALUES ($1, $2)`, member, assoc); err != nil {
			t.Fatalf("owner insert member: %v", err)
		}
		if _, err := pool.Exec(ctx, revisionsInsertSQL, uuid.Must(uuid.NewV7()), assoc, member, "x@example.test", time.Now()); err != nil {
			t.Fatalf("owner insert revision: %v", err)
		}
	}
	var n int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM board_member_revisions`).Scan(&n); err != nil || n != 2 {
		t.Errorf("owner sees %d revisions (err %v), want 2", n, err)
	}
}
