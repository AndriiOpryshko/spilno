package governance_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/AndriiOpryshko/spilno/services/core/internal/governance"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/domain"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/governancetest"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/ports"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/db"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/httpapi"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/testdb"
)

// Handler tests of listCases, createCase, getCase and addCaseDocument (association-onboarding T65, carrying the
// rest of board-protocol T21: AC-25, AC-28), served behind the identity guard exactly as cmd/api does, on two
// seeded associations, with a fake Files port that records what was uploaded.
//
// Association A has a голова (caseHeadEmail) and no storage connection until a test connects it; association B
// has its own голова (rosterBOnlyEmail), a storage connection, and a справа with one document seeded by SQL.

const (
	caseHeadEmail = rosterHeadEmail
	caseWhy       = "спір щодо складу правління"
	caseTitle1    = "Гідроізоляція покрівлі, будинок 1"
	caseTitle2    = "Заміна ліфта, будинок 2"
	caseTitle3    = "Ремонт підʼїзду, будинок 3"

	mebibyte = 1 << 20
)

// Content the contract accepts, recognised by its first bytes.
var (
	casePDF  = []byte("%PDF-1.7\n1 0 obj << /Title (kosztorys) >> endobj\n")
	casePNG  = append([]byte("\x89PNG\r\n\x1a\n"), []byte("\x00\x00\x00\rIHDR....fake-pixels")...)
	caseJPEG = append([]byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10}, []byte("JFIF\x00fake-pixels")...)
)

type caseUpload struct {
	Name, MediaType string
	Data            []byte
}

// caseDrive is a ports.Files that keeps what it was given and hands out storage ids of its own.
type caseDrive struct {
	mu          sync.Mutex
	data        map[string][]byte
	ids         []string // every storage file and revision id handed out
	uploads     []caseUpload
	unavailable bool
}

var _ ports.Files = (*caseDrive)(nil)

func (d *caseDrive) uploaded() []caseUpload {
	d.mu.Lock()
	defer d.mu.Unlock()
	return append([]caseUpload(nil), d.uploads...)
}

func (d *caseDrive) handedOut() []string {
	d.mu.Lock()
	defer d.mu.Unlock()
	return append([]string(nil), d.ids...)
}

func (d *caseDrive) Upload(_ context.Context, name, mediaType string, body io.Reader) (ports.StoredFile, error) {
	d.mu.Lock()
	unavailable := d.unavailable
	d.mu.Unlock()
	if unavailable {
		return ports.StoredFile{}, &domain.DomainError{Code: domain.CodeStorageUnavailable}
	}
	b, err := io.ReadAll(body)
	if err != nil {
		return ports.StoredFile{}, err
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	n := len(d.uploads) + 1
	id, rev := fmt.Sprintf("gdrv-case-file-%d-%s", n, uuid.NewString()), fmt.Sprintf("gdrv-case-rev-%d-%s", n, uuid.NewString())
	d.data[id] = b
	d.ids = append(d.ids, id, rev)
	d.uploads = append(d.uploads, caseUpload{Name: name, MediaType: mediaType, Data: b})
	sum := sha256.Sum256(b)
	return ports.StoredFile{ID: id, Revision: rev, SHA256: hex.EncodeToString(sum[:]), Size: int64(len(b)), MediaType: mediaType}, nil
}

func (d *caseDrive) Open(_ context.Context, id string) (io.ReadCloser, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.unavailable {
		return nil, &domain.DomainError{Code: domain.CodeStorageUnavailable}
	}
	b, ok := d.data[id]
	if !ok {
		return nil, ports.ErrFileNotFound
	}
	return io.NopCloser(bytes.NewReader(b)), nil
}

func (d *caseDrive) Stat(_ context.Context, id string) (ports.FileStatus, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	b, ok := d.data[id]
	if !ok {
		return ports.FileStatus{}, ports.ErrFileNotFound
	}
	sum := sha256.Sum256(b)
	return ports.FileStatus{StoredFile: ports.StoredFile{ID: id, Revision: "rev-" + id, SHA256: hex.EncodeToString(sum[:]), Size: int64(len(b))}}, nil
}

func (d *caseDrive) SetReadOnly(context.Context, string) error { return nil }
func (d *caseDrive) Untrash(context.Context, string) error     { return nil }

type caseEnv struct {
	srv            *httptest.Server
	pool           *pgxpool.Pool
	a, b           string
	cookies        map[string]*http.Cookie
	files          *caseDrive
	headA, headB   uuid.UUID
	foreignCase    uuid.UUID // B's справа
	foreignDoc     uuid.UUID // B's document of it
	foreignDriveID []string
}

type caseResp struct {
	status int
	header http.Header
	body   []byte
}

func (r caseResp) code() string {
	var p struct {
		Code string `json:"code"`
	}
	_ = json.Unmarshal(r.body, &p)
	return p.Code
}

func (r caseResp) detail() string {
	var p struct {
		Detail string `json:"detail"`
	}
	_ = json.Unmarshal(r.body, &p)
	return p.Detail
}

func (r caseResp) pointers() []string {
	var p struct {
		Errors []struct {
			Pointer string `json:"pointer"`
		} `json:"errors"`
	}
	_ = json.Unmarshal(r.body, &p)
	var out []string
	for _, e := range p.Errors {
		out = append(out, e.Pointer)
	}
	return out
}

type caseJSON struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	CreatedBy string `json:"created_by_member_id"`
	CreatedAt string `json:"created_at"`
}

