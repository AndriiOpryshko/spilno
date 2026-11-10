package google

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"sync"
	"testing"

	"golang.org/x/oauth2"
)

const (
	exampleAccountEmail = "osbb@example.test"
	exampleRefreshToken = "1//refresh-token-example"
	exampleAccessToken  = "at-1"
	exampleRedirectURL  = "http://127.0.0.1:8085/callback"
	exampleAuthURL      = "https://accounts.example.test/auth"

	folderMimeType = "application/vnd.google-apps.folder"
	folderQuery    = "name = 'spilno' and mimeType = 'application/vnd.google-apps.folder' and trashed = false"

	// defaultTokenBody is what the token endpoint answers to a consent that granted everything.
	defaultTokenBody = `{"access_token":"at-1","refresh_token":"1//refresh-token-example","token_type":"Bearer","expires_in":3600,"scope":"https://www.googleapis.com/auth/drive.file https://www.googleapis.com/auth/documents"}`
)

// answer is what a fake endpoint replies with.
type answer struct {
	status int
	body   string
}

func ok(body string) answer { return answer{status: http.StatusOK, body: body} }

func (a answer) write(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(a.status)
	_, _ = w.Write([]byte(a.body))
}

// tokenEndpoint is a stand-in for Google's token endpoint: it records the form
// of every request and replies with its answer.
type tokenEndpoint struct {
	URL string

	mu    sync.Mutex
	forms []url.Values
}

func newTokenEndpoint(t *testing.T, reply answer) *tokenEndpoint {
	t.Helper()
	e := &tokenEndpoint{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		e.mu.Lock()
		e.forms = append(e.forms, r.PostForm)
		e.mu.Unlock()
		reply.write(w)
	}))
	t.Cleanup(srv.Close)
	e.URL = srv.URL
	return e
}

// lastForm returns the form of the latest request the endpoint received.
func (e *tokenEndpoint) lastForm(t *testing.T) url.Values {
	t.Helper()
	e.mu.Lock()
	defer e.mu.Unlock()
	if len(e.forms) == 0 {
		t.Fatal("the token endpoint received no request")
	}
	return e.forms[len(e.forms)-1]
}

// driveRequest is one request the Drive fake received.
type driveRequest struct {
	Method        string
	Path          string
	Query         url.Values
	Authorization string
	Body          []byte
}

// driveAnswers is what the Drive fake replies to each of the three calls.
type driveAnswers struct {
	about  answer // GET /drive/v3/about
	list   answer // GET /drive/v3/files
	create answer // POST /drive/v3/files
}

// defaultDriveAnswers is an account that has no spilno folder yet.
func defaultDriveAnswers() driveAnswers {
	return driveAnswers{
		about:  ok(`{"user":{"emailAddress":"osbb@example.test"}}`),
		list:   ok(`{"files":[]}`),
		create: ok(`{"id":"folder-1"}`),
	}
}

// driveFake is a stand-in for the Drive API served under /drive/v3: it records
// every request it receives.
type driveFake struct {
	URL string // the API base, ends with /drive/v3

	mu       sync.Mutex
	requests []driveRequest
}

func newDriveFake(t *testing.T, answers driveAnswers) *driveFake {
	t.Helper()
	d := &driveFake{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		d.mu.Lock()
		d.requests = append(d.requests, driveRequest{
			Method:        r.Method,
			Path:          r.URL.Path,
			Query:         r.URL.Query(),
			Authorization: r.Header.Get("Authorization"),
			Body:          body,
		})
		d.mu.Unlock()

		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/drive/v3/about":
			answers.about.write(w)
		case r.Method == http.MethodGet && r.URL.Path == "/drive/v3/files":
			answers.list.write(w)
		case r.Method == http.MethodPost && r.URL.Path == "/drive/v3/files":
			answers.create.write(w)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	d.URL = srv.URL + "/drive/v3"
	return d
}

// received returns the requests the fake got so far, in order.
func (d *driveFake) received() []driveRequest {
	d.mu.Lock()
	defer d.mu.Unlock()
	return append([]driveRequest(nil), d.requests...)
}

// oauthClient is the OAuth client client-1 / secret-1; its credentials travel in the form.
func oauthClient(tokenURL string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     "client-1",
		ClientSecret: "secret-1",
		RedirectURL:  exampleRedirectURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:   exampleAuthURL,
			TokenURL:  tokenURL,
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}
}

