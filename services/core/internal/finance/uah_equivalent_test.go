package finance_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/AndriiOpryshko/spilno/services/core/internal/finance"
	"github.com/AndriiOpryshko/spilno/services/core/internal/finance/domain"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/testdb"
)

// fakeRates is the rate port of these tests: no call reaches the real NBU.
type fakeRates struct {
	mu    sync.Mutex
	calls int
	// decimal is the rate answered; with step set, every call answers a different rate.
	decimal string
	step    bool
	err     error
}

func (f *fakeRates) Rate(_ context.Context, c domain.Currency, day time.Time) (domain.Rate, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	if f.err != nil {
		return domain.Rate{}, f.err
	}
	dec := f.decimal
	if f.step {
		dec = fmt.Sprintf("50.%04d", f.calls)
	}
	return domain.ParseRate(c, day, dec)
}

func (f *fakeRates) set(decimal string, step bool, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.decimal, f.step, f.err = decimal, step, err
}

func (f *fakeRates) callCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls
}

type equivalent struct {
	AmountKopiyky int64  `json:"amount_kopiyky"`
	Rate          string `json:"rate"`
	RateDate      string `json:"rate_date"`
	Source        string `json:"source"`
}

type contractEq struct {
	ID            string      `json:"id"`
	UAHEquivalent *equivalent `json:"uah_equivalent"`
}

type coded struct {
	Status int    `json:"status"`
	Code   string `json:"code"`
}

const eurBody = `{"counterparty":"Kyivteploenergo","amount_cents":125000,"currency":"EUR","signed_at":"2026-10-20"}`

func contractsURL(srv *httptest.Server, assoc string) string {
	return srv.URL + "/finance/associations/" + assoc + "/contracts"
}

func eqString(e *equivalent) string {
	if e == nil {
		return "null"
	}
	return fmt.Sprintf("%+v", *e)
}

func wantEq(t *testing.T, what string, got *equivalent, want equivalent) {
	t.Helper()
	if got == nil || *got != want {
		t.Errorf("%s: uah_equivalent = %s, want %+v", what, eqString(got), want)
	}
}

var eurAt5053 = equivalent{AmountKopiyky: 6316388, Rate: "50.5311", RateDate: "2026-10-20", Source: "NBU"}

// AC-01: create, get and list carry the same equivalent, with the rate and its date.
func TestContractAnswersCarryTheSameUahEquivalent(t *testing.T) {
	rates := &fakeRates{decimal: "50.5311"}
	srv, _, _ := newGuardedEnv(t, finance.Config{Rates: rates})
	base := contractsURL(srv, assocID)

	var created contractEq
	do(t, "POST", base, eurBody, http.StatusCreated, &created)
	wantEq(t, "create", created.UAHEquivalent, eurAt5053)

	var got contractEq
	do(t, "GET", base+"/"+created.ID, "", http.StatusOK, &got)
	wantEq(t, "get", got.UAHEquivalent, eurAt5053)

	var list struct {
		Items []contractEq `json:"items"`
	}
	do(t, "GET", base, "", http.StatusOK, &list)
	if len(list.Items) != 1 {
		t.Fatalf("list items = %d, want 1", len(list.Items))
	}
	wantEq(t, "list", list.Items[0].UAHEquivalent, eurAt5053)
}

// AC-05: a contract created while the rate was unreadable gets its equivalent on demand,
// and every later answer carries it.
func TestResolveContractUahEquivalentStoresTheEquivalent(t *testing.T) {
	rates := &fakeRates{err: domain.ErrRateUnavailable}
	srv, _, _ := newGuardedEnv(t, finance.Config{Rates: rates})
	base := contractsURL(srv, assocID)

	var created contractEq
	do(t, "POST", base, eurBody, http.StatusCreated, &created)
	if created.UAHEquivalent != nil {
		t.Fatalf("create with a failing rate source: uah_equivalent = %s, want null", eqString(created.UAHEquivalent))
	}

	rates.set("50.5311", false, nil)
	var resolved contractEq
	do(t, "POST", base+"/"+created.ID+"/uah-equivalent", "", http.StatusOK, &resolved)
	if resolved.ID != created.ID {
		t.Errorf("resolved id = %s, want %s", resolved.ID, created.ID)
	}
	wantEq(t, "resolve", resolved.UAHEquivalent, eurAt5053)

	var got contractEq
	do(t, "GET", base+"/"+created.ID, "", http.StatusOK, &got)
	wantEq(t, "get after resolve", got.UAHEquivalent, eurAt5053)
}

