package identity_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
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

const fileAppEdrpou = "12345678"

type fileAppEnv struct {
	srv     *httptest.Server
	pool    *pgxpool.Pool
	assocA  string
	cookies map[string]*http.Cookie
}

// newFileAppEnv serves identity on the app-role pool with three signed-in people: first, second
// (no roster) and head (a член правління of association A).
func newFileAppEnv(t *testing.T) *fileAppEnv {
	t.Helper()
	pool, a, _ := testdb.NewTwoAssociations(t)
	idm := identity.New(pool)
	people := map[string]identity.ProviderIdentity{}
	for _, p := range []string{"first", "second", "head"} {
		people[p] = identity.ProviderIdentity{Issuer: "https://accounts.example.test", Subject: "sub-" + p, Email: p + "@example.test", EmailVerified: true}
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
	env := &fileAppEnv{srv: srv, pool: pool, assocA: a, cookies: map[string]*http.Cookie{}}
	for p := range people {
		c := sessionCookieOf(signInAs(t, srv, p))
		if c == nil {
			t.Fatalf("no session cookie for %s", p)
		}
		env.cookies[p] = c
	}
	return env
}

func fileAppFields() map[string]string {
	return map[string]string{
		"kind": "connection", "edrpou": fileAppEdrpou, "association_name": "Test Association",
		"association_address": "Test Street 1", "applicant_full_name": "Test User",
		"position": "head", "term_start": "2026-03-01",
	}
}

func fileAppMultipart(t *testing.T, fields map[string]string, extract []byte) ([]byte, string) {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	for k, v := range fields {
		if err := w.WriteField(k, v); err != nil {
			t.Fatal(err)
		}
	}
	if extract != nil {
		h := textproto.MIMEHeader{}
		h.Set("Content-Disposition", `form-data; name="extract"; filename="extract.pdf"`)
		h.Set("Content-Type", "application/pdf")
		fw, err := w.CreatePart(h)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = fw.Write(extract)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes(), w.FormDataContentType()
}

func (e *fileAppEnv) file(t *testing.T, as string, fields map[string]string, extract []byte) guardReply {
	t.Helper()
	body, ct := fileAppMultipart(t, fields, extract)
	req, err := http.NewRequest("POST", e.srv.URL+"/identity/applications", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", ct)
	if c := e.cookies[as]; c != nil {
		req.AddCookie(c)
	}
	resp, err := signInNoRedirect.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return guardReply{resp.StatusCode, b}
}

func (e *fileAppEnv) counts(t *testing.T) [3]int {
	t.Helper()
	return [3]int{count(t, e.pool, "applications"), count(t, e.pool, "application_filings"), count(t, e.pool, "application_extracts")}
}

func fileAppPDF(n int) []byte {
	b := bytes.Repeat([]byte{'x'}, n)
	copy(b, "%PDF-1.4\n")
	b[n-1] = 0xff
	return b
}

// AC-01: a complete filing is 201 pending; the stored extract equals the upload byte for byte,
// also for a person who is a член правління of another association.
func TestFileApplicationRecordsPendingAndKeepsTheExtractUnchanged(t *testing.T) {
	for _, who := range []string{"first", "head"} {
		t.Run(who, func(t *testing.T) {
			e := newFileAppEnv(t)
			upload := fileAppPDF(2048)
			got := e.file(t, who, fileAppFields(), upload)
			if got.status != http.StatusCreated {
				t.Fatalf("status = %d (%s), want 201", got.status, got.body)
			}
			var app struct {
				ID, Kind, Edrpou, State string
				FiledAt                 time.Time `json:"filed_at"`
			}
			if err := json.Unmarshal(got.body, &app); err != nil {
				t.Fatal(err)
			}
			if app.State != "pending" || app.Kind != "connection" || app.Edrpou != fileAppEdrpou || app.FiledAt.IsZero() {
				t.Errorf("body = %s, want a pending connection for %s with filed_at", got.body, fileAppEdrpou)
			}
			if c := e.counts(t); c != [3]int{1, 1, 1} {
				t.Fatalf("rows (applications, filings, extracts) = %v, want [1 1 1]", c)
			}
			var content, sum []byte
			var size int64
			err := e.pool.QueryRow(context.Background(), `SELECT content, sha256, size_bytes FROM application_extracts`).Scan(&content, &sum, &size)
			if err != nil {
				t.Fatal(err)
			}
			want := sha256.Sum256(upload)
			if !bytes.Equal(content, upload) || size != int64(len(upload)) || hex.EncodeToString(sum) != hex.EncodeToString(want[:]) {
				t.Errorf("stored extract differs from the upload: len %d vs %d, sha256 %x", len(content), len(upload), sum)
			}
		})
	}
}

// AC-02: fields the contract schema rejects are all named at once (by the shared request validator)
// and nothing is stored.
func TestFileApplicationNamesEverySchemaBrokenFieldAtOnce(t *testing.T) {
	e := newFileAppEnv(t)
	f := fileAppFields()
	f["edrpou"] = "1234"
	f["association_name"] = ""
	delete(f, "applicant_full_name")
	got := e.file(t, "first", f, nil)
	if got.status != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d (%s), want 422", got.status, got.body)
	}
	for _, p := range []string{"/edrpou", "/association_name", "/applicant_full_name", "/extract"} {
		if !strings.Contains(string(got.body), `"`+p+`"`) {
			t.Errorf("errors[] does not name %s: %s", p, got.body)
		}
	}
	if c := e.counts(t); c != [3]int{} {
		t.Errorf("rows = %v, want none stored", c)
	}
}

// AC-02: fields only the domain can judge (a position other than head, a term start after the
// filing date, an extract that is not a PDF or container) are all named at once under
// identity.application_invalid, and nothing is stored.
func TestFileApplicationNamesEveryDomainBrokenFieldAtOnce(t *testing.T) {
	e := newFileAppEnv(t)
	f := fileAppFields()
	f["position"] = "member"
	f["term_start"] = "2999-01-01"
	got := e.file(t, "first", f, []byte("plain text, not a pdf"))
	if got.status != http.StatusUnprocessableEntity || got.code() != "identity.application_invalid" {
		t.Fatalf("status = %d code = %q (%s), want 422 identity.application_invalid", got.status, got.code(), got.body)
	}
	for _, p := range []string{"/position", "/term_start", "/extract"} {
		if !strings.Contains(string(got.body), `"`+p+`"`) {
			t.Errorf("errors[] does not name %s: %s", p, got.body)
		}
	}
	if c := e.counts(t); c != [3]int{} {
		t.Errorf("rows = %v, want none stored", c)
	}
}

// AC-02: a domain-invalid extract (empty of any PDF header) is refused as a field, nothing stored.
func TestFileApplicationRefusesAnExtractThatIsNotAPDF(t *testing.T) {
	e := newFileAppEnv(t)
	got := e.file(t, "first", fileAppFields(), []byte("plain text, not a pdf"))
	if got.status != http.StatusUnprocessableEntity || got.code() != "identity.application_invalid" {
		t.Fatalf("status = %d code = %q (%s), want 422 identity.application_invalid", got.status, got.code(), got.body)
	}
	if c := e.counts(t); c != [3]int{} {
		t.Errorf("rows = %v, want none stored", c)
	}
}

// AC-04: a connected code is refused without any person; AC-30: an ended term is refused.
func TestFileApplicationRefusesAConnectedCodeAndAnEndedTerm(t *testing.T) {
	e := newFileAppEnv(t)
	aid := uuid.MustParse(e.assocA)
	err := db.InAssociation(context.Background(), e.pool, aid, "seed@example.test", func(ctx context.Context) error {
		tx, err := db.From(ctx)
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `UPDATE associations SET edrpou = $2 WHERE id = $1`, aid, fileAppEdrpou)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	got := e.file(t, "first", fileAppFields(), fileAppPDF(64))
	if got.status != http.StatusConflict || got.code() != "identity.association_already_connected" {
		t.Fatalf("connected: status = %d code = %q (%s), want 409 identity.association_already_connected", got.status, got.code(), got.body)
	}
	for _, leak := range []string{"head@example.test", "Association A", "head"} {
		if strings.Contains(string(got.body), leak) {
			t.Errorf("refusal reveals %q: %s", leak, got.body)
		}
	}

	f := fileAppFields()
	f["edrpou"] = "87654321"
	f["term_end"] = "2020-01-01"
	got = e.file(t, "first", f, fileAppPDF(64))
	if got.status != http.StatusConflict || got.code() != "identity.term_ended" {
		t.Fatalf("ended term: status = %d code = %q (%s), want 409 identity.term_ended", got.status, got.code(), got.body)
	}
	if c := e.counts(t); c != [3]int{} {
		t.Errorf("rows = %v, want none stored", c)
	}
}

// AC-04b and AC-04c: a duplicate by the same person is refused; a second person is accepted
// with the very same shape as a lone applicant.
func TestFileApplicationRefusesADuplicateButNotACompetitor(t *testing.T) {
	e := newFileAppEnv(t)
	first := e.file(t, "first", fileAppFields(), fileAppPDF(64))
	if first.status != http.StatusCreated {
		t.Fatalf("first: status = %d (%s), want 201", first.status, first.body)
	}
	dup := e.file(t, "first", fileAppFields(), fileAppPDF(64))
	if dup.status != http.StatusConflict || dup.code() != "identity.application_already_open" {
		t.Fatalf("duplicate: status = %d code = %q (%s), want 409 identity.application_already_open", dup.status, dup.code(), dup.body)
	}
	if c := e.counts(t); c != [3]int{1, 1, 1} {
		t.Errorf("rows after the duplicate = %v, want [1 1 1]", c)
	}
	second := e.file(t, "second", fileAppFields(), fileAppPDF(64))
	if second.status != http.StatusCreated {
		t.Fatalf("competitor: status = %d (%s), want 201", second.status, second.body)
	}
	keys := func(b []byte) []string {
		var m map[string]any
		_ = json.Unmarshal(b, &m)
		var ks []string
		for k := range m {
			ks = append(ks, k)
		}
		return ks
	}
	if len(keys(second.body)) != len(keys(first.body)) {
		t.Errorf("competitor body %s differs in shape from the lone applicant's %s", second.body, first.body)
	}
	if c := e.counts(t); c != [3]int{2, 2, 2} {
		t.Errorf("rows = %v, want [2 2 2]", c)
	}
}

// NFR: 5 MB is filed within 5 s; one byte more is refused and nothing is stored.
func TestFileApplicationAcceptsFiveMegabytesAndRefusesOneByteMore(t *testing.T) {
	e := newFileAppEnv(t)
	start := time.Now()
	got := e.file(t, "first", fileAppFields(), fileAppPDF(5*1024*1024))
	if got.status != http.StatusCreated {
		t.Fatalf("5 MB: status = %d (%s), want 201", got.status, got.body)
	}
	if d := time.Since(start); d > 5*time.Second {
		t.Errorf("5 MB took %s, want under 5s", d)
	}
	before := e.counts(t)
	f := fileAppFields()
	f["edrpou"] = "87654321"
	over := e.file(t, "first", f, fileAppPDF(5*1024*1024+1))
	if over.status < 400 || over.status >= 500 {
		t.Fatalf("5 MB + 1: status = %d, want a 4xx refusal", over.status)
	}
	if after := e.counts(t); after != before {
		t.Errorf("rows %v -> %v, want unchanged after the oversize upload", before, after)
	}
}
