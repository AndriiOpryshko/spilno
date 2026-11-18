package identity_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/AndriiOpryshko/spilno/services/core/internal/governance"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/governancetest"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/db"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/httpapi"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/testdb"
)

type commitEnv struct {
	srv    *httptest.Server
	pool   *pgxpool.Pool
	a      uuid.UUID
	cookie *http.Cookie
}

// newCommitEnv serves two probe routes behind the guard, as cmd/api does. Each inserts a board
// member row in the guard's transaction and then misbehaves in the way its path names.
func newCommitEnv(t *testing.T) *commitEnv {
	t.Helper()
	pool, a, _ := testdb.NewTwoAssociations(t)
	aid := uuid.MustParse(a)
	idm := identity.New(pool)
	idm.EnableSignIn(identity.SignInConfig{
		Provider: &signInFake{people: map[string]identity.ProviderIdentity{
			"member": {Issuer: "https://accounts.example.test", Subject: "sub-member", Email: "member@example.test", EmailVerified: true},
		}},
		OperatorEmail: guardOperatorEmail,
	})
	gov := governance.New(pool, idm.HasStorageConnection, actorAccess)

	insert := func(r *http.Request) error {
		tx, err := db.From(r.Context())
		if err != nil {
			return err
		}
		_, err = tx.Exec(r.Context(), `INSERT INTO board_members (id, association_id) VALUES ($1, $2)`, uuid.Must(uuid.NewV7()), aid)
		return err
	}
	rt := httpapi.NewRouter()
	rt.Group(func(r chi.Router) {
		r.Use(idm.Guard(guardRoster{gov}))
		// Writes, then reports success although the transaction can no longer commit.
		r.Get("/probe/associations/{association_id}/doomed", func(w http.ResponseWriter, r *http.Request) {
			if err := insert(r); err != nil {
				http.Error(w, err.Error(), http.StatusTeapot)
				return
			}
			tx, _ := db.From(r.Context())
			_, _ = tx.Exec(r.Context(), `SELECT 1/0`) // aborts the transaction; the handler ignores it
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"saved":true}`))
		})
		// Writes, then answers 409.
		r.Get("/probe/associations/{association_id}/conflict", func(w http.ResponseWriter, r *http.Request) {
			if err := insert(r); err != nil {
				http.Error(w, err.Error(), http.StatusTeapot)
				return
			}
			w.WriteHeader(http.StatusConflict)
			_, _ = w.Write([]byte(`{"conflict":true}`))
		})
		// Writes, then answers 200.
		r.Get("/probe/associations/{association_id}/fine", func(w http.ResponseWriter, r *http.Request) {
			if err := insert(r); err != nil {
				http.Error(w, err.Error(), http.StatusTeapot)
				return
			}
			_, _ = w.Write([]byte(`{"saved":true}`))
		})
		if err := idm.Routes(r); err != nil {
			t.Fatal(err)
		}
	})
	srv := httptest.NewServer(rt)
	t.Cleanup(srv.Close)

	err := db.InAssociation(context.Background(), pool, aid, "seed@example.test", func(ctx context.Context) error {
		tx, err := db.From(ctx)
		if err != nil {
			return err
		}
		governancetest.Member(t, tx, aid, "member@example.test", "member")
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	c := sessionCookieOf(signInAs(t, srv, "member"))
	if c == nil {
		t.Fatal("sign-in set no session cookie")
	}
	return &commitEnv{srv: srv, pool: pool, a: aid, cookie: c}
}

func (e *commitEnv) get(t *testing.T, path string) (int, string) {
	t.Helper()
	req, _ := http.NewRequest(http.MethodGet, e.srv.URL+"/probe/associations/"+e.a.String()+path, nil)
	req.AddCookie(e.cookie)
	resp, err := signInNoRedirect.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(b)
}

func (e *commitEnv) members(t *testing.T) int {
	t.Helper()
	var n int
	err := db.InAssociation(context.Background(), e.pool, e.a, "seed@example.test", func(ctx context.Context) error {
		tx, err := db.From(ctx)
		if err != nil {
			return err
		}
		return tx.QueryRow(ctx, `SELECT count(*) FROM board_members WHERE association_id = $1`, e.a).Scan(&n)
	})
	if err != nil {
		t.Fatal(err)
	}
	return n
}

// A commit that fails after the handler ran answers one 500, no success body reaches the client,
// and the write is absent.
func TestGuardFailedCommitAnswers500AndNoSuccessBody(t *testing.T) { // AC-25
	e := newCommitEnv(t)
	before := e.members(t)
	status, body := e.get(t, "/doomed")
	if status != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500 (body %q)", status, body)
	}
	if strings.Contains(body, `"saved"`) {
		t.Errorf("the handler's success body reached the client: %q", body)
	}
	if got := e.members(t); got != before {
		t.Errorf("board_members = %d, want %d: the failed write is present", got, before)
	}
}

// A handler that writes and then answers 409 leaves nothing behind.
func TestGuardRollsBackWhenHandlerAnswers4xx(t *testing.T) { // AC-25
	e := newCommitEnv(t)
	before := e.members(t)
	status, _ := e.get(t, "/conflict")
	if status != http.StatusConflict {
		t.Fatalf("status = %d, want 409", status)
	}
	if got := e.members(t); got != before {
		t.Errorf("board_members = %d, want %d: a 409 handler's write was committed", got, before)
	}
}

// Control: a handler that answers 200 keeps its write.
func TestGuardCommitsWhenHandlerAnswers2xx(t *testing.T) { // AC-25
	e := newCommitEnv(t)
	before := e.members(t)
	if status, body := e.get(t, "/fine"); status != http.StatusOK {
		t.Fatalf("status = %d, want 200 (%q)", status, body)
	}
	if got := e.members(t); got != before+1 {
		t.Errorf("board_members = %d, want %d", got, before+1)
	}
}
