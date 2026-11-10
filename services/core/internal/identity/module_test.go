package identity_test

import (
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
