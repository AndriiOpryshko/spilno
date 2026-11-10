package identity_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

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

const guardOperatorEmail = "operator@example.test"

// guardRoster adapts the governance roster to the port the identity guard asks for.
type guardRoster struct{ gov *governance.Module }

func (g guardRoster) CurrentMember(ctx context.Context, association uuid.UUID, email string, date time.Time) (bool, error) {
	_, ok, err := g.gov.CurrentMember(ctx, association, email, date)
	return ok, err
}

type guardEnv struct {
	srv     *httptest.Server
	a, b    string
	cookies map[string]*http.Cookie
}

// newGuardEnv serves identity, governance and finance behind the identity guard, as cmd/api does,
// with associations A and B seeded and these signed-in people: member (on A), expired (on A until
// yesterday), nobody (no roster), operator (guardOperatorEmail; on A's roster only when operatorOnA).
func newGuardEnv(t *testing.T, operatorOnA bool) *guardEnv {
	t.Helper()
	pool, a, b := testdb.NewTwoAssociations(t)
	idm := identity.New(pool)
	people := map[string]identity.ProviderIdentity{}
	for p, email := range map[string]string{"member": "member@example.test", "expired": "expired@example.test", "nobody": "nobody@example.test", "operator": guardOperatorEmail} {
		people[p] = identity.ProviderIdentity{Issuer: "https://accounts.example.test", Subject: "sub-" + p, Email: email, EmailVerified: true}
	}
	idm.EnableSignIn(identity.SignInConfig{Provider: &signInFake{people: people}, OperatorEmail: guardOperatorEmail})
	gov := governance.New(pool)

	r := httpapi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(idm.Guard(guardRoster{gov}))
		r.Get("/probe/associations/{association_id}", guardProbe(pool))
		for _, err := range []error{idm.Routes(r), gov.Routes(r), finance.New(pool).Routes(r)} {
			if err != nil {
				t.Fatal(err)
			}
		}
	})
	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)

	aid := uuid.MustParse(a)
	today := time.Now().UTC().Truncate(24 * time.Hour)
	err := db.InAssociation(context.Background(), pool, aid, "seed@example.test", func(ctx context.Context) error {
		tx, err := db.From(ctx)
		if err != nil {
			return err
		}
		governancetest.Member(t, tx, aid, "member@example.test", "member")
		governancetest.MemberTerm(t, tx, aid, "expired@example.test", "member", today.AddDate(-1, 0, 0), today.AddDate(0, 0, -1))
		if operatorOnA {
			governancetest.Member(t, tx, aid, guardOperatorEmail, "head")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	env := &guardEnv{srv: srv, a: a, b: b, cookies: map[string]*http.Cookie{}}
	for p := range people {
		c := sessionCookieOf(signInAs(t, srv, p))
		if c == nil {
			t.Fatalf("sign-in as %s set no session cookie", p)
		}
		env.cookies[p] = c
	}
	return env
}

// guardProbe reports what a handler behind the guard can see.
func guardProbe(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actor, ok := identity.ActorFrom(r.Context())
		var setting string
		// The handler runs inside the guard's transaction: the setting is visible on db.From.
		if tx, err := db.From(r.Context()); err == nil {
			_ = tx.QueryRow(r.Context(), `SELECT coalesce(current_setting('app.association_id', true), '')`).Scan(&setting)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"has_actor": ok, "email": actor.Email, "member": actor.Member, "setting": setting})
	}
}

type guardReply struct {
	status int
	body   []byte
}

