package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/AndriiOpryshko/spilno/services/core/internal/finance/app"
	"github.com/AndriiOpryshko/spilno/services/core/internal/finance/domain"
	"github.com/AndriiOpryshko/spilno/services/core/internal/finance/ports"
)

// fakeRepo serves List from a slice already in listing order and records the
// calls it gets.
type fakeRepo struct {
	ports.ContractRepository
	rows      []domain.Contract
	gotAfter  *ports.ContractKey
	gotLimit  int
	createdOK bool
	gotEq     *domain.UAHEquivalent

	stored    domain.Contract // served by Get
	getErr    error
	setOK     bool // result of SetUAHEquivalent
	setCalls  int
	setEq     domain.UAHEquivalent
	afterLost *domain.Contract // what Get serves once SetUAHEquivalent ran
}

func (f *fakeRepo) Get(_ context.Context, _ uuid.UUID) (domain.Contract, error) {
	if f.getErr != nil {
		return domain.Contract{}, f.getErr
	}
	if f.setCalls > 0 && f.afterLost != nil {
		return *f.afterLost, nil
	}
	return f.stored, nil
}

func (f *fakeRepo) SetUAHEquivalent(_ context.Context, _ uuid.UUID, eq domain.UAHEquivalent) (domain.Contract, bool, error) {
	f.setCalls++
	f.setEq = eq
	if !f.setOK {
		return domain.Contract{}, false, nil
	}
	c := f.stored
	c.UAHEquivalent = &eq
	return c, true, nil
}

func (f *fakeRepo) Create(_ context.Context, t domain.ContractTerms, eq *domain.UAHEquivalent) (domain.Contract, error) {
	f.createdOK = true
	f.gotEq = eq
	return domain.Contract{ID: uuid.New(), ContractTerms: t, UAHEquivalent: eq}, nil
}

func (f *fakeRepo) List(_ context.Context, after *ports.ContractKey, limit int) ([]domain.Contract, error) {
	f.gotAfter, f.gotLimit = after, limit
	start := 0
	if after != nil {
		for i, c := range f.rows {
			if c.ID == after.ID {
				start = i + 1
			}
		}
	}
	return f.rows[start:min(start+limit, len(f.rows))], nil
}

func rows(n int) []domain.Contract {
	base := time.Date(2026, 10, 21, 12, 0, 0, 123456000, time.UTC)
	out := make([]domain.Contract, n)
	for i := range out {
		out[i] = domain.Contract{ID: uuid.New(), CreatedAt: base.Add(-time.Duration(i) * time.Minute)}
	}
	return out
}

func TestCreateRejectsInvalidTermsBeforeStoring(t *testing.T) {
	repo := &fakeRepo{}
	_, err := app.NewContractService(repo, &fakeRates{}).Create(context.Background(), domain.ContractTerms{})
	var v *domain.ValidationError
	if !errors.As(err, &v) {
		t.Fatalf("err = %v, want *ValidationError", err)
	}
	if repo.createdOK {
		t.Error("invalid terms reached the repository")
	}
}

func TestListWalksAllPages(t *testing.T) {
	repo := &fakeRepo{rows: rows(5)}
	svc := app.NewContractService(repo, &fakeRates{})

	var seen []uuid.UUID
	cursor := ""
	for range 10 {
		page, err := svc.List(context.Background(), 2, cursor)
		if err != nil {
			t.Fatal(err)
		}
		for _, c := range page.Items {
			seen = append(seen, c.ID)
		}
		if page.NextCursor == "" {
			break
		}
		if repo.gotLimit != 3 {
			t.Errorf("repo limit = %d, want page size + 1", repo.gotLimit)
		}
		cursor = page.NextCursor
	}
	if len(seen) != 5 {
		t.Fatalf("saw %d contracts, want 5", len(seen))
	}
	for i, c := range repo.rows {
		if seen[i] != c.ID {
			t.Errorf("item %d = %v, want %v", i, seen[i], c.ID)
		}
	}
}

