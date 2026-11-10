package identity_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/oauth2"

	"github.com/AndriiOpryshko/spilno/services/core/internal/identity"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/db"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/testdb"
)

const (
	exampleAccountEmail = "osbb@example.test"
	exampleRefreshToken = "1//refresh-token-example"
	secondRefreshToken  = "1//second-token"
	scopeDriveFile      = "https://www.googleapis.com/auth/drive.file"
	scopeDocuments      = "https://www.googleapis.com/auth/documents"
)

// testAssociation is the association the migrations create.
var testAssociation = uuid.MustParse("00000000-0000-7000-8000-000000000001")

// googleTestKey returns a 32-byte key; different seeds give different keys.
func googleTestKey(seed byte) []byte {
	key := make([]byte, 32)
	for i := range key {
		key[i] = seed + byte(i)
	}
	return key
}

// tokenEndpoint is a stand-in for Google's token endpoint: it records the form
// of every request and answers with the access token at-1.
type tokenEndpoint struct {
	URL string

	mu    sync.Mutex
	forms []url.Values
}

func newTokenEndpoint(t *testing.T) *tokenEndpoint {
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
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"at-1","token_type":"Bearer","expires_in":3600}`))
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

// oauthClient is the OAuth client client-1 / secret-1; its credentials travel in the form.
func oauthClient(tokenURL string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     "client-1",
		ClientSecret: "secret-1",
		Endpoint:     oauth2.Endpoint{TokenURL: tokenURL, AuthStyle: oauth2.AuthStyleInParams},
	}
}

// enabledModule returns a module on pool with the Google connection turned on.
func enabledModule(t *testing.T, pool *pgxpool.Pool, keyID string, key []byte, tokenURL string) *identity.Module {
	t.Helper()
	m := identity.New(pool)
	if err := m.EnableGoogle(identity.GoogleConfig{KeyID: keyID, Key: key, OAuth: oauthClient(tokenURL)}); err != nil {
		t.Fatalf("EnableGoogle(key id %q, %d-byte key): unexpected error %v", keyID, len(key), err)
	}
	return m
}

// insertOperator stores the account of whoever gives the consent and returns its id.
func insertOperator(t *testing.T, pool *pgxpool.Pool) uuid.UUID {
	t.Helper()
	id, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	_, err = pool.Exec(t.Context(),
		`INSERT INTO accounts (id, oidc_issuer, oidc_subject, email)
		 VALUES ($1, $2, $3, $4)`,
		id.String(), "https://issuer.example.test", "operator-subject", "operator@example.test")
	if err != nil {
		t.Fatalf("insert the operator account: %v", err)
	}
	return id
}

// exampleConnection is the consent of osbb@example.test given by operator.
func exampleConnection(operator uuid.UUID) identity.GoogleConnection {
	return identity.GoogleConnection{
		AssociationID: testAssociation,
		AccountEmail:  exampleAccountEmail,
		RefreshToken:  exampleRefreshToken,
		Scopes:        []string{scopeDriveFile, scopeDocuments},
		AuthorizedBy:  operator,
	}
}

// connectionRow is one google_connections row as it is stored.
type connectionRow struct {
	ID                string
	AccountEmail      string
	Ciphertext        []byte
	Nonce             []byte
	KeyID             string
	Scopes            string
	DriveRootFolderID *string
	AuthorizedBy      string
}

// connectionRows reads the association's google_connections rows with plain SQL.
func connectionRows(t *testing.T, pool *pgxpool.Pool) []connectionRow {
	t.Helper()
	var out []connectionRow
	// google_connections is under row-level security: the app-role pool sees rows only inside
	// the association's own transaction.
	err := db.InAssociation(t.Context(), pool, testAssociation, "", func(ctx context.Context) error {
		q, err := db.From(ctx)
		if err != nil {
			return err
		}
		rows, err := q.Query(ctx,
			`SELECT id::text, google_account_email, refresh_token_ciphertext, refresh_token_nonce,
			        encryption_key_id, scopes, drive_root_folder_id, authorized_by_account_id::text
			   FROM google_connections
			  WHERE association_id = $1`,
			testAssociation.String())
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var r connectionRow
			if err := rows.Scan(&r.ID, &r.AccountEmail, &r.Ciphertext, &r.Nonce,
				&r.KeyID, &r.Scopes, &r.DriveRootFolderID, &r.AuthorizedBy); err != nil {
				return err
			}
			out = append(out, r)
		}
		return rows.Err()
	})
	if err != nil {
		t.Fatalf("read google_connections: %v", err)
	}
	return out
}

// onlyConnectionRow fails the test unless the association has exactly one row.
func onlyConnectionRow(t *testing.T, pool *pgxpool.Pool) connectionRow {
	t.Helper()
	rows := connectionRows(t, pool)
	if len(rows) != 1 {
		t.Fatalf("the association has %d google_connections rows, want exactly 1", len(rows))
	}
	return rows[0]
}

// refreshedToken asks m for the association's token source and calls Token() on it.
func refreshedToken(t *testing.T, m *identity.Module) *oauth2.Token {
	t.Helper()
	ts, err := m.GoogleTokenSource(t.Context(), testAssociation)
	if err != nil {
		t.Fatalf("GoogleTokenSource: unexpected error %v", err)
	}
	if ts == nil {
		t.Fatal("GoogleTokenSource = nil token source with nil error")
	}
	tok, err := ts.Token()
	if err != nil {
		t.Fatalf("Token(): unexpected error %v", err)
	}
	if tok == nil {
		t.Fatal("Token() = nil token with nil error")
	}
	return tok
}

// AC-1: Given a module enabled with key id `k1` and a 32-byte key, When `ConnectGoogle` stores the account
// `osbb@example.test` with the refresh token `1//refresh-token-example` and the scopes
// `https://www.googleapis.com/auth/drive.file` and `https://www.googleapis.com/auth/documents`, Then the
// association has exactly one `google_connections` row with that email, the two scopes joined by one space,
// `encryption_key_id` `k1`, a 12-byte nonce, a ciphertext that does not contain the token's bytes, a `NULL`
// `drive_root_folder_id` and the operator's account in `authorized_by_account_id`.
func TestGoogleConnection_AC1_ConnectStoresTheSealedToken(t *testing.T) {
	pool := testdb.NewApp(t)
	operator := insertOperator(t, pool)
	m := enabledModule(t, pool, "k1", googleTestKey(1), newTokenEndpoint(t).URL)

	if err := m.ConnectGoogle(t.Context(), exampleConnection(operator)); err != nil {
		t.Fatalf("ConnectGoogle: unexpected error %v", err)
	}

	row := onlyConnectionRow(t, pool)
	if row.AccountEmail != exampleAccountEmail {
		t.Errorf("google_account_email = %q, want %q", row.AccountEmail, exampleAccountEmail)
	}
	if want := scopeDriveFile + " " + scopeDocuments; row.Scopes != want {
		t.Errorf("scopes = %q, want %q", row.Scopes, want)
	}
	if row.KeyID != "k1" {
		t.Errorf("encryption_key_id = %q, want %q", row.KeyID, "k1")
	}
	if len(row.Nonce) != 12 {
		t.Errorf("refresh_token_nonce is %d bytes, want 12", len(row.Nonce))
	}
	if len(row.Ciphertext) == 0 {
		t.Error("refresh_token_ciphertext is empty")
	}
	if bytes.Contains(row.Ciphertext, []byte(exampleRefreshToken)) {
		t.Error("refresh_token_ciphertext contains the refresh token in plain text")
	}
	if row.DriveRootFolderID != nil {
		t.Errorf("drive_root_folder_id = %q, want NULL", *row.DriveRootFolderID)
	}
	if row.AuthorizedBy != operator.String() {
		t.Errorf("authorized_by_account_id = %s, want the operator's account %s", row.AuthorizedBy, operator)
	}
}

