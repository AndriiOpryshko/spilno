package governance_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/AndriiOpryshko/spilno/services/core/internal/governance"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/governancetest"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/db"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/httpapi"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/testdb"
)

const guardedMemberEmail = "member@example.test"

type guardedFakeProvider struct{}

func (guardedFakeProvider) AuthCodeURL(state, nonce, challenge string) string {
	return "https://accounts.example.test/auth?" + url.Values{"state": {state}}.Encode()
}

func (guardedFakeProvider) Exchange(context.Context, string, string, string) (identity.ProviderIdentity, error) {
	return identity.ProviderIdentity{Issuer: "https://accounts.example.test", Subject: "sub-member", Email: guardedMemberEmail, EmailVerified: true}, nil
}

type guardedRoster struct{ gov *governance.Module }

func (g guardedRoster) CurrentMember(ctx context.Context, association uuid.UUID, email string, date time.Time) (bool, error) {
	_, ok, err := g.gov.CurrentMember(ctx, association, email, date)
	return ok, err
}

// newGuardedServer serves governance behind the identity guard, as cmd/api does. Every request
// carries the session of a person who is on the roster of both seeded associations.
func newGuardedServer(t *testing.T, mount func(chi.Router, *governance.Module) error) *httptest.Server {
	t.Helper()
	pool := testdb.NewApp(t)
	idm := identity.New(pool)
	idm.EnableSignIn(identity.SignInConfig{Provider: guardedFakeProvider{}})
	gov := governance.New(pool)
	r := httpapi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(idm.Guard(guardedRoster{gov}))
		for _, err := range []error{idm.Routes(r), mount(r, gov)} {
			if err != nil {
				t.Fatal(err)
			}
		}
	})
	for _, a := range []string{assocA, assocB} {
		id := uuid.MustParse(a)
		err := db.InAssociation(context.Background(), pool, id, "seed@example.test", func(ctx context.Context) error {
			tx, err := db.From(ctx)
			if err != nil {
				return err
			}
			governancetest.Member(t, tx, id, guardedMemberEmail, "member")
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	cookie := guardedSignIn(t, r)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		req.AddCookie(cookie)
		r.ServeHTTP(w, req)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func guardedSignIn(t *testing.T, h http.Handler) *http.Cookie {
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
	t.Fatal(errors.New("sign-in set no session cookie"))
	return nil
}