func TestListCursorKeepsMicroseconds(t *testing.T) {
	repo := &fakeRepo{rows: rows(2)}
	svc := app.NewContractService(repo, &fakeRates{})
	page, err := svc.List(context.Background(), 1, "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.List(context.Background(), 1, page.NextCursor); err != nil {
		t.Fatal(err)
	}
	want := repo.rows[0]
	if repo.gotAfter == nil || repo.gotAfter.ID != want.ID || !repo.gotAfter.CreatedAt.Equal(want.CreatedAt) {
		t.Errorf("after = %+v, want {%v %v}", repo.gotAfter, want.CreatedAt, want.ID)
	}
}

func TestListLimitBounds(t *testing.T) {
	tests := []struct{ limit, wantRepo int }{
		{0, app.DefaultPageSize + 1},
		{-3, app.DefaultPageSize + 1},
		{1, 2},
		{app.MaxPageSize, app.MaxPageSize + 1},
		{app.MaxPageSize + 1, app.MaxPageSize + 1},
	}
	for _, tt := range tests {
		repo := &fakeRepo{}
		if _, err := app.NewContractService(repo, &fakeRates{}).List(context.Background(), tt.limit, ""); err != nil {
			t.Fatal(err)
		}
		if repo.gotLimit != tt.wantRepo {
			t.Errorf("List(limit=%d): repo limit = %d, want %d", tt.limit, repo.gotLimit, tt.wantRepo)
		}
	}
}

func TestListRejectsForeignCursor(t *testing.T) {
	svc := app.NewContractService(&fakeRepo{}, &fakeRates{})
	for _, c := range []string{"not base64!", "bm9jb2xvbg", "eDpub3QtYS11dWlk"} {
		if _, err := svc.List(context.Background(), 10, c); !errors.Is(err, app.ErrInvalidCursor) {
			t.Errorf("List(cursor=%q) err = %v, want ErrInvalidCursor", c, err)
		}
	}
}

// fakeRates answers with rate/err, or blocks until the context ends when block is set.
type fakeRates struct {
	rate  domain.Rate
	err   error
	block bool
	calls int
}

func (f *fakeRates) Rate(ctx context.Context, c domain.Currency, day time.Time) (domain.Rate, error) {
	f.calls++
	if f.block {
		<-ctx.Done()
		return domain.Rate{}, ctx.Err()
	}
	if f.err != nil {
		return domain.Rate{}, f.err
	}
	r := f.rate
	r.Currency, r.Day = c, day
	return r, nil
}

func terms(cur domain.Currency) domain.ContractTerms {
	return domain.ContractTerms{
		Counterparty: "Test Counterparty",
		AmountCents:  125000,
		Currency:     cur,
		SignedAt:     time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC),
	}
}

func TestRateTimeoutIsThreeSeconds(t *testing.T) {
	if app.RateTimeout != 3*time.Second {
		t.Errorf("RateTimeout = %v, want 3s", app.RateTimeout)
	}
}

func TestCreateForeignContractStoresEquivalent(t *testing.T) {
	rate, err := domain.ParseRate(domain.CurrencyEUR, time.Time{}, "50.5311")
	if err != nil {
		t.Fatal(err)
	}
	repo, rates := &fakeRepo{}, &fakeRates{rate: rate}
	c, err := app.NewContractService(repo, rates).Create(context.Background(), terms(domain.CurrencyEUR))
	if err != nil {
		t.Fatal(err)
	}
	if repo.gotEq == nil || repo.gotEq.AmountKopiyky != 6316388 || repo.gotEq.Rate.Decimal() != "50.5311" {
		t.Fatalf("stored equivalent = %+v, want 6316388 kopiyky at 50.5311", repo.gotEq)
	}
	if c.UAHEquivalent == nil {
		t.Error("returned contract has no equivalent")
	}
}

func TestCreateHryvniaContractDoesNotAskRates(t *testing.T) {
	repo, rates := &fakeRepo{}, &fakeRates{}
	if _, err := app.NewContractService(repo, rates).Create(context.Background(), terms(domain.CurrencyUAH)); err != nil {
		t.Fatal(err)
	}
	if rates.calls != 0 {
		t.Errorf("rate port called %d times for a UAH contract", rates.calls)
	}
	if !repo.createdOK || repo.gotEq != nil {
		t.Errorf("created=%v equivalent=%+v, want stored without equivalent", repo.createdOK, repo.gotEq)
	}
}

