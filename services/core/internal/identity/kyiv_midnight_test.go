package identity_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/go-chi/chi/v5"

	"github.com/AndriiOpryshko/spilno/services/core/internal/governance"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/httpapi"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/testdb"
)

// Contract this test pins: Module.SetClock(func() time.Time) replaces the clock the guard, the card
// filing, the card, the memberships and the storage consent take «today» from; default time.Now.
// «Today» is always the Europe/Kyiv calendar day, never the UTC day.

// AC-09, AC-19, AC-29: at 00:40 Kyiv on 2026-11-05 (22:40 UTC on 2026-11-04) an approval with term
// start 2026-11-05 admits the new голова at the next request and refuses the previous one.
func TestApprovalTransfer_AtKyivMidnightTheTermStartDayIsAlreadyToday(t *testing.T) {
	pinned := time.Date(2026, 11, 5, 0, 40, 0, 0, time.FixedZone("EET", 2*60*60))
	pool, _, _ := testdb.NewTwoAssociations(t)
	gov := governance.New(pool)
	idm := identity.New(pool)
	idm.SetClock(func() time.Time { return pinned })
	idm.EnableApprovals(identity.ApprovalConfig{
		SeatHead: func(ctx context.Context, association uuid.UUID, fullName, signInEmail string, termStart, termEnd time.Time, decisionID uuid.UUID) (uuid.UUID, error) {
			return gov.SeatHead(ctx, association, governance.Person{FullName: fullName, SignInEmail: signInEmail}, governance.Term{Start: termStart, End: termEnd}, decisionID)
		},
		EndTerm: gov.EndTerm,
	})
	people := map[string]identity.ProviderIdentity{}
	for p, email := range map[string]string{"first": "first@example.test", "second": "second@example.test", "operator": guardOperatorEmail} {
		people[p] = identity.ProviderIdentity{Issuer: "https://accounts.example.test", Subject: "sub-" + p, Email: email, EmailVerified: true}
	}
	idm.EnableSignIn(identity.SignInConfig{Provider: &signInFake{people: people}, OperatorEmail: guardOperatorEmail})
	r := httpapi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(idm.Guard(guardRoster{gov}))
		r.Get("/probe/associations/{association_id}", guardProbe(pool))
		if err := idm.Routes(r); err != nil {
			t.Fatal(err)
		}
	})
	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)
	e := &reviewEnv{srv: srv, pool: pool, cookies: map[string]*http.Cookie{}}
	for p := range people {
		c := sessionCookieOf(signInAs(t, srv, p))
		if c == nil {
			t.Fatalf("no session cookie for %s", p)
		}
		e.cookies[p] = c
	}

	assoc, _ := transferApprovalSeedHead(t, e)
	if status, body := transferApprovalProbe(e, t, "first", assoc); status != http.StatusOK {
		t.Fatalf("previous head before the approval = %d %s, want 200", status, body)
	}
	f := transferApprovalFile(t, e, "second", fileAppEdrpou, "", "2026-11-05")
	approvalRecordCheck(t, e, f, true, fileAppEdrpou, "Test User")
	if status, body := approvalApprove(e, t, "operator", f, ""); status != http.StatusOK {
		t.Fatalf("approveApplication (transfer) = %d, want 200: %s", status, body)
	}

	if status, body := transferApprovalProbe(e, t, "second", assoc); status != http.StatusOK {
		t.Errorf("new head at 00:40 Kyiv on the term start = %d %s, want 200", status, body)
	}
	status, body := transferApprovalProbe(e, t, "first", assoc)
	if status != http.StatusForbidden || assocStateProblemCode(body) != "identity.not_current_member" {
		t.Errorf("previous head at 00:40 Kyiv on the term start = %d %s, want 403 identity.not_current_member", status, body)
	}
}
