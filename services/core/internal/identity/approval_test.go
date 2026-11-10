package identity_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/AndriiOpryshko/spilno/services/core/internal/governance"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/db"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/httpapi"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/testdb"
)

// Contract this test pins for the wiring (like StorageConnectionConfig): identity cannot import
// governance, so the approval receives the roster step as a function. cmd/api passes
// governance.Module.SeatHead through it; the call runs inside the approval's transaction.
//
//	func (m *Module) EnableApprovals(cfg ApprovalConfig)
//	type ApprovalConfig struct {
//		SeatHead func(ctx context.Context, association uuid.UUID, fullName, signInEmail string, termStart time.Time, decisionID uuid.UUID) (uuid.UUID, error)
//	}

// newApprovalEnv is newReviewEnv with approvals enabled; seat nil seats through governance.SeatHead.
func newApprovalEnv(t *testing.T, seat func(ctx context.Context, association uuid.UUID, fullName, signInEmail string, termStart time.Time, decisionID uuid.UUID) (uuid.UUID, error)) *reviewEnv {
	t.Helper()
	pool, _, _ := testdb.NewTwoAssociations(t)
	gov := governance.New(pool)
	if seat == nil {
		seat = func(ctx context.Context, association uuid.UUID, fullName, signInEmail string, termStart time.Time, decisionID uuid.UUID) (uuid.UUID, error) {
			return gov.SeatHead(ctx, association, governance.Person{FullName: fullName, SignInEmail: signInEmail}, governance.Term{Start: termStart}, decisionID)
		}
	}
	idm := identity.New(pool)
	idm.EnableApprovals(identity.ApprovalConfig{SeatHead: seat})
	people := map[string]identity.ProviderIdentity{}
	for p, email := range map[string]string{"first": "first@example.test", "second": "second@example.test", "head": "head@example.test", "operator": guardOperatorEmail} {
		people[p] = identity.ProviderIdentity{Issuer: "https://accounts.example.test", Subject: "sub-" + p, Email: email, EmailVerified: true}
	}
	idm.EnableSignIn(identity.SignInConfig{Provider: &signInFake{people: people}, OperatorEmail: guardOperatorEmail})
	r := httpapi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(idm.Guard(guardRoster{gov}))
		if err := idm.Routes(r); err != nil {
			t.Fatal(err)
		}
	})
	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)
	env := &reviewEnv{srv: srv, pool: pool, cookies: map[string]*http.Cookie{}}
	for p := range people {
		c := sessionCookieOf(signInAs(t, srv, p))
		if c == nil {
			t.Fatalf("no session cookie for %s", p)
		}
		env.cookies[p] = c
	}
	return env
}

type approvalFiled struct{ app, filing string }

// approvalFile files a заявка as `as` for the standard code and returns its ids.
func approvalFile(t *testing.T, e *reviewEnv, as string) approvalFiled {
	t.Helper()
	id := e.fileAs(t, as, fileAppFields(), fileAppPDF(512))
	return approvalFiled{app: id, filing: e.review(t, id).Filings[0].ID}
}

// approvalRecordCheck records the оператор's check; code and head are the extract's values.
func approvalRecordCheck(t *testing.T, e *reviewEnv, f approvalFiled, valid bool, code, head string) {
	t.Helper()
	day := reviewKyivDay(1)
	b, _ := json.Marshal(map[string]any{
		"filing_id": f.filing, "signature_valid": valid, "signer_full_name": "Test User", "signed_on": day,
		"head_full_name": head, "extract_edrpou": code, "extract_dated_on": day,
	})
	status, _, body := e.do(t, "operator", http.MethodPost, "/identity/operator/applications/"+f.app+"/checks", string(b))
	if status != http.StatusCreated {
		t.Fatalf("recordExtractCheck = %d: %s", status, body)
	}
}

func approvalGoodCheck(t *testing.T, e *reviewEnv, f approvalFiled) {
	t.Helper()
	approvalRecordCheck(t, e, f, true, fileAppEdrpou, "Test User")
}

func approvalApprove(e *reviewEnv, t *testing.T, as string, f approvalFiled, document string) (int, []byte) {
	t.Helper()
	m := map[string]any{"filing_id": f.filing}
	if document != "" {
		m["relied_on_document"] = document
	}
	b, _ := json.Marshal(m)
	status, _, body := e.do(t, as, http.MethodPost, "/identity/operator/applications/"+f.app+"/approval", string(b))
	return status, body
}

func approvalScalar[T any](t *testing.T, e *reviewEnv, q string, args ...any) T {
	t.Helper()
	var v T
	if err := e.pool.QueryRow(context.Background(), q, args...).Scan(&v); err != nil {
		t.Fatalf("%s: %v", q, err)
	}
	return v
}

