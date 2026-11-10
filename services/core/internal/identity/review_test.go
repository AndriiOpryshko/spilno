package identity_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/AndriiOpryshko/spilno/services/core/internal/governance"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/governancetest"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/db"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/httpapi"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/testdb"
)

type reviewEnv struct {
	srv     *httptest.Server
	pool    *pgxpool.Pool
	cookies map[string]*http.Cookie
}

// newReviewEnv serves identity behind the guard on the app-role pool with four signed-in people:
// operator, first and second (applicants, no roster) and head (a член правління of association A).
func newReviewEnv(t *testing.T) *reviewEnv {
	t.Helper()
	pool, a, _ := testdb.NewTwoAssociations(t)
	idm := identity.New(pool)
	people := map[string]identity.ProviderIdentity{}
	for p, email := range map[string]string{"first": "first@example.test", "second": "second@example.test", "head": "head@example.test", "operator": guardOperatorEmail} {
		people[p] = identity.ProviderIdentity{Issuer: "https://accounts.example.test", Subject: "sub-" + p, Email: email, EmailVerified: true}
	}
	idm.EnableSignIn(identity.SignInConfig{Provider: &signInFake{people: people}, OperatorEmail: guardOperatorEmail})
	r := httpapi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(idm.Guard(guardRoster{governance.New(pool)}))
		if err := idm.Routes(r); err != nil {
			t.Fatal(err)
		}
	})
	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)
	aid := uuid.MustParse(a)
	err := db.InAssociation(context.Background(), pool, aid, "seed@example.test", func(ctx context.Context) error {
		tx, err := db.From(ctx)
		if err != nil {
			return err
		}
		governancetest.Member(t, tx, aid, "head@example.test", "head")
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	env := &reviewEnv{srv: srv, pool: pool, cookies: map[string]*http.Cookie{}}
	for p := range people {
		c := sessionCookieOf(signInAs(t, srv, p))
		if c == nil {
			t.Fatalf("no session cookie for %s", p)
		}
		env.cookies[p] = c
	}
	return env
}

func (e *reviewEnv) do(t *testing.T, as, method, path, body string) (int, http.Header, []byte) {
	t.Helper()
	req, err := http.NewRequest(method, e.srv.URL+path, bytes.NewBufferString(body))
	if err != nil {
		t.Fatal(err)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if c := e.cookies[as]; c != nil {
		req.AddCookie(c)
	}
	resp, err := signInNoRedirect.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, resp.Header, b
}

// fileAs files a заявка on the wire and returns its id.
func (e *reviewEnv) fileAs(t *testing.T, as string, fields map[string]string, extract []byte) string {
	t.Helper()
	body, ct := fileAppMultipart(t, fields, extract)
	req, _ := http.NewRequest("POST", e.srv.URL+"/identity/applications", bytes.NewReader(body))
	req.Header.Set("Content-Type", ct)
	req.AddCookie(e.cookies[as])
	resp, err := signInNoRedirect.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("filing as %s = %d (%s), want 201", as, resp.StatusCode, b)
	}
	var app struct{ ID string }
	if err := json.Unmarshal(b, &app); err != nil {
		t.Fatal(err)
	}
	return app.ID
}

type reviewView struct {
	Application struct {
		ID    string `json:"id"`
		State string `json:"state"`
	} `json:"application"`
	ApplicantEmail string `json:"applicant_email"`
	Filings        []struct {
		ID                string `json:"id"`
		ApplicantFullName string `json:"applicant_full_name"`
		Position          string `json:"position"`
		TermStart         string `json:"term_start"`
		Extract           struct {
			OriginalName string `json:"original_name"`
			SizeBytes    int    `json:"size_bytes"`
		} `json:"extract"`
	} `json:"filings"`
	Checks []struct {
		ID             string `json:"id"`
		FilingID       string `json:"filing_id"`
		SignatureValid bool   `json:"signature_valid"`
		SignerFullName string `json:"signer_full_name"`
		SignedOn       string `json:"signed_on"`
		HeadFullName   string `json:"head_full_name"`
		ExtractEdrpou  string `json:"extract_edrpou"`
		ExtractDatedOn string `json:"extract_dated_on"`
		CheckedAt      string `json:"checked_at"`
	} `json:"checks"`
	Warnings []struct {
		Kind   string `json:"kind"`
		Detail string `json:"detail"`
	} `json:"warnings"`
	Competing []string `json:"competing_application_ids"`
}

func (e *reviewEnv) review(t *testing.T, id string) reviewView {
	t.Helper()
	status, _, body := e.do(t, "operator", http.MethodGet, "/identity/operator/applications/"+id, "")
	if status != http.StatusOK {
		t.Fatalf("getApplicationForReview = %d, want 200: %s", status, body)
	}
	var v reviewView
	if err := json.Unmarshal(body, &v); err != nil {
		t.Fatal(err)
	}
	return v
}

func reviewCheckBody(filing, signer, head, dated string) string {
	b, _ := json.Marshal(map[string]any{
		"filing_id": filing, "signature_valid": true, "signer_full_name": signer, "signed_on": dated,
		"head_full_name": head, "extract_edrpou": fileAppEdrpou, "extract_dated_on": dated,
	})
	return string(b)
}

func reviewKyivDay(daysAgo int) string {
	loc, err := time.LoadLocation("Europe/Kyiv")
	if err != nil {
		loc = time.UTC
	}
	return time.Now().In(loc).AddDate(0, 0, -daysAgo).Format("2006-01-02")
}

func reviewWarningKinds(v reviewView) []string {
	var k []string
	for _, w := range v.Warnings {
		k = append(k, w.Kind)
	}
	return k
}

// AC-10, AC-11, AC-11b: the review shows typed values with no checks until one is recorded, the
// check is a new row with its five values and time, a corrected check is a second row, and the
// warnings name the differing values, on recording and on reopening, without blocking.
func TestReview_RecordsChecksAsRowsAndRaisesAdvisoryWarnings(t *testing.T) {
	e := newReviewEnv(t)
	upload := fileAppPDF(1024)
	id := e.fileAs(t, "first", fileAppFields(), upload)

	v := e.review(t, id)
	if v.Application.State != "pending" || v.ApplicantEmail != "first@example.test" {
		t.Fatalf("review application = %+v / %q", v.Application, v.ApplicantEmail)
	}
	if len(v.Filings) != 1 || v.Filings[0].ApplicantFullName != "Test User" || v.Filings[0].Position != "head" || v.Filings[0].TermStart != "2026-03-01" || v.Filings[0].Extract.SizeBytes != len(upload) {
		t.Fatalf("review filings = %+v", v.Filings)
	}
	if len(v.Checks) != 0 || len(v.Warnings) != 0 || len(v.Competing) != 0 {
		t.Fatalf("before a check: checks=%d warnings=%v competing=%v, want none", len(v.Checks), v.Warnings, v.Competing)
	}
	filing := v.Filings[0].ID

	// The check carries the five values, and the warnings stay off for the same person and a fresh extract.
	status, _, body := e.do(t, "operator", http.MethodPost, "/identity/operator/applications/"+id+"/checks", reviewCheckBody(filing, "Test User", "Test User", reviewKyivDay(1)))
	if status != http.StatusCreated {
		t.Fatalf("recordExtractCheck = %d, want 201: %s", status, body)
	}
	v = e.review(t, id)
	if len(v.Checks) != 1 {
		t.Fatalf("checks after one recording = %d, want 1", len(v.Checks))
	}
	c := v.Checks[0]
	if c.FilingID != filing || !c.SignatureValid || c.SignerFullName != "Test User" || c.HeadFullName != "Test User" || c.ExtractEdrpou != fileAppEdrpou || c.ExtractDatedOn != reviewKyivDay(1) || c.SignedOn != reviewKyivDay(1) || c.CheckedAt == "" {
		t.Fatalf("recorded check = %+v", c)
	}
	if len(v.Warnings) != 0 {
		t.Fatalf("warnings for a matching, fresh check = %v, want none", v.Warnings)
	}
	firstID := c.ID

	// AC-11b: 30 days -> no warning; 31 days -> warning (recorded on the response and on reopening).
	status, _, body = e.do(t, "operator", http.MethodPost, "/identity/operator/applications/"+id+"/checks", reviewCheckBody(filing, "Test User", "Test User", reviewKyivDay(30)))
	if status != http.StatusCreated {
		t.Fatalf("check at 30 days = %d: %s", status, body)
	}
	if k := reviewWarningKinds(e.review(t, id)); len(k) != 0 {
		t.Fatalf("warnings for a 30-day extract = %v, want none", k)
	}
	status, _, body = e.do(t, "operator", http.MethodPost, "/identity/operator/applications/"+id+"/checks", reviewCheckBody(filing, "Test User", "Test User", reviewKyivDay(31)))
	if status != http.StatusCreated {
		t.Fatalf("check at 31 days = %d: %s", status, body)
	}
	v = e.review(t, id)
	if k := reviewWarningKinds(v); len(k) != 1 || k[0] != "extract_age" || !strings.Contains(v.Warnings[0].Detail, reviewKyivDay(31)) {
		t.Fatalf("warnings for a 31-day extract = %+v, want one extract_age naming the date", v.Warnings)
	}

	// AC-11: a different head names the differing value; the latest check counts, the earlier rows stay.
	status, _, body = e.do(t, "operator", http.MethodPost, "/identity/operator/applications/"+id+"/checks", reviewCheckBody(filing, "Test User", "Other Person", reviewKyivDay(1)))
	if status != http.StatusCreated {
		t.Fatalf("check with another head = %d: %s", status, body)
	}
	v = e.review(t, id)
	if k := reviewWarningKinds(v); len(k) != 1 || k[0] != "name_mismatch" || !strings.Contains(v.Warnings[0].Detail, "head") {
		t.Fatalf("warnings = %+v, want one name_mismatch naming the head", v.Warnings)
	}
	if len(v.Checks) != 4 || v.Checks[0].ID != firstID || v.Checks[0].HeadFullName != "Test User" {
		t.Fatalf("checks = %+v, want 4 rows with the first unchanged", v.Checks)
	}
	if n := count(t, e.pool, "application_checks"); n != 4 {
		t.Fatalf("extract_checks rows = %d, want 4", n)
	}
	status, _, body = e.do(t, "operator", http.MethodPost, "/identity/operator/applications/"+id+"/checks", reviewCheckBody(filing, "Test User", "Test User", reviewKyivDay(1)))
	if status != http.StatusCreated {
		t.Fatalf("corrected check = %d: %s", status, body)
	}
	if k := reviewWarningKinds(e.review(t, id)); len(k) != 0 {
		t.Fatalf("warnings after the corrected check = %v, want none", k)
	}
}

// AC-04c: two pending заявки for one code each list the other, the queue lists both oldest
// first, and the extract is downloadable as an attachment byte for byte.
func TestReview_QueueCompetingApplicationsAndExtractDownload(t *testing.T) {
	e := newReviewEnv(t)
	upload := fileAppPDF(1500)
	one := e.fileAs(t, "first", fileAppFields(), upload)
	two := e.fileAs(t, "second", fileAppFields(), fileAppPDF(900))

	if got := e.review(t, one).Competing; len(got) != 1 || got[0] != two {
		t.Fatalf("competing of first = %v, want [%s]", got, two)
	}
	if got := e.review(t, two).Competing; len(got) != 1 || got[0] != one {
		t.Fatalf("competing of second = %v, want [%s]", got, one)
	}

	status, _, body := e.do(t, "operator", http.MethodGet, "/identity/operator/applications?limit=100", "")
	if status != http.StatusOK {
		t.Fatalf("listApplicationsForReview = %d: %s", status, body)
	}
	var page struct {
		Items []struct {
			ID    string `json:"id"`
			State string `json:"state"`
		} `json:"items"`
	}
	if err := json.Unmarshal(body, &page); err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 2 || page.Items[0].ID != one || page.Items[1].ID != two || page.Items[0].State != "pending" {
		t.Fatalf("queue = %s, want [%s %s] oldest first", body, one, two)
	}

	filing := e.review(t, one).Filings[0].ID
	status, hdr, got := e.do(t, "operator", http.MethodGet, "/identity/operator/applications/"+one+"/filings/"+filing+"/extract", "")
	if status != http.StatusOK || !bytes.Equal(got, upload) {
		t.Fatalf("downloadExtractForReview = %d, %d bytes, want 200 and the %d uploaded bytes", status, len(got), len(upload))
	}
	if !strings.HasPrefix(hdr.Get("Content-Disposition"), "attachment") {
		t.Fatalf("Content-Disposition = %q, want attachment", hdr.Get("Content-Disposition"))
	}
}

// AC-16: a член правління, an applicant and a person with nothing get 403 identity.not_operator on
// all four operations, the same body whether the application exists or not.
func TestReview_NonOperatorsAreRefusedOnAllFourOperations(t *testing.T) {
	e := newReviewEnv(t)
	id := e.fileAs(t, "first", fileAppFields(), fileAppPDF(512))
	filing := e.review(t, id).Filings[0].ID
	missing := uuid.NewString()
	check := reviewCheckBody(filing, "Test User", "Test User", reviewKyivDay(1))
	before := count(t, e.pool, "application_checks")

	for _, who := range []string{"head", "first", "second", "anonymous"} {
		for _, target := range []string{id, missing} {
			base := "/identity/operator/applications/" + target
			ops := []struct{ method, path, body string }{
				{"GET", "/identity/operator/applications", ""},
				{"GET", base, ""},
				{"GET", base + "/filings/" + filing + "/extract", ""},
				{"POST", base + "/checks", check},
			}
			for _, op := range ops {
				status, _, body := e.do(t, who, op.method, op.path, op.body)
				if who == "anonymous" {
					if status != http.StatusUnauthorized {
						t.Errorf("%s %s as anonymous = %d, want 401", op.method, op.path, status)
					}
					continue
				}
				if status != http.StatusForbidden || assocStateProblemCode(body) != "identity.not_operator" {
					t.Errorf("%s %s as %s = %d %s, want 403 identity.not_operator", op.method, op.path, who, status, body)
				}
				if strings.Contains(string(body), id) || strings.Contains(string(body), "Test Association") {
					t.Errorf("%s %s as %s leaks the queue: %s", op.method, op.path, who, body)
				}
			}
		}
	}
	if after := count(t, e.pool, "application_checks"); after != before {
		t.Fatalf("extract_checks rows = %d after refused attempts, want %d", after, before)
	}
}
