package finance_test

import (
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"testing"

	"github.com/AndriiOpryshko/spilno/services/core/internal/finance"
	"github.com/AndriiOpryshko/spilno/services/core/internal/finance/domain"
)

type eqPage struct {
	Items      []contractEq `json:"items"`
	NextCursor *string      `json:"next_cursor"`
}

func createEq(t *testing.T, base, body string) string {
	t.Helper()
	var c contractEq
	do(t, "POST", base, body, http.StatusCreated, &c)
	return c.ID
}

func contractBody(currency string, i int) string {
	return fmt.Sprintf(`{"counterparty":"Vendor %d","amount_cents":%d,"currency":%q,"signed_at":"2026-10-20"}`, i, 1000+i, currency)
}

// walk follows next_cursor from first and returns every id it saw, in order.
func walk(t *testing.T, first string) []string {
	t.Helper()
	var seen []string
	next := first
	for pages := 0; ; pages++ {
		if pages > 20 {
			t.Fatal("listing did not end")
		}
		var p eqPage
		do(t, "GET", next, "", http.StatusOK, &p)
		for _, c := range p.Items {
			seen = append(seen, c.ID)
		}
		if p.NextCursor == nil {
			return seen
		}
		next = first + "&cursor=" + url.QueryEscape(*p.NextCursor)
	}
}

// Issue #70 AC-1, AC-2: uah_equivalent=missing answers only the foreign-currency
// contracts without an equivalent; without it the list is unchanged.
func TestListContractsMissingUahEquivalent(t *testing.T) {
	rates := &fakeRates{decimal: "50.5311"}
	srv, _, _ := newGuardedEnv(t, finance.Config{Rates: rates})
	base := contractsURL(srv, assocID)

	uah := createEq(t, base, contractBody("UAH", 1))
	eur := createEq(t, base, contractBody("EUR", 2))
	rates.set("", false, domain.ErrRateUnavailable)
	usd := createEq(t, base, contractBody("USD", 3))

	var missing eqPage
	do(t, "GET", base+"?uah_equivalent=missing", "", http.StatusOK, &missing)
	if len(missing.Items) != 1 || missing.Items[0].ID != usd || missing.NextCursor != nil {
		t.Errorf("missing page = %+v, want only the USD contract %s and next_cursor null", missing, usd)
	}

	var all eqPage
	do(t, "GET", base, "", http.StatusOK, &all)
	var got []string
	for _, c := range all.Items {
		got = append(got, c.ID)
	}
	if want := []string{usd, eur, uah}; !slices.Equal(got, want) {
		t.Errorf("unfiltered = %v, want every contract newest first %v", got, want)
	}
}

// Issue #70 AC-3: the filtered listing pages newest first with limit and
// cursor, returns each missing contract once and skips none.
func TestListContractsMissingUahEquivalentPages(t *testing.T) {
	rates := &fakeRates{decimal: "50.5311"}
	srv, _, _ := newGuardedEnv(t, finance.Config{Rates: rates})
	base := contractsURL(srv, assocID)

	var want []string
	for i := range 7 {
		switch i % 3 {
		case 0:
			rates.set("50.5311", false, nil)
			createEq(t, base, contractBody("EUR", i))
		case 1:
			createEq(t, base, contractBody("UAH", i))
		default:
			rates.set("", false, domain.ErrRateUnavailable)
			want = append([]string{createEq(t, base, contractBody("USD", i))}, want...)
		}
	}

	for _, limit := range []int{1, 2, len(want)} {
		got := walk(t, fmt.Sprintf("%s?uah_equivalent=missing&limit=%d", base, limit))
		if !slices.Equal(got, want) {
			t.Errorf("limit=%d: walked %v, want newest first %v", limit, got, want)
		}
	}
}