func TestCreateStoresWithoutEquivalentWhenRateFails(t *testing.T) {
	old := app.RateTimeout
	app.RateTimeout = 100 * time.Millisecond
	t.Cleanup(func() { app.RateTimeout = old })

	cases := map[string]*fakeRates{
		"not published": {err: domain.ErrRateNotPublished},
		"unavailable":   {err: domain.ErrRateUnavailable},
		"slow":          {block: true},
	}
	for name, rates := range cases {
		t.Run(name, func(t *testing.T) {
			repo := &fakeRepo{}
			start := time.Now()
			_, err := app.NewContractService(repo, rates).Create(context.Background(), terms(domain.CurrencyUSD))
			if err != nil {
				t.Fatalf("Create failed over the rate: %v", err)
			}
			if rates.calls != 1 {
				t.Errorf("rate port called %d times, want 1", rates.calls)
			}
			if !repo.createdOK || repo.gotEq != nil {
				t.Errorf("created=%v equivalent=%+v, want stored without equivalent", repo.createdOK, repo.gotEq)
			}
			if d := time.Since(start); d > time.Second {
				t.Errorf("Create took %v, want within the budget", d)
			}
		})
	}
}

// An equivalent that does not fit int64 is not stored; the contract still is.
func TestCreateStoresWithoutEquivalentWhenOutOfRange(t *testing.T) {
	rate, err := domain.ParseRate(domain.CurrencyEUR, time.Time{}, "50.5311")
	if err != nil {
		t.Fatal(err)
	}
	huge := terms(domain.CurrencyEUR)
	huge.AmountCents = 200_000_000_000_000_000
	repo := &fakeRepo{}
	if _, err := app.NewContractService(repo, &fakeRates{rate: rate}).Create(context.Background(), huge); err != nil {
		t.Fatalf("Create failed over the equivalent: %v", err)
	}
	if !repo.createdOK || repo.gotEq != nil {
		t.Errorf("created=%v equivalent=%+v, want stored without equivalent", repo.createdOK, repo.gotEq)
	}
}

func TestCreateInvalidTermsDoNotAskRates(t *testing.T) {
	rates := &fakeRates{}
	bad := terms(domain.CurrencyEUR)
	bad.AmountCents = 0
	if _, err := app.NewContractService(&fakeRepo{}, rates).Create(context.Background(), bad); err == nil {
		t.Fatal("want validation error")
	}
	if rates.calls != 0 {
		t.Errorf("rate port called %d times for invalid terms", rates.calls)
	}
}

func storedContract(cur domain.Currency, eq *domain.UAHEquivalent) domain.Contract {
	return domain.Contract{ID: uuid.New(), ContractTerms: terms(cur), UAHEquivalent: eq}
}

func wantCode(t *testing.T, err error, code string) {
	t.Helper()
	var de *domain.DomainError
	if !errors.As(err, &de) || de.Code != code {
		t.Fatalf("err = %v, want *DomainError with code %s", err, code)
	}
}

func TestResolveUAHEquivalentFillsMissingEquivalent(t *testing.T) {
	rate, err := domain.ParseRate(domain.CurrencyEUR, time.Time{}, "50.5311")
	if err != nil {
		t.Fatal(err)
	}
	repo := &fakeRepo{stored: storedContract(domain.CurrencyEUR, nil), setOK: true}
	rates := &fakeRates{rate: rate}
	c, err := app.NewContractService(repo, rates).ResolveUAHEquivalent(context.Background(), repo.stored.ID)
	if err != nil {
		t.Fatal(err)
	}
	if rates.calls != 1 || repo.setCalls != 1 {
		t.Errorf("rate calls=%d set calls=%d, want 1 and 1", rates.calls, repo.setCalls)
	}
	if c.UAHEquivalent == nil || c.UAHEquivalent.AmountKopiyky != 6316388 || c.UAHEquivalent.Rate.Decimal() != "50.5311" {
		t.Fatalf("returned equivalent = %+v, want 6316388 kopiyky at 50.5311", c.UAHEquivalent)
	}
}

