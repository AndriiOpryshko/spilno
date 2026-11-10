package identity_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/oauth2"

	"github.com/AndriiOpryshko/spilno/services/core/internal/governance"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/governancetest"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/db"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/google"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/httpapi"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/testdb"
)

// storageT33Env serves identity behind the guard on the app-role pool with a fake Google (token
// endpoint and Drive) and the storage connection turned on (Flow 10, ADR-0008).
type storageT33Env struct {
	srv     *httptest.Server
	pool    *pgxpool.Pool
	a, b    string
	cookies map[string]*http.Cookie
	ids     map[string]uuid.UUID // account id by person

	mu          sync.Mutex
	driveEmail  string // the account the fake Drive says gave the consent
	refuseCodes bool
}

const storageT33Token = "1//storage-refresh-token"

func newStorageT33Env(t *testing.T) *storageT33Env {
	t.Helper()
	pool, a, b := testdb.NewTwoAssociations(t)
	env := &storageT33Env{pool: pool, a: a, b: b, cookies: map[string]*http.Cookie{}, ids: map[string]uuid.UUID{}, driveEmail: "osbb@example.test"}

	gfake := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		env.mu.Lock()
		email, refuse := env.driveEmail, env.refuseCodes
		env.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/token":
			if refuse {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"error":"invalid_grant"}`))
				return
			}
			_, _ = w.Write([]byte(`{"access_token":"at-1","refresh_token":"` + storageT33Token + `","token_type":"Bearer","expires_in":3600,"scope":"https://www.googleapis.com/auth/drive.file https://www.googleapis.com/auth/documents"}`))
		case r.URL.Path == "/drive/v3/about":
			_, _ = w.Write([]byte(`{"user":{"emailAddress":"` + email + `"}}`))
		case r.Method == http.MethodGet && r.URL.Path == "/drive/v3/files":
			_, _ = w.Write([]byte(`{"files":[]}`))
		case r.Method == http.MethodPost && r.URL.Path == "/drive/v3/files":
			_, _ = w.Write([]byte(`{"id":"folder-1"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(gfake.Close)

	oauth := &oauth2.Config{
		ClientID: "client-1", ClientSecret: "secret-1", RedirectURL: "http://localhost:5180/storage/callback",
		Endpoint: oauth2.Endpoint{AuthURL: "https://accounts.example.test/o/oauth2/auth", TokenURL: gfake.URL + "/token", AuthStyle: oauth2.AuthStyleInParams},
	}
	idm := identity.New(pool)
	if err := idm.EnableGoogle(identity.GoogleConfig{KeyID: "k1", Key: googleTestKey(1), OAuth: oauth}); err != nil {
		t.Fatal(err)
	}
	gov := governance.New(pool)
	idm.EnableStorageConnection(identity.StorageConnectionConfig{
		Consent: google.Consent{OAuth: oauth, DriveURL: gfake.URL + "/drive/v3"},
		Members: gov,
	})

	people := map[string]identity.ProviderIdentity{}
	for p, email := range map[string]string{"member": "member@example.test", "other": "Other@Example.Test", "memberB": "memberb@example.test", "nobody": "nobody@example.test", "operator": guardOperatorEmail} {
		people[p] = identity.ProviderIdentity{Issuer: "https://accounts.example.test", Subject: "sub-" + p, Email: email, EmailVerified: true}
	}
	idm.EnableSignIn(identity.SignInConfig{Provider: &signInFake{people: people}, OperatorEmail: guardOperatorEmail})
	r := httpapi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(idm.Guard(guardRoster{gov}))
		if err := idm.Routes(r); err != nil {
			t.Fatal(err)
		}
	})
	env.srv = httptest.NewServer(r)
	t.Cleanup(env.srv.Close)

	for p := range people {
		c := sessionCookieOf(signInAs(t, env.srv, p))
		if c == nil {
			t.Fatalf("no session cookie for %s", p)
		}
		env.cookies[p] = c
	}
	for p, email := range map[string]string{"member": "member@example.test", "other": "other@example.test"} {
		seedStorageT33Member(t, pool, a, email)
		_ = p
	}
	seedStorageT33Member(t, pool, b, "memberb@example.test")
	for p, email := range map[string]string{"member": "member@example.test", "other": "other@example.test"} {
		var id uuid.UUID
		if err := pool.QueryRow(t.Context(), `SELECT id FROM accounts WHERE lower(email) = $1`, email).Scan(&id); err != nil {
			t.Fatalf("account of %s: %v", p, err)
		}
		env.ids[p] = id
	}
	return env
}

