// Package platform_test holds the release gate of SAD §10 QG-1: no association's data reaches
// a person who is not on its roster.
package platform_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/AndriiOpryshko/spilno/services/core/internal/finance"
	fingen "github.com/AndriiOpryshko/spilno/services/core/internal/finance/adapters/http/gen"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance"
	govgen "github.com/AndriiOpryshko/spilno/services/core/internal/governance/adapters/http/gen"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/governancetest"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity"
	idgen "github.com/AndriiOpryshko/spilno/services/core/internal/identity/adapters/http/gen"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/db"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/httpapi"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/testdb"
)

const (
	isoOperatorEmail = "operator@example.test"
	isoBOnlyEmail    = "only-b-board-member@example.test"
	isoBName         = "Association B"
)

type isoRoster struct{ gov *governance.Module }

func (g isoRoster) CurrentMember(ctx context.Context, association uuid.UUID, email string, date time.Time) (bool, error) {
	_, ok, err := g.gov.CurrentMember(ctx, association, email, date)
	return ok, err
}

type isoProvider struct {
	people map[string]identity.ProviderIdentity
}

func (isoProvider) AuthCodeURL(state, nonce, challenge string) string {
	q := url.Values{"state": {state}, "nonce": {nonce}, "code_challenge": {challenge}}
	return "https://accounts.example.test/o/oauth2/auth?" + q.Encode()
}

func (f isoProvider) Exchange(_ context.Context, code, _, _ string) (identity.ProviderIdentity, error) {
	p, ok := f.people[code]
	if !ok {
		return identity.ProviderIdentity{}, errors.New("unknown code")
	}
	return p, nil
}

var isoNoRedirect = &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}

type isoEnv struct {
	srv     *httptest.Server
	a, b    string
	cookies map[string]*http.Cookie
}

// isoSignIn runs start and callback for the fake person behind code and returns the session cookie.
func isoSignIn(t *testing.T, srv *httptest.Server, code string) *http.Cookie {
	t.Helper()
	resp, err := isoNoRedirect.Get(srv.URL + "/identity/sign-in")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	loc, err := url.Parse(resp.Header.Get("Location"))
	if err != nil {
		t.Fatal(err)
	}
	req, _ := http.NewRequest("GET", srv.URL+"/identity/sign-in/callback?code="+code+"&state="+url.QueryEscape(loc.Query().Get("state")), nil)
	for _, c := range resp.Cookies() {
		req.AddCookie(c)
	}
	cb, err := isoNoRedirect.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	cb.Body.Close()
	for _, c := range cb.Cookies() {
		if c.Name == "spilno_session" && c.Value != "" {
			return c
		}
	}
	t.Fatalf("sign-in as %s set no session cookie", code)
	return nil
}

