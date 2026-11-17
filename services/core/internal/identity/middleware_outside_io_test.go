package identity_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
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

// T93: the operations with outside I/O (addCaseDocument, downloadFile, completeStorageConnection) run
// after the guard's short check transaction: no pooled connection is held while the request body is read
// and while the handler does its outside call, and db.From fails loudly there.

type ioSample struct {
	atEntry, afterFirstByte int32
	fromErr                 error
}

func TestGuardOutsideIOHoldsNoConnection(t *testing.T) {
	pool, a, _ := testdb.NewTwoAssociations(t)
	idm := identity.New(pool)
	idm.EnableSignIn(identity.SignInConfig{
		Provider:      &signInFake{people: map[string]identity.ProviderIdentity{"member": {Issuer: "https://accounts.example.test", Subject: "sub-member", Email: "member@example.test", EmailVerified: true}}},
		OperatorEmail: guardOperatorEmail,
	})
	gov := governance.New(pool, idm.HasStorageConnection, actorAccess)

	samples := make(chan ioSample, 1)
	probe := func(w http.ResponseWriter, r *http.Request) {
		var s ioSample
		s.atEntry = pool.Stat().AcquiredConns()
		_, s.fromErr = db.From(r.Context())
		one := make([]byte, 1)
		_, _ = io.ReadFull(r.Body, one)
		s.afterFirstByte = pool.Stat().AcquiredConns()
		samples <- s
		_, _ = io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusNoContent)
	}

	r := httpapi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(idm.Guard(guardRoster{gov}))
		if err := idm.Routes(r); err != nil {
			t.Fatal(err)
		}
		// the probes are registered after the real routes so that they replace them
		r.Post("/governance/associations/{association_id}/cases/{case_id}/documents", probe)
		r.Get("/governance/associations/{association_id}/files/{file_id}", probe)
		r.Post("/identity/associations/{association_id}/storage-connection", probe)
	})
	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)

	aid := uuid.MustParse(a)
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
	cookie := sessionCookieOf(signInAs(t, srv, "member"))
	if cookie == nil {
		t.Fatal("no session cookie")
	}

	base := "/governance/associations/" + a
	cases := []struct{ op, method, path string }{
		{"addCaseDocument", "POST", base + "/cases/" + uuid.NewString() + "/documents"},
		{"downloadFile", "GET", base + "/files/" + uuid.NewString()},
		{"completeStorageConnection", "POST", "/identity/associations/" + a + "/storage-connection"},
	}
	for _, c := range cases {
		t.Run(c.op, func(t *testing.T) {
			pr, pw := io.Pipe()
			req, err := http.NewRequest(c.method, srv.URL+c.path, pr)
			if err != nil {
				t.Fatal(err)
			}
			req.AddCookie(cookie)
			done := make(chan error, 1)
			go func() {
				resp, err := signInNoRedirect.Do(req)
				if err == nil {
					_, _ = io.Copy(io.Discard, resp.Body)
					resp.Body.Close()
				}
				done <- err
			}()
			_, _ = pw.Write([]byte("x"))
			var s ioSample
			select {
			case s = <-samples:
			case <-time.After(10 * time.Second):
				t.Fatal("handler never read the body")
			}
			_ = pw.Close()
			if err := <-done; err != nil {
				t.Fatal(err)
			}
			if s.atEntry != 0 || s.afterFirstByte != 0 {
				t.Errorf("%s: pooled connections held = %d at handler entry, %d while the body is being read; want 0 and 0", c.op, s.atEntry, s.afterFirstByte)
			}
			if !errors.Is(s.fromErr, db.ErrNoTransaction) {
				t.Errorf("%s: db.From outside the handler's own transaction = %v, want db.ErrNoTransaction", c.op, s.fromErr)
			}
		})
	}
}