func seedStorageT33Member(t *testing.T, pool *pgxpool.Pool, assoc, email string) {
	t.Helper()
	aid := uuid.MustParse(assoc)
	err := db.InAssociation(context.Background(), pool, aid, "seed@example.test", func(ctx context.Context) error {
		tx, err := db.From(ctx)
		if err != nil {
			return err
		}
		governancetest.Member(t, tx, aid, email, "member")
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func (e *storageT33Env) do(t *testing.T, as, method, path, body string) (int, http.Header, []byte) {
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

func storageT33Path(assoc string) string {
	return "/identity/associations/" + assoc + "/storage-connection"
}

// start runs startStorageConnection and returns the state in the consent link.
func (e *storageT33Env) start(t *testing.T, as, assoc string) string {
	t.Helper()
	status, hdr, body := e.do(t, as, "GET", storageT33Path(assoc)+"/authorization", "")
	if status != http.StatusFound {
		t.Fatalf("startStorageConnection as %s: status %d, want 302; body %s", as, status, body)
	}
	u, err := url.Parse(hdr.Get("Location"))
	if err != nil || !strings.HasPrefix(hdr.Get("Location"), "https://accounts.example.test/o/oauth2/auth") {
		t.Fatalf("Location = %q, want the consent link", hdr.Get("Location"))
	}
	state := u.Query().Get("state")
	if state == "" {
		t.Fatalf("the consent link %q carries no state", u)
	}
	return state
}

func (e *storageT33Env) complete(t *testing.T, as, assoc, code, state string) (int, []byte) {
	t.Helper()
	status, _, body := e.do(t, as, "POST", storageT33Path(assoc), `{"code":"`+code+`","state":"`+state+`"}`)
	return status, body
}

// rows returns (account email, authorized_by) of the association's stored connections.
func (e *storageT33Env) rows(t *testing.T, assoc string) [][2]string {
	t.Helper()
	var out [][2]string
	err := db.InAssociation(context.Background(), e.pool, uuid.MustParse(assoc), "test@example.test", func(ctx context.Context) error {
		tx, err := db.From(ctx)
		if err != nil {
			return err
		}
		rows, err := tx.Query(ctx, `SELECT google_account_email, authorized_by_account_id::text FROM google_connections WHERE association_id = $1`, assoc)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var r [2]string
			if err := rows.Scan(&r[0], &r[1]); err != nil {
				return err
			}
			out = append(out, r)
		}
		return rows.Err()
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func (e *storageT33Env) wantNothingStored(t *testing.T, assoc string) {
	t.Helper()
	if rows := e.rows(t, assoc); len(rows) != 0 {
		t.Errorf("stored connections = %v, want none", rows)
	}
}

func storageT33Status(t *testing.T, body []byte) (connected bool, email string) {
	t.Helper()
	var s struct {
		Connected bool    `json:"connected"`
		Email     *string `json:"google_account_email"`
	}
	if err := json.Unmarshal(body, &s); err != nil {
		t.Fatalf("decode %s: %v", body, err)
	}
	if s.Email != nil {
		email = *s.Email
	}
	return s.Connected, email
}

func TestStorageConnectionIsStoredWithAccountAndMember(t *testing.T) { // AC-27
	env := newStorageT33Env(t)

	if status, _, body := env.do(t, "member", "GET", storageT33Path(env.a), ""); status != 200 {
		t.Fatalf("getStorageConnection before: status %d, body %s", status, body)
	} else if c, _ := storageT33Status(t, body); c {
		t.Errorf("connected before any consent = true, want false")
	}

	state := env.start(t, "member", env.a)
	status, body := env.complete(t, "member", env.a, "code-1", state)
	if status != http.StatusCreated {
		t.Fatalf("completeStorageConnection: status %d, want 201; body %s", status, body)
	}
	if c, email := storageT33Status(t, body); !c || email != "osbb@example.test" {
		t.Errorf("answer = connected %v, account %q; want true, osbb@example.test", c, email)
	}
	rows := env.rows(t, env.a)
	if len(rows) != 1 || rows[0][0] != "osbb@example.test" || rows[0][1] != env.ids["member"].String() {
		t.Errorf("stored rows = %v, want one row of osbb@example.test authorized by the member's account %s", rows, env.ids["member"])
	}

	// every current member of the association reads it
	status, _, body = env.do(t, "other", "GET", storageT33Path(env.a), "")
	if c, email := storageT33Status(t, body); status != 200 || !c || email != "osbb@example.test" {
		t.Errorf("another member reads: status %d connected %v account %q; want 200, true, osbb@example.test", status, c, email)
	}
}

func TestStorageConnectionRefusesAMemberSignInAddress(t *testing.T) { // AC-28b
	env := newStorageT33Env(t)
	env.driveEmail = "OTHER@example.TEST" // another member's sign-in address in other letter case

	state := env.start(t, "member", env.a)
	status, body := env.complete(t, "member", env.a, "code-1", state)
	if status != http.StatusConflict || assocStateProblemCode(body) != "identity.storage_account_personal" {
		t.Errorf("status %d code %q, want 409 identity.storage_account_personal; body %s", status, assocStateProblemCode(body), body)
	}
	env.wantNothingStored(t, env.a)
}

func TestStorageConnectionRefusesABadStateOrCode(t *testing.T) {
	env := newStorageT33Env(t)

	t.Run("unknown state", func(t *testing.T) {
		status, _ := env.complete(t, "member", env.a, "code-1", "not-a-state")
		if status != 422 {
			t.Errorf("status %d, want 422", status)
		}
		env.wantNothingStored(t, env.a)
	})
	t.Run("state of another session", func(t *testing.T) {
		state := env.start(t, "member", env.a)
		if status, _ := env.complete(t, "other", env.a, "code-1", state); status != 422 {
			t.Errorf("status %d, want 422", status)
		}
		env.wantNothingStored(t, env.a)
	})
	t.Run("state of another association", func(t *testing.T) {
		state := env.start(t, "memberB", env.b)
		if status, _ := env.complete(t, "member", env.a, "code-1", state); status != 422 {
			t.Errorf("status %d, want 422", status)
		}
		env.wantNothingStored(t, env.a)
		env.wantNothingStored(t, env.b)
	})
	t.Run("replayed state", func(t *testing.T) {
		state := env.start(t, "member", env.a)
		if status, body := env.complete(t, "member", env.a, "code-1", state); status != 201 {
			t.Fatalf("first completion: status %d, body %s", status, body)
		}
		env.do(t, "member", "GET", storageT33Path(env.a), "")
		before := env.rows(t, env.a)
		env.mu.Lock()
		env.driveEmail = "second@example.test"
		env.mu.Unlock()
		if status, _ := env.complete(t, "member", env.a, "code-2", state); status != 422 {
			t.Errorf("replay: status %d, want 422", status)
		}
		if after := env.rows(t, env.a); len(after) != 1 || after[0] != before[0] {
			t.Errorf("rows after replay = %v, want unchanged %v", after, before)
		}
	})
	t.Run("code the provider refuses", func(t *testing.T) {
		env := newStorageT33Env(t)
		env.mu.Lock()
		env.refuseCodes = true
		env.mu.Unlock()
		state := env.start(t, "member", env.a)
		if status, _ := env.complete(t, "member", env.a, "bad", state); status != 422 {
			t.Errorf("status %d, want 422", status)
		}
		env.wantNothingStored(t, env.a)
	})
}

func TestStorageConnectionIsForCurrentMembersOnly(t *testing.T) {
	env := newStorageT33Env(t)
	for _, as := range []string{"nobody", "memberB"} {
		if status, _, _ := env.do(t, as, "GET", storageT33Path(env.a)+"/authorization", ""); status != 403 {
			t.Errorf("start as %s: status %d, want 403", as, status)
		}
		if status, _ := env.complete(t, as, env.a, "code-1", "s"); status != 403 {
			t.Errorf("complete as %s: status %d, want 403", as, status)
		}
	}
	env.wantNothingStored(t, env.a)
}

func TestLaterStorageConnectionReplacesTheRow(t *testing.T) {
	env := newStorageT33Env(t)
	if status, body := env.complete(t, "member", env.a, "c1", env.start(t, "member", env.a)); status != 201 {
		t.Fatalf("first: status %d, body %s", status, body)
	}
	env.driveEmail = "second@example.test"
	if status, body := env.complete(t, "other", env.a, "c2", env.start(t, "other", env.a)); status != 201 {
		t.Fatalf("second: status %d, body %s", status, body)
	}
	rows := env.rows(t, env.a)
	if len(rows) != 1 || rows[0][0] != "second@example.test" || rows[0][1] != env.ids["other"].String() {
		t.Errorf("rows = %v, want one row of second@example.test by the second member", rows)
	}
}

// storageT33Full is the whole StorageConnection answer, nulls kept visible.
type storageT33Full struct {
	ConnectedAt *time.Time `json:"connected_at"`
	ConnectedBy *string    `json:"connected_by_member_id"`
}

func storageT33Decode(t *testing.T, body []byte) storageT33Full {
	t.Helper()
	var s storageT33Full
	if err := json.Unmarshal(body, &s); err != nil {
		t.Fatalf("decode %s: %v", body, err)
	}
	return s
}

func (e *storageT33Env) memberID(t *testing.T, assoc, email string) string {
	t.Helper()
	var id string
	err := db.InAssociation(context.Background(), e.pool, uuid.MustParse(assoc), "test@example.test", func(ctx context.Context) error {
		tx, err := db.From(ctx)
		if err != nil {
			return err
		}
		return tx.QueryRow(ctx, `SELECT board_member_id::text FROM board_member_revisions WHERE association_id = $1 AND lower(sign_in_email) = $2`, assoc, email).Scan(&id)
	})
	if err != nil {
		t.Fatalf("board member of %s: %v", email, err)
	}
	return id
}

func TestStorageConnectionAnswersWhenAndByWhichBoardMember(t *testing.T) { // AC-27
	env := newStorageT33Env(t)
	before := time.Now().Add(-2 * time.Second)
	if status, body := env.complete(t, "member", env.a, "c1", env.start(t, "member", env.a)); status != 201 {
		t.Fatalf("complete: status %d, body %s", status, body)
	}
	after := time.Now().Add(2 * time.Second)
	wantMember := env.memberID(t, env.a, "member@example.test")

	for name, read := range map[string]func() []byte{
		"getStorageConnection": func() []byte {
			_, _, body := env.do(t, "other", "GET", storageT33Path(env.a), "")
			return body
		},
	} {
		got := storageT33Decode(t, read())
		if got.ConnectedAt == nil || got.ConnectedAt.Before(before) || got.ConnectedAt.After(after) {
			t.Errorf("%s: connected_at = %v, want the time the connection was stored (between %v and %v)", name, got.ConnectedAt, before, after)
		}
		if got.ConnectedBy == nil || *got.ConnectedBy != wantMember {
			t.Errorf("%s: connected_by_member_id = %v, want the board member id %s (not the account id %s)", name, got.ConnectedBy, wantMember, env.ids["member"])
		}
	}
}

func TestStorageConnectionNamesNoMemberOnceTheyLeftTheRoster(t *testing.T) { // AC-27
	env := newStorageT33Env(t)
	if status, body := env.complete(t, "member", env.a, "c1", env.start(t, "member", env.a)); status != 201 {
		t.Fatalf("complete: status %d, body %s", status, body)
	}
	err := db.InAssociation(context.Background(), env.pool, uuid.MustParse(env.a), "test@example.test", func(ctx context.Context) error {
		tx, err := db.From(ctx)
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `UPDATE board_member_revisions SET term_end = current_date - 1 WHERE association_id = $1 AND lower(sign_in_email) = 'member@example.test'`, env.a)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	_, _, body := env.do(t, "other", "GET", storageT33Path(env.a), "")
	got := storageT33Decode(t, body)
	if got.ConnectedBy != nil {
		t.Errorf("connected_by_member_id = %s, want null for a person no longer on the roster", *got.ConnectedBy)
	}
	if got.ConnectedAt == nil {
		t.Errorf("connected_at = null, want the stored time")
	}
}
