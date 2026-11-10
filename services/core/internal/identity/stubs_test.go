package identity_test

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// Every new onboarding operation is stubbed with 501 problem+json until its
// ports task lands; the session is not enforced yet, so no cookie is sent. A
// 404 here means the operation is missing from the merged contract.
func TestNewOperationsAreStubbed(t *testing.T) {
	srv := newServer(t)

	app, assoc, filing := uuid.NewString(), uuid.NewString(), uuid.NewString()
	const (
		jsonCT = "application/json"
		pdf    = "%PDF-1.4 test"
	)
	approveBody := `{"filing_id":"` + filing + `","relied_on_document":"test document"}`
	rejectBody := `{"filing_id":"` + filing + `","reason":"test reason"}`
	disableBody := `{"reason":"test reason"}`
	storageBody := `{"code":"test-code","state":"test-state"}`

	refile, refileCT := multipartBody(t, map[string]string{
		"association_name": "Test Association", "association_address": "Test Street 1",
		"applicant_full_name": "Test User", "position": "head", "term_start": "2026-03-01",
	}, pdf)

	a := "/identity/applications/" + app
	op := "/identity/operator/applications/" + app
	as := "/identity/associations/" + assoc
	tests := []struct {
		name, method, path, ct string
		body                   []byte
	}{
		{"listMemberships", "GET", "/identity/memberships", "", nil},
		{"listMyApplications", "GET", "/identity/applications", "", nil},
		{"getMyApplication", "GET", a, "", nil},
		{"refileApplication", "POST", a + "/filings", refileCT, refile},
		{"downloadMyExtract", "GET", a + "/filings/" + filing + "/extract", "", nil},
		{"approveApplication", "POST", op + "/approval", jsonCT, []byte(approveBody)},
		{"rejectApplication", "POST", op + "/rejection", jsonCT, []byte(rejectBody)},
		{"listAssociationsForOperator", "GET", "/identity/operator/associations", "", nil},
		{"disableAssociation", "POST", "/identity/operator/associations/" + assoc + "/disabling", jsonCT, []byte(disableBody)},
		{"enableAssociation", "POST", "/identity/operator/associations/" + assoc + "/enabling", "", nil},
		{"getAssociationCard", "GET", as, "", nil},
		{"downloadConfirmationExtract", "GET", as + "/extracts/" + uuid.NewString(), "", nil},
		{"getStorageConnection", "GET", as + "/storage-connection", "", nil},
		{"completeStorageConnection", "POST", as + "/storage-connection", jsonCT, []byte(storageBody)},
		{"startStorageConnection", "GET", as + "/storage-connection/authorization", "", nil},
	}
	if len(tests) != 15 {
		t.Fatalf("table has %d operations, want the 15 stubbed ones", len(tests))
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

func multipartBody(t *testing.T, fields map[string]string, extract string) ([]byte, string) {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	for k, v := range fields {
		if err := w.WriteField(k, v); err != nil {
			t.Fatal(err)
		}
	}
	// CreateFormFile would declare application/octet-stream, which the contract does not list for the extract.
	h := textproto.MIMEHeader{}
	h.Set("Content-Disposition", `form-data; name="extract"; filename="extract.pdf"`)
	h.Set("Content-Type", "application/pdf")
	fw, err := w.CreatePart(h)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fw.Write([]byte(extract)); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes(), strings.TrimSpace(w.FormDataContentType())
}