// AC-2: Given that connection, When `GoogleTokenSource` is asked for the association and `Token()` is called,
// Then the token endpoint received `grant_type=refresh_token`, `refresh_token=1//refresh-token-example`,
// `client_id=client-1` and `client_secret=secret-1`, and the returned access token is `at-1`.
func TestGoogleConnection_AC2_TokenSourceExchangesTheStoredRefreshToken(t *testing.T) {
	pool := testdb.NewApp(t)
	operator := insertOperator(t, pool)
	endpoint := newTokenEndpoint(t)
	m := enabledModule(t, pool, "k1", googleTestKey(1), endpoint.URL)
	if err := m.ConnectGoogle(t.Context(), exampleConnection(operator)); err != nil {
		t.Fatalf("ConnectGoogle: unexpected error %v", err)
	}

	tok := refreshedToken(t, m)

	form := endpoint.lastForm(t)
	for _, want := range []struct{ field, value string }{
		{"grant_type", "refresh_token"},
		{"refresh_token", exampleRefreshToken},
		{"client_id", "client-1"},
		{"client_secret", "secret-1"},
	} {
		if got := form.Get(want.field); got != want.value {
			t.Errorf("the token endpoint received %s=%q, want %q", want.field, got, want.value)
		}
	}
	if tok.AccessToken != "at-1" {
		t.Errorf("access token = %q, want %q", tok.AccessToken, "at-1")
	}
}

