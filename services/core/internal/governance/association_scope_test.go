package governance_test

import (
	"net/http"
	"strings"
	"testing"

	fingen "github.com/AndriiOpryshko/spilno/services/core/internal/finance/adapters/http/gen"
	govgen "github.com/AndriiOpryshko/spilno/services/core/internal/governance/adapters/http/gen"
)

const (
	assocA = "0191a000-0000-7000-8000-00000000000a"
	assocB = "0191a000-0000-7000-8000-00000000000b"
)

// A meeting created under one association is not listed under another (AC-04d).
func TestMeetingsAreIsolatedPerAssociation(t *testing.T) {
	srv := newServer(t)
	var created map[string]any
	do(t, "POST", srv.URL+"/governance/associations/"+assocA+"/meetings", validBody(), http.StatusCreated, &created)

	type page struct {
		Items []map[string]any `json:"items"`
	}
	var a, b page
	do(t, "GET", srv.URL+"/governance/associations/"+assocA+"/meetings", nil, http.StatusOK, &a)
	do(t, "GET", srv.URL+"/governance/associations/"+assocB+"/meetings", nil, http.StatusOK, &b)
	if len(a.Items) != 1 || a.Items[0]["id"] != created["id"] {
		t.Errorf("list under A = %v, want the created meeting", a.Items)
	}
	if len(b.Items) != 0 {
		t.Errorf("list under B = %v, want empty", b.Items)
	}
}

// Every governance and finance operation carries {association_id} in its path.
func TestAssociationScopedOperationsCarryAssociationID(t *testing.T) {
	for name, load := range map[string]struct {
		prefix string
		spec   func() (specPaths, error)
	}{
		"governance": {"/governance/associations/{association_id}/", func() (specPaths, error) {
			s, err := govgen.GetSpec()
			if err != nil {
				return nil, err
			}
			return pathsOf(s.Paths.Map()), nil
		}},
		"finance": {"/finance/associations/{association_id}/", func() (specPaths, error) {
			s, err := fingen.GetSpec()
			if err != nil {
				return nil, err
			}
			return pathsOf(s.Paths.Map()), nil
		}},
	} {
		t.Run(name, func(t *testing.T) {
			paths, err := load.spec()
			if err != nil {
				t.Fatal(err)
			}
			seen := 0
			for p := range paths {
				if !strings.HasPrefix(p, "/"+name+"/") {
					continue // other contexts share the embedded spec
				}
				seen++
				if !strings.HasPrefix(p, load.prefix) {
					t.Errorf("path %s does not start with %s", p, load.prefix)
				}
			}
			if seen == 0 {
				t.Fatalf("no %s paths in the spec", name)
			}
		})
	}
}

type specPaths map[string]struct{}

func pathsOf[V any](m map[string]V) specPaths {
	out := specPaths{}
	for p := range m {
		out[p] = struct{}{}
	}
	return out
}
