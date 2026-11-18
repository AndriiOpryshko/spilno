package identity_test

import (
	"sort"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/AndriiOpryshko/spilno/services/core/internal/identity"
)

// T124: the outside-I/O list is not opt-in. Every association-scoped operation whose request body is
// multipart/form-data in the live contract runs outside the guard's check transaction, so a new upload
// operation cannot buffer its body while a pooled connection is held.
func TestEveryAssociationMultipartOperationIsOutsideIO(t *testing.T) {
	doc, err := openapi3.NewLoader().LoadFromFile("../../api/openapi.yaml")
	if err != nil {
		t.Fatalf("load live contract: %v", err)
	}
	var missing []string
	seen := 0
	for path, item := range doc.Paths.Map() {
		if !strings.Contains(path, "{association_id}") {
			continue
		}
		for method, op := range item.Operations() {
			if op.RequestBody == nil || op.RequestBody.Value == nil || op.RequestBody.Value.Content.Get("multipart/form-data") == nil {
				continue
			}
			seen++
			if key := method + " " + path; !identity.OutsideIOOperations[key] {
				missing = append(missing, op.OperationID+" ("+key+")")
			}
		}
	}
	if seen == 0 {
		t.Fatal("found no association-scoped multipart operation in the contract; the walk is broken")
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Errorf("multipart operations missing from the outside-I/O list: %s", strings.Join(missing, ", "))
	}
}
