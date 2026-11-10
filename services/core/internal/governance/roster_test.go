package governance_test

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/AndriiOpryshko/spilno/services/core/internal/governance"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/governancetest"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/db"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/testdb"
)

// Expected API (governance/module.go), all on the transaction in ctx:
//
//	type Membership struct { AssociationID uuid.UUID; Position string } // "head" | "member"
//	func (m *Module) CurrentMember(ctx, association uuid.UUID, email string, date time.Time) (Membership, bool, error)
//	func (m *Module) MembershipsOf(ctx, email string, date time.Time) ([]Membership, error)
//	func (m *Module) MemberSignInAddresses(ctx, association uuid.UUID) ([]string, error)

func today() time.Time { return time.Now().UTC().Truncate(24 * time.Hour) }

func TestRosterFacade(t *testing.T) {
	a, b := uuid.MustParse(assocA), uuid.MustParse(assocB)
	day := func(n int) time.Time { return today().AddDate(0, 0, n) }

	setup := func(t *testing.T) (*governance.Module, func(assoc uuid.UUID, fn func(ctx context.Context, tx db.DBTX))) {
		pool := testdb.New(t)
		m := governance.New(pool)
		return m, func(assoc uuid.UUID, fn func(ctx context.Context, tx db.DBTX)) {
			t.Helper()
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
	}

	t.Run("a term that ended yesterday is not current, one ending today is", func(t *testing.T) {
		m, run := setup(t)
		run(a, func(ctx context.Context, tx db.DBTX) {
			governancetest.MemberTerm(t, tx, a, "old@example.test", "member", day(-30), day(-1))
			governancetest.MemberTerm(t, tx, a, "last@example.test", "member", day(-30), day(0))
			_, ok, err := m.CurrentMember(ctx, a, "old@example.test", today())
			if err != nil || ok {
				t.Errorf("ended yesterday: ok=%v err=%v, want not current", ok, err)
			}
			_, ok, err = m.CurrentMember(ctx, a, "last@example.test", today())
			if err != nil || !ok {
				t.Errorf("ends today: ok=%v err=%v, want current", ok, err)
			}
		})
	})

	t.Run("email is compared case-insensitively", func(t *testing.T) {
		m, run := setup(t)
		run(a, func(ctx context.Context, tx db.DBTX) {
			governancetest.Member(t, tx, a, "Head@Example.test", "head")
			got, ok, err := m.CurrentMember(ctx, a, "head@example.test", today())
			if err != nil || !ok || got.Position != "head" {
				t.Errorf("CurrentMember = %+v ok=%v err=%v, want head", got, ok, err)
			}
		})
	})

	t.Run("the same address on two rosters gives two memberships with their positions", func(t *testing.T) {
		m, run := setup(t)
		run(a, func(ctx context.Context, tx db.DBTX) { governancetest.Member(t, tx, a, "both@example.test", "head") })
		run(b, func(ctx context.Context, tx db.DBTX) { governancetest.Member(t, tx, b, "both@example.test", "member") })
		run(a, func(ctx context.Context, tx db.DBTX) {
			got, err := m.MembershipsOf(ctx, "both@example.test", today())
			if err != nil {
				t.Fatal(err)
			}
			pos := map[uuid.UUID]string{}
			for _, ms := range got {
				pos[ms.AssociationID] = ms.Position
			}
			if len(got) != 2 || pos[a] != "head" || pos[b] != "member" {
				t.Errorf("MembershipsOf = %+v, want head in A and member in B", got)
			}
		})
	})

	t.Run("a person off every roster has no membership", func(t *testing.T) {
		m, run := setup(t)
		run(a, func(ctx context.Context, tx db.DBTX) {
			governancetest.MemberTerm(t, tx, a, "gone@example.test", "member", day(-30), day(-1))
			got, err := m.MembershipsOf(ctx, "gone@example.test", today())
			if err != nil || len(got) != 0 {
				t.Errorf("MembershipsOf = %+v err=%v, want none", got, err)
			}
		})
	})

	t.Run("sign-in addresses are the current members of one association only", func(t *testing.T) {
		m, run := setup(t)
		run(a, func(ctx context.Context, tx db.DBTX) {
			governancetest.Member(t, tx, a, "one@example.test", "head")
			governancetest.Member(t, tx, a, "two@example.test", "member")
			governancetest.MemberTerm(t, tx, a, "ended@example.test", "member", day(-30), day(-1))
		})
		run(b, func(ctx context.Context, tx db.DBTX) { governancetest.Member(t, tx, b, "other@example.test", "head") })
		run(a, func(ctx context.Context, tx db.DBTX) {
			got, err := m.MemberSignInAddresses(ctx, a)
			if err != nil {
				t.Fatal(err)
			}
			slices.Sort(got)
			if !slices.Equal(got, []string{"one@example.test", "two@example.test"}) {
				t.Errorf("MemberSignInAddresses(A) = %v, want the two current members of A", got)
			}
		})
	})
}