func TestResolveUAHEquivalentReturnsStoredOneWithoutAskingRates(t *testing.T) {
	rate, _ := domain.ParseRate(domain.CurrencyEUR, time.Time{}, "50.5311")
	eq := domain.UAHEquivalent{AmountKopiyky: 1, Rate: rate}
	repo := &fakeRepo{stored: storedContract(domain.CurrencyEUR, &eq)}
	rates := &fakeRates{}
	c, err := app.NewContractService(repo, rates).ResolveUAHEquivalent(context.Background(), repo.stored.ID)
	if err != nil {
		t.Fatal(err)
	}
	if rates.calls != 0 || repo.setCalls != 0 {
		t.Errorf("rate calls=%d set calls=%d, want none", rates.calls, repo.setCalls)
	}
	if c.UAHEquivalent == nil || c.UAHEquivalent.AmountKopiyky != 1 {
		t.Errorf("equivalent = %+v, want the stored one unchanged", c.UAHEquivalent)
	}
}

func TestResolveUAHEquivalentRefusesHryvniaContract(t *testing.T) {
	repo := &fakeRepo{stored: storedContract(domain.CurrencyUAH, nil)}
	rates := &fakeRates{}
	_, err := app.NewContractService(repo, rates).ResolveUAHEquivalent(context.Background(), repo.stored.ID)
	wantCode(t, err, domain.CodeEquivalentNotApplicable)
	if rates.calls != 0 || repo.setCalls != 0 {
		t.Errorf("rate calls=%d set calls=%d, want none", rates.calls, repo.setCalls)
	}
}

func TestResolveUAHEquivalentRateFailures(t *testing.T) {
	old := app.RateTimeout
	app.RateTimeout = 100 * time.Millisecond
	t.Cleanup(func() { app.RateTimeout = old })

	cases := map[string]struct {
		rates *fakeRates
		code  string
	}{
		"not published": {&fakeRates{err: domain.ErrRateNotPublished}, domain.CodeRateNotPublished},
		"unavailable":   {&fakeRates{err: domain.ErrRateUnavailable}, domain.CodeRateUnavailable},
		"slow":          {&fakeRates{block: true}, domain.CodeRateUnavailable},
	}
	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			repo := &fakeRepo{stored: storedContract(domain.CurrencyUSD, nil), setOK: true}
			_, err := app.NewContractService(repo, tt.rates).ResolveUAHEquivalent(context.Background(), repo.stored.ID)
			wantCode(t, err, tt.code)
			if repo.setCalls != 0 {
				t.Error("an equivalent was stored after a failed lookup")
			}
		})
	}
}

func TestResolveUAHEquivalentRefusesOutOfRange(t *testing.T) {
	rate, err := domain.ParseRate(domain.CurrencyEUR, time.Time{}, "50.5311")
	if err != nil {
		t.Fatal(err)
	}
	c := storedContract(domain.CurrencyEUR, nil)
	c.AmountCents = 200_000_000_000_000_000
	repo := &fakeRepo{stored: c, setOK: true}
	_, err = app.NewContractService(repo, &fakeRates{rate: rate}).ResolveUAHEquivalent(context.Background(), c.ID)
	wantCode(t, err, domain.CodeEquivalentOutOfRange)
	if repo.setCalls != 0 {
		t.Error("an out-of-range equivalent was stored")
	}
}

func TestResolveUAHEquivalentLostRaceReturnsStoredContract(t *testing.T) {
	rate, _ := domain.ParseRate(domain.CurrencyEUR, time.Time{}, "50.5311")
	winner := domain.UAHEquivalent{AmountKopiyky: 42, Rate: rate}
	won := storedContract(domain.CurrencyEUR, &winner)
	repo := &fakeRepo{stored: storedContract(domain.CurrencyEUR, nil), setOK: false, afterLost: &won}
	c, err := app.NewContractService(repo, &fakeRates{rate: rate}).ResolveUAHEquivalent(context.Background(), repo.stored.ID)
	if err != nil {
		t.Fatal(err)
	}
	if c.UAHEquivalent == nil || c.UAHEquivalent.AmountKopiyky != 42 {
		t.Fatalf("equivalent = %+v, want the winner's stored 42", c.UAHEquivalent)
	}
}

func TestResolveUAHEquivalentUnknownContract(t *testing.T) {
	repo := &fakeRepo{getErr: domain.ErrContractNotFound}
	_, err := app.NewContractService(repo, &fakeRates{}).ResolveUAHEquivalent(context.Background(), uuid.New())
	if !errors.Is(err, domain.ErrContractNotFound) {
		t.Fatalf("err = %v, want ErrContractNotFound", err)
	}
}
