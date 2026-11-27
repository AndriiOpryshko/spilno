package finance_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/AndriiOpryshko/spilno/services/core/internal/finance"
)

// AC-10: with RatesMCPURL set the module reads rates over MCP at that endpoint.
func TestModuleReadsRatesFromTheConfiguredMCPURL(t *testing.T) {
	var hits atomic.Int32
	mcp := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		http.Error(w, "down", http.StatusServiceUnavailable)
	}))
	defer mcp.Close()

	srv, _, _ := newGuardedEnv(t, finance.Config{RatesMCPURL: mcp.URL})
	do(t, "POST", contractsURL(srv, assocID), eurBody, http.StatusCreated, nil)

	if hits.Load() == 0 {
		t.Error("the MCP endpoint received no request: RatesMCPURL is not wired to the rate provider")
	}
}

// AC-04: with neither Rates nor RatesMCPURL the module still serves contracts and answers
// 503 finance.rate_unavailable on the resolve operation.
func TestModuleWithoutRateSourceAnswersRateUnavailable(t *testing.T) {
	srv, _, _ := newGuardedEnv(t, finance.Config{})
	base := contractsURL(srv, assocID)

	var created contractEq
	do(t, "POST", base, eurBody, http.StatusCreated, &created)

	var p coded
	do(t, "POST", base+"/"+created.ID+"/uah-equivalent", "", http.StatusServiceUnavailable, &p)
	if p.Code != "finance.rate_unavailable" {
		t.Errorf("code = %q, want finance.rate_unavailable", p.Code)
	}
}

// T32 (AC-10): an endpoint net/http would print escaped is refused when the module is built: the source
// stays disabled, a warning names the refusal, no lookup goes out and no secret reaches the log or a body.
func TestModuleRefusesARateSourceAddressThatNetHTTPPrintsEscaped(t *testing.T) {
	tests := []struct {
		name   string
		url    func(srv *httptest.Server) string
		secret string
	}{
		{"quote in query", func(s *httptest.Server) string { return s.URL + `/mcp?token="quoted-secret"` }, "quoted-secret"},
		{"quote in password", func(s *httptest.Server) string {
			return strings.Replace(s.URL, "http://", `http://u:pa"ss-secret@`, 1) + "/mcp"
		}, "ss-secret"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var hits atomic.Int32
			mcp := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				hits.Add(1)
				http.Error(w, "down", http.StatusServiceUnavailable)
			}))
			defer mcp.Close()

			var logs bytes.Buffer
			prev := slog.Default()
			slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
			t.Cleanup(func() { slog.SetDefault(prev) })

			srv, _, _ := newGuardedEnv(t, finance.Config{RatesMCPURL: tc.url(mcp)})
			base := contractsURL(srv, assocID)
			var created contractEq
			do(t, "POST", base, eurBody, http.StatusCreated, &created)
			var p coded
			var raw json.RawMessage
			do(t, "POST", base+"/"+created.ID+"/uah-equivalent", "", http.StatusServiceUnavailable, &raw)
			if err := json.Unmarshal(raw, &p); err != nil {
				t.Fatal(err)
			}

			if p.Code != "finance.rate_unavailable" {
				t.Errorf("code = %q, want finance.rate_unavailable", p.Code)
			}
			if n := hits.Load(); n != 0 {
				t.Errorf("the refused endpoint received %d requests, want 0", n)
			}
			out := logs.String()
			if !strings.Contains(out, "level=WARN") || !strings.Contains(out, "refus") {
				t.Errorf("log %q, want a warning naming the refusal", out)
			}
			for _, leak := range []string{tc.secret, "pa\"", "u:pa"} {
				if strings.Contains(out, leak) || strings.Contains(string(raw), leak) {
					t.Errorf("log %q or body %s leaks %q", out, raw, leak)
				}
			}
		})
	}
}
