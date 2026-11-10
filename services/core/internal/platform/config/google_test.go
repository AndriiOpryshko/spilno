package config

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

const (
	envGoogleClientID     = "GOOGLE_OAUTH_CLIENT_ID"
	envGoogleClientSecret = "GOOGLE_OAUTH_CLIENT_SECRET"
	envGoogleTokenKeyID   = "GOOGLE_TOKEN_KEY_ID"
	envGoogleTokenKey     = "GOOGLE_TOKEN_KEY"

	testGoogleClientID     = "client-1"
	testGoogleClientSecret = "secret-1"
	testGoogleTokenKeyID   = "k1"
	// testGoogleTokenKey is base64.StdEncoding of the 32 bytes 0x00 to 0x1f.
	testGoogleTokenKey = "AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8="
)

// googleEnv returns the four variables of AC-2.
func googleEnv() map[string]string {
	return map[string]string{
		envGoogleClientID:     testGoogleClientID,
		envGoogleClientSecret: testGoogleClientSecret,
		envGoogleTokenKeyID:   testGoogleTokenKeyID,
		envGoogleTokenKey:     testGoogleTokenKey,
	}
}

// setGoogleEnv sets all four variables, so the developer's environment does not leak into the test.
func setGoogleEnv(t *testing.T, vars map[string]string) {
	t.Helper()
	for _, name := range []string{envGoogleClientID, envGoogleClientSecret, envGoogleTokenKeyID, envGoogleTokenKey} {
		t.Setenv(name, vars[name])
	}
}

// testGoogleKeyBytes returns the 32 bytes 0x00 to 0x1f in order.
func testGoogleKeyBytes() []byte {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	return key
}

// AC-1: Given none of the four variables is set, Then `LoadGoogle` returns a nil configuration and a nil error.
func TestLoadGoogle_AC1_NoneSetReturnsNilConfigAndNilError(t *testing.T) {
	setGoogleEnv(t, map[string]string{})

	got, err := LoadGoogle()

	if err != nil {
		t.Errorf("LoadGoogle() error = %v, want nil", err)
	}
	if got != nil {
		t.Errorf("LoadGoogle() = %+v, want a nil configuration", got)
	}
}

// AC-2: Given `GOOGLE_OAUTH_CLIENT_ID=client-1`, `GOOGLE_OAUTH_CLIENT_SECRET=secret-1`,
// `GOOGLE_TOKEN_KEY_ID=k1` and `GOOGLE_TOKEN_KEY=AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8=`, Then `LoadGoogle`
// returns those three strings and a 32-byte key whose bytes are `0x00` to `0x1f` in order.
func TestLoadGoogle_AC2_AllSetReturnsStringsAndDecodedKey(t *testing.T) {
	setGoogleEnv(t, googleEnv())

	got, err := LoadGoogle()

	if err != nil {
		t.Fatalf("LoadGoogle() error = %v, want nil", err)
	}
	if got == nil {
		t.Fatal("LoadGoogle() = nil, want a configuration")
	}
	if got.ClientID != testGoogleClientID {
		t.Errorf("ClientID = %q, want %q", got.ClientID, testGoogleClientID)
	}
	if got.ClientSecret != testGoogleClientSecret {
		t.Errorf("ClientSecret = %q, want %q", got.ClientSecret, testGoogleClientSecret)
	}
	if got.TokenKeyID != testGoogleTokenKeyID {
		t.Errorf("TokenKeyID = %q, want %q", got.TokenKeyID, testGoogleTokenKeyID)
	}
	if len(got.TokenKey) != 32 {
		t.Errorf("len(TokenKey) = %d, want 32", len(got.TokenKey))
	}
	if !bytes.Equal(got.TokenKey, testGoogleKeyBytes()) {
		t.Errorf("TokenKey is not the bytes 0x00 to 0x1f in order")
	}
}