// newConsent is a consent wired to the two fakes.
func newConsent(token *tokenEndpoint, drive *driveFake) Consent {
	return Consent{OAuth: oauthClient(token.URL), DriveURL: drive.URL}
}

// bothScopes is what every grant holds.
func bothScopes() []string { return []string{ScopeDriveFile, ScopeDocuments} }

func assertZeroGrant(t *testing.T, got Grant) {
	t.Helper()
	if !reflect.DeepEqual(got, Grant{}) {
		t.Errorf("grant = %+v, want the zero grant", got)
	}
}

func assertNoDriveRequest(t *testing.T, drive *driveFake) {
	t.Helper()
	if got := drive.received(); len(got) != 0 {
		t.Errorf("Drive received %d request(s), want none; first: %s %s", len(got), got[0].Method, got[0].Path)
	}
}

// AC-1: Given that consent, When AuthURL("state-1") is called with OAuth.Scopes
// holding some other scope, Then the link starts with
// https://accounts.example.test/auth? and its query has response_type=code,
// client_id=client-1, redirect_uri=http://127.0.0.1:8085/callback,
// state=state-1, access_type=offline, prompt=consent and scope equal to the two
// scopes joined by one space; and OAuth.Scopes is unchanged.
func TestConsent_AC1_AuthURLAsksForOfflineConsentOnBothScopes(t *testing.T) {
	cfg := oauthClient("https://accounts.example.test/token")
	cfg.Scopes = []string{"openid"}
	c := Consent{OAuth: cfg, DriveURL: "https://drive.example.test/drive/v3"}

	link := c.AuthURL("state-1")

	if !strings.HasPrefix(link, exampleAuthURL+"?") {
		t.Errorf("AuthURL(%q) = %q, want it to start with %q", "state-1", link, exampleAuthURL+"?")
	}
	u, err := url.Parse(link)
	if err != nil {
		t.Fatalf("AuthURL(%q) = %q is not a URL: %v", "state-1", link, err)
	}
	query := u.Query()
	want := map[string]string{
		"response_type": "code",
		"client_id":     "client-1",
		"redirect_uri":  exampleRedirectURL,
		"state":         "state-1",
		"access_type":   "offline",
		"prompt":        "consent",
		"scope":         ScopeDriveFile + " " + ScopeDocuments,
	}
	for name, value := range want {
		if got := query[name]; len(got) != 1 || got[0] != value {
			t.Errorf("query %s = %q, want exactly %q", name, got, value)
		}
	}
	if !reflect.DeepEqual(cfg.Scopes, []string{"openid"}) {
		t.Errorf("OAuth.Scopes = %q after AuthURL, want it unchanged: %q", cfg.Scopes, []string{"openid"})
	}
}