func (e *guardEnv) call(t *testing.T, as, method, path string) guardReply {
	t.Helper()
	req, err := http.NewRequest(method, e.srv.URL+path, bytes.NewReader(nil))
	if err != nil {
		t.Fatal(err)
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
	return guardReply{resp.StatusCode, b}
}

func (r guardReply) code() string {
	var p struct {
		Code string `json:"code"`
	}
	_ = json.Unmarshal(r.body, &p)
	return p.Code
}

func (e *guardEnv) meetings(assoc string) string {
	return "/governance/associations/" + assoc + "/meetings"
}

// AC-04d, AC-07: a member of A is served under A and refused under B, with the very answer an unknown id gets.
func TestGuardServesTheRosterOfThePathAssociationOnly(t *testing.T) {
	e := newGuardEnv(t, false)
	if got := e.call(t, "member", "GET", e.meetings(e.a)); got.status != http.StatusOK {
		t.Fatalf("member under A: status = %d (%s), want 200", got.status, got.body)
	}
	underB := e.call(t, "member", "GET", e.meetings(e.b))
	unknown := e.call(t, "member", "GET", e.meetings(uuid.NewString()))
	if underB.status != http.StatusForbidden || underB.code() != "identity.not_current_member" {
		t.Fatalf("member under B: status = %d code = %q, want 403 identity.not_current_member", underB.status, underB.code())
	}
	if unknown.status != underB.status || !bytes.Equal(unknown.body, underB.body) {
		t.Errorf("unknown id answers %d %s, want the byte-equal answer of B: %d %s", unknown.status, unknown.body, underB.status, underB.body)
	}
}

// AC-07: no roster entry, an ended term, and no session are refused before any resource is read.
func TestGuardRefusesEveryoneOffTheRoster(t *testing.T) {
	e := newGuardEnv(t, false)
	want := e.call(t, "member", "GET", e.meetings(e.b))
	for _, who := range []string{"nobody", "expired"} {
		got := e.call(t, who, "GET", e.meetings(e.a))
		if got.status != http.StatusForbidden || got.code() != "identity.not_current_member" {
			t.Errorf("%s under A: status = %d code = %q, want 403 identity.not_current_member", who, got.status, got.code())
		}
		if !bytes.Equal(got.body, want.body) {
			t.Errorf("%s under A answers %s, want the same body as an unknown id: %s", who, got.body, want.body)
		}
	}
	got := e.call(t, "", "GET", e.meetings(e.a))
	if got.status != http.StatusUnauthorized || got.code() != "identity.session_required" {
		t.Errorf("no session: status = %d code = %q, want 401 identity.session_required", got.status, got.code())
	}
}

// AC-16b: the оператор off the roster is a stranger; on A's roster they are a member of A and nothing more.
func TestGuardGivesTheOperatorOnlyTheirRoster(t *testing.T) {
	off := newGuardEnv(t, false)
	if got := off.call(t, "operator", "GET", off.meetings(off.a)); got.status != http.StatusForbidden || got.code() != "identity.not_current_member" {
		t.Errorf("operator off the roster under A: status = %d code = %q, want 403 identity.not_current_member", got.status, got.code())
	}
	on := newGuardEnv(t, true)
	if got := on.call(t, "operator", "GET", on.meetings(on.a)); got.status != http.StatusOK {
		t.Errorf("operator on A's roster under A: status = %d (%s), want 200", got.status, got.body)
	}
	if got := on.call(t, "operator", "GET", on.meetings(on.b)); got.status != http.StatusForbidden || got.code() != "identity.not_current_member" {
		t.Errorf("operator on A's roster under B: status = %d code = %q, want 403 identity.not_current_member", got.status, got.code())
	}
}

// Every association-scoped operation of governance, finance and identity is guarded.
func TestGuardCoversEveryAssociationScopedOperation(t *testing.T) {
	e := newGuardEnv(t, false)
	idSpec, _ := idgen.GetSpec()
	govSpec, _ := govgen.GetSpec()
	finSpec, _ := fingen.GetSpec()
	param := regexp.MustCompile(`\{[^}]+\}`)
	seen := 0
	for _, spec := range []struct {
		prefix string
		paths  map[string][]string
	}{
		{"/governance/", opsOf(govSpec.Paths.Map())},
		{"/finance/", opsOf(finSpec.Paths.Map())},
		{"/identity/", opsOf(idSpec.Paths.Map())},
	} {
		for path, methods := range spec.paths {
			if !strings.HasPrefix(path, spec.prefix) || !strings.Contains(path, "{association_id}") || strings.HasPrefix(path, "/identity/operator/") {
				continue
			}
			for _, m := range methods {
				seen++
				url := strings.Replace(path, "{association_id}", e.b, 1)
				url = param.ReplaceAllString(url, uuid.NewString())
				got := e.call(t, "member", m, url)
				if got.status != http.StatusForbidden || got.code() != "identity.not_current_member" {
					t.Errorf("%s %s as a member of A only: status = %d code = %q, want 403 identity.not_current_member", m, path, got.status, got.code())
				}
			}
		}
	}
	if seen < 10 {
		t.Fatalf("checked %d association-scoped operations, want every one of the three contexts", seen)
	}
}

// AC-16: a person who is not the оператор is refused on every /identity/operator operation; the оператор is let in.
func TestGuardRefusesEveryOperatorOperationToANonOperator(t *testing.T) {
	e := newGuardEnv(t, false)
	idSpec, _ := idgen.GetSpec()
	param := regexp.MustCompile(`\{[^}]+\}`)
	seen := 0
	for path, methods := range opsOf(idSpec.Paths.Map()) {
		if !strings.HasPrefix(path, "/identity/operator/") {
			continue
		}
		url := param.ReplaceAllString(path, uuid.NewString())
		for _, m := range methods {
			seen++
			for _, who := range []string{"member", "nobody"} {
				got := e.call(t, who, m, url)
				if got.status != http.StatusForbidden || got.code() != "identity.not_operator" {
					t.Errorf("%s %s as %s: status = %d code = %q, want 403 identity.not_operator", m, path, who, got.status, got.code())
				}
			}
			if got := e.call(t, "operator", m, url); got.status == http.StatusForbidden || got.status == http.StatusUnauthorized {
				t.Errorf("%s %s as the operator: status = %d, want it past the guard", m, path, got.status)
			}
			if got := e.call(t, "", m, url); got.status != http.StatusUnauthorized {
				t.Errorf("%s %s without a session: status = %d, want 401", m, path, got.status)
			}
		}
	}
	if seen < 9 { // the contract has 9 operations under /identity/operator/
		t.Fatalf("checked %d operator operations, want all of the spec", seen)
	}
}

// The handler behind the guard sees the actor and a transaction scoped to the path association.
func TestGuardPutsTheActorAndTheScopeBehindTheHandler(t *testing.T) {
	e := newGuardEnv(t, false)
	got := e.call(t, "member", "GET", "/probe/associations/"+e.a)
	if got.status != http.StatusOK {
		t.Fatalf("probe under A: status = %d (%s), want 200", got.status, got.body)
	}
	var seen struct {
		HasActor bool   `json:"has_actor"`
		Email    string `json:"email"`
		Member   bool   `json:"member"`
		Setting  string `json:"setting"`
	}
	if err := json.Unmarshal(got.body, &seen); err != nil {
		t.Fatal(err)
	}
	if !seen.HasActor || seen.Email != "member@example.test" || !seen.Member {
		t.Errorf("ActorFrom = %+v, want the member member@example.test", seen)
	}
	if seen.Setting != e.a {
		t.Errorf("current_setting('app.association_id') = %q, want %q", seen.Setting, e.a)
	}
}

func opsOf(m map[string]*openapi3.PathItem) map[string][]string {
	out := map[string][]string{}
	for p, item := range m {
		for method := range item.Operations() {
			out[p] = append(out[p], method)
		}
	}
	return out
}