// newIsoEnv builds the router as cmd/api does on two seeded associations with: onlyA (A's roster),
// both (A's and B's), operator (isoOperatorEmail, A's roster only), and a person who is on B only.
func newIsoEnv(t *testing.T) *isoEnv {
	t.Helper()
	pool, a, b := testdb.NewTwoAssociations(t)
	people := map[string]identity.ProviderIdentity{}
	for p, email := range map[string]string{
		"onlyA": "only-a@example.test", "both": "both@example.test",
		"operator": isoOperatorEmail, "onlyB": isoBOnlyEmail,
	} {
		people[p] = identity.ProviderIdentity{Issuer: "https://accounts.example.test", Subject: "sub-" + p, Email: email, EmailVerified: true}
	}
	idm := identity.New(pool)
	idm.EnableSignIn(identity.SignInConfig{Provider: isoProvider{people}, OperatorEmail: isoOperatorEmail})
	gov := governance.New(pool)

	r := httpapi.NewRouter()
	var routeErr error
	r.Group(func(r chi.Router) {
		r.Use(idm.Guard(isoRoster{gov}))
		for _, err := range []error{finance.New(pool).Routes(r), gov.Routes(r), idm.Routes(r)} {
			if err != nil {
				routeErr = err
			}
		}
	})
	if routeErr != nil {
		t.Fatal(routeErr)
	}
	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)

	seed := func(assoc string, emails ...string) {
		id := uuid.MustParse(assoc)
		err := db.InAssociation(context.Background(), pool, id, "seed@example.test", func(ctx context.Context) error {
			tx, err := db.From(ctx)
			if err != nil {
				return err
			}
			for _, e := range emails {
				governancetest.Member(t, tx, id, e, "member")
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	seed(a, "only-a@example.test", "both@example.test", isoOperatorEmail)
	seed(b, "both@example.test", isoBOnlyEmail)

	env := &isoEnv{srv: srv, a: a, b: b, cookies: map[string]*http.Cookie{}}
	for p := range people {
		env.cookies[p] = isoSignIn(t, srv, p)
	}
	return env
}

type isoReply struct {
	status int
	body   []byte
}

func (e *isoEnv) call(t *testing.T, as, method, path string) isoReply {
	t.Helper()
	req, err := http.NewRequest(method, e.srv.URL+path, bytes.NewReader(nil))
	if err != nil {
		t.Fatal(err)
	}
	req.AddCookie(e.cookies[as])
	resp, err := isoNoRedirect.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return isoReply{resp.StatusCode, b}
}

func (r isoReply) refusedAsStranger() bool {
	return r.status == http.StatusForbidden && bytes.Contains(r.body, []byte(`"identity.not_current_member"`))
}

func (r isoReply) pastTheGuard() bool {
	return r.status != http.StatusUnauthorized && r.status != http.StatusForbidden
}

type isoOp struct{ method, path string }

// isoScopedOperations is every roster-scoped operation of the contract: an {association_id} path
// outside /identity/operator/ (those are the оператор's, guarded by who they are, not by a roster).
func isoScopedOperations(t *testing.T) []isoOp {
	t.Helper()
	var ops []isoOp
	for _, load := range []func() (*openapi3.T, error){idgen.GetSpec, govgen.GetSpec, fingen.GetSpec} {
		spec, err := load()
		if err != nil {
			t.Fatal(err)
		}
		for path, item := range spec.Paths.Map() {
			if !strings.Contains(path, "{association_id}") || strings.HasPrefix(path, "/identity/operator/") {
				continue
			}
			for m := range item.Operations() {
				ops = append(ops, isoOp{m, path})
			}
		}
	}
	sort.Slice(ops, func(i, j int) bool { return ops[i].path+ops[i].method < ops[j].path+ops[j].method })
	return ops
}

var isoParam = regexp.MustCompile(`\{[^}]+\}`)

func (o isoOp) url(association string, sameID string) string {
	u := strings.Replace(o.path, "{association_id}", association, 1)
	return isoParam.ReplaceAllString(u, sameID)
}

// QG-1 (AC-04d, AC-07, AC-16b): every association-scoped operation of the contract, called with
// B's id and with A's id by a person on A only, a person on both rosters and the оператор on A only.
func TestIsolationEveryScopedOperationKeepsAssociationBFromPeopleOffItsRoster(t *testing.T) {
	e := newIsoEnv(t)
	ops := isoScopedOperations(t)
	if len(ops) < 30 {
		t.Fatalf("found %d association-scoped operations in the contract, want every one of identity, governance and finance", len(ops))
	}
	leaks := []string{e.b, isoBName, isoBOnlyEmail}
	missingID := uuid.NewString()
	for _, op := range ops {
		id := uuid.NewString()
		unknown := e.call(t, "onlyA", op.method, op.url(missingID, id))
		for _, who := range []string{"onlyA", "operator"} {
			label := who + " " + op.method + " " + op.path

			underB := e.call(t, who, op.method, op.url(e.b, id))
			if !underB.refusedAsStranger() {
				t.Errorf("%s under B: status = %d body = %s, want 403 identity.not_current_member", label, underB.status, underB.body)
			}
			// The same body as an id that does not exist: B's existence is not revealed.
			if underB.status != unknown.status || bytes.Contains(underB.body, []byte(e.b)) {
				t.Errorf("%s under B answers %d %s, want the answer of an unknown id without B's id", label, underB.status, underB.body)
			}
			underA := e.call(t, who, op.method, op.url(e.a, id))
			if !underA.pastTheGuard() {
				t.Errorf("%s under A: status = %d body = %s, want it served past the guard", label, underA.status, underA.body)
			}
			for name, got := range map[string]isoReply{"under B": underB, "under A": underA} {
				for _, leak := range leaks {
					if bytes.Contains(got.body, []byte(leak)) {
						t.Errorf("%s %s: body carries %q of association B: %s", label, name, leak, got.body)
					}
				}
			}
		}
		label := "both " + op.method + " " + op.path
		for name, assoc := range map[string]string{"A": e.a, "B": e.b} {
			if got := e.call(t, "both", op.method, op.url(assoc, id)); !got.pastTheGuard() {
				t.Errorf("%s under %s: status = %d body = %s, want it past the guard", label, name, got.status, got.body)
			}
		}
		// Under A the person on both rosters must still see nothing of B.
		if got := e.call(t, "both", op.method, op.url(e.a, id)); bytes.Contains(got.body, []byte(isoBOnlyEmail)) {
			t.Errorf("%s under A: body carries a person of B only: %s", label, got.body)
		}
	}
}