func approvalAssociations(t *testing.T, e *reviewEnv) int {
	t.Helper()
	return approvalScalar[int](t, e, `SELECT count(*) FROM associations`)
}

func approvalState(t *testing.T, e *reviewEnv, app string) string {
	t.Helper()
	return approvalScalar[string](t, e, `SELECT state FROM applications WHERE id = $1`, app)
}

func approvalDecisions(t *testing.T, e *reviewEnv) int {
	t.Helper()
	return approvalScalar[int](t, e, `SELECT count(*) FROM application_decisions`)
}

// approvalMember reads the roster of the association as at date.
func approvalMember(t *testing.T, e *reviewEnv, assoc, email string, date time.Time) (governance.Membership, bool) {
	t.Helper()
	var (
		m  governance.Membership
		ok bool
	)
	id := uuid.MustParse(assoc)
	err := db.InAssociation(context.Background(), e.pool, id, "test@example.test", func(ctx context.Context) (err error) {
		m, ok, err = governance.New(e.pool).CurrentMember(ctx, id, email, date)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	return m, ok
}

// approvalRefused asserts a refusal that changed nothing.
func approvalRefused(t *testing.T, e *reviewEnv, f approvalFiled, status int, body []byte, wantStatus int, wantCode string, assocsBefore, decisionsBefore int) {
	t.Helper()
	if status != wantStatus || (wantCode != "" && assocStateProblemCode(body) != wantCode) {
		t.Fatalf("approve = %d %s, want %d %s", status, body, wantStatus, wantCode)
	}
	if got := approvalState(t, e, f.app); got != "pending" {
		t.Errorf("application state = %q after a refusal, want pending", got)
	}
	if n := approvalAssociations(t, e); n != assocsBefore {
		t.Errorf("associations = %d after a refusal, want %d", n, assocsBefore)
	}
	if n := approvalDecisions(t, e); n != decisionsBefore {
		t.Errorf("decisions = %d after a refusal, want %d", n, decisionsBefore)
	}
}

// AC-08, AC-09: approving creates the association with the filing's code, name and address, the
// decision, and the applicant as head from the term start, all at once.
func TestApproval_CreatesAssociationDecisionAndSeatsTheHead(t *testing.T) {
	e := newApprovalEnv(t, nil)
	f := approvalFile(t, e, "first")
	approvalGoodCheck(t, e, f)
	assocsBefore := approvalAssociations(t, e)

	status, body := approvalApprove(e, t, "operator", f, "")
	if status != http.StatusOK {
		t.Fatalf("approveApplication = %d, want 200: %s", status, body)
	}
	raw := map[string]any{}
	_ = json.Unmarshal(body, &raw)
	decID, _ := raw["id"].(string)
	filingID, _ := raw["filing_id"].(string)
	outcome, _ := raw["outcome"].(string)
	assoc, _ := raw["association_id"].(string)
	decidedAt, _ := raw["decided_at"].(string)
	if decID == "" || filingID != f.filing || outcome != "approved" || assoc == "" || decidedAt == "" {
		t.Fatalf("decision = %s, want approved for filing %s with an association", body, f.filing)
	}

	if n := approvalAssociations(t, e); n != assocsBefore+1 {
		t.Fatalf("associations = %d, want %d", n, assocsBefore+1)
	}
	var edrpou, name, address, state string
	var confirmedBy uuid.UUID
	if err := e.pool.QueryRow(context.Background(), `SELECT edrpou, name, address, state, confirmed_by_decision_id FROM associations WHERE id = $1`, assoc).
		Scan(&edrpou, &name, &address, &state, &confirmedBy); err != nil {
		t.Fatalf("created association: %v", err)
	}
	if edrpou != fileAppEdrpou || name != "Test Association" || address != "Test Street 1" || state != "connected" || confirmedBy.String() != decID {
		t.Fatalf("association = %q %q %q %q confirmed by %s, want the filing's values, connected, confirmed by %s", edrpou, name, address, state, confirmedBy, decID)
	}
	var by, outcomeRow string
	var checkID *uuid.UUID
	if err := e.pool.QueryRow(context.Background(), `SELECT decided_by_email, outcome, check_id FROM application_decisions WHERE id = $1 AND association_id = $2 AND filing_id = $3`, decID, assoc, f.filing).
		Scan(&by, &outcomeRow, &checkID); err != nil {
		t.Fatalf("decision row: %v", err)
	}
	if by != guardOperatorEmail || outcomeRow != "approved" || checkID == nil {
		t.Fatalf("decision row = %q %q check=%v, want the оператор, approved, the check relied on", by, outcomeRow, checkID)
	}
	if got := approvalState(t, e, f.app); got != "approved" {
		t.Fatalf("application state = %q, want approved", got)
	}
	if got := approvalScalar[string](t, e, `SELECT association_id::text FROM applications WHERE id = $1`, f.app); got != assoc {
		t.Fatalf("application association = %s, want %s", got, assoc)
	}

	// AC-09: head from the term start (2026-03-01), not before.
	m, ok := approvalMember(t, e, assoc, "first@example.test", time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC))
	if !ok || m.Position != "head" {
		t.Fatalf("member on the term start = %+v ok=%v, want head", m, ok)
	}
	if _, ok := approvalMember(t, e, assoc, "first@example.test", time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC)); ok {
		t.Fatal("applicant is a member the day before the term start")
	}
}

