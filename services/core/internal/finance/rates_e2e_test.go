package finance_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/AndriiOpryshko/spilno/services/core/internal/finance"
)

// AC-11, AC-01, AC-10: the chain finance -> MCP adapter -> the real nbu-rates binary -> a fake NBU.
// No request leaves the machine: the binary's NBU_BASE_URL points at an httptest server.
func TestUahEquivalentThroughTheRealNBURatesServer(t *testing.T) {
	if testing.Short() {
		t.Skip("builds and starts mcp/nbu-rates")
	}
	dir, err := filepath.Abs("../../../../mcp/nbu-rates")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Skipf("mcp/nbu-rates is absent (%v): the end-to-end chain cannot be built", err)
	}

	bin := filepath.Join(t.TempDir(), "nbu-rates")
	if out, err := exec.Command("go", "build", "-C", dir, "-o", bin, ".").CombinedOutput(); err != nil {
		t.Fatalf("go build mcp/nbu-rates: %v\n%s", err, out)
	}

	// The fake NBU publishes EUR on 2026-10-20 only; any other day has no row, as on a real holiday.
	var upstream atomic.Int32
	nbu := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstream.Add(1)
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("valcode") == "EUR" && r.URL.Query().Get("date") == "20261020" {
			_, _ = w.Write([]byte(`[{"r030":978,"txt":"Євро","rate":50.5311,"cc":"EUR","exchangedate":"20.10.2026"}]`))
			return
		}
		_, _ = w.Write([]byte(`[]`))
	}))
	t.Cleanup(nbu.Close)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, bin, "-http", addr)
	cmd.Env = append(os.Environ(), "NBU_BASE_URL="+nbu.URL, "NBU_RATES_HTTP_ADDR=")
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		cancel()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cancel()
		_ = cmd.Wait()
	})

	deadline := time.Now().Add(15 * time.Second)
	for {
		resp, err := http.Get("http://" + addr + "/healthz")
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				break
			}
		}
		if time.Now().After(deadline) {
			t.Fatalf("nbu-rates did not answer /healthz on %s", addr)
		}
		time.Sleep(50 * time.Millisecond)
	}

	srv, _, _ := newGuardedEnv(t, finance.Config{RatesMCPURL: fmt.Sprintf("http://%s/mcp", addr)})
	base := contractsURL(srv, assocID)

	// A published day: the equivalent is computed from the rate the real server read.
	var created contractEq
	do(t, "POST", base, eurBody, http.StatusCreated, &created)
	wantEq(t, "create", created.UAHEquivalent, eurAt5053)
	var got contractEq
	do(t, "GET", base+"/"+created.ID, "", http.StatusOK, &got)
	wantEq(t, "get", got.UAHEquivalent, eurAt5053)

	// A day the fake NBU does not publish: created without an equivalent, then 409 on resolve.
	const holidayBody = `{"counterparty":"Kyivteploenergo","amount_cents":125000,"currency":"EUR","signed_at":"2026-10-24"}`
	var holiday contractEq
	do(t, "POST", base, holidayBody, http.StatusCreated, &holiday)
	if holiday.UAHEquivalent != nil {
		t.Fatalf("unpublished day: uah_equivalent = %s, want null", eqString(holiday.UAHEquivalent))
	}
	var p coded
	do(t, "POST", base+"/"+holiday.ID+"/uah-equivalent", "", http.StatusConflict, &p)
	if p.Code != "finance.rate_not_published" {
		t.Errorf("code = %q, want finance.rate_not_published", p.Code)
	}
	if upstream.Load() < 2 {
		t.Errorf("the fake NBU received %d requests, want one per day: the rate did not come through the real server", upstream.Load())
	}
}
