package identity_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/AndriiOpryshko/spilno/services/core/internal/governance"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/db"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/httpapi"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/testdb"
)

type assocStateEnv struct {
	srv     *httptest.Server
	pool    *pgxpool.Pool
	a, b    string
	cookies map[string]*http.Cookie
}

// newAssocStateEnv serves identity behind the guard on the app-role pool with associations A and B
// and three signed-in people: operator, member (on A's roster) and nobody.
func newAssocStateEnv(t *testing.T) *assocStateEnv {
	t.Helper()
	pool, a, b := testdb.NewTwoAssociations(t)
	idm := identity.New(pool)
	people := map[string]identity.ProviderIdentity{}
	for p, email := range map[string]string{"member": "member@example.test", "nobody": "nobody@example.test", "operator": guardOperatorEmail} {
		people[p] = identity.ProviderIdentity{Issuer: "https://accounts.example.test", Subject: "sub-" + p, Email: email, EmailVerified: true}
	}
	idm.EnableSignIn(identity.SignInConfig{Provider: &signInFake{people: people}, OperatorEmail: guardOperatorEmail})
	gov := governance.New(pool)
	r := httpapi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(idm.Guard(guardRoster{gov}))
		if err := idm.Routes(r); err != nil {
			t.Fatal(err)
		}
	})
	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)
	env := &assocStateEnv{srv: srv, pool: pool, a: a, b: b, cookies: map[string]*http.Cookie{}}
	for p := range people {
		c := sessionCookieOf(signInAs(t, srv, p))
		if c == nil {
			t.Fatalf("no session cookie for %s", p)
		}
		env.cookies[p] = c
	}
	return env
}

func (e *assocStateEnv) do(t *testing.T, as, method, path, body string) (int, []byte) {
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
	return resp.StatusCode, b
}

func assocStateProblemCode(b []byte) string {
	var p struct {
		Code string `json:"code"`
	}
	_ = json.Unmarshal(b, &p)
	return p.Code
}

