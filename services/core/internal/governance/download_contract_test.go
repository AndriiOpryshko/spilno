package governance_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

const liveContractPath = "../../api/openapi.yaml"

// downloadForbidden returns the 403 response of downloadFile in the live contract.
func downloadForbidden(t *testing.T) *openapi3.Response {
	t.Helper()
	doc, err := openapi3.NewLoader().LoadFromFile(liveContractPath)
	if err != nil {
		t.Fatalf("load live contract: %v", err)
	}
	for _, item := range doc.Paths.Map() {
		if op := item.Get; op != nil && op.OperationID == "downloadFile" {
			ref := op.Responses.Value("403")
			if ref == nil || ref.Value == nil {
				t.Fatal("downloadFile has no 403 response")
			}
			return ref.Value
		}
	}
	t.Fatal("downloadFile not found in the live contract")
	return nil
}

// AC-25: the 403 of downloadFile lists both codes, each with an example.
func TestDownloadFileContractAdvertisesBothForbiddenCodes(t *testing.T) {
	content := downloadForbidden(t).Content.Get("application/problem+json")
	if content == nil {
		t.Fatal("downloadFile 403 has no application/problem+json content")
	}
	codes := map[string]bool{}
	for name, ex := range content.Examples {
		if m, ok := ex.Value.Value.(map[string]any); ok {
			if c, ok := m["code"].(string); ok {
				codes[c] = true
			}
		}
		_ = name
	}
	for _, want := range []string{"identity.not_current_member", "identity.association_disabled"} {
		if !codes[want] {
			t.Errorf("downloadFile 403 examples carry codes %v, want an example of %s", codes, want)
		}
	}
}

// requireMatchesContractExample compares the status, title and code of the response with the named
// example of the downloadFile 403.
func requireMatchesContractExample(t *testing.T, got download, exampleName string) {
	t.Helper()
	content := downloadForbidden(t).Content.Get("application/problem+json")
	if content == nil {
		t.Fatal("downloadFile 403 has no application/problem+json content")
	}
	var ex *openapi3.ExampleRef
	for _, e := range content.Examples {
		if m, ok := e.Value.Value.(map[string]any); ok && m["code"] == map[string]string{"AssociationDisabled": "identity.association_disabled"}[exampleName] {
			ex = e
		}
	}
	if ex == nil {
		t.Fatalf("downloadFile 403 has no %s example", exampleName)
	}
	want, _ := ex.Value.Value.(map[string]any)
	var have map[string]any
	_ = json.Unmarshal(got.body, &have)
	for _, k := range []string{"type", "title", "status", "code"} {
		if wj, _ := json.Marshal(want[k]); string(wj) != mustJSON(have[k]) {
			t.Errorf("response %s = %s, contract example has %s", k, mustJSON(have[k]), wj)
		}
	}
}

func mustJSON(v any) string { b, _ := json.Marshal(v); return string(b) }

// Q08: no governance source outside domain/errors.go spells a problem code as a string literal.
func TestGovernanceHasNoProblemCodeLiteralsOutsideTheRegistry(t *testing.T) {
	re := regexp.MustCompile(`"(governance|identity)\.[a-z][a-z0-9_]*"`)
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		slash := filepath.ToSlash(path)
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") ||
			strings.Contains(slash, "/gen/") || slash == "domain/errors.go" {
			return nil
		}
		raw, rerr := os.ReadFile(path)
		if rerr != nil {
			return rerr
		}
		for _, m := range re.FindAllString(string(raw), -1) {
			t.Errorf("%s spells the problem code %s as a literal; use the constant", slash, m)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// T129 (review of 2026-11-17-2, R09): the guard's refusal codes belong to identity and reach governance
// through cmd/api; no governance source, the registry in domain/errors.go included, spells an identity.* code.
func TestGovernanceSourcesSpellNoIdentityCode(t *testing.T) {
	re := regexp.MustCompile(`"identity\.[a-z][a-z0-9_]*"`)
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		slash := filepath.ToSlash(path)
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") || strings.Contains(slash, "/gen/") {
			return nil
		}
		raw, rerr := os.ReadFile(path)
		if rerr != nil {
			return rerr
		}
		for _, m := range re.FindAllString(string(raw), -1) {
			t.Errorf("%s spells the identity code %s; take it from the access lookup cmd/api injects", slash, m)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
