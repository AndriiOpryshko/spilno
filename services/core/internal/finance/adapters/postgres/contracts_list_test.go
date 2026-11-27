package postgres_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/AndriiOpryshko/spilno/services/core/internal/finance/adapters/postgres"
	"github.com/AndriiOpryshko/spilno/services/core/internal/finance/domain"
	"github.com/AndriiOpryshko/spilno/services/core/internal/finance/ports"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/testdb"
)

var missingOnly = ports.ContractFilter{MissingUAHEquivalent: true}

func withCurrency(c domain.Currency) domain.ContractTerms {
	t := usdTerms
	t.Currency = c
	return t
}

// Issue #70 AC-1, AC-2: the filter keeps only foreign-currency contracts
// without an equivalent; the zero filter lists every contract.
func TestContractsListMissingUAHEquivalent(t *testing.T) {
	pool, assoc, _ := testdb.NewTwoAssociations(t)
	repo := postgres.NewContracts()
	eq := usdEquivalent()
	inAssoc(t, pool, assoc, func(ctx context.Context) {
		create := func(terms domain.ContractTerms, eq *domain.UAHEquivalent) uuid.UUID {
			c, err := repo.Create(ctx, terms, eq)
			if err != nil {
				t.Fatal(err)
			}
			return c.ID
		}
		create(withCurrency(domain.CurrencyUAH), nil)
		create(withCurrency(domain.CurrencyEUR), &eq)
		usd := create(withCurrency(domain.CurrencyUSD), nil)

		got, err := repo.List(ctx, missingOnly, nil, 10)
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 || got[0].ID != usd {
			t.Errorf("missing = %v, want only the USD contract %v", ids(got), usd)
		}

		all, err := repo.List(ctx, ports.ContractFilter{}, nil, 10)
		if err != nil {
			t.Fatal(err)
		}
		if len(all) != 3 {
			t.Errorf("unfiltered = %d contracts, want 3", len(all))
		}
	})
}

// Issue #70 AC-3: the filtered keyset walk returns each missing contract once,
// newest first, and skips none.
func TestContractsListMissingUAHEquivalentPages(t *testing.T) {
	pool, assoc, _ := testdb.NewTwoAssociations(t)
	repo := postgres.NewContracts()
	eq := usdEquivalent()
	inAssoc(t, pool, assoc, func(ctx context.Context) {
		want := map[uuid.UUID]bool{}
		for i := range 6 {
			terms := withCurrency(domain.CurrencyEUR)
			if i%2 == 0 {
				if _, err := repo.Create(ctx, terms, &eq); err != nil {
					t.Fatal(err)
				}
			}
			c, err := repo.Create(ctx, terms, nil)
			if err != nil {
				t.Fatal(err)
			}
			want[c.ID] = true
		}

		var seen []domain.Contract
		var after *ports.ContractKey
		for range 20 {
			page, err := repo.List(ctx, missingOnly, after, 2)
			if err != nil {
				t.Fatal(err)
			}
			if len(page) == 0 {
				break
			}
			seen = append(seen, page...)
			last := page[len(page)-1]
			after = &ports.ContractKey{CreatedAt: last.CreatedAt, ID: last.ID}
		}

		if len(seen) != len(want) {
			t.Fatalf("walked %d contracts, want %d", len(seen), len(want))
		}
		for i, c := range seen {
			if !want[c.ID] {
				t.Errorf("walk returned %v twice or one that has an equivalent", c.ID)
			}
			delete(want, c.ID)
			if i > 0 && !before(seen[i-1], c) {
				t.Errorf("item %d is not older than item %d", i, i-1)
			}
		}
	})
}

// before reports whether a comes before b in the newest-first order.
func before(a, b domain.Contract) bool {
	if !a.CreatedAt.Equal(b.CreatedAt) {
		return a.CreatedAt.After(b.CreatedAt)
	}
	return a.ID.String() > b.ID.String()
}

func ids(cs []domain.Contract) []uuid.UUID {
	out := make([]uuid.UUID, len(cs))
	for i, c := range cs {
		out[i] = c.ID
	}
	return out
}
