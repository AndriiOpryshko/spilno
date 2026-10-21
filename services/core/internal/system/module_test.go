package system_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/httpapi"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/testdb"
	"github.com/AndriiOpryshko/spilno/services/core/internal/system"
)

// newServer serves the context through the same router, validator and
// generated handlers as production, on a real migrated database.
func newServer(t *testing.T) *httptest.Server {
	t.Helper()
	r := httpapi.NewRouter()
	if err := system.New(testdb.New(t)).Routes(r); err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)
	return srv
}

func TestHealth(t *testing.T) {
	srv := newServer(t)

	resp, err := http.Get(srv.URL + "/system/health")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	var body struct {
		Status   string `json:"status"`
		Database struct {
			Status  string `json:"status"`
			Version string `json:"version"`
		} `json:"database"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Status != "ok" || body.Database.Status != "up" || body.Database.Version == "" {
		t.Errorf("body = %+v, want ok/up with a version", body)
	}
}

func TestUnknownRouteIsProblem(t *testing.T) {
	srv := newServer(t)

	resp, err := http.Get(srv.URL + "/system/nope")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "application/problem+json" {
		t.Errorf("Content-Type = %q, want application/problem+json", ct)
	}
}