// AC-06: computed once. A repeat answers the stored value without asking the source, even
// when the source now says something else.
func TestResolveContractUahEquivalentIsComputedOnce(t *testing.T) {
	rates := &fakeRates{err: domain.ErrRateUnavailable}
	srv, _, _ := newGuardedEnv(t, finance.Config{Rates: rates})
	base := contractsURL(srv, assocID)

	var created contractEq
	do(t, "POST", base, eurBody, http.StatusCreated, &created)
	rates.set("50.5311", false, nil)
	var first contractEq
	do(t, "POST", base+"/"+created.ID+"/uah-equivalent", "", http.StatusOK, &first)
	wantEq(t, "first", first.UAHEquivalent, eurAt5053)

	calls := rates.callCount()
	rates.set("99.9999", false, nil)
	var second contractEq
	do(t, "POST", base+"/"+created.ID+"/uah-equivalent", "", http.StatusOK, &second)
	wantEq(t, "second", second.UAHEquivalent, eurAt5053)
	if got := rates.callCount(); got != calls {
		t.Errorf("rate source asked %d times after the equivalent was stored, want %d", got, calls)
	}
}

// AC-06: parallel calls end with one stored value that every caller was given.
func TestResolveContractUahEquivalentConcurrentCallsStoreOneValue(t *testing.T) {
	rates := &fakeRates{err: domain.ErrRateUnavailable}
	srv, _, _ := newGuardedEnv(t, finance.Config{Rates: rates})
	base := contractsURL(srv, assocID)

	var created contractEq
	do(t, "POST", base, eurBody, http.StatusCreated, &created)
	// Every lookup answers a different rate, so two stored values would differ.
	rates.set("", true, nil)

	const n = 8
	answers := make([]contractEq, n)
	var wg sync.WaitGroup
	for i := range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, err := http.Post(base+"/"+created.ID+"/uah-equivalent", "application/json", nil)
			if err != nil {
				t.Errorf("call %d: %v", i, err)
				return
			}
			defer resp.Body.Close()
			b, _ := io.ReadAll(resp.Body)
			if resp.StatusCode != http.StatusOK {
				t.Errorf("call %d: status = %d, want 200; body %s", i, resp.StatusCode, b)
				return
			}
			if err := json.Unmarshal(b, &answers[i]); err != nil {
				t.Errorf("call %d: decode %s: %v", i, b, err)
			}
		}()
	}
	wg.Wait()

	var stored contractEq
	do(t, "GET", base+"/"+created.ID, "", http.StatusOK, &stored)
	if stored.UAHEquivalent == nil {
		t.Fatal("stored uah_equivalent = null, want the resolved value")
	}
	for i, a := range answers {
		if a.UAHEquivalent == nil || *a.UAHEquivalent != *stored.UAHEquivalent {
			t.Errorf("call %d answered %s, want the stored %s", i, eqString(a.UAHEquivalent), eqString(stored.UAHEquivalent))
		}
	}
}

// AC-07: every refusal is problem+json with its code and leaves the contract unchanged.
func TestResolveContractUahEquivalentRefusals(t *testing.T) {
	const uahBody = `{"counterparty":"A","amount_cents":100,"currency":"UAH","signed_at":"2026-10-20"}`
	tests := []struct {
		name       string
		disabled   bool
		body       string
		rateErr    error
		wantStatus int
		wantCode   string
	}{
		{"uah contract", false, uahBody, domain.ErrRateUnavailable, 409, "finance.equivalent_not_applicable"},
		{"rate not published", false, eurBody, domain.ErrRateNotPublished, 409, "finance.rate_not_published"},
		{"rate source down", false, eurBody, domain.ErrRateUnavailable, 503, "finance.rate_unavailable"},
		{"rate source not configured", true, eurBody, nil, 503, "finance.rate_unavailable"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rates := &fakeRates{err: domain.ErrRateUnavailable}
			cfg := finance.Config{Rates: rates}
			if tt.disabled {
				cfg = finance.Config{}
			}
			srv, _, _ := newGuardedEnv(t, cfg)
			base := contractsURL(srv, assocID)

			var created contractEq
			do(t, "POST", base, tt.body, http.StatusCreated, &created)
			rates.set("", false, tt.rateErr)

			var p coded
			do(t, "POST", base+"/"+created.ID+"/uah-equivalent", "", tt.wantStatus, &p)
			if p.Code != tt.wantCode || p.Status != tt.wantStatus {
				t.Errorf("problem = %+v, want status %d code %s", p, tt.wantStatus, tt.wantCode)
			}

			var got contractEq
			do(t, "GET", base+"/"+created.ID, "", http.StatusOK, &got)
			if got.UAHEquivalent != nil {
				t.Errorf("contract after the refusal: uah_equivalent = %s, want null", eqString(got.UAHEquivalent))
			}
		})
	}
}

