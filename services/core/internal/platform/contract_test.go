package platform_test

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

const (
	liveContract  = "../../api/openapi.yaml"
	deltaContract = "../../../../docs/features/association-onboarding/contracts/openapi.yaml"
	govGenFile    = "../governance/adapters/http/gen/governance.gen.go"
)

func loadContract(t *testing.T, path string) *openapi3.T {
	t.Helper()
	doc, err := openapi3.NewLoader().LoadFromFile(path)
	if err != nil {
		t.Fatalf("load %s: %v", path, err)
	}
	return doc
}

func operationByID(t *testing.T, doc *openapi3.T, id string) *openapi3.Operation {
	t.Helper()
	for _, item := range doc.Paths.Map() {
		for _, op := range item.Operations() {
			if op.OperationID == id {
				return op
			}
		}
	}
	t.Fatalf("operation %s not found", id)
	return nil
}

func response(t *testing.T, op *openapi3.Operation, status string) *openapi3.Response {
	t.Helper()
	ref := op.Responses.Value(status)
	if ref == nil || ref.Value == nil {
		t.Fatalf("%s advertises no %s response", op.OperationID, status)
	}
	return ref.Value
}

// T130 (R08, AC-02): every operation that takes a file advertises the 413 the body limit answers, as
// problem+json over the shared Problem schema.
func TestFileOperationsAdvertisePayloadTooLarge(t *testing.T) {
	doc := loadContract(t, liveContract)
	for _, id := range []string{"fileApplication", "refileApplication", "addCaseDocument"} {
		t.Run(id, func(t *testing.T) {
			r := response(t, operationByID(t, doc, id), "413")
			mt := r.Content.Get("application/problem+json")
			if mt == nil || mt.Schema == nil || mt.Schema.Ref != "#/components/schemas/Problem" {
				t.Errorf("%s 413 content = %+v, want application/problem+json over #/components/schemas/Problem", id, mt)
			}
		})
	}
}

// T130 (R08, AC-05): the sign-in cookie is on the contract. startSignIn sets it on the 302 and no longer
// claims server-side state; completeSignIn clears it, and its 401 names a missing or foreign cookie.
func TestSignInOperationsDescribeThePendingCookie(t *testing.T) {
	doc := loadContract(t, liveContract)
	start := operationByID(t, doc, "startSignIn")
	if _, ok := response(t, start, "302").Headers["Set-Cookie"]; !ok {
		t.Error("startSignIn 302 lists no Set-Cookie header")
	}
	if d := strings.ToLower(start.Description); strings.Contains(d, "server-side") || !strings.Contains(d, "cookie") {
		t.Errorf("startSignIn description = %q, want the sealed cookie and no server-side state", start.Description)
	}
	complete := operationByID(t, doc, "completeSignIn")
	if _, ok := response(t, complete, "302").Headers["Set-Cookie"]; !ok {
		t.Error("completeSignIn 302 lists no Set-Cookie header")
	}
	if d := strings.ToLower(complete.Description); !strings.Contains(d, "cookie") {
		t.Errorf("completeSignIn description = %q, want the pending cookie it clears", complete.Description)
	}
	if d := strings.ToLower(*response(t, complete, "401").Description); !strings.Contains(d, "cookie") {
		t.Errorf("completeSignIn 401 = %q, want it to name a missing or foreign cookie", d)
	}
}

// T130 (R08, AC-19): the 422 of updateBoardMember names the refusal on /term_end of an ended entry.
func TestUpdateBoardMemberNamesTheTermEndRefusal(t *testing.T) {
	doc := loadContract(t, liveContract)
	d := *response(t, operationByID(t, doc, "updateBoardMember"), "422").Description
	if !strings.Contains(d, "/term_end") {
		t.Errorf("updateBoardMember 422 = %q, want it to name the refusal on /term_end", d)
	}
}

// T130 (R10, AC-26): decision_effect is a nullable string enum by $ref, not an enum holding null.
func TestDecisionEffectIsAStringEnumByRef(t *testing.T) {
	doc := loadContract(t, liveContract)
	effect := doc.Components.Schemas["DecisionEffect"]
	if effect == nil || effect.Value == nil {
		t.Fatal("components.schemas has no DecisionEffect")
	}
	if !effect.Value.Type.Is("string") || len(effect.Value.Enum) != 2 || effect.Value.Enum[0] != "seated" || effect.Value.Enum[1] != "ended" {
		t.Errorf("DecisionEffect = type %v enum %v, want a string enum [seated ended]", effect.Value.Type, effect.Value.Enum)
	}
	p := doc.Components.Schemas["BoardMember"].Value.Properties["decision_effect"]
	if p == nil || p.Value == nil || len(p.Value.OneOf) != 2 {
		t.Fatalf("BoardMember.decision_effect = %+v, want a oneOf of the DecisionEffect $ref and null", p)
	}
	if p.Value.OneOf[0].Ref != "#/components/schemas/DecisionEffect" || !p.Value.OneOf[1].Value.Type.Is("null") {
		t.Errorf("BoardMember.decision_effect oneOf = [%s, %v], want [#/components/schemas/DecisionEffect, null]",
			p.Value.OneOf[0].Ref, p.Value.OneOf[1].Value.Type)
	}
}

// T130 (R10): the generated governance code carries no constant for a null enum member, and so
// Valid() accepts no "<nil>".
func TestGeneratedGovernanceCodeHasNoNilConstant(t *testing.T) {
	src, err := os.ReadFile(govGenFile)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(src), "<nil>") {
		t.Error(`governance.gen.go contains "<nil>"; a null enum member leaked into a generated constant`)
	}
}

func asJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// T130 (R11): the delta carries BoardMember.decision_effect and the 403 of downloadFile, and equals
// the live contract for every operation and schema it carries.
func TestDeltaEqualsTheLiveContract(t *testing.T) {
	live, delta := loadContract(t, liveContract), loadContract(t, deltaContract)

	if bm := delta.Components.Schemas["BoardMember"]; bm == nil || bm.Value.Properties["decision_effect"] == nil {
		t.Error("delta has no BoardMember.decision_effect")
	}
	if delta.Components.Schemas["DecisionEffect"] == nil {
		t.Error("delta has no DecisionEffect schema")
	}
	found := false
	for _, item := range delta.Paths.Map() {
		if op := item.Get; op != nil && op.OperationID == "downloadFile" {
			found = true
			if op.Responses.Value("403") == nil {
				t.Error("delta downloadFile has no 403 response")
			}
		}
	}
	if !found {
		t.Error("delta has no downloadFile operation")
	}

	for path, item := range delta.Paths.Map() {
		liveItem := live.Paths.Find(path)
		if liveItem == nil {
			t.Errorf("delta path %s is not in the live contract", path)
			continue
		}
		for method, op := range item.Operations() {
			lop := liveItem.GetOperation(method)
			if lop == nil {
				t.Errorf("delta %s %s is not in the live contract", method, path)
				continue
			}
			if a, b := asJSON(t, op), asJSON(t, lop); a != b {
				t.Errorf("%s %s (%s) differs between the delta and the live contract", method, path, op.OperationID)
			}
		}
	}
	for name, schema := range delta.Components.Schemas {
		ls := live.Components.Schemas[name]
		if ls == nil {
			t.Errorf("delta schema %s is not in the live contract", name)
			continue
		}
		if a, b := asJSON(t, schema), asJSON(t, ls); a != b {
			t.Errorf("schema %s differs between the delta and the live contract", name)
		}
	}
}