// AC-3: Given that connection, When `ConnectGoogle` is called again with the refresh token `1//second-token`
// and the folder id `folder-1`, Then the association still has exactly one row, its `id` differs from the first
// one, `drive_root_folder_id` is `folder-1`, and a new token source sends `refresh_token=1//second-token`.
func TestGoogleConnection_AC3_ConnectAgainReplacesTheRow(t *testing.T) {
	pool := testdb.NewApp(t)
	operator := insertOperator(t, pool)
	endpoint := newTokenEndpoint(t)
	m := enabledModule(t, pool, "k1", googleTestKey(1), endpoint.URL)
	if err := m.ConnectGoogle(t.Context(), exampleConnection(operator)); err != nil {
		t.Fatalf("ConnectGoogle: unexpected error %v", err)
	}
	first := onlyConnectionRow(t, pool)

	again := exampleConnection(operator)
	again.RefreshToken = secondRefreshToken
	again.DriveRootFolderID = "folder-1"
	if err := m.ConnectGoogle(t.Context(), again); err != nil {
		t.Fatalf("ConnectGoogle again: unexpected error %v", err)
	}

	second := onlyConnectionRow(t, pool)
	if second.ID == first.ID {
		t.Errorf("id is still %s, want a new id for the replaced row", second.ID)
	}
	if second.DriveRootFolderID == nil {
		t.Error("drive_root_folder_id = NULL, want \"folder-1\"")
	} else if *second.DriveRootFolderID != "folder-1" {
		t.Errorf("drive_root_folder_id = %q, want %q", *second.DriveRootFolderID, "folder-1")
	}

	refreshedToken(t, m)
	if got := endpoint.lastForm(t).Get("refresh_token"); got != secondRefreshToken {
		t.Errorf("the token endpoint received refresh_token=%q, want %q", got, secondRefreshToken)
	}
}

// AC-4: Given an enabled module and an association with no connection, Then `GoogleTokenSource` returns
// `ErrGoogleNotConnected`.
func TestGoogleConnection_AC4_TokenSourceWithoutConnection(t *testing.T) {
	pool := testdb.NewApp(t)
	m := enabledModule(t, pool, "k1", googleTestKey(1), newTokenEndpoint(t).URL)

	_, err := m.GoogleTokenSource(t.Context(), testAssociation)

	if !errors.Is(err, identity.ErrGoogleNotConnected) {
		t.Fatalf("GoogleTokenSource error = %v, want ErrGoogleNotConnected", err)
	}
}

// AC-5: Given a connection stored under key id `k1`, When a second module on the same database is enabled
// with key id `k1` and a different key, or with key id `k2` and the same key, Then `GoogleTokenSource` returns
// `ErrGoogleTokenUnreadable` in both cases.
func TestGoogleConnection_AC5_TokenSealedUnderAnotherKeyIsUnreadable(t *testing.T) {
	pool := testdb.NewApp(t)
	operator := insertOperator(t, pool)
	endpoint := newTokenEndpoint(t)
	first := enabledModule(t, pool, "k1", googleTestKey(1), endpoint.URL)
	if err := first.ConnectGoogle(t.Context(), exampleConnection(operator)); err != nil {
		t.Fatalf("ConnectGoogle: unexpected error %v", err)
	}

	cases := []struct {
		name  string
		keyID string
		key   []byte
	}{
		{"key id k1 and a different key", "k1", googleTestKey(2)},
		{"key id k2 and the same key", "k2", googleTestKey(1)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			second := enabledModule(t, pool, tc.keyID, tc.key, endpoint.URL)

			_, err := second.GoogleTokenSource(t.Context(), testAssociation)

			if !errors.Is(err, identity.ErrGoogleTokenUnreadable) {
				t.Fatalf("GoogleTokenSource error = %v, want ErrGoogleTokenUnreadable", err)
			}
		})
	}
}