// changes returns the (to_state, reason, changed_by_email) rows recorded for the association.
func (e *assocStateEnv) changes(t *testing.T, assoc string) [][3]string {
	t.Helper()
	var out [][3]string
	err := db.InAssociation(context.Background(), e.pool, uuid.MustParse(assoc), "test@example.test", func(ctx context.Context) error {
		tx, err := db.From(ctx)
		if err != nil {
			return err
		}
		rows, err := tx.Query(ctx, `SELECT to_state, coalesce(reason, ''), changed_by_email FROM association_state_changes WHERE association_id = $1 ORDER BY created_at, id`, assoc)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var r [3]string
			if err := rows.Scan(&r[0], &r[1], &r[2]); err != nil {
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

// stateInList reads the association's state and its page entry through listAssociationsForOperator.
func (e *assocStateEnv) stateInList(t *testing.T, id string) string {
	t.Helper()
	status, body := e.do(t, "operator", http.MethodGet, "/identity/operator/associations?limit=100", "")
	if status != http.StatusOK {
		t.Fatalf("listAssociationsForOperator status = %d, want 200: %s", status, body)
	}
	var page struct {
		Items []struct {
			ID    string `json:"id"`
			State string `json:"state"`
		} `json:"items"`
	}
	if err := json.Unmarshal(body, &page); err != nil {
		t.Fatal(err)
	}
	for _, it := range page.Items {
		if it.ID == id {
			return it.State
		}
	}
	t.Fatalf("association %s not in the operator list: %s", id, body)
	return ""
}

func TestAssociationState_OperatorDisablesAndEnables(t *testing.T) {
	e := newAssocStateEnv(t)
	disable := "/identity/operator/associations/" + e.a + "/disabling"
	enable := "/identity/operator/associations/" + e.a + "/enabling"

	if got := e.stateInList(t, e.a); got != "connected" {
		t.Fatalf("initial state = %q, want connected", got)
	}

	// AC-22: empty reason refused, nothing recorded.
	status, body := e.do(t, "operator", http.MethodPost, disable, `{"reason":""}`)
	if status != http.StatusUnprocessableEntity || assocStateProblemCode(body) != "identity.reason_required" {
		t.Fatalf("empty reason = %d %s, want 422 identity.reason_required", status, body)
	}
	if got := e.changes(t, e.a); len(got) != 0 || e.stateInList(t, e.a) != "connected" {
		t.Fatalf("refused empty reason left rows %v / state changed", got)
	}

	// AC-21: disabled with the reason, the оператор and the time.
	status, body = e.do(t, "operator", http.MethodPost, disable, `{"reason":"спір щодо складу правління"}`)
	if status != http.StatusOK {
		t.Fatalf("disableAssociation = %d, want 200: %s", status, body)
	}
	var ch struct {
		AssociationID string  `json:"association_id"`
		State         string  `json:"state"`
		Reason        *string `json:"reason"`
		ChangedAt     string  `json:"changed_at"`
	}
	if err := json.Unmarshal(body, &ch); err != nil {
		t.Fatal(err)
	}
	if ch.AssociationID != e.a || ch.State != "disabled" || ch.Reason == nil || *ch.Reason != "спір щодо складу правління" || ch.ChangedAt == "" {
		t.Fatalf("disable response = %s", body)
	}
	if got := e.stateInList(t, e.a); got != "disabled" {
		t.Fatalf("state after disabling = %q, want disabled", got)
	}
	rows := e.changes(t, e.a)
	if len(rows) != 1 || rows[0] != [3]string{"disabled", "спір щодо складу правління", guardOperatorEmail} {
		t.Fatalf("state changes after disabling = %v", rows)
	}

	// AC-24: disabling a disabled association is refused, nothing recorded.
	status, body = e.do(t, "operator", http.MethodPost, disable, `{"reason":"ще раз"}`)
	if status != http.StatusConflict || assocStateProblemCode(body) != "identity.association_state_unchanged" {
		t.Fatalf("disable twice = %d %s, want 409 identity.association_state_unchanged", status, body)
	}
	if got := e.changes(t, e.a); len(got) != 1 {
		t.Fatalf("refused second disabling changed rows: %v", got)
	}

	// AC-23: enabled again; the enabling is a row.
	status, body = e.do(t, "operator", http.MethodPost, enable, "")
	if status != http.StatusOK {
		t.Fatalf("enableAssociation = %d, want 200: %s", status, body)
	}
	if got := e.stateInList(t, e.a); got != "connected" {
		t.Fatalf("state after enabling = %q, want connected", got)
	}
	rows = e.changes(t, e.a)
	if len(rows) != 2 || rows[1][0] != "connected" || rows[1][2] != guardOperatorEmail {
		t.Fatalf("state changes after enabling = %v", rows)
	}

	// AC-24: enabling a connected association is refused, nothing recorded.
	status, body = e.do(t, "operator", http.MethodPost, enable, "")
	if status != http.StatusConflict || assocStateProblemCode(body) != "identity.association_state_unchanged" {
		t.Fatalf("enable twice = %d %s, want 409 identity.association_state_unchanged", status, body)
	}
	if got := e.changes(t, e.a); len(got) != 2 {
		t.Fatalf("refused second enabling changed rows: %v", got)
	}

	// The other association is untouched.
	if got := e.stateInList(t, e.b); got != "connected" || len(e.changes(t, e.b)) != 0 {
		t.Fatalf("association B changed: state %q", got)
	}
}

// AC-16: anyone but the оператор gets 403 on all three and nothing changes.
func TestAssociationState_NonOperatorRefused(t *testing.T) {
	e := newAssocStateEnv(t)
	base := "/identity/operator/associations"
	for _, as := range []string{"member", "nobody"} {
		for _, c := range []struct{ method, path, body string }{
			{http.MethodGet, base, ""},
			{http.MethodPost, base + "/" + e.a + "/disabling", `{"reason":"x"}`},
			{http.MethodPost, base + "/" + e.a + "/enabling", ""},
		} {
			status, body := e.do(t, as, c.method, c.path, c.body)
			if status != http.StatusForbidden || assocStateProblemCode(body) != "identity.not_operator" {
				t.Errorf("%s %s as %s = %d %s, want 403 identity.not_operator", c.method, c.path, as, status, body)
			}
		}
	}
	if got := e.changes(t, e.a); len(got) != 0 {
		t.Fatalf("refused calls recorded rows: %v", got)
	}
	if got := e.stateInList(t, e.a); got != "connected" {
		t.Fatalf("state = %q, want connected", got)
	}
}
