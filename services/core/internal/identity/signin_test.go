package identity_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/AndriiOpryshko/spilno/services/core/internal/identity"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/config"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/httpapi"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/testdb"
)

const signInOperatorEmail = "Operator@Example.Test"

// signInFake is the in-process sign-in provider: the authorization code names
// the person it resolves to. It implements identity.SignInProvider.
type signInFake struct {
	people map[string]identity.ProviderIdentity
}

func (f *signInFake) AuthCodeURL(state, nonce, codeChallenge string) string {
	q := url.Values{"state": {state}, "nonce": {nonce}, "code_challenge": {codeChallenge}}
	return "https://accounts.example.test/o/oauth2/auth?" + q.Encode()
}

func (f *signInFake) Exchange(_ context.Context, code, _ /*verifier*/, _ /*nonce*/ string) (identity.ProviderIdentity, error) {
	p, ok := f.people[code]
	if !ok {
		return identity.ProviderIdentity{}, errors.New("unknown code")
	}
	return p, nil
}

func newSignInServer(t *testing.T) (*httptest.Server, *pgxpool.Pool) {
	t.Helper()
	pool := testdb.New(t)
	m := identity.New(pool)
	m.EnableSignIn(identity.SignInConfig{
		Provider: &signInFake{people: map[string]identity.ProviderIdentity{
			"applicant":  {Issuer: "https://accounts.example.test", Subject: "sub-applicant", Email: "applicant@example.test", EmailVerified: true},
			"operator":   {Issuer: "https://accounts.example.test", Subject: "sub-operator", Email: "operator@example.test", EmailVerified: true},
			"unverified": {Issuer: "https://accounts.example.test", Subject: "sub-unverified", Email: "stranger@example.test", EmailVerified: false},
		}},
		OperatorEmail: signInOperatorEmail,
	})
	r := httpapi.NewRouter()
	if err := m.Routes(r); err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)
	return srv, pool
}

var signInNoRedirect = &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}

// signInAs runs start + callback for the fake person behind code and returns the callback response.
func signInAs(t *testing.T, srv *httptest.Server, code string) *http.Response {
	t.Helper()
	resp, err := signInNoRedirect.Get(srv.URL + "/identity/sign-in")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("startSignIn status = %d, want 302", resp.StatusCode)
	}
	loc, err := url.Parse(resp.Header.Get("Location"))
	if err != nil {
		t.Fatal(err)
	}
	state := loc.Query().Get("state")
	if state == "" {
		t.Fatalf("startSignIn Location %q carries no state", loc)
	}
	req, _ := http.NewRequest("GET", srv.URL+"/identity/sign-in/callback?code="+code+"&state="+url.QueryEscape(state), nil)
	for _, c := range resp.Cookies() { // a pre-login cookie, if the implementation uses one
		req.AddCookie(c)
	}
	cb, err := signInNoRedirect.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { cb.Body.Close() })
	return cb
}

func sessionCookieOf(resp *http.Response) *http.Cookie {
	for _, c := range resp.Cookies() {
		if c.Name == "spilno_session" && c.Value != "" {
			return c
		}
	}
	return nil
}

func count(t *testing.T, pool *pgxpool.Pool, table string) int {
	t.Helper()
	var n int
	if err := pool.QueryRow(context.Background(), "SELECT count(*) FROM "+table).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func getSessionWith(t *testing.T, srv *httptest.Server, c *http.Cookie) (*http.Response, map[string]any) {
	t.Helper()
	req, _ := http.NewRequest("GET", srv.URL+"/identity/session", nil)
	if c != nil {
		req.AddCookie(c)
	}
	resp, err := signInNoRedirect.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var body map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&body)
	return resp, body
}

