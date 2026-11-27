package finance_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/AndriiOpryshko/spilno/services/core/internal/finance"
	"github.com/AndriiOpryshko/spilno/services/core/internal/finance/adapters/postgres"
	"github.com/AndriiOpryshko/spilno/services/core/internal/finance/domain"
	"github.com/AndriiOpryshko/spilno/services/core/internal/finance/ports"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/db"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/testdb"
)

// Issue #70 AC-4: contracts without an equivalent in another association are
// never listed, and a person off that association's roster gets 403
// identity.not_current_member. The member is on A's roster only; B holds its
// own contract without an equivalent, stored under B's row-level security.
func TestListContractsMissingUahEquivalentIsScopedToTheAssociation(t *testing.T) {
	pool, a, b := testdb.NewTwoAssociations(t)
	srv, _, _ := newGuardedEnvOn(t, pool, finance.Config{Rates: &fakeRates{err: domain.ErrRateUnavailable}}, a)

	inA := createEq(t, contractsURL(srv, a), contractBody("USD", 1))

	bID := uuid.MustParse(b)
	var inB uuid.UUID
	err := db.InAssociation(context.Background(), pool, bID, "seed@example.test", func(ctx context.Context) error {
		c, err := postgres.NewContracts().Create(ports.WithAssociation(ctx, bID), domain.ContractTerms{
			Counterparty: "Test Counterparty",
			AmountCents:  5000,
			Currency:     domain.CurrencyEUR,
			SignedAt:     time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC),
		}, nil)
		inB = c.ID
		return err
	})
	if err != nil {
		t.Fatal(err)
	}

	var page eqPage
	do(t, "GET", contractsURL(srv, a)+"?uah_equivalent=missing", "", http.StatusOK, &page)
	if len(page.Items) != 1 || page.Items[0].ID != inA {
		t.Errorf("missing under A = %+v, want only A's contract %s (B's is %s)", page.Items, inA, inB)
	}

	var p coded
	do(t, "GET", contractsURL(srv, b)+"?uah_equivalent=missing", "", http.StatusForbidden, &p)
	if p.Code != "identity.not_current_member" {
		t.Errorf("non-member of B: code = %q, want identity.not_current_member", p.Code)
	}
}
