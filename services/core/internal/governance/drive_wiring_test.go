package governance_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
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

// T97 (AC-26, AC-28, review N09): with SPILNO_DEV_STORAGE unset, an association's documents live in the Drive folder
// of its Google connection, reached with its token. The resolver is the one cmd/api builds:
// governance.NewDriveResolver over identity.GoogleStorage and governance.NewDriveFiles.

// wiringDrive is a fake Google (token endpoint and Drive) that records whom each upload was for.
type wiringDrive struct {
	srv *httptest.Server

	mu       sync.Mutex
	sessions map[string]wiringSession
	files    map[string][]byte
	uploads  []wiringSession // in order, with the bearer token and parent folder of each

	tokenCalls int           // requests to the token endpoint
	hang       bool          // when set, Drive metadata and media reads never answer
	release    chan struct{} // closed to free hanging handlers
	once       sync.Once
}

type wiringSession struct {
	Auth   string
	Parent string
}

func newWiringDrive(t *testing.T) *wiringDrive {
	t.Helper()
	d := &wiringDrive{sessions: map[string]wiringSession{}, files: map[string][]byte{}, release: make(chan struct{})}
	mux := http.NewServeMux()
	// Token endpoint: the access token names the refresh token it was minted from.
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		d.mu.Lock()
		d.tokenCalls++
		d.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"access_token":"at-for-%s","token_type":"Bearer","expires_in":3600}`, r.PostForm.Get("refresh_token"))
	})
	mux.HandleFunc("/upload/drive/v3/files", func(w http.ResponseWriter, r *http.Request) {
		var meta struct {
			Parents []string `json:"parents"`
		}
		_ = json.NewDecoder(r.Body).Decode(&meta)
		d.mu.Lock()
		defer d.mu.Unlock()
		id := fmt.Sprintf("s%d", len(d.sessions)+1)
		s := wiringSession{Auth: r.Header.Get("Authorization")}
		if len(meta.Parents) == 1 {
			s.Parent = meta.Parents[0]
		}
		d.sessions[id] = s
		w.Header().Set("Location", d.srv.URL+"/session/"+id)
	})
	mux.HandleFunc("/session/", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		sum := sha256.Sum256(body)
		d.mu.Lock()
		defer d.mu.Unlock()
		id := strings.TrimPrefix(r.URL.Path, "/session/")
		s := d.sessions[id]
		s.Auth = firstNonEmpty(r.Header.Get("Authorization"), s.Auth)
		d.uploads = append(d.uploads, s)
		d.files["f-"+id] = body
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"id":"f-%s","headRevisionId":"rev-1","sha256Checksum":"%s"}`, id, hex.EncodeToString(sum[:]))
	})
	mux.HandleFunc("/drive/v3/files/", func(w http.ResponseWriter, r *http.Request) {
		d.mu.Lock()
		hang := d.hang
		d.mu.Unlock()
		if hang {
			select {
			case <-r.Context().Done():
			case <-d.release:
			}
			return
		}
		d.mu.Lock()
		body, ok := d.files[strings.TrimPrefix(r.URL.Path, "/drive/v3/files/")]
		d.mu.Unlock()
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
	d.srv = httptest.NewServer(mux)
	t.Cleanup(d.srv.Close)
	t.Cleanup(d.unhang)
	return d
}

func (d *wiringDrive) tokenRequests() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.tokenCalls
}

func (d *wiringDrive) unhang() { d.once.Do(func() { close(d.release) }) }

func (d *wiringDrive) setHang(v bool) {
	d.mu.Lock()
	d.hang = v
	d.mu.Unlock()
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

func (d *wiringDrive) uploaded() []wiringSession {
	d.mu.Lock()
	defer d.mu.Unlock()
	return append([]wiringSession(nil), d.uploads...)
}

// wiringEnv is governance behind the identity guard with Google enabled against the fake and no dev storage.
type wiringEnv struct {
	*caseEnv
	drive *wiringDrive
	idm   *identity.Module
	// useResolver replaces the storage resolver the running API asks.
	useResolver func(governance.FilesResolver)
	// otherKeyModule is a second identity module on the same database sealed under another key: it cannot read
	// the tokens idm stored.
	otherKeyModule func(t *testing.T) *identity.Module
}

func newWiringEnv(t *testing.T) *wiringEnv {
	t.Helper()
	drive := newWiringDrive(t)
	pool, a, b := testdb.NewTwoAssociations(t)
	idm := identity.New(pool)
	googleCfg := func(key byte) identity.GoogleConfig {
		return identity.GoogleConfig{KeyID: "k1", Key: bytes.Repeat([]byte{key}, 32), OAuth: &oauth2.Config{
			ClientID: "client-1", ClientSecret: "secret-1",
			Endpoint: oauth2.Endpoint{TokenURL: drive.srv.URL + "/token", AuthStyle: oauth2.AuthStyleInParams},
		}}
	}
	people := map[string]identity.ProviderIdentity{}
	for name, email := range map[string]string{"head": caseHeadEmail, "bhead": rosterBOnlyEmail} {
		people[name] = identity.ProviderIdentity{Issuer: "https://accounts.example.test", Subject: "sub-" + name, Email: email, EmailVerified: true}
	}
	idm.EnableSignIn(identity.SignInConfig{Provider: rosterProvider{people}})
	if err := idm.EnableGoogle(googleCfg(0)); err != nil {
		t.Fatal(err)
	}
	apiURL, uploadURL := drive.srv.URL+"/drive/v3", drive.srv.URL+"/upload/drive/v3"
	resolve := governance.NewDriveResolver(idm, identity.ErrGoogleNotConnected, apiURL, uploadURL)
	var resMu sync.Mutex
	current := resolve
	swappable := func(ctx context.Context, assoc uuid.UUID) (governance.Files, error) {
		resMu.Lock()
		f := current
		resMu.Unlock()
		return f(ctx, assoc)
	}
	gov := governance.New(pool, idm.HasStorageConnection, actorAccess).WithFiles(swappable)
	r := httpapi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(idm.Guard(guardedRoster{gov}))
		for _, err := range []error{idm.Routes(r), gov.Routes(r)} {
			if err != nil {
				t.Fatal(err)
			}
		}
	})
	e := &caseEnv{pool: pool, a: a, b: b, cookies: map[string]*http.Cookie{}}
	for assoc, email := range map[string]string{a: caseHeadEmail, b: rosterBOnlyEmail} {
		id := uuid.MustParse(assoc)
		e.seed(t, id, func(ctx context.Context, tx db.DBTX) { governancetest.Member(t, tx, id, email, "head") })
	}
	e.srv = httptest.NewServer(r)
	t.Cleanup(e.srv.Close)
	for name := range people {
		e.cookies[name] = rosterSignIn(t, r, name)
	}
	return &wiringEnv{caseEnv: e, drive: drive, idm: idm,
		useResolver: func(f governance.FilesResolver) { resMu.Lock(); current = f; resMu.Unlock() },
		otherKeyModule: func(t *testing.T) *identity.Module {
			other := identity.New(pool)
			if err := other.EnableGoogle(googleCfg(1)); err != nil {
				t.Fatal(err)
			}
			return other
		}}
}

