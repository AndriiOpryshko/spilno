package finance_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/AndriiOpryshko/spilno/services/core/internal/finance"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/httpapi"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/testdb"
)

// newServer serves the context through the same router, validator and
// generated handlers as production, on a real migrated database.
func newServer(t *testing.T) *httptest.Server {
	t.Helper()
	r := httpapi.NewRouter()
	if err := finance.New(testdb.New(t)).Routes(r); err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)
	return srv
}

type contract struct {
	ID           string  `json:"id"`
	Counterparty string  `json:"counterparty"`
	AmountCents  int64   `json:"amount_cents"`
	Currency     string  `json:"currency"`
	SignedAt     string  `json:"signed_at"`
	EndsAt       *string `json:"ends_at"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

type problem struct {
	Status int `json:"status"`
	Errors []struct {
		Pointer string `json:"pointer"`
	} `json:"errors"`
}

// do sends a request and decodes the JSON response into out, failing the test
// unless the status is want.
func do(t *testing.T, method, url, body string, want int, out any) {
	t.Helper()
	req, err := http.NewRequest(method, url, bytes.NewBufferString(body))
	if err != nil {
		t.Fatal(err)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var raw json.RawMessage
	_ = json.NewDecoder(resp.Body).Decode(&raw)
	if resp.StatusCode != want {
		t.Fatalf("%s %s: status = %d, want %d; body %s", method, url, resp.StatusCode, want, raw)
	}
	if want >= 400 {
		if ct := resp.Header.Get("Content-Type"); ct != "application/problem+json" {
			t.Errorf("Content-Type = %q, want application/problem+json", ct)
		}
	}
	if out != nil {
		if err := json.Unmarshal(raw, out); err != nil {
			t.Fatalf("decode %s: %v", raw, err)
		}
	}
}

func pointers(p problem) []string {
	var out []string
	for _, e := range p.Errors {
		out = append(out, e.Pointer)
	}
	return out
}

func TestCreateAndGetContract(t *testing.T) {
	srv := newServer(t)

	var created contract
	do(t, "POST", srv.URL+"/finance/contracts",
		`{"counterparty":"Kyivteploenergo","amount_cents":125000,"currency":"UAH","signed_at":"2026-01-15","ends_at":"2026-12-31"}`,
		http.StatusCreated, &created)
	if created.ID == "" || created.CreatedAt == "" || created.UpdatedAt == "" {
		t.Errorf("created = %+v, want id and timestamps", created)
	}
	if created.SignedAt != "2026-01-15" || created.EndsAt == nil || *created.EndsAt != "2026-12-31" {
		t.Errorf("dates = %s..%v, want 2026-01-15..2026-12-31", created.SignedAt, created.EndsAt)
	}

	var got contract
	do(t, "GET", srv.URL+"/finance/contracts/"+created.ID, "", http.StatusOK, &got)
	if got != (contract{ID: created.ID, Counterparty: "Kyivteploenergo", AmountCents: 125000, Currency: "UAH",
		SignedAt: "2026-01-15", EndsAt: got.EndsAt, CreatedAt: created.CreatedAt, UpdatedAt: created.UpdatedAt}) ||
		*got.EndsAt != "2026-12-31" {
		t.Errorf("got = %+v, want the created contract", got)
	}
}

func TestOpenEndedContractHasNullEndsAt(t *testing.T) {
	srv := newServer(t)

	var raw map[string]any
	do(t, "POST", srv.URL+"/finance/contracts",
		`{"counterparty":"Lift service","amount_cents":1,"currency":"EUR","signed_at":"2026-03-01"}`,
		http.StatusCreated, &raw)
	if v, ok := raw["ends_at"]; !ok || v != nil {
		t.Errorf("ends_at = %v (present %v), want null", v, ok)
	}
}

func TestCreateContractRejectsInvalidTerms(t *testing.T) {
	srv := newServer(t)

	tests := []struct {
		name, body  string
		wantPointer string
	}{
		{"ends before signed", `{"counterparty":"A","amount_cents":1,"currency":"UAH","signed_at":"2026-02-01","ends_at":"2026-01-31"}`, "/ends_at"},
		{"zero amount", `{"counterparty":"A","amount_cents":0,"currency":"UAH","signed_at":"2026-02-01"}`, "/amount_cents"},
		{"unknown currency", `{"counterparty":"A","amount_cents":1,"currency":"GBP","signed_at":"2026-02-01"}`, "/currency"},
		{"missing counterparty", `{"amount_cents":1,"currency":"UAH","signed_at":"2026-02-01"}`, "/counterparty"},
		{"empty counterparty", `{"counterparty":"","amount_cents":1,"currency":"UAH","signed_at":"2026-02-01"}`, "/counterparty"},
		{"missing signed_at", `{"counterparty":"A","amount_cents":1,"currency":"UAH"}`, "/signed_at"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p problem
			do(t, "POST", srv.URL+"/finance/contracts", tt.body, http.StatusUnprocessableEntity, &p)
			if tt.wantPointer != "" && !containsPointer(p, tt.wantPointer) {
				t.Errorf("pointers = %v, want %s", pointers(p), tt.wantPointer)
			}
		})
	}
}

func containsPointer(p problem, want string) bool {
	for _, got := range pointers(p) {
		if got == want {
			return true
		}
	}
	return false
}

func TestGetContractErrors(t *testing.T) {
	srv := newServer(t)

	do(t, "GET", srv.URL+"/finance/contracts/0b0e6f2a-5c1d-4c4e-9b8a-0f6c2d1e3a4b", "", http.StatusNotFound, nil)
	do(t, "GET", srv.URL+"/finance/contracts/not-a-uuid", "", http.StatusUnprocessableEntity, nil)
}

func TestListContractsPagesNewestFirst(t *testing.T) {
	srv := newServer(t)

	var ids []string
	for i := range 5 {
		var c contract
		do(t, "POST", srv.URL+"/finance/contracts",
			fmt.Sprintf(`{"counterparty":"Vendor %d","amount_cents":%d,"currency":"USD","signed_at":"2026-01-01"}`, i, i+1),
			http.StatusCreated, &c)
		ids = append([]string{c.ID}, ids...)
	}

	type page struct {
		Items      []contract `json:"items"`
		NextCursor *string    `json:"next_cursor"`
	}
	var seen []string
	next := srv.URL + "/finance/contracts?limit=2"
	for pages := 0; ; pages++ {
		if pages > 5 {
			t.Fatal("listing did not end")
		}
		var p page
		do(t, "GET", next, "", http.StatusOK, &p)
		for _, c := range p.Items {
			seen = append(seen, c.ID)
		}
		if p.NextCursor == nil {
			break
		}
		next = srv.URL + "/finance/contracts?limit=2&cursor=" + url.QueryEscape(*p.NextCursor)
	}
	if fmt.Sprint(seen) != fmt.Sprint(ids) {
		t.Errorf("listed %v, want newest first %v", seen, ids)
	}

	var all page
	do(t, "GET", srv.URL+"/finance/contracts", "", http.StatusOK, &all)
	if len(all.Items) != 5 || all.NextCursor != nil {
		t.Errorf("default page: %d items, next %v; want 5 and null", len(all.Items), all.NextCursor)
	}

	// A page filled exactly to the limit is still the last one when nothing follows.
	var exact page
	do(t, "GET", srv.URL+"/finance/contracts?limit=5", "", http.StatusOK, &exact)
	if len(exact.Items) != 5 || exact.NextCursor != nil {
		t.Errorf("limit=5 page: %d items, next %v; want 5 and null", len(exact.Items), exact.NextCursor)
	}
}

func TestListContractsEmptyIsArray(t *testing.T) {
	srv := newServer(t)

	var raw map[string]json.RawMessage
	do(t, "GET", srv.URL+"/finance/contracts", "", http.StatusOK, &raw)
	if string(raw["items"]) != "[]" || string(raw["next_cursor"]) != "null" {
		t.Errorf("body = items %s, next_cursor %s; want [] and null", raw["items"], raw["next_cursor"])
	}
}

func TestListContractsRejectsBadParams(t *testing.T) {
	srv := newServer(t)

	for _, q := range []string{"limit=0", "limit=101", "cursor=garbage"} {
		var p problem
		do(t, "GET", srv.URL+"/finance/contracts?"+q, "", http.StatusUnprocessableEntity, &p)
		if len(p.Errors) == 0 {
			t.Errorf("%s: no field errors", q)
		}
	}
}

func TestUnknownRouteIsProblem(t *testing.T) {
	srv := newServer(t)

	do(t, "GET", srv.URL+"/finance/nope", "", http.StatusNotFound, nil)
}
