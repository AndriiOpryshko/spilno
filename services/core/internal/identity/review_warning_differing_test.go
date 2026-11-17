package identity_test

import (
	"encoding/json"
	"net/http"
	"reflect"
	"testing"
)

type warningWire struct {
	Kind      string   `json:"kind"`
	Differing []string `json:"differing"`
}

func (e *reviewEnv) warningsAfterCheck(t *testing.T, signer, head string, daysAgo int) []warningWire {
	t.Helper()
	id := e.fileAs(t, "first", fileAppFields(), fileAppPDF(512))
	filing := e.review(t, id).Filings[0].ID
	status, _, body := e.do(t, "operator", http.MethodPost, "/identity/operator/applications/"+id+"/checks", reviewCheckBody(filing, signer, head, reviewKyivDay(daysAgo)))
	if status != http.StatusCreated {
		t.Fatalf("recordExtractCheck = %d: %s", status, body)
	}
	_, _, raw := e.do(t, "operator", http.MethodGet, "/identity/operator/applications/"+id, "")
	var v struct {
		Warnings []json.RawMessage `json:"warnings"`
	}
	if err := json.Unmarshal(raw, &v); err != nil {
		t.Fatal(err)
	}
	var out []warningWire
	for _, w := range v.Warnings {
		var ww warningWire
		if err := json.Unmarshal(w, &ww); err != nil {
			t.Fatal(err)
		}
		var keys map[string]json.RawMessage
		_ = json.Unmarshal(w, &keys)
		if _, ok := keys["differing"]; !ok {
			t.Fatalf("warning %s has no `differing` member: %s", ww.Kind, w)
		}
		out = append(out, ww)
	}
	return out
}

// AC-11: the warning carries which of the three values differ, so the page can name only those.
// The typed name of fileAppFields is "Test User".
func TestReview_NameWarningSaysWhichValuesDiffer(t *testing.T) {
	cases := []struct {
		name         string
		signer, head string
		want         []string
	}{
		{"only the head differs", "Test User", "Other Person", []string{"head"}},
		{"only the signer differs", "Other Person", "Test User", []string{"signer"}},
		{"only the typed name differs", "Other Person", "Other Person", []string{"typed"}},
		{"all three differ", "Second Person", "Third Person", []string{"typed", "signer", "head"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := newReviewEnv(t)
			ws := e.warningsAfterCheck(t, tc.signer, tc.head, 1)
			if len(ws) != 1 || ws[0].Kind != "name_mismatch" || !reflect.DeepEqual(ws[0].Differing, tc.want) {
				t.Fatalf("warnings = %+v, want one name_mismatch differing %v", ws, tc.want)
			}
		})
	}
}

// AC-11b: an old extract differs in no name; the member is there and empty, not null.
func TestReview_ExtractAgeWarningHasNoDifferingNames(t *testing.T) {
	e := newReviewEnv(t)
	ws := e.warningsAfterCheck(t, "Test User", "Test User", 31)
	if len(ws) != 1 || ws[0].Kind != "extract_age" || ws[0].Differing == nil || len(ws[0].Differing) != 0 {
		t.Fatalf("warnings = %+v, want one extract_age with an empty, non-null differing", ws)
	}
}