// AC-2: Given the default answers, When Complete gets the code code-1, Then the
// grant is the email osbb@example.test, the refresh token
// 1//refresh-token-example, the two scopes and the folder folder-1; the token
// endpoint received grant_type=authorization_code, code=code-1,
// redirect_uri=http://127.0.0.1:8085/callback, client_id=client-1 and
// client_secret=secret-1; Drive received, in this order, GET /drive/v3/about
// with fields=user(emailAddress), GET /drive/v3/files with the q above and POST
// /drive/v3/files whose JSON body has name spilno and mimeType
// application/vnd.google-apps.folder; and each of the three carried
// Authorization: Bearer at-1.
func TestConsent_AC2_CompleteExchangesCodeReadsEmailAndCreatesFolder(t *testing.T) {
	token := newTokenEndpoint(t, ok(defaultTokenBody))
	drive := newDriveFake(t, defaultDriveAnswers())

	got, err := newConsent(token, drive).Complete(t.Context(), "code-1")
	if err != nil {
		t.Fatalf("Complete(%q): unexpected error %v", "code-1", err)
	}

	want := Grant{
		AccountEmail:      exampleAccountEmail,
		RefreshToken:      exampleRefreshToken,
		Scopes:            bothScopes(),
		DriveRootFolderID: "folder-1",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("grant = %+v, want %+v", got, want)
	}

	form := token.lastForm(t)
	for name, value := range map[string]string{
		"grant_type":    "authorization_code",
		"code":          "code-1",
		"redirect_uri":  exampleRedirectURL,
		"client_id":     "client-1",
		"client_secret": "secret-1",
	} {
		if got := form.Get(name); got != value {
			t.Errorf("token endpoint form %s = %q, want %q", name, got, value)
		}
	}

	requests := drive.received()
	if len(requests) != 3 {
		t.Fatalf("Drive received %d request(s), want 3: about, list, create", len(requests))
	}
	about, list, create := requests[0], requests[1], requests[2]

	if about.Method != http.MethodGet || about.Path != "/drive/v3/about" {
		t.Errorf("first Drive request = %s %s, want GET /drive/v3/about", about.Method, about.Path)
	}
	if got := about.Query.Get("fields"); got != "user(emailAddress)" {
		t.Errorf("about fields = %q, want %q", got, "user(emailAddress)")
	}

	if list.Method != http.MethodGet || list.Path != "/drive/v3/files" {
		t.Errorf("second Drive request = %s %s, want GET /drive/v3/files", list.Method, list.Path)
	}
	if got := list.Query.Get("q"); got != folderQuery {
		t.Errorf("list q = %q, want %q", got, folderQuery)
	}

	if create.Method != http.MethodPost || create.Path != "/drive/v3/files" {
		t.Errorf("third Drive request = %s %s, want POST /drive/v3/files", create.Method, create.Path)
	}
	var folder struct {
		Name     string `json:"name"`
		MimeType string `json:"mimeType"`
	}
	if err := json.Unmarshal(create.Body, &folder); err != nil {
		t.Errorf("create body %q is not JSON: %v", create.Body, err)
	}
	if folder.Name != RootFolderName {
		t.Errorf("create body name = %q, want %q", folder.Name, RootFolderName)
	}
	if folder.MimeType != folderMimeType {
		t.Errorf("create body mimeType = %q, want %q", folder.MimeType, folderMimeType)
	}

	for i, r := range requests {
		if r.Authorization != "Bearer "+exampleAccessToken {
			t.Errorf("Drive request %d (%s %s) Authorization = %q, want %q",
				i+1, r.Method, r.Path, r.Authorization, "Bearer "+exampleAccessToken)
		}
	}
}

// AC-3: Given Drive lists {"files":[{"id":"folder-0"},{"id":"folder-9"}]}, When
// Complete runs, Then the grant's folder is folder-0 and Drive received no POST.
func TestConsent_AC3_CompleteReusesTheFirstListedFolder(t *testing.T) {
	answers := defaultDriveAnswers()
	answers.list = ok(`{"files":[{"id":"folder-0"},{"id":"folder-9"}]}`)
	token := newTokenEndpoint(t, ok(defaultTokenBody))
	drive := newDriveFake(t, answers)

	got, err := newConsent(token, drive).Complete(t.Context(), "code-1")
	if err != nil {
		t.Fatalf("Complete(%q): unexpected error %v", "code-1", err)
	}

	if got.DriveRootFolderID != "folder-0" {
		t.Errorf("grant folder = %q, want %q", got.DriveRootFolderID, "folder-0")
	}
	for _, r := range drive.received() {
		if r.Method == http.MethodPost {
			t.Errorf("Drive received %s %s, want no POST when a folder is listed", r.Method, r.Path)
		}
	}
}

// AC-4: Given the token endpoint answers without refresh_token, Then Complete
// returns ErrNoRefreshToken, the zero grant, and Drive received no request.
func TestConsent_AC4_CompleteRejectsAnswerWithoutRefreshToken(t *testing.T) {
	token := newTokenEndpoint(t, ok(`{"access_token":"at-1","token_type":"Bearer","expires_in":3600,"scope":"https://www.googleapis.com/auth/drive.file https://www.googleapis.com/auth/documents"}`))
	drive := newDriveFake(t, defaultDriveAnswers())

	got, err := newConsent(token, drive).Complete(t.Context(), "code-1")

	if !errors.Is(err, ErrNoRefreshToken) {
		t.Errorf("Complete(%q) error = %v, want ErrNoRefreshToken", "code-1", err)
	}
	assertZeroGrant(t, got)
	assertNoDriveRequest(t, drive)
}

