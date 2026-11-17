package governance_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/AndriiOpryshko/spilno/services/core/internal/governance"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/httpapi"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/testdb"
)

// AC-26 (ADR-0007): a download that reaches the handler without an actor in the context is refused with the
// same 403 as a person off the roster; a missing actor must not read as "not read-only".
func TestDownloadFileRefusesARequestWithoutAnActor(t *testing.T) {
	pool, a, _ := testdb.NewTwoAssociations(t)
	idm := identity.New(pool)
	gov := governance.New(pool, idm.HasStorageConnection, actorAccess).WithFiles(func(context.Context, uuid.UUID) (governance.Files, error) {
		return &driveFiles{data: map[string][]byte{}}, nil
	})
	r := httpapi.NewRouter()
	r.Group(func(r chi.Router) {
		// no idm.Guard: nothing puts an actor in the context
		if err := gov.Routes(r); err != nil {
			t.Fatal(err)
		}
	})
	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)

	res, err := http.Get(srv.URL + "/governance/associations/" + a + "/files/" + uuid.NewString())
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusForbidden || !strings.Contains(string(body), `"identity.not_current_member"`) {
		t.Fatalf("download without an actor: status = %d body %s, want 403 identity.not_current_member", res.StatusCode, body)
	}
}

// AC-26: the governance HTTP adapter imports nothing from identity.
func TestGovernanceAdaptersDoNotDependOnIdentity(t *testing.T) {
	out, err := exec.Command("go", "list", "-deps", "./adapters/...").CombinedOutput()
	if err != nil {
		t.Fatalf("go list: %v\n%s", err, out)
	}
	for _, p := range strings.Fields(string(out)) {
		if strings.Contains(p, "/internal/identity") {
			t.Errorf("governance adapters depend on %s, want no identity package", p)
		}
	}
}
