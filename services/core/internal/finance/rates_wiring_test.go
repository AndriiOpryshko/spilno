package finance_test

import (
	"net/http"
	"net/http/httptest"
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