// AC-3: Given the variables of AC-2 with exactly one of them empty (four cases), Then `LoadGoogle` returns
// `ErrGoogleIncomplete` and a nil configuration each time; the error's text contains the name of the empty
// variable and contains neither `secret-1` nor the key string.
func TestLoadGoogle_AC3_OneEmptyReturnsErrGoogleIncomplete(t *testing.T) {
	for _, empty := range []string{envGoogleClientID, envGoogleClientSecret, envGoogleTokenKeyID, envGoogleTokenKey} {
		t.Run(empty, func(t *testing.T) {
			vars := googleEnv()
			vars[empty] = ""
			setGoogleEnv(t, vars)

			got, err := LoadGoogle()

			if got != nil {
				t.Errorf("LoadGoogle() returned a configuration, want nil")
			}
			if !errors.Is(err, ErrGoogleIncomplete) {
				t.Fatalf("LoadGoogle() error = %v, want ErrGoogleIncomplete", err)
			}
			text := err.Error()
			if !strings.Contains(text, empty) {
				t.Errorf("error text %q does not name the empty variable %s", text, empty)
			}
			if strings.Contains(text, testGoogleClientSecret) {
				t.Errorf("error text contains the client secret")
			}
			if strings.Contains(text, testGoogleTokenKey) {
				t.Errorf("error text contains the key string")
			}
		})
	}
}

// AC-4: Given the variables of AC-2 with `GOOGLE_TOKEN_KEY` set to `not-base64!`, to
// `AAECAwQFBgcICQoLDA0ODw==` (16 bytes) or to `AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8g` (33 bytes), Then
// `LoadGoogle` returns `ErrGoogleTokenKey` and a nil configuration each time, and the error's text does not
// contain the value.
func TestLoadGoogle_AC4_BadTokenKeyReturnsErrGoogleTokenKey(t *testing.T) {
	cases := []struct {
		name  string
		value string
	}{
		{name: "not base64", value: "not-base64!"},
		{name: "16 bytes", value: "AAECAwQFBgcICQoLDA0ODw=="},
		{name: "33 bytes", value: "AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8g"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			vars := googleEnv()
			vars[envGoogleTokenKey] = tc.value
			setGoogleEnv(t, vars)

			got, err := LoadGoogle()

			if got != nil {
				t.Errorf("LoadGoogle() returned a configuration, want nil")
			}
			if !errors.Is(err, ErrGoogleTokenKey) {
				t.Fatalf("LoadGoogle() error = %v, want ErrGoogleTokenKey", err)
			}
			if strings.Contains(err.Error(), tc.value) {
				t.Errorf("error text contains the GOOGLE_TOKEN_KEY value")
			}
		})
	}
}

// AC-5: Given the configuration of AC-2, When `OAuth("http://127.0.0.1:8085/callback")` is called, Then the
// client has `ClientID` `client-1`, `ClientSecret` `secret-1`, that redirect URL, the auth URL
// `https://accounts.google.com/o/oauth2/auth` and the token URL `https://oauth2.googleapis.com/token`; and
// `OAuth("")` has an empty redirect URL.
func TestGoogleOAuth_AC5_ClientAgainstGoogleEndpoints(t *testing.T) {
	const redirectURL = "http://127.0.0.1:8085/callback"
	cfg := Google{
		ClientID:     testGoogleClientID,
		ClientSecret: testGoogleClientSecret,
		TokenKeyID:   testGoogleTokenKeyID,
		TokenKey:     testGoogleKeyBytes(),
	}

	got := cfg.OAuth(redirectURL)

	if got == nil {
		t.Fatalf("OAuth(%q) = nil, want a client", redirectURL)
	}
	if got.ClientID != testGoogleClientID {
		t.Errorf("ClientID = %q, want %q", got.ClientID, testGoogleClientID)
	}
	if got.ClientSecret != testGoogleClientSecret {
		t.Errorf("ClientSecret = %q, want %q", got.ClientSecret, testGoogleClientSecret)
	}
	if got.RedirectURL != redirectURL {
		t.Errorf("RedirectURL = %q, want %q", got.RedirectURL, redirectURL)
	}
	if want := "https://accounts.google.com/o/oauth2/auth"; got.Endpoint.AuthURL != want {
		t.Errorf("Endpoint.AuthURL = %q, want %q", got.Endpoint.AuthURL, want)
	}
	if want := "https://oauth2.googleapis.com/token"; got.Endpoint.TokenURL != want {
		t.Errorf("Endpoint.TokenURL = %q, want %q", got.Endpoint.TokenURL, want)
	}

	refreshOnly := cfg.OAuth("")

	if refreshOnly == nil {
		t.Fatal(`OAuth("") = nil, want a client`)
	}
	if refreshOnly.RedirectURL != "" {
		t.Errorf(`OAuth("").RedirectURL = %q, want empty`, refreshOnly.RedirectURL)
	}
}