// AC-07: an unknown contract is a plain 404 problem.
func TestResolveContractUahEquivalentUnknownContract(t *testing.T) {
	srv, _, _ := newGuardedEnv(t, finance.Config{Rates: &fakeRates{decimal: "50.5311"}})
	do(t, "POST", contractsURL(srv, assocID)+"/0b0e6f2a-5c1d-4c4e-9b8a-0f6c2d1e3a4b/uah-equivalent", "", http.StatusNotFound, nil)
}

// AC-08: no session is 401, a stranger and a disabled association are 403, and another
// association's contract is a 404 like a missing one.
func TestResolveContractUahEquivalentAuthorization(t *testing.T) {
	rates := &fakeRates{err: domain.ErrRateUnavailable}
	srv, _, router := newGuardedEnv(t, finance.Config{Rates: rates})
	base := contractsURL(srv, assocID)

	var created contractEq
	do(t, "POST", base, eurBody, http.StatusCreated, &created)
	rates.set("50.5311", false, nil)
	path := "/finance/associations/%s/contracts/%s/uah-equivalent"

	t.Run("no session", func(t *testing.T) {
		bare := httptest.NewServer(router)
		defer bare.Close()
		resp, err := http.Post(bare.URL+fmt.Sprintf(path, assocID, created.ID), "application/json", nil)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusUnauthorized || !strings.Contains(string(b), "identity.session_required") {
			t.Errorf("status = %d body = %s, want 401 identity.session_required", resp.StatusCode, b)
		}
	})

	t.Run("not a member", func(t *testing.T) {
		var p coded
		do(t, "POST", srv.URL+fmt.Sprintf(path, "0191a000-0000-7000-8000-0000000000ff", created.ID), "", http.StatusForbidden, &p)
		if p.Code != "identity.not_current_member" {
			t.Errorf("code = %q, want identity.not_current_member", p.Code)
		}
	})

	t.Run("another association", func(t *testing.T) {
		var p, missing coded
		do(t, "POST", srv.URL+fmt.Sprintf(path, otherAssocID, created.ID), "", http.StatusNotFound, &p)
		do(t, "POST", srv.URL+fmt.Sprintf(path, otherAssocID, "0b0e6f2a-5c1d-4c4e-9b8a-0f6c2d1e3a4b"), "", http.StatusNotFound, &missing)
		if p != missing {
			t.Errorf("answer for another association's id = %+v, want the one of a missing id %+v", p, missing)
		}
		var got contractEq
		do(t, "GET", base+"/"+created.ID, "", http.StatusOK, &got)
		if got.UAHEquivalent != nil {
			t.Errorf("contract of A after a call under B: uah_equivalent = %s, want null", eqString(got.UAHEquivalent))
		}
	})
}

// AC-08: a member of a disabled association is refused with identity.association_disabled.
func TestResolveContractUahEquivalentOnADisabledAssociation(t *testing.T) {
	rates := &fakeRates{err: domain.ErrRateUnavailable}
	pool, a, _ := testdb.NewTwoAssociations(t)
	srv, _, _ := newGuardedEnvOn(t, pool, finance.Config{Rates: rates}, a)
	base := contractsURL(srv, a)

	var created contractEq
	do(t, "POST", base, eurBody, http.StatusCreated, &created)
	rates.set("50.5311", false, nil)
	do(t, "POST", srv.URL+"/identity/operator/associations/"+a+"/disabling",
		`{"reason":"спір щодо складу правління"}`, http.StatusOK, nil)

	var p coded
	do(t, "POST", base+"/"+created.ID+"/uah-equivalent", "", http.StatusForbidden, &p)
	if p.Code != "identity.association_disabled" {
		t.Errorf("code = %q, want identity.association_disabled", p.Code)
	}
}