type caseFile struct {
	ID         string  `json:"id"`
	Name       string  `json:"original_name"`
	MediaType  string  `json:"media_type"`
	Size       int64   `json:"size_bytes"`
	SHA256     string  `json:"sha256"`
	UploadedBy *string `json:"uploaded_by_member_id"`
	CreatedAt  string  `json:"created_at"`
}

type caseDoc struct {
	ID         string   `json:"id"`
	CaseID     string   `json:"case_id"`
	File       caseFile `json:"file"`
	Previous   *string  `json:"previous_version_id"`
	DocumentOn *string  `json:"document_date"`
	CreatedBy  string   `json:"created_by_member_id"`
	CreatedAt  string   `json:"created_at"`
}

type caseDetail struct {
	caseJSON
	Documents []caseDoc `json:"documents"`
}

type casePage struct {
	Items      []caseJSON `json:"items"`
	NextCursor *string    `json:"next_cursor"`
}

func newCaseEnv(t *testing.T) *caseEnv {
	t.Helper()
	pool, a, b := testdb.NewTwoAssociations(t)
	idm := identity.New(pool)
	people := map[string]identity.ProviderIdentity{}
	for name, email := range map[string]string{"head": caseHeadEmail, "bhead": rosterBOnlyEmail, "operator": rosterOperatorEmail} {
		people[name] = identity.ProviderIdentity{Issuer: "https://accounts.example.test", Subject: "sub-" + name, Email: email, EmailVerified: true}
	}
	idm.EnableSignIn(identity.SignInConfig{Provider: rosterProvider{people}, OperatorEmail: rosterOperatorEmail})
	e := &caseEnv{pool: pool, a: a, b: b, cookies: map[string]*http.Cookie{}, files: &caseDrive{data: map[string][]byte{}}}
	gov := governance.New(pool, idm.HasStorageConnection, actorAccess).WithFiles(e.files)
	r := httpapi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(idm.Guard(guardedRoster{gov}))
		for _, err := range []error{idm.Routes(r), gov.Routes(r)} {
			if err != nil {
				t.Fatal(err)
			}
		}
	})

	aid, bid := uuid.MustParse(a), uuid.MustParse(b)
	memberOf := func(assoc uuid.UUID, email string) (id uuid.UUID) {
		e.seed(t, assoc, func(ctx context.Context, tx db.DBTX) {
			governancetest.Member(t, tx, assoc, email, "head")
			if err := tx.QueryRow(ctx, `SELECT board_member_id FROM board_member_revisions WHERE association_id = $1 AND sign_in_email = $2`, assoc, email).Scan(&id); err != nil {
				t.Fatal(err)
			}
		})
		return id
	}
	e.headA, e.headB = memberOf(aid, caseHeadEmail), memberOf(bid, rosterBOnlyEmail)

	e.srv = httptest.NewServer(r)
	t.Cleanup(e.srv.Close)
	for name := range people {
		e.cookies[name] = rosterSignIn(t, r, name)
	}

	// B keeps a connected storage and a справа with one document of its own.
	e.connectStorage(t, bid)
	e.foreignCase, e.foreignDoc = uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7())
	fileID := uuid.Must(uuid.NewV7())
	e.foreignDriveID = []string{"gdrv-file-b-case", "gdrv-rev-b-case", "gdrv-folder-b-case"}
	e.seed(t, bid, func(ctx context.Context, tx db.DBTX) {
		sum := sha256.Sum256(casePDF)
		for _, q := range []struct {
			sql  string
			args []any
		}{
			{`INSERT INTO cases (id, association_id, title, drive_folder_id, created_by_member_id) VALUES ($1, $2, 'Справа ОСББ Б', $3, $4)`, []any{e.foreignCase, bid, e.foreignDriveID[2], e.headB}},
			{`INSERT INTO files (id, association_id, drive_file_id, drive_revision_id, sha256, size_bytes, media_type, original_name, uploaded_by_member_id)
				VALUES ($1, $2, $3, $4, $5, $6, 'application/pdf', 'b-contract.pdf', $7)`, []any{fileID, bid, e.foreignDriveID[0], e.foreignDriveID[1], sum[:], len(casePDF), e.headB}},
			{`INSERT INTO case_documents (id, association_id, case_id, file_id, created_by_member_id) VALUES ($1, $2, $3, $4, $5)`, []any{e.foreignDoc, bid, e.foreignCase, fileID, e.headB}},
		} {
			if _, err := tx.Exec(ctx, q.sql, q.args...); err != nil {
				t.Fatal(err)
			}
		}
	})
	return e
}