// AC-05: a verified person on no roster gets an account and a session; sign-out revokes it.
func TestSignInGivesAVerifiedPersonAnAccountAndSession(t *testing.T) {
	srv, pool := newSignInServer(t)

	cb := signInAs(t, srv, "applicant")
	if cb.StatusCode != http.StatusFound {
		t.Fatalf("callback status = %d, want 302", cb.StatusCode)
	}
	c := sessionCookieOf(cb)
	if c == nil {
		t.Fatal("callback set no spilno_session cookie")
	}
	if !c.HttpOnly {
		t.Error("session cookie is not HttpOnly")
	}
	if n := count(t, pool, "accounts"); n != 1 {
		t.Errorf("accounts = %d, want 1", n)
	}
	if n := count(t, pool, "sessions"); n != 1 {
		t.Errorf("sessions = %d, want 1", n)
	}

	resp, body := getSessionWith(t, srv, c)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("getSession status = %d, want 200", resp.StatusCode)
	}
	if body["email"] != "applicant@example.test" || body["is_operator"] != false {
		t.Errorf("principal = %v, want applicant@example.test with is_operator false", body)
	}

	req, _ := http.NewRequest("POST", srv.URL+"/identity/sign-out", nil)
	req.AddCookie(c)
	out, err := signInNoRedirect.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	out.Body.Close()
	if out.StatusCode != http.StatusNoContent {
		t.Errorf("signOut status = %d, want 204", out.StatusCode)
	}
	if n := count(t, pool, "sessions"); n != 0 {
		t.Errorf("sessions after signOut = %d, want 0", n)
	}
	resp, body = getSessionWith(t, srv, c)
	if resp.StatusCode != http.StatusUnauthorized || body["code"] != "identity.session_required" {
		t.Errorf("getSession after signOut = %d %v, want 401 identity.session_required", resp.StatusCode, body["code"])
	}
}

// AC-16: is_operator is true only for SPILNO_OPERATOR_EMAIL, in any letter case.
func TestPrincipalIsOperatorOnlyForConfiguredAddress(t *testing.T) {
	srv, _ := newSignInServer(t)

	op := sessionCookieOf(signInAs(t, srv, "operator"))
	if op == nil {
		t.Fatal("operator got no session cookie")
	}
	_, body := getSessionWith(t, srv, op)
	if body["is_operator"] != true {
		t.Errorf("operator principal = %v, want is_operator true", body)
	}

	other := sessionCookieOf(signInAs(t, srv, "applicant"))
	_, body = getSessionWith(t, srv, other)
	if body["is_operator"] != false {
		t.Errorf("other principal = %v, want is_operator false", body)
	}
}

// AC-05: an address the provider does not mark verified gets no account and no session.
func TestUnverifiedAddressGetsNothing(t *testing.T) {
	srv, pool := newSignInServer(t)

	cb := signInAs(t, srv, "unverified")
	if cb.StatusCode == http.StatusFound && strings.HasPrefix(cb.Header.Get("Location"), "/") && sessionCookieOf(cb) != nil {
		t.Error("unverified address got a session cookie")
	}
	if sessionCookieOf(cb) != nil {
		t.Error("unverified address got a session cookie")
	}
	if cb.StatusCode < 400 {
		t.Errorf("callback status = %d, want a refusal", cb.StatusCode)
	}
	if n := count(t, pool, "accounts"); n != 0 {
		t.Errorf("accounts = %d, want 0", n)
	}
	if n := count(t, pool, "sessions"); n != 0 {
		t.Errorf("sessions = %d, want 0", n)
	}
}

// No cookie, or a cookie that names no session, is 401 identity.session_required.
func TestNoSessionIsRefused(t *testing.T) {
	srv, _ := newSignInServer(t)

	for name, c := range map[string]*http.Cookie{"no cookie": nil, "unknown token": {Name: "spilno_session", Value: "not-a-session"}} {
		resp, body := getSessionWith(t, srv, c)
		if resp.StatusCode != http.StatusUnauthorized || body["code"] != "identity.session_required" {
			t.Errorf("%s: getSession = %d %v, want 401 identity.session_required", name, resp.StatusCode, body["code"])
		}
	}
}

// The development sign-in shim is off in the production configuration; the
// operator address is read from SPILNO_OPERATOR_EMAIL.
func TestProductionConfigHasNoDevShim(t *testing.T) {
	t.Setenv("SPILNO_DEV_SIGN_IN", "")
	t.Setenv("SPILNO_OPERATOR_EMAIL", "operator@example.test")
	cfg := config.Load()
	if cfg.DevSignIn {
		t.Error("DevSignIn is on without being asked for")
	}
	if cfg.OperatorEmail != "operator@example.test" {
		t.Errorf("OperatorEmail = %q, want the SPILNO_OPERATOR_EMAIL value", cfg.OperatorEmail)
	}
}
