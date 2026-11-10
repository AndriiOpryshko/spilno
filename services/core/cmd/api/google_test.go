package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"golang.org/x/oauth2"

	"github.com/AndriiOpryshko/spilno/services/core/internal/governance"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/httpapi"
)

const wireGoogleT33Path = "/identity/associations/0198a0a0-0000-7000-8000-000000000001/storage-connection"

// wireGoogle decides the Google wiring: nil leaves EnableGoogle and EnableStorageConnection off
// (the operations answer 501), a configuration turns both on.
func wireGoogleT33Status(t *testing.T, gc *identity.GoogleConfig) (wired bool, status int) {
	t.Helper()
	idm := identity.New(nil)
	idm.EnableSignIn(identity.SignInConfig{Provider: identity.DevSignInProvider("user@example.test")})
	wired, err := wireGoogle(idm, governance.New(nil), gc)
	if err != nil {
		t.Fatalf("wireGoogle: %v", err)
	}
	r := httpapi.NewRouter()
	r.Group(func(r chi.Router) {
		if err := idm.Routes(r); err != nil {
			t.Fatal(err)
		}
	})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, wireGoogleT33Path, nil))
	return wired, rec.Code
}

func TestWireGoogleEnablesStorageConnectionOnlyWithConfiguration(t *testing.T) { // AC-27
	if wired, status := wireGoogleT33Status(t, nil); wired || status != http.StatusNotImplemented {
		t.Errorf("without configuration: wired %v, status %d; want false, 501", wired, status)
	}
	gc := &identity.GoogleConfig{KeyID: "k1", Key: make([]byte, 32), OAuth: &oauth2.Config{ClientID: "c", ClientSecret: "s"}}
	if wired, status := wireGoogleT33Status(t, gc); !wired || status == http.StatusNotImplemented {
		t.Errorf("with configuration: wired %v, status %d; want true and not 501", wired, status)
	}
}

func TestWireGoogleRefusesAnInvalidConfiguration(t *testing.T) {
	idm := identity.New(nil)
	_, err := wireGoogle(idm, governance.New(nil), &identity.GoogleConfig{KeyID: "k1", Key: make([]byte, 16), OAuth: &oauth2.Config{}})
	if err == nil {
		t.Errorf("wireGoogle with a 16-byte key: nil error, want the EnableGoogle error")
	}
}