// connect stores the association's Google connection; folder "" leaves the spilno folder uncreated.
func (e *wiringEnv) connect(t *testing.T, assoc, refreshToken, folder string) {
	t.Helper()
	id := uuid.MustParse(assoc)
	var by uuid.UUID
	e.seed(t, id, func(ctx context.Context, tx db.DBTX) {
		if err := tx.QueryRow(ctx, `SELECT id FROM accounts LIMIT 1`).Scan(&by); err != nil {
			t.Fatal(err)
		}
	})
	err := e.idm.ConnectGoogle(context.Background(), identity.GoogleConnection{
		AssociationID: id, AccountEmail: "osbb@example.test", RefreshToken: refreshToken,
		Scopes: []string{"https://www.googleapis.com/auth/drive.file"}, DriveRootFolderID: folder, AuthorizedBy: by,
	})
	if err != nil {
		t.Fatalf("ConnectGoogle: %v", err)
	}
}

// attach creates a справа and uploads one PDF to it; it returns the response of the upload.
func (e *wiringEnv) attach(t *testing.T, as, assoc string, data []byte) caseResp {
	t.Helper()
	c := e.mustCreate(t, as, assoc, caseTitle1)
	return e.upload(t, as, assoc, c.ID, "contract.pdf", "application/pdf", data, nil)
}

func TestDocumentsLiveInTheDriveFolderOfTheAssociationsConnection(t *testing.T) {
	e := newWiringEnv(t)
	e.connect(t, e.a, "1//token-A", "folder-A")
	e.connect(t, e.b, "1//token-B", "folder-B")
	pdfA := append([]byte("%PDF-1.7 A "), bytes.Repeat([]byte{0, 1, 2, 255}, 100)...)
	pdfB := []byte("%PDF-1.7 B only")

	for _, tc := range []struct {
		as, assoc, token, folder string
		data                     []byte
	}{
		{"head", e.a, "at-for-1//token-A", "folder-A", pdfA},
		{"bhead", e.b, "at-for-1//token-B", "folder-B", pdfB},
	} {
		got := e.attach(t, tc.as, tc.assoc, tc.data)
		if got.status != http.StatusCreated {
			t.Fatalf("upload for %s: status = %d, want 201; body %s", tc.folder, got.status, got.body)
		}
		doc := decodeCase[caseDoc](t, got)
		dl := e.json(t, tc.as, "GET", "/governance/associations/"+tc.assoc+"/files/"+doc.File.ID, "")
		if dl.status != http.StatusOK || !bytes.Equal(dl.body, tc.data) {
			t.Errorf("download for %s: status = %d, %d bytes, want 200 and the %d uploaded bytes", tc.folder, dl.status, len(dl.body), len(tc.data))
		}
		ups := e.drive.uploaded()
		last := ups[len(ups)-1]
		if last.Auth != "Bearer "+tc.token || last.Parent != tc.folder {
			t.Errorf("upload for %s reached Drive as %+v, want token %q and folder %q", tc.folder, last, "Bearer "+tc.token, tc.folder)
		}
	}
	if n := len(e.drive.uploaded()); n != 2 {
		t.Errorf("uploads to Drive = %d, want 2 (one per association, two folders)", n)
	}
}

func TestAnAssociationWithoutAConnectionOrAFolderGetsNoDrive(t *testing.T) {
	e := newWiringEnv(t)
	e.connect(t, e.b, "1//token-B", "") // connected, but no spilno folder

	got := e.attach(t, "head", e.a, []byte("%PDF-1.7 x"))
	if got.status != http.StatusConflict || got.code() != "governance.drive_not_connected" {
		t.Errorf("no connection: status = %d code = %q, want 409 governance.drive_not_connected; body %s", got.status, got.code(), got.body)
	}
	got = e.attach(t, "bhead", e.b, []byte("%PDF-1.7 x"))
	if got.status != http.StatusServiceUnavailable || got.code() != "governance.storage_unavailable" {
		t.Errorf("no folder: status = %d code = %q, want 503 governance.storage_unavailable; body %s", got.status, got.code(), got.body)
	}
	if n := len(e.drive.uploaded()); n != 0 {
		t.Errorf("uploads to Drive = %d, want 0", n)
	}
}
