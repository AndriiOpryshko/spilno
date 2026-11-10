package identity_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"
)

// Every new onboarding operation is stubbed with 501 problem+json until its
// ports task lands; the session is not enforced yet, so no cookie is sent. A
// 404 here means the operation is missing from the merged contract.
func TestNewOperationsAreStubbed(t *testing.T) {
	srv := newServer(t)

	assoc := uuid.NewString()
	const jsonCT = "application/json"
	disableBody := `{"reason":"test reason"}`

	tests := []struct {
		name, method, path, ct string
		body                   []byte
	}{
		{"listAssociationsForOperator", "GET", "/identity/operator/associations", "", nil},
		{"disableAssociation", "POST", "/identity/operator/associations/" + assoc + "/disabling", jsonCT, []byte(disableBody)},
		{"enableAssociation", "POST", "/identity/operator/associations/" + assoc + "/enabling", "", nil},
	}
	if len(tests) != 3 {
		t.Fatalf("table has %d operations, want the 3 stubbed ones", len(tests))
	}

	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(tc.method, srv.URL+tc.path, bytes.NewReader(tc.body))
			if err != nil {
				t.Fatal(err)
			}
			if tc.ct != "" {
				req.Header.Set("Content-Type", tc.ct)
			}
			resp, err := client.Do(req)
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
			var p struct {
				Status int `json:"status"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
				t.Fatal(err)
			}
			if p.Status != http.StatusNotImplemented {
				t.Errorf("body status = %d, want 501", p.Status)
			}
		})
	}
}
