package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/AndriiOpryshko/spilno/services/core/internal/finance/adapters/postgres"
	"github.com/AndriiOpryshko/spilno/services/core/internal/finance/domain"
	"github.com/AndriiOpryshko/spilno/services/core/internal/finance/ports"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/db"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/testdb"
)

var usdTerms = domain.ContractTerms{
	Counterparty: "Test Counterparty",
	AmountCents:  10000,
	Currency:     domain.CurrencyUSD,
	SignedAt:     time.Date(2026, 11, 20, 0, 0, 0, 0, time.UTC),
}

func usdEquivalent() domain.UAHEquivalent {
	return domain.UAHEquivalent{
		AmountKopiyky: 41_500,
		Rate:          domain.Rate{Currency: domain.CurrencyUSD, Day: usdTerms.SignedAt, Micros: 41_500_000},
	}
}

func inAssoc(t *testing.T, pool *pgxpool.Pool, assoc string, fn func(ctx context.Context)) {
	t.Helper()
	id := uuid.MustParse(assoc)
	err := db.InAssociation(context.Background(), pool, id, "", func(ctx context.Context) error {
		fn(ports.WithAssociation(ctx, id))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func assertEq(t *testing.T, got *domain.UAHEquivalent, want domain.UAHEquivalent) {
	t.Helper()
	if got == nil {
		t.Fatalf("UAHEquivalent = nil, want kopiykas %d rate %d", want.AmountKopiyky, want.Rate.Micros)
	}
	if got.AmountKopiyky != want.AmountKopiyky || got.Rate.Micros != want.Rate.Micros {
		t.Fatalf("UAHEquivalent = %d kopiykas rate %d, want %d rate %d",
			got.AmountKopiyky, got.Rate.Micros, want.AmountKopiyky, want.Rate.Micros)
	}
}

// AC-01: an equivalent stored at creation reads back with the same kopiykas and rate.
func TestContractsCreateWithEquivalentReadsBack(t *testing.T) {
	pool, assoc, _ := testdb.NewTwoAssociations(t)
	repo := postgres.NewContracts()
	eq := usdEquivalent()
	inAssoc(t, pool, assoc, func(ctx context.Context) {
		created, err := repo.Create(ctx, usdTerms, &eq)
		if err != nil {
			t.Fatal(err)
		}
		got, err := repo.Get(ctx, created.ID)
		if err != nil {
			t.Fatal(err)
		}
		assertEq(t, got.UAHEquivalent, eq)
		list, err := repo.List(ctx, ports.ContractFilter{}, nil, 10)
		if err != nil || len(list) != 1 {
			t.Fatalf("List = %v, %v", list, err)
		}
		assertEq(t, list[0].UAHEquivalent, eq)
	})
}

// AC-01: a contract created without an equivalent reads back with none.
func TestContractsCreateWithoutEquivalentReadsBackNone(t *testing.T) {
	pool, assoc, _ := testdb.NewTwoAssociations(t)
	repo := postgres.NewContracts()
	inAssoc(t, pool, assoc, func(ctx context.Context) {
		created, err := repo.Create(ctx, usdTerms, nil)
		if err != nil {
			t.Fatal(err)
		}
		got, err := repo.Get(ctx, created.ID)
		if err != nil {
			t.Fatal(err)
		}
		if got.UAHEquivalent != nil {
			t.Fatalf("UAHEquivalent = %+v, want nil", got.UAHEquivalent)
		}
	})
}

// AC-06: the equivalent is set once on a contract that has none.
func TestContractsSetUAHEquivalentOnContractWithNone(t *testing.T) {
	pool, assoc, _ := testdb.NewTwoAssociations(t)
	repo := postgres.NewContracts()
	eq := usdEquivalent()
	inAssoc(t, pool, assoc, func(ctx context.Context) {
		created, err := repo.Create(ctx, usdTerms, nil)
		if err != nil {
			t.Fatal(err)
		}
		c, ok, err := repo.SetUAHEquivalent(ctx, created.ID, eq)
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Fatal("SetUAHEquivalent ok = false, want true on a contract with no equivalent")
		}
		assertEq(t, c.UAHEquivalent, eq)
		got, err := repo.Get(ctx, created.ID)
		if err != nil {
			t.Fatal(err)
		}
		assertEq(t, got.UAHEquivalent, eq)
	})
}

// AC-06: an equivalent once set is never changed.
func TestContractsSetUAHEquivalentChangesNothingWhenSet(t *testing.T) {
	pool, assoc, _ := testdb.NewTwoAssociations(t)
	repo := postgres.NewContracts()
	first := usdEquivalent()
	inAssoc(t, pool, assoc, func(ctx context.Context) {
		created, err := repo.Create(ctx, usdTerms, &first)
		if err != nil {
			t.Fatal(err)
		}
		other := domain.UAHEquivalent{AmountKopiyky: 99_999, Rate: domain.Rate{Currency: domain.CurrencyUSD, Day: usdTerms.SignedAt, Micros: 99_999_000}}
		_, ok, err := repo.SetUAHEquivalent(ctx, created.ID, other)
		if err != nil {
			t.Fatal(err)
		}
		if ok {
			t.Fatal("SetUAHEquivalent ok = true, want false on a contract that already has an equivalent")
		}
		got, err := repo.Get(ctx, created.ID)
		if err != nil {
			t.Fatal(err)
		}
		assertEq(t, got.UAHEquivalent, first)
	})
}

// AC-06: another association's contract is not found and stays untouched.
func TestContractsSetUAHEquivalentIgnoresOtherAssociation(t *testing.T) {
	pool, a, b := testdb.NewTwoAssociations(t)
	repo := postgres.NewContracts()
	var id uuid.UUID
	inAssoc(t, pool, a, func(ctx context.Context) {
		created, err := repo.Create(ctx, usdTerms, nil)
		if err != nil {
			t.Fatal(err)
		}
		id = created.ID
	})
	inAssoc(t, pool, b, func(ctx context.Context) {
		_, ok, err := repo.SetUAHEquivalent(ctx, id, usdEquivalent())
		if err != nil {
			t.Fatal(err)
		}
		if ok {
			t.Fatal("SetUAHEquivalent ok = true, want false for another association's contract")
		}
		if _, err := repo.Get(ctx, id); !errors.Is(err, domain.ErrContractNotFound) {
			t.Fatalf("Get err = %v, want ErrContractNotFound", err)
		}
	})
	inAssoc(t, pool, a, func(ctx context.Context) {
		got, err := repo.Get(ctx, id)
		if err != nil {
			t.Fatal(err)
		}
		if got.UAHEquivalent != nil {
			t.Fatalf("UAHEquivalent = %+v, want nil", got.UAHEquivalent)
		}
	})
}
