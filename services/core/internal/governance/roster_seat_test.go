package governance_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/AndriiOpryshko/spilno/services/core/internal/governance"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/governancetest"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/db"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/testdb"
)

// Expected API (governance/module.go), both on the transaction in ctx and refusing to run without one:
//
//	type Person = struct{ FullName, SignInEmail string }  // governance.Person
//	type Term = domain.Term                                // governance.Term
//	func (m *Module) SeatHead(ctx, association uuid.UUID, person Person, term Term, decisionID uuid.UUID) (memberID uuid.UUID, err error)
//	func (m *Module) EndTerm(ctx, association, memberID uuid.UUID, date time.Time, decisionID uuid.UUID) error
//
// SeatHead: a person off the roster gets one new entry (position head, term from term.Start,
// decision reference); a person with a current entry gets a new revision of the same entry
// (head, the entry's own term kept, new decision). EndTerm: a revision ending the whole entry
// on date, clamped to the entry's first day, with the decision; other entries untouched.

type seatSnapshot struct{ entries, revisions int }

func seatCount(t *testing.T, ctx context.Context, tx db.DBTX, assoc uuid.UUID) seatSnapshot {
	t.Helper()
	var s seatSnapshot
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM board_members WHERE association_id = $1`, assoc).Scan(&s.entries); err != nil {
		t.Fatal(err)
	}
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM board_member_revisions WHERE association_id = $1`, assoc).Scan(&s.revisions); err != nil {
		t.Fatal(err)
	}
	return s
}

// seatLatest returns the newest revision of an entry as a comparable string plus position and decision.
func seatLatest(t *testing.T, ctx context.Context, tx db.DBTX, member uuid.UUID) (position string, decision *uuid.UUID, start string, end *string) {
	t.Helper()
	var s time.Time
	var e *time.Time
	err := tx.QueryRow(ctx, `SELECT position, decision_id, term_start, term_end FROM board_member_revisions
		WHERE board_member_id = $1 ORDER BY created_at DESC, id DESC LIMIT 1`, member).Scan(&position, &decision, &s, &e)
	if err != nil {
		t.Fatal(err)
	}
	start = s.Format("2006-01-02")
	if e != nil {
		f := e.Format("2006-01-02")
		end = &f
	}
	return
}