// AC-5: Given the token endpoint answers with scope equal to
// https://www.googleapis.com/auth/drive.file only, Then Complete returns
// ErrScopeNotGranted and Drive received no request; and given it answers with
// no scope field at all, Then Complete succeeds with both scopes in the grant.
func TestConsent_AC5_CompleteChecksTheGrantedScopes(t *testing.T) {
	t.Run("one scope unticked", func(t *testing.T) {
		token := newTokenEndpoint(t, ok(`{"access_token":"at-1","refresh_token":"1//refresh-token-example","token_type":"Bearer","expires_in":3600,"scope":"https://www.googleapis.com/auth/drive.file"}`))
		drive := newDriveFake(t, defaultDriveAnswers())

		_, err := newConsent(token, drive).Complete(t.Context(), "code-1")

		if !errors.Is(err, ErrScopeNotGranted) {
			t.Errorf("Complete(%q) error = %v, want ErrScopeNotGranted", "code-1", err)
		}
		assertNoDriveRequest(t, drive)
	})

	t.Run("no scope field", func(t *testing.T) {
		token := newTokenEndpoint(t, ok(`{"access_token":"at-1","refresh_token":"1//refresh-token-example","token_type":"Bearer","expires_in":3600}`))
		drive := newDriveFake(t, defaultDriveAnswers())

		got, err := newConsent(token, drive).Complete(t.Context(), "code-1")
		if err != nil {
			t.Fatalf("Complete(%q): unexpected error %v", "code-1", err)
		}

		if !reflect.DeepEqual(got.Scopes, bothScopes()) {
			t.Errorf("grant scopes = %q, want %q", got.Scopes, bothScopes())
		}
	})
}

// AC-6: Given the token endpoint answers 400 with {"error":"invalid_grant"},
// Then Complete returns ErrCodeExchangeFailed, the zero grant, and Drive
// received no request.
func TestConsent_AC6_CompleteReportsFailedCodeExchange(t *testing.T) {
	token := newTokenEndpoint(t, answer{status: http.StatusBadRequest, body: `{"error":"invalid_grant"}`})
	drive := newDriveFake(t, defaultDriveAnswers())

	got, err := newConsent(token, drive).Complete(t.Context(), "code-1")

	if !errors.Is(err, ErrCodeExchangeFailed) {
		t.Errorf("Complete(%q) error = %v, want ErrCodeExchangeFailed", "code-1", err)
	}
	assertZeroGrant(t, got)
	assertNoDriveRequest(t, drive)
}

// AC-7: Given the about call answers 503, or it answers {"user":{}}, or the
// create call answers 403, Then Complete returns ErrDriveUnavailable and the
// zero grant each time, and the error's text contains neither
// 1//refresh-token-example nor at-1.
func TestConsent_AC7_CompleteReportsDriveFailureWithoutTokens(t *testing.T) {
	cases := []struct {
		name   string
		change func(*driveAnswers)
	}{
		{
			name: "about answers 503",
			change: func(a *driveAnswers) {
				a.about = answer{status: http.StatusServiceUnavailable, body: `{"error":{"code":503,"message":"backend error"}}`}
			},
		},
		{
			name:   "about answers no email",
			change: func(a *driveAnswers) { a.about = ok(`{"user":{}}`) },
		},
		{
			name: "create answers 403",
			change: func(a *driveAnswers) {
				a.create = answer{status: http.StatusForbidden, body: `{"error":{"code":403,"message":"forbidden"}}`}
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			answers := defaultDriveAnswers()
			tc.change(&answers)
			token := newTokenEndpoint(t, ok(defaultTokenBody))
			drive := newDriveFake(t, answers)

			got, err := newConsent(token, drive).Complete(t.Context(), "code-1")

			if !errors.Is(err, ErrDriveUnavailable) {
				t.Fatalf("Complete(%q) error = %v, want ErrDriveUnavailable", "code-1", err)
			}
			assertZeroGrant(t, got)
			for _, secret := range []string{exampleRefreshToken, exampleAccessToken} {
				if strings.Contains(err.Error(), secret) {
					t.Errorf("error text %q contains the token %q", err.Error(), secret)
				}
			}
		})
	}
}