// AC-03, AC-03b, AC-12: every guard refuses with its code and changes nothing; a document of 500
// characters is stored, 501 is a broken field.
func TestApproval_GuardsRefuseAndLeaveTheApplicationPending(t *testing.T) {
	e := newApprovalEnv(t, nil)
	assocs := approvalAssociations(t, e)
	refuse := func(name string, prep func(f approvalFiled), document string, wantStatus int, wantCode string) {
		t.Run(name, func(t *testing.T) {
			f := approvalFile(t, e, "first")
			// this filing must not compete with the next case, whatever the outcome.
			defer func() {
				if _, err := e.pool.Exec(context.Background(), `UPDATE applications SET state = 'rejected' WHERE id = $1`, f.app); err != nil {
					t.Error(err)
				}
			}()
			prep(f)
			status, body := approvalApprove(e, t, "operator", f, document)
			approvalRefused(t, e, f, status, body, wantStatus, wantCode, assocs, 0)
		})
	}
	refuse("no check", func(approvalFiled) {}, "", 409, "identity.signature_not_checked")
	refuse("signature found invalid", func(f approvalFiled) { approvalRecordCheck(t, e, f, false, fileAppEdrpou, "Test User") }, "", 409, "identity.signature_not_checked")
	refuse("extract of another code", func(f approvalFiled) { approvalRecordCheck(t, e, f, true, "87654321", "Test User") }, "", 409, "identity.extract_code_mismatch")
	refuse("name mismatch without a document", func(f approvalFiled) { approvalRecordCheck(t, e, f, true, fileAppEdrpou, "Other Person") }, "", 409, "identity.mismatch_needs_document")
	refuse("document longer than 500", func(f approvalFiled) { approvalRecordCheck(t, e, f, true, fileAppEdrpou, "Other Person") }, strings.Repeat("д", 501), 422, "")

	t.Run("name mismatch with a document of 500 characters", func(t *testing.T) {
		f := approvalFile(t, e, "first")
		approvalRecordCheck(t, e, f, true, fileAppEdrpou, "Other Person")
		doc := strings.Repeat("д", 500)
		status, body := approvalApprove(e, t, "operator", f, doc)
		if status != http.StatusOK {
			t.Fatalf("approve with a 500-character document = %d, want 200: %s", status, body)
		}
		if got := approvalScalar[string](t, e, `SELECT relied_on_document FROM application_decisions WHERE filing_id = $1`, f.filing); got != doc {
			t.Fatalf("stored document has %d characters, want the 500 sent", len([]rune(got)))
		}
	})
}

// AC-15: a decided заявка is not decided again.
func TestApproval_ADecidedApplicationIsRefusedWithNoSecondDecision(t *testing.T) {
	e := newApprovalEnv(t, nil)
	f := approvalFile(t, e, "first")
	approvalGoodCheck(t, e, f)
	if status, body := approvalApprove(e, t, "operator", f, ""); status != http.StatusOK {
		t.Fatalf("first approval = %d: %s", status, body)
	}
	assocs, decisions := approvalAssociations(t, e), approvalDecisions(t, e)
	status, body := approvalApprove(e, t, "operator", f, "")
	if status != http.StatusConflict || assocStateProblemCode(body) != "identity.application_not_pending" {
		t.Fatalf("second approval = %d %s, want 409 identity.application_not_pending", status, body)
	}
	if approvalAssociations(t, e) != assocs || approvalDecisions(t, e) != decisions {
		t.Fatal("a refused repeat created an association or a decision")
	}
}