func seatMemberOf(t *testing.T, ctx context.Context, tx db.DBTX, assoc uuid.UUID, email string) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	err := tx.QueryRow(ctx, `SELECT m.id FROM board_members m JOIN board_member_revisions r ON r.board_member_id = m.id
		WHERE m.association_id = $1 AND r.sign_in_email = $2 LIMIT 1`, assoc, email).Scan(&id)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func TestRosterSeatAndEnd(t *testing.T) {
	a := uuid.MustParse(assocA)
	day := func(n int) time.Time { return today().AddDate(0, 0, n) }
	person := governance.Person{FullName: "Test User", SignInEmail: "new-head@example.test"}

	setup := func(t *testing.T) (*governance.Module, func(fn func(ctx context.Context, tx db.DBTX) error) error) {
		pool := testdb.New(t)
		m := governance.New(pool)
		return m, func(fn func(ctx context.Context, tx db.DBTX) error) error {
			return db.InAssociation(context.Background(), pool, a, "actor@example.test", func(ctx context.Context) error {
				tx, err := db.From(ctx)
				if err != nil {
					return err
				}
				return fn(ctx, tx)
			})
		}
	}

	t.Run("seat on an empty roster creates one head entry with the decision", func(t *testing.T) {
		m, run := setup(t)
		decision := uuid.Must(uuid.NewV7())
		err := run(func(ctx context.Context, tx db.DBTX) error {
			id, err := m.SeatHead(ctx, a, person, governance.Term{Start: day(-3)}, decision)
			if err != nil {
				return fmt.Errorf("SeatHead: %w", err)
			}
			if got := seatCount(t, ctx, tx, a); got != (seatSnapshot{1, 1}) {
				t.Errorf("entries/revisions = %+v, want {1 1}", got)
			}
			pos, dec, start, end := seatLatest(t, ctx, tx, id)
			if pos != "head" || dec == nil || *dec != decision || start != day(-3).Format("2006-01-02") || end != nil {
				t.Errorf("revision = %s %v %s %v, want head %s %s open", pos, dec, start, end, decision, day(-3).Format("2006-01-02"))
			}
			if mm, ok, err := m.CurrentMember(ctx, a, person.SignInEmail, today()); err != nil || !ok || mm.Position != "head" {
				t.Errorf("CurrentMember = %+v ok=%v err=%v, want head", mm, ok, err)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("re-seat of a current member keeps the entry and its term", func(t *testing.T) {
		m, run := setup(t)
		decision := uuid.Must(uuid.NewV7())
		err := run(func(ctx context.Context, tx db.DBTX) error {
			governancetest.MemberTerm(t, tx, a, person.SignInEmail, "member", day(-100), time.Time{})
			existing := seatMemberOf(t, ctx, tx, a, person.SignInEmail)
			id, err := m.SeatHead(ctx, a, person, governance.Term{Start: day(-1)}, decision)
			if err != nil {
				return fmt.Errorf("SeatHead: %w", err)
			}
			if id != existing {
				t.Errorf("entry id = %s, want the same entry %s", id, existing)
			}
			if got := seatCount(t, ctx, tx, a); got != (seatSnapshot{1, 2}) {
				t.Errorf("entries/revisions = %+v, want {1 2}", got)
			}
			pos, dec, start, end := seatLatest(t, ctx, tx, id)
			if pos != "head" || dec == nil || *dec != decision || start != day(-100).Format("2006-01-02") || end != nil {
				t.Errorf("revision = %s %v %s %v, want head %s with term start %s", pos, dec, start, end, decision, day(-100).Format("2006-01-02"))
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("EndTerm ends one whole entry and leaves the others", func(t *testing.T) {
		m, run := setup(t)
		decision := uuid.Must(uuid.NewV7())
		err := run(func(ctx context.Context, tx db.DBTX) error {
			governancetest.MemberTerm(t, tx, a, "prev-head@example.test", "head", day(-100), time.Time{})
			governancetest.MemberTerm(t, tx, a, "other@example.test", "member", day(-100), time.Time{})
			prev := seatMemberOf(t, ctx, tx, a, "prev-head@example.test")
			other := seatMemberOf(t, ctx, tx, a, "other@example.test")
			p0, d0, s0, e0 := seatLatest(t, ctx, tx, other)
			before := seatCount(t, ctx, tx, a)
			if err := m.EndTerm(ctx, a, prev, day(-1), decision); err != nil {
				return fmt.Errorf("EndTerm: %w", err)
			}
			if _, ok, err := m.CurrentMember(ctx, a, "prev-head@example.test", today()); err != nil || ok {
				t.Errorf("ended member current today: ok=%v err=%v, want not current", ok, err)
			}
			if _, ok, _ := m.CurrentMember(ctx, a, "prev-head@example.test", day(-1)); !ok {
				t.Error("ended member must still be current on the end date")
			}
			pos, dec, _, end := seatLatest(t, ctx, tx, prev)
			if pos != "head" || dec == nil || *dec != decision || end == nil || *end != day(-1).Format("2006-01-02") {
				t.Errorf("ended revision = %s %v end=%v, want head %s ending %s", pos, dec, end, decision, day(-1).Format("2006-01-02"))
			}
			if got := seatCount(t, ctx, tx, a); got.entries != before.entries || got.revisions != before.revisions+1 {
				t.Errorf("entries/revisions = %+v, want one more revision than %+v", got, before)
			}
			p1, d1, s1, e1 := seatLatest(t, ctx, tx, other)
			if p1 != p0 || fmt.Sprint(d1) != fmt.Sprint(d0) || s1 != s0 || fmt.Sprint(e1) != fmt.Sprint(e0) {
				t.Errorf("other entry changed: %s %v %s %v -> %s %v %s %v", p0, d0, s0, e0, p1, d1, s1, e1)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("EndTerm before the entry's first day is clamped to it", func(t *testing.T) {
		m, run := setup(t)
		err := run(func(ctx context.Context, tx db.DBTX) error {
			governancetest.MemberTerm(t, tx, a, "prev-head@example.test", "head", day(-5), time.Time{})
			prev := seatMemberOf(t, ctx, tx, a, "prev-head@example.test")
			if err := m.EndTerm(ctx, a, prev, day(-20), uuid.Must(uuid.NewV7())); err != nil {
				return fmt.Errorf("EndTerm: %w", err)
			}
			_, _, _, end := seatLatest(t, ctx, tx, prev)
			if end == nil || *end != day(-5).Format("2006-01-02") {
				t.Errorf("term_end = %v, want %s", end, day(-5).Format("2006-01-02"))
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("a failure in the caller leaves the roster unchanged", func(t *testing.T) {
		m, run := setup(t)
		var before seatSnapshot
		if err := run(func(ctx context.Context, tx db.DBTX) error {
			governancetest.MemberTerm(t, tx, a, "prev-head@example.test", "head", day(-100), time.Time{})
			return nil
		}); err != nil {
			t.Fatal(err)
		}
		_ = run(func(ctx context.Context, tx db.DBTX) error {
			before = seatCount(t, ctx, tx, a)
			prev := seatMemberOf(t, ctx, tx, a, "prev-head@example.test")
			decision := uuid.Must(uuid.NewV7())
			if err := m.EndTerm(ctx, a, prev, day(-1), decision); err != nil {
				t.Errorf("EndTerm: %v", err)
			}
			if _, err := m.SeatHead(ctx, a, person, governance.Term{Start: day(0)}, decision); err != nil {
				t.Errorf("SeatHead: %v", err)
			}
			return fmt.Errorf("caller fails after both calls")
		})
		if err := run(func(ctx context.Context, tx db.DBTX) error {
			if got := seatCount(t, ctx, tx, a); got != before || before.entries != 1 || before.revisions != 1 {
				t.Errorf("after rollback = %+v, before = %+v, want {1 1}", got, before)
			}
			if _, ok, _ := m.CurrentMember(ctx, a, "prev-head@example.test", today()); !ok {
				t.Error("previous head must still be current after rollback")
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("both refuse to run without a transaction", func(t *testing.T) {
		m, _ := setup(t)
		ctx := context.Background()
		if _, err := m.SeatHead(ctx, a, person, governance.Term{Start: day(-1)}, uuid.Must(uuid.NewV7())); err == nil {
			t.Error("SeatHead without a transaction: want an error")
		}
		if err := m.EndTerm(ctx, a, uuid.Must(uuid.NewV7()), day(-1), uuid.Must(uuid.NewV7())); err == nil {
			t.Error("EndTerm without a transaction: want an error")
		}
	})
}
