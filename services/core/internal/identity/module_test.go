package identity_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AndriiOpryshko/spilno/services/core/internal/identity"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/httpapi"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/testdb"
)

// newServer serves the context through the same router, validator and
// generated handlers as production, on a real migrated database.
func newServer(t *testing.T) *httptest.Server {
	t.Helper()
	r := httpapi.NewRouter()
	if err := identity.New(testdb.New(t)).Routes(r); err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)
	return srv
}

// The skeleton stubs every operation with 501. A 401 here would mean the
// session is already enforced, which only T11 does.
func TestSessionIsStubbedAndNotEnforced(t *testing.T) {
	srv := newServer(t)

	resp, err := http.Get(srv.URL + "/identity/session")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotImplemented {
		t.Fatalf("status = %d, want 501", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "application/problem+json" {
		t.Errorf("Content-Type = %q, want application/problem+json", ct)
	}
	var body struct {
		Status int `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Status != http.StatusNotImplemented {
		t.Errorf("body status = %d, want 501", body.Status)
	}
}