// AC-16c: the оператор's own filing stays pending whatever the check says.
func TestApproval_OperatorCannotApproveTheirOwnFiling(t *testing.T) {
	e := newApprovalEnv(t, nil)
	f := approvalFile(t, e, "operator")
	approvalGoodCheck(t, e, f)
	assocs, decisions := approvalAssociations(t, e), approvalDecisions(t, e)
	status, body := approvalApprove(e, t, "operator", f, "")
	approvalRefused(t, e, f, status, body, http.StatusConflict, "identity.operator_own_application", assocs, decisions)
}

// AC-16: only the оператор decides.
func TestApproval_NonOperatorsAreRefusedAndNothingChanges(t *testing.T) {
	e := newApprovalEnv(t, nil)
	f := approvalFile(t, e, "first")
	approvalGoodCheck(t, e, f)
	assocs, decisions := approvalAssociations(t, e), approvalDecisions(t, e)
	for _, who := range []string{"head", "first", "second"} {
		status, body := approvalApprove(e, t, who, f, "")
		approvalRefused(t, e, f, status, body, http.StatusForbidden, "identity.not_operator", assocs, decisions)
	}
}

// AC-04c: of two competing заявки only one creates the association; the other is refused as
// already connected and stays pending.
func TestApproval_CompetingApplicationsCreateExactlyOneAssociation(t *testing.T) {
	t.Run("sequential", func(t *testing.T) {
		e := newApprovalEnv(t, nil)
		one, two := approvalFile(t, e, "first"), approvalFile(t, e, "second")
		approvalGoodCheck(t, e, one)
		approvalGoodCheck(t, e, two)
		if status, body := approvalApprove(e, t, "operator", one, ""); status != http.StatusOK {
			t.Fatalf("first approval = %d: %s", status, body)
		}
		assocs, decisions := approvalAssociations(t, e), approvalDecisions(t, e)
		status, body := approvalApprove(e, t, "operator", two, "")
		approvalRefused(t, e, two, status, body, http.StatusConflict, "identity.association_already_connected", assocs, decisions)
	})

	t.Run("racing", func(t *testing.T) {
		e := newApprovalEnv(t, nil)
		one, two := approvalFile(t, e, "first"), approvalFile(t, e, "second")
		approvalGoodCheck(t, e, one)
		approvalGoodCheck(t, e, two)
		assocs := approvalAssociations(t, e)

		type reply struct {
			status int
			body   []byte
		}
		var wg sync.WaitGroup
		replies := make([]reply, 2)
		start := make(chan struct{})
		for i, f := range []approvalFiled{one, two} {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-start
				s, b := approvalApprove(e, t, "operator", f, "")
				replies[i] = reply{s, b}
			}()
		}
		close(start)
		wg.Wait()

		ok, conflict := 0, 0
		for _, r := range replies {
			switch {
			case r.status == http.StatusOK:
				ok++
			case r.status == http.StatusConflict && assocStateProblemCode(r.body) == "identity.association_already_connected":
				conflict++
			default:
				t.Errorf("racing approval = %d %s, want 200 or 409 identity.association_already_connected", r.status, r.body)
			}
		}
		if ok != 1 || conflict != 1 {
			t.Fatalf("racing approvals: %d ok, %d conflict, want 1 and 1", ok, conflict)
		}
		if n := approvalAssociations(t, e); n != assocs+1 {
			t.Fatalf("associations = %d, want %d", n, assocs+1)
		}
		if n := approvalScalar[int](t, e, `SELECT count(*) FROM associations WHERE edrpou = $1`, fileAppEdrpou); n != 1 {
			t.Fatalf("associations with the code = %d, want 1", n)
		}
		if n := approvalScalar[int](t, e, `SELECT count(*) FROM applications WHERE state = 'pending'`); n != 1 {
			t.Fatalf("pending applications = %d, want the loser still pending", n)
		}
	})
}

// The roster step failing rolls everything back: no association, no decision, the заявка pending.
func TestApproval_RosterFailureRollsBackTheWholeApproval(t *testing.T) {
	e := newApprovalEnv(t, func(context.Context, uuid.UUID, string, string, time.Time, uuid.UUID) (uuid.UUID, error) {
		return uuid.Nil, errors.New("roster unavailable")
	})
	f := approvalFile(t, e, "first")
	approvalGoodCheck(t, e, f)
	assocs, decisions := approvalAssociations(t, e), approvalDecisions(t, e)

	status, body := approvalApprove(e, t, "operator", f, "")
	approvalRefused(t, e, f, status, body, http.StatusInternalServerError, "", assocs, decisions)
	if n := approvalScalar[int](t, e, `SELECT count(*) FROM associations WHERE edrpou = $1`, fileAppEdrpou); n != 0 {
		t.Fatalf("associations with the code = %d after a rollback, want 0", n)
	}
}
