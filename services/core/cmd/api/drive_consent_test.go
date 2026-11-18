package main

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
	"strings"
	"sync"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"golang.org/x/oauth2"

	"github.com/AndriiOpryshko/spilno/services/core/internal/governance"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/governancetest"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/db"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/httpapi"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/testdb"
)

// T112 (AC-26, AC-27, AC-28, AC-28b, review R01, R02): the consent and the Drive resolver below are the ones
// wireGoogle and driveFiles build, the functions cmd/api runs; the test re-types neither.

const consentMemberEmail = "member@example.test"

// consentGoogle is a fake Google: token endpoint, Drive about/files and resumable upload.
type consentGoogle struct {
	srv *httptest.Server

	mu       sync.Mutex
	sessions map[string]string // upload session id -> bearer token it was opened with
	files    map[string][]byte
	parents  []string // parent folder of each upload
}

func newConsentGoogle(t *testing.T) *consentGoogle {
	t.Helper()
	g := &consentGoogle{sessions: map[string]string{}, files: map[string][]byte{}}
	mux := http.NewServeMux()
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		w.Header().Set("Content-Type", "application/json")
		if r.PostForm.Get("grant_type") == "refresh_token" {
			_, _ = fmt.Fprintf(w, `{"access_token":"at-for-%s","token_type":"Bearer","expires_in":3600}`, r.PostForm.Get("refresh_token"))
			return
		}
		_, _ = w.Write([]byte(`{"access_token":"at-1","refresh_token":"1//consent-refresh","token_type":"Bearer","expires_in":3600,"scope":"https://www.googleapis.com/auth/drive.file https://www.googleapis.com/auth/documents"}`))
	})
	mux.HandleFunc("/drive/v3/about", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"user":{"emailAddress":"osbb@example.test"}}`))
	})
	mux.HandleFunc("/drive/v3/files", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			_, _ = w.Write([]byte(`{"id":"folder-consent"}`))
			return
		}
		_, _ = w.Write([]byte(`{"files":[]}`))
	})
	mux.HandleFunc("/upload/drive/v3/files", func(w http.ResponseWriter, r *http.Request) {
		var meta struct {
			Parents []string `json:"parents"`
		}
		_ = json.NewDecoder(r.Body).Decode(&meta)
		g.mu.Lock()
		defer g.mu.Unlock()
		id := fmt.Sprintf("s%d", len(g.sessions)+1)
		g.sessions[id] = r.Header.Get("Authorization")
		if len(meta.Parents) == 1 {
			g.parents = append(g.parents, meta.Parents[0])
		}
		w.Header().Set("Location", g.srv.URL+"/session/"+id)
	})
	mux.HandleFunc("/session/", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		sum := sha256.Sum256(body)
		id := strings.TrimPrefix(r.URL.Path, "/session/")
		g.mu.Lock()
		g.files["f-"+id] = body
		g.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"id":"f-%s","headRevisionId":"rev-1","sha256Checksum":"%s"}`, id, hex.EncodeToString(sum[:]))
	})
	mux.HandleFunc("/drive/v3/files/", func(w http.ResponseWriter, r *http.Request) {
		g.mu.Lock()
		body, ok := g.files[strings.TrimPrefix(r.URL.Path, "/drive/v3/files/")]
		g.mu.Unlock()
		if !ok {
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("alt") == "media" {
			_, _ = w.Write(body)
			return
		}
		sum := sha256.Sum256(body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"id":"x","headRevisionId":"rev-1","sha256Checksum":"%s","size":"%d","mimeType":"application/pdf","trashed":false}`, hex.EncodeToString(sum[:]), len(body))
	})
	g.srv = httptest.NewServer(mux)
	t.Cleanup(g.srv.Close)
	return g
}

type consentProvider struct{}

func (consentProvider) AuthCodeURL(state, _, _ string) string {
	return "https://accounts.example.test/auth?" + url.Values{"state": {state}}.Encode()
}

func (consentProvider) Exchange(context.Context, string, string, string) (identity.ProviderIdentity, error) {
	return identity.ProviderIdentity{Issuer: "https://accounts.example.test", Subject: "sub-member", Email: consentMemberEmail, EmailVerified: true}, nil
}

func consentActorAccess(ctx context.Context) (bool, string, bool) {
	a, ok := identity.ActorFrom(ctx)
	return a.ReadOnly, a.DisabledReason, ok
}

func TestConsentBuiltByWireGoogleCompletesAndDocumentsGoThroughTheDriveResolver(t *testing.T) {
	g := newConsentGoogle(t)
	pool, a, _ := testdb.NewTwoAssociations(t)
	assoc := uuid.MustParse(a)

	idm := identity.New(pool)
	idm.EnableSignIn(identity.SignInConfig{Provider: consentProvider{}})
	gov := governance.New(pool, idm.HasStorageConnection, consentActorAccess)
	gc := &identity.GoogleConfig{KeyID: "k1", Key: make([]byte, 32), OAuth: &oauth2.Config{
		ClientID: "client-1", ClientSecret: "secret-1", RedirectURL: "http://localhost:5180/storage/callback",
		Endpoint: oauth2.Endpoint{AuthURL: "https://accounts.example.test/o/oauth2/auth", TokenURL: g.srv.URL + "/token", AuthStyle: oauth2.AuthStyleInParams},
	}}
	apiURL, uploadURL := g.srv.URL+"/drive/v3", g.srv.URL+"/upload/drive/v3"
	// The same two calls run does: wireGoogle with the configured Drive address, then the resolver.
	wired, err := wireGoogle(idm, gov, gc, apiURL)
	if err != nil || !wired {
		t.Fatalf("wireGoogle: wired %v, err %v", wired, err)
	}
	gov.WithFiles(driveFiles(idm, apiURL, uploadURL))

	err = db.InAssociation(context.Background(), pool, assoc, "seed@example.test", func(ctx context.Context) error {
		tx, err := db.From(ctx)
		if err != nil {
			return err
		}
		governancetest.Member(t, tx, assoc, consentMemberEmail, "head")
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	r := httpapi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(idm.Guard(rosterOf{gov}))
		for _, err := range []error{idm.Routes(r), gov.Routes(r)} {
			if err != nil {
				t.Fatal(err)
			}
		}
	})
	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)
	cookie := consentSignIn(t, r)

	do := func(method, path, contentType string, body io.Reader) (int, http.Header, []byte) {
		t.Helper()
		req, err := http.NewRequest(method, srv.URL+path, body)
		if err != nil {
			t.Fatal(err)
		}
		if contentType != "" {
			req.Header.Set("Content-Type", contentType)
		}
		req.AddCookie(cookie)
		resp, err := (&http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}).Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = resp.Body.Close() }()
		raw, _ := io.ReadAll(resp.Body)
		return resp.StatusCode, resp.Header, raw
	}

	// AC-27: consent link, then the code exchange, the account read and the folder.
	base := "/identity/associations/" + a + "/storage-connection"
	status, hdr, body := do("GET", base+"/authorization", "", nil)
	if status != http.StatusFound {
		t.Fatalf("startStorageConnection: status %d, want 302; body %s", status, body)
	}
	loc, err := url.Parse(hdr.Get("Location"))
	if err != nil || loc.Query().Get("state") == "" {
		t.Fatalf("Location %q carries no state", hdr.Get("Location"))
	}
	payload, _ := json.Marshal(map[string]string{"code": "code-1", "state": loc.Query().Get("state")})
	if status, _, body = do("POST", base, "application/json", bytes.NewReader(payload)); status/100 != 2 {
		t.Fatalf("completeStorageConnection: status %d, want 2xx (a consent built without the Drive address fails after the code exchange); body %s", status, body)
	}
	var folder *string
	err = db.InAssociation(context.Background(), pool, assoc, "test@example.test", func(ctx context.Context) error {
		tx, err := db.From(ctx)
		if err != nil {
			return err
		}
		return tx.QueryRow(ctx, `SELECT drive_root_folder_id FROM google_connections WHERE association_id = $1`, assoc).Scan(&folder)
	})
	if err != nil || folder == nil || *folder != "folder-consent" {
		t.Fatalf("connection folder = %v (err %v), want %q", folder, err, "folder-consent")
	}

	// AC-26, AC-28, AC-28b: a document goes up to that folder with the connection's token and comes back.
	cbody, _ := json.Marshal(map[string]string{"title": "Гідроізоляція покрівлі"})
	status, _, body = do("POST", "/governance/associations/"+a+"/cases", "application/json", bytes.NewReader(cbody))
	if status != http.StatusCreated {
		t.Fatalf("createCase: status %d, want 201; body %s", status, body)
	}
	var c struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(body, &c)
	pdf := append([]byte("%PDF-1.7 consent "), bytes.Repeat([]byte{0, 1, 2, 255}, 50)...)
	mbody, ct := consentMultipart(t, "contract.pdf", pdf)
	status, _, body = do("POST", "/governance/associations/"+a+"/cases/"+c.ID+"/documents", ct, mbody)
	if status != http.StatusCreated {
		t.Fatalf("upload: status %d, want 201; body %s", status, body)
	}
	var doc struct {
		File struct {
			ID string `json:"id"`
		} `json:"file"`
	}
	_ = json.Unmarshal(body, &doc)
	status, _, body = do("GET", "/governance/associations/"+a+"/files/"+doc.File.ID, "", nil)
	if status != http.StatusOK || !bytes.Equal(body, pdf) {
		t.Errorf("download: status %d, %d bytes, want 200 and the %d uploaded bytes", status, len(body), len(pdf))
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if len(g.parents) != 1 || g.parents[0] != "folder-consent" {
		t.Errorf("upload parents = %v, want [folder-consent]", g.parents)
	}
	for id, auth := range g.sessions {
		if auth != "Bearer at-for-1//consent-refresh" {
			t.Errorf("upload session %s opened with %q, want the connection's token", id, auth)
		}
	}
}

func consentMultipart(t *testing.T, filename string, data []byte) (io.Reader, string) {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	h := textproto.MIMEHeader{}
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename=%q`, filename))
	h.Set("Content-Type", "application/pdf")
	fw, err := w.CreatePart(h)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = fw.Write(data)
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return &buf, w.FormDataContentType()
}

func consentSignIn(t *testing.T, h http.Handler) *http.Cookie {
	t.Helper()
	start := httptest.NewRecorder()
	h.ServeHTTP(start, httptest.NewRequest("GET", "/identity/sign-in", nil))
	loc, err := url.Parse(start.Header().Get("Location"))
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("GET", "/identity/sign-in/callback?code=member&state="+url.QueryEscape(loc.Query().Get("state")), nil)
	for _, c := range start.Result().Cookies() {
		req.AddCookie(c)
	}
	cb := httptest.NewRecorder()
	h.ServeHTTP(cb, req)
	for _, c := range cb.Result().Cookies() {
		if c.Name == "spilno_session" && c.Value != "" {
			return c
		}
	}
	t.Fatal("sign-in set no session cookie")
	return nil
}
