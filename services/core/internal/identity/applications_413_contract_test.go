package identity_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

// T130 (R08, AC-02): the 413 body of an oversized filing is the one the live contract advertises for
// fileApplication: the status is a declared response, the body matches its problem+json schema, and
// the title and status equal those of its example.
func TestOversizeFilingAnswersTheAdvertised413(t *testing.T) {
	doc, err := openapi3.NewLoader().LoadFromFile("../../api/openapi.yaml")
	if err != nil {
		t.Fatalf("load live contract: %v", err)
	}
	var advertised *openapi3.Response
	for _, item := range doc.Paths.Map() {
		if op := item.Post; op != nil && op.OperationID == "fileApplication" {
			if ref := op.Responses.Value("413"); ref != nil {
				advertised = ref.Value
			}
		}
	}
	if advertised == nil {
		t.Fatal("fileApplication advertises no 413 response")
	}
	mt := advertised.Content.Get("application/problem+json")
	if mt == nil || mt.Schema == nil {
		t.Fatal("fileApplication 413 has no application/problem+json schema")
	}

	e := newFileAppEnv(t)
	got := e.file(t, "first", fileAppFields(), fileAppPDF(20*1024*1024))
	if got.status != http.StatusRequestEntityTooLarge {
		t.Fatalf("20 MB filing: status = %d (%.200s), want 413", got.status, got.body)
	}
	var body any
	if err := json.Unmarshal(got.body, &body); err != nil {
		t.Fatalf("body %.200s is not JSON: %v", got.body, err)
	}
	if err := mt.Schema.Value.VisitJSON(body); err != nil {
		t.Errorf("413 body %.200s does not match the advertised schema: %v", got.body, err)
	}
	if len(mt.Examples) == 0 {
		t.Error("fileApplication 413 carries no example of the body")
	}
	for name, ex := range mt.Examples {
		want, _ := ex.Value.Value.(map[string]any)
		have, _ := body.(map[string]any)
		for _, k := range []string{"title", "status"} {
			if w, g := asString(want[k]), asString(have[k]); w != g {
				t.Errorf("413 example %s: %s = %s, the filing answered %s", name, k, w, g)
			}
		}
	}
}

func asString(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