func (e *caseEnv) seed(t *testing.T, assoc uuid.UUID, fn func(ctx context.Context, tx db.DBTX)) {
	t.Helper()
	err := db.InAssociation(context.Background(), e.pool, assoc, "seed@example.test", func(ctx context.Context) error {
		tx, err := db.From(ctx)
		if err != nil {
			return err
		}
		fn(ctx, tx)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// connectStorage records the association's storage connection, on behalf of any signed-in account.
func (e *caseEnv) connectStorage(t *testing.T, assoc uuid.UUID) {
	t.Helper()
	e.seed(t, assoc, func(ctx context.Context, tx db.DBTX) {
		if _, err := tx.Exec(ctx, `INSERT INTO google_connections
			(id, association_id, google_account_email, refresh_token_ciphertext, refresh_token_nonce,
			 encryption_key_id, scopes, authorized_by_account_id)
			SELECT $1, $2, 'osbb@example.test', '\x00', '\x00', 'k1', 'drive.file', id FROM accounts LIMIT 1`,
			uuid.Must(uuid.NewV7()), assoc); err != nil {
			t.Fatal(err)
		}
	})
}

func (e *caseEnv) count(t *testing.T, assoc, table string) int {
	t.Helper()
	var n int
	id := uuid.MustParse(assoc)
	e.seed(t, id, func(ctx context.Context, tx db.DBTX) {
		if err := tx.QueryRow(ctx, `SELECT count(*) FROM `+table+` WHERE association_id = $1`, id).Scan(&n); err != nil {
			t.Fatal(err)
		}
	})
	return n
}

func (e *caseEnv) send(t *testing.T, as, method, path, contentType string, body io.Reader) caseResp {
	t.Helper()
	req, err := http.NewRequest(method, e.srv.URL+path, body)
	if err != nil {
		t.Fatal(err)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	req.AddCookie(e.cookies[as])
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	return caseResp{resp.StatusCode, resp.Header, raw}
}

func (e *caseEnv) json(t *testing.T, as, method, path, body string) caseResp {
	t.Helper()
	ct := ""
	if body != "" {
		ct = "application/json"
	}
	return e.send(t, as, method, path, ct, strings.NewReader(body))
}

func caseBase(assoc string) string { return "/governance/associations/" + assoc + "/cases" }

func (e *caseEnv) create(t *testing.T, as, assoc, title string) caseResp {
	t.Helper()
	b, _ := json.Marshal(map[string]string{"title": title})
	return e.json(t, as, "POST", caseBase(assoc), string(b))
}

func (e *caseEnv) mustCreate(t *testing.T, as, assoc, title string) caseJSON {
	t.Helper()
	got := e.create(t, as, assoc, title)
	if got.status != http.StatusCreated {
		t.Fatalf("createCase %q: status = %d, want 201; body %s", title, got.status, got.body)
	}
	return decodeCase[caseJSON](t, got)
}

func decodeCase[T any](t *testing.T, r caseResp) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(r.body, &v); err != nil {
		t.Fatalf("decode %s: %v", r.body, err)
	}
	return v
}

type partSpec struct {
	field, filename, contentType string
	data                         []byte
}

func caseMultipart(t *testing.T, fields map[string]string, file *partSpec) (io.Reader, string) {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	for k, v := range fields {
		if err := w.WriteField(k, v); err != nil {
			t.Fatal(err)
		}
	}
	if file != nil {
		h := textproto.MIMEHeader{}
		h.Set("Content-Disposition", fmt.Sprintf(`form-data; name=%q; filename=%q`, file.field, file.filename))
		h.Set("Content-Type", file.contentType)
		fw, err := w.CreatePart(h)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := fw.Write(file.data); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return &buf, w.FormDataContentType()
}

func (e *caseEnv) upload(t *testing.T, as, assoc, caseID, filename, contentType string, data []byte, fields map[string]string) caseResp {
	t.Helper()
	body, ct := caseMultipart(t, fields, &partSpec{"file", filename, contentType, data})
	return e.send(t, as, "POST", caseBase(assoc)+"/"+caseID+"/documents", ct, body)
}

func (e *caseEnv) open(t *testing.T, as, assoc, caseID string) caseResp {
	t.Helper()
	return e.json(t, as, "GET", caseBase(assoc)+"/"+caseID, "")
}

// noStorageIDs fails when the response carries a storage file, revision or folder id, or the field names of one.
func (e *caseEnv) noStorageIDs(t *testing.T, what string, got caseResp) {
	t.Helper()
	all := string(got.body)
	for k, vs := range got.header {
		all += "\n" + k + ": " + strings.Join(vs, ",")
	}
	for _, id := range append(e.files.handedOut(), e.foreignDriveID...) {
		if strings.Contains(all, id) {
			t.Errorf("%s: the response carries the storage id %q", what, id)
		}
	}
	for _, key := range []string{"drive_file_id", "drive_revision_id", "drive_folder_id"} {
		if strings.Contains(all, key) {
			t.Errorf("%s: the response carries the field %q", what, key)
		}
	}
}

var sha256Hex = regexp.MustCompile(`^[0-9a-f]{64}$`)

// AC-25 (the справа exists to attach to): a член правління creates a справа without a connected storage, sees
// it in the list newest first, and opens it with an empty list of documents. Nothing storage-side is on the wire.
func TestCasesCreateListAndOpen(t *testing.T) {
	e := newCaseEnv(t)
	before := time.Now().Add(-time.Minute)

	first := e.mustCreate(t, "head", e.a, caseTitle1)
	second := e.mustCreate(t, "head", e.a, caseTitle2)
	if first.ID == "" || first.ID == second.ID {
		t.Fatalf("created ids = %q, %q, want two different ids", first.ID, second.ID)
	}
	if first.Title != caseTitle1 || first.CreatedBy != e.headA.String() {
		t.Errorf("created case = %+v, want title %q created by the голова %s", first, caseTitle1, e.headA)
	}
	if at, err := time.Parse(time.RFC3339, first.CreatedAt); err != nil || at.Before(before) || at.After(time.Now().Add(time.Minute)) {
		t.Errorf("created_at = %q (%v), want a timestamp of now", first.CreatedAt, err)
	}

	listed := e.json(t, "head", "GET", caseBase(e.a), "")
	if listed.status != http.StatusOK {
		t.Fatalf("listCases: status = %d, want 200; body %s", listed.status, listed.body)
	}
	page := decodeCase[casePage](t, listed)
	if len(page.Items) != 2 || page.Items[0].ID != second.ID || page.Items[1].ID != first.ID {
		t.Errorf("listed cases = %+v, want [%s, %s] (newest first)", page.Items, second.ID, first.ID)
	}
	if page.NextCursor != nil {
		t.Errorf("next_cursor = %q, want null on the only page", *page.NextCursor)
	}
	if !bytes.Contains(listed.body, []byte(`"next_cursor":null`)) {
		t.Errorf("list body %s, want the key next_cursor present as null", listed.body)
	}
	e.noStorageIDs(t, "listCases", listed)

	opened := e.open(t, "head", e.a, first.ID)
	if opened.status != http.StatusOK {
		t.Fatalf("getCase: status = %d, want 200; body %s", opened.status, opened.body)
	}
	got := decodeCase[caseDetail](t, opened)
	if got.ID != first.ID || got.Title != caseTitle1 || got.CreatedBy != e.headA.String() || got.CreatedAt == "" {
		t.Errorf("opened case = %+v, want the created one", got.caseJSON)
	}
	if !bytes.Contains(opened.body, []byte(`"documents":[]`)) {
		t.Errorf("opened body %s, want documents as an empty array", opened.body)
	}
	e.noStorageIDs(t, "getCase", opened)
}

// The list pages by cursor, newest first, with the shared page shape; a cursor the endpoint did not issue is 422.
func TestListCasesPagesByCursor(t *testing.T) {
	e := newCaseEnv(t)
	c1 := e.mustCreate(t, "head", e.a, caseTitle1)
	c2 := e.mustCreate(t, "head", e.a, caseTitle2)
	c3 := e.mustCreate(t, "head", e.a, caseTitle3)

	p1 := e.json(t, "head", "GET", caseBase(e.a)+"?limit=2", "")
	if p1.status != http.StatusOK {
		t.Fatalf("listCases?limit=2: status = %d, want 200; body %s", p1.status, p1.body)
	}
	page1 := decodeCase[casePage](t, p1)
	if len(page1.Items) != 2 || page1.Items[0].ID != c3.ID || page1.Items[1].ID != c2.ID || page1.NextCursor == nil {
		t.Fatalf("first page = %+v cursor %v, want [%s, %s] and a next_cursor", page1.Items, page1.NextCursor, c3.ID, c2.ID)
	}
	p2 := e.json(t, "head", "GET", caseBase(e.a)+"?limit=2&cursor="+url.QueryEscape(*page1.NextCursor), "")
	if p2.status != http.StatusOK {
		t.Fatalf("listCases second page: status = %d, want 200; body %s", p2.status, p2.body)
	}
	page2 := decodeCase[casePage](t, p2)
	if len(page2.Items) != 1 || page2.Items[0].ID != c1.ID || page2.NextCursor != nil {
		t.Errorf("second page = %+v cursor %v, want [%s] and a null next_cursor", page2.Items, page2.NextCursor, c1.ID)
	}

	bad := e.json(t, "head", "GET", caseBase(e.a)+"?cursor=not-a-cursor", "")
	if bad.status != http.StatusUnprocessableEntity {
		t.Errorf("listCases with a foreign cursor: status = %d, want 422; body %s", bad.status, bad.body)
	}
}

// createCase names every broken field at once and creates nothing; getCase of an unknown id is 404.
func TestCreateCaseRefusesABrokenTitleAndGetCaseHidesUnknownIDs(t *testing.T) {
	e := newCaseEnv(t)
	before := e.count(t, e.a, "cases")
	for _, tc := range []struct{ name, body string }{
		{"empty", `{"title":""}`},
		{"blank", `{"title":"   "}`},
		{"tab and newline", `{"title":"\t \n"}`},
		{"over 200 chars", `{"title":"` + strings.Repeat("я", 201) + `"}`},
		{"missing", `{}`},
	} {
		got := e.json(t, "head", "POST", caseBase(e.a), tc.body)
		if got.status != http.StatusUnprocessableEntity {
			t.Errorf("createCase with a %s title: status = %d, want 422; body %s", tc.name, got.status, got.body)
			continue
		}
		// The shared request validator names a missing property by the object's own (empty) pointer.
		if p := got.pointers(); tc.name != "missing" && (len(p) != 1 || !strings.HasSuffix(p[0], "title")) {
			t.Errorf("createCase with a %s title: pointers = %v, want the title named", tc.name, p)
		}
	}
	if after := e.count(t, e.a, "cases"); after != before {
		t.Errorf("cases = %d, want %d: a refused createCase must create nothing", after, before)
	}

	for name, id := range map[string]string{"unknown id": uuid.NewString()} {
		got := e.open(t, "head", e.a, id)
		if got.status != http.StatusNotFound {
			t.Errorf("getCase with an %s: status = %d, want 404; body %s", name, got.status, got.body)
		}
	}
}

// AC-28: with no storage connected addCaseDocument answers 409 governance.drive_not_connected, says any член
// правління can connect it, and attaches nothing. The storage guard is the first step: it answers before the
// справа is even looked up, and the 409 is no obstacle to createCase.
func TestAddCaseDocumentRefusedUntilTheStorageIsConnected(t *testing.T) {
	e := newCaseEnv(t)
	c := e.mustCreate(t, "head", e.a, caseTitle1)
	for name, id := range map[string]string{"an existing справа": c.ID, "an unknown справа": uuid.NewString()} {
		got := e.upload(t, "head", e.a, id, "kosztorys.pdf", "application/pdf", casePDF, nil)
		if got.status != http.StatusConflict || got.code() != "governance.drive_not_connected" {
			t.Fatalf("upload into %s before the storage is connected: status = %d code = %q, want 409 governance.drive_not_connected; body %s", name, got.status, got.code(), got.body)
		}
		if !strings.Contains(got.detail(), "Спершу підключіть сховище документів ОСББ") || !strings.Contains(got.detail(), "будь-який член правління") {
			t.Errorf("detail = %q, want the connect-the-storage message that says any член правління can do it", got.detail())
		}
	}
	if n := len(e.files.uploaded()); n != 0 {
		t.Errorf("uploads reaching the storage = %d, want 0", n)
	}
	for _, table := range []string{"files", "case_documents"} {
		if n := e.count(t, e.a, table); n != 0 {
			t.Errorf("%s rows = %d, want 0: nothing is attached", table, n)
		}
	}
}

// AC-28 (the connected side): once the storage is connected a document goes through the fake Files and is
// stored with its checksum; a newer version is a new row that names its predecessor; the old row stays; a
// download of the stored file returns the uploaded bytes; no storage id is on any response.
func TestAddCaseDocumentStoresVersionsThroughFiles(t *testing.T) {
	e := newCaseEnv(t)
	c := e.mustCreate(t, "head", e.a, caseTitle1)
	e.connectStorage(t, uuid.MustParse(e.a))

	v1 := e.upload(t, "head", e.a, c.ID, "kosztorys.pdf", "application/pdf", casePDF, map[string]string{"document_date": "2026-10-18"})
	if v1.status != http.StatusCreated {
		t.Fatalf("addCaseDocument: status = %d, want 201; body %s", v1.status, v1.body)
	}
	d1 := decodeCase[caseDoc](t, v1)
	sum := sha256.Sum256(casePDF)
	if d1.ID == "" || d1.CaseID != c.ID || d1.CreatedBy != e.headA.String() || d1.CreatedAt == "" {
		t.Errorf("document = %+v, want an id in case %s created by %s", d1, c.ID, e.headA)
	}
	if d1.File.ID == "" || d1.File.Name != "kosztorys.pdf" || d1.File.MediaType != "application/pdf" || d1.File.Size != int64(len(casePDF)) || d1.File.SHA256 != hex.EncodeToString(sum[:]) {
		t.Errorf("document file = %+v, want kosztorys.pdf, application/pdf, %d bytes, sha256 %x", d1.File, len(casePDF), sum)
	}
	if !sha256Hex.MatchString(d1.File.SHA256) {
		t.Errorf("file sha256 = %q, want 64 lowercase hex digits", d1.File.SHA256)
	}
	if d1.File.UploadedBy == nil || *d1.File.UploadedBy != e.headA.String() {
		t.Errorf("file uploaded_by_member_id = %v, want %s", d1.File.UploadedBy, e.headA)
	}
	if d1.Previous != nil || d1.DocumentOn == nil || *d1.DocumentOn != "2026-10-18" {
		t.Errorf("document previous/date = %v/%v, want null/2026-10-18", d1.Previous, d1.DocumentOn)
	}
	e.noStorageIDs(t, "addCaseDocument", v1)

	ups := e.files.uploaded()
	if len(ups) != 1 || ups[0].Name != "kosztorys.pdf" || !bytes.Equal(ups[0].Data, casePDF) {
		t.Fatalf("uploads reaching the storage = %d (%+v), want exactly kosztorys.pdf with the uploaded bytes", len(ups), ups)
	}

	// A newer version, with no date: a new row naming the previous version; the null date stays on the wire.
	revised := append(append([]byte(nil), casePDF...), []byte("revised estimate")...)
	v2 := e.upload(t, "head", e.a, c.ID, "kosztorys-v2.pdf", "application/pdf", revised, map[string]string{"previous_version_id": d1.ID})
	if v2.status != http.StatusCreated {
		t.Fatalf("addCaseDocument (newer version): status = %d, want 201; body %s", v2.status, v2.body)
	}
	d2 := decodeCase[caseDoc](t, v2)
	if d2.ID == d1.ID || d2.File.ID == d1.File.ID || d2.Previous == nil || *d2.Previous != d1.ID {
		t.Errorf("newer version = %+v, want a new document and file naming %s as previous_version_id", d2, d1.ID)
	}
	if !bytes.Contains(v2.body, []byte(`"document_date":null`)) {
		t.Errorf("newer version body %s, want the key document_date present as null", v2.body)
	}
	e.noStorageIDs(t, "addCaseDocument (newer version)", v2)

	opened := e.open(t, "head", e.a, c.ID)
	if opened.status != http.StatusOK {
		t.Fatalf("getCase: status = %d, want 200; body %s", opened.status, opened.body)
	}
	got := decodeCase[caseDetail](t, opened)
	byID := map[string]caseDoc{}
	for _, d := range got.Documents {
		byID[d.ID] = d
	}
	if len(got.Documents) != 2 || byID[d1.ID].File.Name != "kosztorys.pdf" || byID[d2.ID].Previous == nil || *byID[d2.ID].Previous != d1.ID {
		t.Errorf("opened documents = %+v, want both versions: the old row kept and the new one naming it", got.Documents)
	}
	e.noStorageIDs(t, "getCase with documents", opened)
	if n := e.count(t, e.a, "files"); n != 2 {
		t.Errorf("files rows = %d, want 2 (one per version)", n)
	}

	// The stored file is the one uploaded: downloading it returns its bytes, as an attachment.
	for _, tc := range []struct {
		doc  caseDoc
		want []byte
	}{{d1, casePDF}, {d2, revised}} {
		dl := e.json(t, "head", "GET", "/governance/associations/"+e.a+"/files/"+tc.doc.File.ID, "")
		if dl.status != http.StatusOK || !bytes.Equal(dl.body, tc.want) {
			t.Errorf("download of %s: status = %d body %q, want 200 and the uploaded bytes", tc.doc.File.Name, dl.status, dl.body)
		}
	}
}

// The media type is read from the content: a PNG and a JPEG are accepted, whatever the part header claims.
func TestAddCaseDocumentAcceptsPDFJPEGAndPNGByContent(t *testing.T) {
	e := newCaseEnv(t)
	c := e.mustCreate(t, "head", e.a, caseTitle1)
	e.connectStorage(t, uuid.MustParse(e.a))
	for _, tc := range []struct {
		name, mediaType string
		data            []byte
	}{{"scan.png", "image/png", casePNG}, {"photo.jpg", "image/jpeg", caseJPEG}, {"invoice.pdf", "application/pdf", casePDF}} {
		got := e.upload(t, "head", e.a, c.ID, tc.name, tc.mediaType, tc.data, nil)
		if got.status != http.StatusCreated {
			t.Errorf("upload of %s: status = %d, want 201; body %s", tc.name, got.status, got.body)
			continue
		}
		if d := decodeCase[caseDoc](t, got); d.File.MediaType != tc.mediaType || d.File.Name != tc.name {
			t.Errorf("upload of %s: file = %+v, want media type %s", tc.name, d.File, tc.mediaType)
		}
	}
}

// Uploads over 20 MB, and content that is not PDF, JPEG or PNG whatever its name and declared type, answer 422
// governance.file_invalid naming /file, and attach nothing. Exactly 20 MiB is still accepted.
func TestAddCaseDocumentRefusesAFileThatIsTooBigOrTheWrongType(t *testing.T) {
	e := newCaseEnv(t)
	c := e.mustCreate(t, "head", e.a, caseTitle1)
	e.connectStorage(t, uuid.MustParse(e.a))

	pad := func(n int) []byte {
		return append([]byte("%PDF-1.7\n"), bytes.Repeat([]byte{'x'}, n-len("%PDF-1.7\n"))...)
	}
	for _, tc := range []struct {
		name, filename, declared string
		data                     []byte
	}{
		{"one byte over 20 MiB", "huge.pdf", "application/pdf", pad(20*mebibyte + 1)},
		{"html named pdf", "invoice.pdf", "application/pdf", []byte("<!DOCTYPE html><html><script>alert(1)</script></html>")},
		{"executable named png", "photo.png", "image/png", []byte("MZ\x90\x00\x03\x00\x00\x00 this is a windows executable")},
		{"plain text named jpeg", "photo.jpg", "image/jpeg", []byte("just some text, not a picture")},
		{"empty file", "empty.pdf", "application/pdf", nil},
	} {
		before := e.count(t, e.a, "files")
		got := e.upload(t, "head", e.a, c.ID, tc.filename, tc.declared, tc.data, nil)
		if got.status != http.StatusUnprocessableEntity || got.code() != "governance.file_invalid" {
			t.Errorf("upload of %s: status = %d code = %q, want 422 governance.file_invalid; body %s", tc.name, got.status, got.code(), got.body)
			continue
		}
		if p := got.pointers(); len(p) == 0 || p[0] != "/file" {
			t.Errorf("upload of %s: pointers = %v, want /file", tc.name, p)
		}
		if after := e.count(t, e.a, "files"); after != before {
			t.Errorf("upload of %s: files rows = %d, want %d", tc.name, after, before)
		}
	}
	if n := e.count(t, e.a, "case_documents"); n != 0 {
		t.Errorf("case_documents rows = %d after every refusal, want 0", n)
	}

	got := e.upload(t, "head", e.a, c.ID, "big.pdf", "application/pdf", pad(20*mebibyte), nil)
	if got.status != http.StatusCreated {
		t.Fatalf("upload of exactly 20 MiB: status = %d, want 201; body %.300s", got.status, got.body)
	}
	if d := decodeCase[caseDoc](t, got); d.File.Size != 20*mebibyte {
		t.Errorf("size_bytes = %d, want %d", d.File.Size, 20*mebibyte)
	}
}

// A request that names no file, an unknown справа, a справа of another association, or a previous version that
// is not a document of this справа is refused and attaches nothing.
func TestAddCaseDocumentRefusesAMissingFileUnknownCaseAndForeignPreviousVersion(t *testing.T) {
	e := newCaseEnv(t)
	c := e.mustCreate(t, "head", e.a, caseTitle1)
	other := e.mustCreate(t, "head", e.a, caseTitle2)
	e.connectStorage(t, uuid.MustParse(e.a))
	inOther := decodeCase[caseDoc](t, e.upload(t, "head", e.a, other.ID, "other.pdf", "application/pdf", casePDF, nil))
	if inOther.ID == "" {
		t.Fatal("setup: the document of the other справа was not stored")
	}
	filesBefore, docsBefore := e.count(t, e.a, "files"), e.count(t, e.a, "case_documents")

	body, ct := caseMultipart(t, map[string]string{"document_date": "2026-10-18"}, nil)
	got := e.send(t, "head", "POST", caseBase(e.a)+"/"+c.ID+"/documents", ct, body)
	if got.status != http.StatusUnprocessableEntity {
		t.Errorf("upload with no file: status = %d, want 422; body %s", got.status, got.body)
	}

	for name, id := range map[string]string{"an unknown справа": uuid.NewString(), "a справа of association B": e.foreignCase.String()} {
		got := e.upload(t, "head", e.a, id, "x.pdf", "application/pdf", casePDF, nil)
		if got.status != http.StatusNotFound {
			t.Errorf("upload into %s: status = %d, want 404; body %s", name, got.status, got.body)
		}
	}

	for name, prev := range map[string]string{
		"a document of another справа": inOther.ID,
		"an unknown document":          uuid.NewString(),
		"a document of association B":  e.foreignDoc.String(),
	} {
		got := e.upload(t, "head", e.a, c.ID, "v2.pdf", "application/pdf", casePDF, map[string]string{"previous_version_id": prev})
		if got.status != http.StatusUnprocessableEntity {
			t.Errorf("upload naming %s as previous_version_id: status = %d, want 422; body %s", name, got.status, got.body)
		}
	}
	if files, docs := e.count(t, e.a, "files"), e.count(t, e.a, "case_documents"); files != filesBefore || docs != docsBefore {
		t.Errorf("files/case_documents rows = %d/%d, want %d/%d unchanged", files, docs, filesBefore, docsBefore)
	}
}

// Flow 8: a storage outage answers 503 governance.storage_unavailable and commits nothing; so does a missing
// Files wiring (no dev storage, no Drive adapter).
func TestAddCaseDocumentAnswersStorageUnavailable(t *testing.T) {
	e := newCaseEnv(t)
	c := e.mustCreate(t, "head", e.a, caseTitle1)
	e.connectStorage(t, uuid.MustParse(e.a))

	e.files.mu.Lock()
	e.files.unavailable = true
	e.files.mu.Unlock()
	got := e.upload(t, "head", e.a, c.ID, "kosztorys.pdf", "application/pdf", casePDF, nil)
	if got.status != http.StatusServiceUnavailable || got.code() != "governance.storage_unavailable" {
		t.Fatalf("upload during an outage: status = %d code = %q, want 503 governance.storage_unavailable; body %s", got.status, got.code(), got.body)
	}
	for _, table := range []string{"files", "case_documents"} {
		if n := e.count(t, e.a, table); n != 0 {
			t.Errorf("%s rows = %d after the outage, want 0", table, n)
		}
	}
	opened := decodeCase[caseDetail](t, e.open(t, "head", e.a, c.ID))
	if len(opened.Documents) != 0 {
		t.Errorf("documents after the outage = %+v, want none", opened.Documents)
	}
}

// AC-25: while the association is disabled every operation of the справи answers 403
// identity.association_disabled with the reason, creates nothing and attaches nothing, whether or not the
// storage is connected.
func TestCasesAreRefusedWhileTheAssociationIsDisabled(t *testing.T) {
	e := newCaseEnv(t)
	c := e.mustCreate(t, "head", e.a, caseTitle1)
	e.connectStorage(t, uuid.MustParse(e.a))
	if got := e.upload(t, "head", e.a, c.ID, "kosztorys.pdf", "application/pdf", casePDF, nil); got.status != http.StatusCreated {
		t.Fatalf("setup upload: status = %d, want 201; body %s", got.status, got.body)
	}

	disabling, _ := json.Marshal(map[string]string{"reason": caseWhy})
	if got := e.json(t, "operator", "POST", "/identity/operator/associations/"+e.a+"/disabling", string(disabling)); got.status != http.StatusOK {
		t.Fatalf("disableAssociation: status = %d, want 200; body %s", got.status, got.body)
	}
	casesBefore, filesBefore, docsBefore, upsBefore := e.count(t, e.a, "cases"), e.count(t, e.a, "files"), e.count(t, e.a, "case_documents"), len(e.files.uploaded())

	title, _ := json.Marshal(map[string]string{"title": caseTitle2})
	ops := map[string]caseResp{
		"listCases":       e.json(t, "head", "GET", caseBase(e.a), ""),
		"createCase":      e.json(t, "head", "POST", caseBase(e.a), string(title)),
		"getCase":         e.open(t, "head", e.a, c.ID),
		"addCaseDocument": e.upload(t, "head", e.a, c.ID, "more.pdf", "application/pdf", casePDF, nil),
	}
	for name, got := range ops {
		if got.status != http.StatusForbidden || got.code() != "identity.association_disabled" || !strings.Contains(got.detail(), caseWhy) {
			t.Errorf("%s while disabled: status = %d code = %q detail = %q, want 403 identity.association_disabled carrying %q; body %s", name, got.status, got.code(), got.detail(), caseWhy, got.body)
		}
	}
	if n := e.count(t, e.a, "cases"); n != casesBefore {
		t.Errorf("cases = %d, want %d", n, casesBefore)
	}
	if files, docs := e.count(t, e.a, "files"), e.count(t, e.a, "case_documents"); files != filesBefore || docs != docsBefore {
		t.Errorf("files/case_documents = %d/%d, want %d/%d", files, docs, filesBefore, docsBefore)
	}
	if n := len(e.files.uploaded()); n != upsBefore {
		t.Errorf("uploads reaching the storage = %d, want %d", n, upsBefore)
	}
}

// AC-25 (the storage guard and the disabled state): a disabled association with no storage answers the
// disabled refusal, not the storage one; the member check comes first.
func TestAddCaseDocumentInADisabledAssociationWithoutStorageAnswersDisabled(t *testing.T) {
	e := newCaseEnv(t)
	c := e.mustCreate(t, "head", e.a, caseTitle1)
	disabling, _ := json.Marshal(map[string]string{"reason": caseWhy})
	if got := e.json(t, "operator", "POST", "/identity/operator/associations/"+e.a+"/disabling", string(disabling)); got.status != http.StatusOK {
		t.Fatalf("disableAssociation: status = %d, want 200; body %s", got.status, got.body)
	}
	got := e.upload(t, "head", e.a, c.ID, "kosztorys.pdf", "application/pdf", casePDF, nil)
	if got.status != http.StatusForbidden || got.code() != "identity.association_disabled" {
		t.Errorf("upload: status = %d code = %q, want 403 identity.association_disabled; body %s", got.status, got.code(), got.body)
	}
}

// Isolation across two associations: each голова lists and opens only the справи of their own association; a
// справа id of the other association answers 404 on their own path and 403 identity.not_current_member on the
// other's path; a document cannot be attached to the other's справа; rows carry the association that wrote them.
func TestCasesAreIsolatedBetweenAssociations(t *testing.T) {
	e := newCaseEnv(t)
	e.connectStorage(t, uuid.MustParse(e.a))
	ca := e.mustCreate(t, "head", e.a, caseTitle1)
	docA := decodeCase[caseDoc](t, e.upload(t, "head", e.a, ca.ID, "a-contract.pdf", "application/pdf", casePDF, nil))
	cb := e.mustCreate(t, "bhead", e.b, caseTitle3)
	if cb.CreatedBy != e.headB.String() {
		t.Errorf("B's справа created_by_member_id = %s, want B's голова %s", cb.CreatedBy, e.headB)
	}

	listA := decodeCase[casePage](t, e.json(t, "head", "GET", caseBase(e.a), ""))
	if len(listA.Items) != 1 || listA.Items[0].ID != ca.ID {
		t.Errorf("A's list = %+v, want only A's справа %s", listA.Items, ca.ID)
	}
	listB := e.json(t, "bhead", "GET", caseBase(e.b), "")
	pageB := decodeCase[casePage](t, listB)
	ids := map[string]bool{}
	for _, it := range pageB.Items {
		ids[it.ID] = true
	}
	if len(pageB.Items) != 2 || !ids[cb.ID] || !ids[e.foreignCase.String()] || ids[ca.ID] {
		t.Errorf("B's list = %+v, want B's seeded справа and the one created in B, never A's %s", pageB.Items, ca.ID)
	}
	e.noStorageIDs(t, "B's list", listB)

	if got := e.open(t, "head", e.a, e.foreignCase.String()); got.status != http.StatusNotFound {
		t.Errorf("getCase of B's справа on A's path: status = %d, want 404; body %s", got.status, got.body)
	}
	if got := e.open(t, "bhead", e.b, ca.ID); got.status != http.StatusNotFound {
		t.Errorf("getCase of A's справа on B's path: status = %d, want 404; body %s", got.status, got.body)
	}
	if got := e.open(t, "head", e.b, e.foreignCase.String()); got.status != http.StatusForbidden || got.code() != "identity.not_current_member" {
		t.Errorf("getCase on B's path by A's голова: status = %d code = %q, want 403 identity.not_current_member", got.status, got.code())
	}
	filesB, docsB := e.count(t, e.b, "files"), e.count(t, e.b, "case_documents")
	if got := e.upload(t, "bhead", e.b, ca.ID, "x.pdf", "application/pdf", casePDF, nil); got.status != http.StatusNotFound {
		t.Errorf("upload into A's справа on B's path: status = %d, want 404; body %s", got.status, got.body)
	}
	if got := e.upload(t, "bhead", e.b, cb.ID, "x.pdf", "application/pdf", casePDF, map[string]string{"previous_version_id": docA.ID}); got.status != http.StatusUnprocessableEntity {
		t.Errorf("upload naming A's document as previous version: status = %d, want 422; body %s", got.status, got.body)
	}
	if files, docs := e.count(t, e.b, "files"), e.count(t, e.b, "case_documents"); files != filesB || docs != docsB {
		t.Errorf("B's files/case_documents = %d/%d, want %d/%d unchanged", files, docs, filesB, docsB)
	}
	if n := e.count(t, e.a, "cases"); n != 1 {
		t.Errorf("A's cases = %d, want 1", n)
	}
	if n := e.count(t, e.b, "cases"); n != 2 {
		t.Errorf("B's cases = %d, want 2 (seeded and created)", n)
	}
}
