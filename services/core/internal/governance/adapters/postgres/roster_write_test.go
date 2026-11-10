package postgres_test

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/adapters/postgres"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/domain"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/db"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/testdb"
)

// Expected API (adapters/postgres/roster_write.go), on the transaction in ctx:
//
//	func (Roster) WriteRevision(ctx, association, boardMember uuid.UUID, entry domain.BoardMemberEntry) error
//
// It validates entry in Go (position must be head or member, never empty) before any SQL,
// then inserts one board_member_revisions row with position and decision_id.

var assoc = uuid.MustParse("0191a000-0000-7000-8000-00000000000a")

func inAssoc(t *testing.T, fn func(ctx context.Context, tx db.DBTX)) {
	t.Helper()
	pool := testdb.New(t)
	err := db.InAssociation(context.Background(), pool, assoc, "actor@example.test", func(ctx context.Context) error {
		tx, err := db.From(ctx)
		if err != nil {
			return err
		}
		fn(ctx, tx)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func newMember(t *testing.T, ctx context.Context, tx db.DBTX) uuid.UUID {
	t.Helper()
	id := uuid.Must(uuid.NewV7())
	if _, err := tx.Exec(ctx, `INSERT INTO board_members (id, association_id) VALUES ($1, $2)`, id, assoc); err != nil {
		t.Fatal(err)
	}
	return id
}

func entry(position domain.Position, decision string) domain.BoardMemberEntry {
	start := time.Now().UTC().Truncate(24*time.Hour).AddDate(-1, 0, 0)
	return domain.BoardMemberEntry{
		FullName: "Test User", SignInEmail: "head@example.test",
		Term: domain.Term{Start: start}, Position: position, DecisionID: decision,
	}
}

func TestWriteRevision(t *testing.T) {
	roster := postgres.NewRoster()

	t.Run("position and decision id are stored and read back", func(t *testing.T) {
		inAssoc(t, func(ctx context.Context, tx db.DBTX) {
			member := newMember(t, ctx, tx)
			decision := uuid.Must(uuid.NewV7())
			if err := roster.WriteRevision(ctx, assoc, member, entry(domain.PositionHead, decision.String())); err != nil {
				t.Fatalf("WriteRevision: %v", err)
			}
			got, ok, err := roster.CurrentMember(ctx, assoc, "head@example.test", time.Now().UTC())
			if err != nil || !ok || got.Position != "head" {
				t.Errorf("CurrentMember = %+v ok=%v err=%v, want head", got, ok, err)
			}
			var stored uuid.UUID
			if err := tx.QueryRow(ctx, `SELECT decision_id FROM board_member_revisions WHERE board_member_id = $1`, member).Scan(&stored); err != nil {
				t.Fatalf("read decision_id: %v", err)
			}
			if stored != decision {
				t.Errorf("decision_id = %s, want %s", stored, decision)
			}
		})
	})

	t.Run("an empty position is refused in Go and writes no row", func(t *testing.T) {
		inAssoc(t, func(ctx context.Context, tx db.DBTX) {
			member := newMember(t, ctx, tx)
			err := roster.WriteRevision(ctx, assoc, member, entry("", ""))
			var coded *domain.CodedFieldsError
			if !errors.As(err, &coded) || !slices.Contains(coded.Pointers(), "/position") {
				t.Errorf("err = %v, want CodedFieldsError naming /position", err)
			}
			var n int
			if err := tx.QueryRow(ctx, `SELECT count(*) FROM board_member_revisions WHERE board_member_id = $1`, member).Scan(&n); err != nil {
				t.Fatal(err)
			}
			if n != 0 {
				t.Errorf("rows = %d, want 0", n)
			}
		})
	})

	t.Run("an unknown position is refused", func(t *testing.T) {
		inAssoc(t, func(ctx context.Context, tx db.DBTX) {
			member := newMember(t, ctx, tx)
			err := roster.WriteRevision(ctx, assoc, member, entry("chair", ""))
			var coded *domain.CodedFieldsError
			if !errors.As(err, &coded) || !slices.Contains(coded.Pointers(), "/position") {
				t.Errorf("err = %v, want CodedFieldsError naming /position", err)
			}
		})
	})
}