// AC-6: Given a module without `EnableGoogle`, Then `ConnectGoogle` and `GoogleTokenSource` return
// `ErrGoogleNotConfigured`; and `EnableGoogle` with a 16-byte key, an empty key id, a nil `OAuth`, an empty
// `ClientID` or an empty `Endpoint.TokenURL` returns `ErrGoogleConfigInvalid` and leaves the module not
// configured.
func TestGoogleConnection_AC6_NotConfiguredAndInvalidConfig(t *testing.T) {
	pool := testdb.NewApp(t)
	operator := insertOperator(t, pool)
	tokenURL := newTokenEndpoint(t).URL

	// assertNotConfigured checks both methods of m refuse to work.
	assertNotConfigured := func(t *testing.T, m *identity.Module) {
		t.Helper()
		if err := m.ConnectGoogle(t.Context(), exampleConnection(operator)); !errors.Is(err, identity.ErrGoogleNotConfigured) {
			t.Errorf("ConnectGoogle error = %v, want ErrGoogleNotConfigured", err)
		}
		if _, err := m.GoogleTokenSource(t.Context(), testAssociation); !errors.Is(err, identity.ErrGoogleNotConfigured) {
			t.Errorf("GoogleTokenSource error = %v, want ErrGoogleNotConfigured", err)
		}
	}

	t.Run("without EnableGoogle", func(t *testing.T) {
		assertNotConfigured(t, identity.New(pool))
	})

	withoutClientID := oauthClient(tokenURL)
	withoutClientID.ClientID = ""
	invalid := []struct {
		name string
		cfg  identity.GoogleConfig
	}{
		{"a 16-byte key", identity.GoogleConfig{KeyID: "k1", Key: googleTestKey(1)[:16], OAuth: oauthClient(tokenURL)}},
		{"an empty key id", identity.GoogleConfig{KeyID: "", Key: googleTestKey(1), OAuth: oauthClient(tokenURL)}},
		{"a nil OAuth", identity.GoogleConfig{KeyID: "k1", Key: googleTestKey(1), OAuth: nil}},
		{"an empty ClientID", identity.GoogleConfig{KeyID: "k1", Key: googleTestKey(1), OAuth: withoutClientID}},
		{"an empty Endpoint.TokenURL", identity.GoogleConfig{KeyID: "k1", Key: googleTestKey(1), OAuth: oauthClient("")}},
	}
	for _, tc := range invalid {
		t.Run("EnableGoogle with "+tc.name, func(t *testing.T) {
			m := identity.New(pool)

			if err := m.EnableGoogle(tc.cfg); !errors.Is(err, identity.ErrGoogleConfigInvalid) {
				t.Errorf("EnableGoogle error = %v, want ErrGoogleConfigInvalid", err)
			}

			assertNotConfigured(t, m)
		})
	}
}

// AC-7: Given an enabled module, When `ConnectGoogle` gets an empty refresh token, an empty account email or
// no scopes, Then it returns `ErrGoogleConnectionInvalid` each time and the association has no row.
func TestGoogleConnection_AC7_InvalidConnectionWritesNothing(t *testing.T) {
	pool := testdb.NewApp(t)
	operator := insertOperator(t, pool)
	m := enabledModule(t, pool, "k1", googleTestKey(1), newTokenEndpoint(t).URL)

	cases := []struct {
		name   string
		change func(c *identity.GoogleConnection)
	}{
		{"an empty refresh token", func(c *identity.GoogleConnection) { c.RefreshToken = "" }},
		{"an empty account email", func(c *identity.GoogleConnection) { c.AccountEmail = "" }},
		{"no scopes", func(c *identity.GoogleConnection) { c.Scopes = nil }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := exampleConnection(operator)
			tc.change(&c)

			err := m.ConnectGoogle(t.Context(), c)

			if !errors.Is(err, identity.ErrGoogleConnectionInvalid) {
				t.Errorf("ConnectGoogle error = %v, want ErrGoogleConnectionInvalid", err)
			}
			if rows := connectionRows(t, pool); len(rows) != 0 {
				t.Errorf("the association has %d google_connections rows, want none", len(rows))
			}
		})
	}
}
