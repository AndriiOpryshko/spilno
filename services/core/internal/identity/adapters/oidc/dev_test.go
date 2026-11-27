package oidc

import (
	"net/url"
	"testing"
)

// A browser reaches the API through the web app's /api prefix, so the shim must send it back to the
// callback at the address the browser can open, not at the API's own path.
func TestDevAuthCodeURLReturnsToTheConfiguredCallback(t *testing.T) {
	d := Dev{Email: "operator@example.test", Callback: "http://localhost:5180/api/identity/sign-in/callback"}
	u, err := url.Parse(d.AuthCodeURL("state-1", "", ""))
	if err != nil {
		t.Fatal(err)
	}
	if got := u.Scheme + "://" + u.Host + u.Path; got != d.Callback {
		t.Fatalf("callback address = %q, want %q", got, d.Callback)
	}
	if u.Query().Get("code") != d.Email || u.Query().Get("state") != "state-1" {
		t.Fatalf("query = %q, want the address as the code and the state", u.RawQuery)
	}
}

func TestDevAuthCodeURLDefaultsToTheAPIPath(t *testing.T) {
	u, err := url.Parse(Dev{Email: "user@example.test"}.AuthCodeURL("s", "", ""))
	if err != nil {
		t.Fatal(err)
	}
	if u.Path != "/identity/sign-in/callback" || u.Host != "" {
		t.Fatalf("address = %q, want the API's own callback path", u.String())
	}
}
