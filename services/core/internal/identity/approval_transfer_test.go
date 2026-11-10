package identity_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/governancetest"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/db"
)

// Contract this test pins for the wiring (next to SeatHead): the end of the previous голова's
// entry is a second function, run in the same transaction; cmd/api passes governance.Module.EndTerm.
//
//	type ApprovalConfig struct {
//		SeatHead func(...)
//		EndTerm  func(ctx context.Context, association, memberID uuid.UUID, date time.Time, decisionID uuid.UUID) error
//	}

const transferApprovalHandAssoc = "00000000-0000-7000-8000-0000000000a1"

func transferApprovalDay(s string) time.Time {
	d, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return d
}

// transferApprovalSeedHead connects the standard code through an approved заявка by "first", so
// "first" is the голова confirmed by an approval; returns the association and the decision ids.
func transferApprovalSeedHead(t *testing.T, e *reviewEnv) (assoc, decision string) {
	t.Helper()
	f := approvalFile(t, e, "first")
	approvalGoodCheck(t, e, f)
	status, body := approvalApprove(e, t, "operator", f, "")
	if status != http.StatusOK {
		t.Fatalf("seeding approval = %d: %s", status, body)
	}
	assoc = approvalScalar[string](t, e, `SELECT association_id::text FROM applications WHERE id = $1`, f.app)
	decision = approvalScalar[string](t, e, `SELECT id::text FROM application_decisions WHERE filing_id = $1`, f.filing)
	return assoc, decision
}

// transferApprovalFile files a запит на передачу as `as` with the term start; assoc "" files by code.
func transferApprovalFile(t *testing.T, e *reviewEnv, as, code, assoc, termStart string) approvalFiled {
	t.Helper()
	fields := transferFields(code)
	fields["term_start"] = termStart
	if assoc != "" {
		fields["association_id"] = assoc
	}
	id := e.fileAs(t, as, fields, fileAppPDF(256))
	return approvalFiled{app: id, filing: e.review(t, id).Filings[0].ID}
}

func transferApprovalDecision(t *testing.T, e *reviewEnv, f approvalFiled) string {
	t.Helper()
	return approvalScalar[string](t, e, `SELECT id::text FROM application_decisions WHERE filing_id = $1`, f.filing)
}

// transferApprovalInRoster runs q on the roster of the association (row-level security).
func transferApprovalInRoster(t *testing.T, e *reviewEnv, assoc string, q func(ctx context.Context, tx db.DBTX)) {
	t.Helper()
	err := db.InAssociation(context.Background(), e.pool, uuid.MustParse(assoc), "test@example.test", func(ctx context.Context) error {
		tx, err := db.From(ctx)
		if err != nil {
			return err
		}
		q(ctx, tx)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func transferApprovalSeedMember(t *testing.T, e *reviewEnv, assoc, email, position string) {
	t.Helper()
	transferApprovalInRoster(t, e, assoc, func(_ context.Context, tx db.DBTX) {
		governancetest.Member(t, tx, uuid.MustParse(assoc), email, position)
	})
}

type transferApprovalRevision struct {
	Position string
	Decision *uuid.UUID
	TermEnd  *time.Time
	Entries  int
}

// transferApprovalLatest reads the latest revision of the person's roster entry and the number of entries.
func transferApprovalLatest(t *testing.T, e *reviewEnv, assoc, email string) transferApprovalRevision {
	t.Helper()
	var r transferApprovalRevision
	transferApprovalInRoster(t, e, assoc, func(ctx context.Context, tx db.DBTX) {
		err := tx.QueryRow(ctx, `SELECT position, decision_id, term_end FROM board_member_revisions
			WHERE association_id = $1 AND sign_in_email = $2 ORDER BY id DESC LIMIT 1`, assoc, email).Scan(&r.Position, &r.Decision, &r.TermEnd)
		if err != nil {
			t.Errorf("latest revision of %s: %v", email, err)
		}
		if err := tx.QueryRow(ctx, `SELECT count(DISTINCT board_member_id) FROM board_member_revisions WHERE association_id = $1 AND sign_in_email = $2`, assoc, email).Scan(&r.Entries); err != nil {
			t.Errorf("entries of %s: %v", email, err)
		}
	})
	return r
}

func transferApprovalProbe(e *reviewEnv, t *testing.T, as, assoc string) (int, []byte) {
	t.Helper()
	status, _, body := e.do(t, as, http.MethodGet, "/probe/associations/"+assoc, "")
	return status, body
}

func transferApprovalSameDay(a *time.Time, want string) bool {
	return a != nil && a.Format("2006-01-02") == want
}

// AC-19, AC-20: term start 2026-11-05: the new голова is current from that day, the previous entry
// ends the day before, the other entries are unchanged, both revisions carry the decision, and the
// previous голова's session opened before the approval is refused at its next request.
func TestApprovalTransfer_SeatsTheNewHeadAndEndsThePreviousEntry(t *testing.T) {
	e := newApprovalEnv(t, nil)
	assoc, firstDecision := transferApprovalSeedHead(t, e)
	transferApprovalSeedMember(t, e, assoc, "other@example.test", "member")
	if status, body := transferApprovalProbe(e, t, "first", assoc); status != http.StatusOK {
		t.Fatalf("previous head before the approval = %d %s, want 200", status, body)
	}
	f := transferApprovalFile(t, e, "second", fileAppEdrpou, "", "2026-11-05")
	approvalRecordCheck(t, e, f, true, fileAppEdrpou, "Test User")

	status, body := approvalApprove(e, t, "operator", f, "")
	if status != http.StatusOK {
		t.Fatalf("approveApplication (transfer) = %d, want 200: %s", status, body)
	}
	dec := transferApprovalDecision(t, e, f)
	if dec == firstDecision {
		t.Fatal("the transfer reused the first decision")
	}
	if got := approvalState(t, e, f.app); got != "approved" {
		t.Errorf("application state = %q, want approved", got)
	}

	if m, ok := approvalMember(t, e, assoc, "second@example.test", transferApprovalDay("2026-11-05")); !ok || m.Position != "head" {
		t.Errorf("new head on 2026-11-05 = %+v ok=%v, want head", m, ok)
	}
	if _, ok := approvalMember(t, e, assoc, "second@example.test", transferApprovalDay("2026-11-04")); ok {
		t.Error("new head is a member before the term start")
	}
	if _, ok := approvalMember(t, e, assoc, "first@example.test", transferApprovalDay("2026-11-04")); !ok {
		t.Error("previous head is not a member on 2026-11-04, the last day of the term")
	}
	if _, ok := approvalMember(t, e, assoc, "first@example.test", transferApprovalDay("2026-11-05")); ok {
		t.Error("previous head is still a member on the new term start")
	}
	prev := transferApprovalLatest(t, e, assoc, "first@example.test")
	if !transferApprovalSameDay(prev.TermEnd, "2026-11-04") || prev.Decision == nil || prev.Decision.String() != dec {
		t.Errorf("previous revision = %+v, want ended 2026-11-04 under decision %s", prev, dec)
	}
	next := transferApprovalLatest(t, e, assoc, "second@example.test")
	if next.Position != "head" || next.Decision == nil || next.Decision.String() != dec {
		t.Errorf("new revision = %+v, want head under decision %s", next, dec)
	}
	if other := transferApprovalLatest(t, e, assoc, "other@example.test"); other.Position != "member" || other.TermEnd != nil || other.Decision != nil {
		t.Errorf("other member's entry = %+v, want untouched", other)
	}

	status, body = transferApprovalProbe(e, t, "first", assoc)
	if status != http.StatusForbidden || assocStateProblemCode(body) != "identity.not_current_member" {
		t.Errorf("previous head's next request = %d %s, want 403 identity.not_current_member", status, body)
	}
	if status, body := transferApprovalProbe(e, t, "second", assoc); status != http.StatusOK {
		t.Errorf("new head's request = %d %s, want 200", status, body)
	}
}

// AC-19: an applicant who is already on the roster gets no second entry, the entry becomes head.
func TestApprovalTransfer_AnApplicantOnTheRosterIsReseatedWithoutASecondEntry(t *testing.T) {
	e := newApprovalEnv(t, nil)
	assoc, _ := transferApprovalSeedHead(t, e)
	transferApprovalSeedMember(t, e, assoc, "second@example.test", "member")
	f := transferApprovalFile(t, e, "second", fileAppEdrpou, "", "2026-11-05")
	approvalRecordCheck(t, e, f, true, fileAppEdrpou, "Test User")
	if status, body := approvalApprove(e, t, "operator", f, ""); status != http.StatusOK {
		t.Fatalf("approveApplication (transfer) = %d, want 200: %s", status, body)
	}
	got := transferApprovalLatest(t, e, assoc, "second@example.test")
	if got.Entries != 1 || got.Position != "head" || got.TermEnd != nil || got.Decision == nil || got.Decision.String() != transferApprovalDecision(t, e, f) {
		t.Errorf("applicant's entry = %+v, want one open head entry under the decision", got)
	}
	if _, ok := approvalMember(t, e, assoc, "first@example.test", time.Now()); ok {
		t.Error("previous head is still current")
	}
}

// AC-19b: two запити pending; approving one leaves the other pending; approving it later ends the
// entry of the голова confirmed by the first approval.
func TestApprovalTransfer_TheCompetingRequestStaysPendingAndEndsTheFirstFaction(t *testing.T) {
	e := newApprovalEnv(t, nil)
	assoc, _ := transferApprovalSeedHead(t, e)
	one := transferApprovalFile(t, e, "second", fileAppEdrpou, "", "2026-11-05")
	two := transferApprovalFile(t, e, "head", fileAppEdrpou, "", "2026-11-08")
	approvalRecordCheck(t, e, one, true, fileAppEdrpou, "Test User")
	approvalRecordCheck(t, e, two, true, fileAppEdrpou, "Test User")

	if status, body := approvalApprove(e, t, "operator", one, ""); status != http.StatusOK {
		t.Fatalf("first transfer = %d: %s", status, body)
	}
	if got := approvalState(t, e, two.app); got != "pending" {
		t.Fatalf("competing application state = %q, want pending", got)
	}
	if status, body := approvalApprove(e, t, "operator", two, ""); status != http.StatusOK {
		t.Fatalf("competing transfer = %d, want 200: %s", status, body)
	}
	prev := transferApprovalLatest(t, e, assoc, "second@example.test")
	if !transferApprovalSameDay(prev.TermEnd, "2026-11-07") {
		t.Errorf("first transfer's head entry = %+v, want ended 2026-11-07", prev)
	}
	if m, ok := approvalMember(t, e, assoc, "head@example.test", transferApprovalDay("2026-11-08")); !ok || m.Position != "head" {
		t.Errorf("competing head on 2026-11-08 = %+v ok=%v, want head", m, ok)
	}
}

// AC-19b edge: a term start on or before the previous entry's first day never ends it before that day.
func TestApprovalTransfer_ThePreviousEntryNeverEndsBeforeItsFirstDay(t *testing.T) {
	e := newApprovalEnv(t, nil)
	assoc, _ := transferApprovalSeedHead(t, e)
	one := transferApprovalFile(t, e, "second", fileAppEdrpou, "", "2026-11-05")
	two := transferApprovalFile(t, e, "head", fileAppEdrpou, "", "2026-11-05")
	approvalRecordCheck(t, e, one, true, fileAppEdrpou, "Test User")
	approvalRecordCheck(t, e, two, true, fileAppEdrpou, "Test User")
	for _, f := range []approvalFiled{one, two} {
		if status, body := approvalApprove(e, t, "operator", f, ""); status != http.StatusOK {
			t.Fatalf("transfer = %d, want 200: %s", status, body)
		}
	}
	prev := transferApprovalLatest(t, e, assoc, "second@example.test")
	if !transferApprovalSameDay(prev.TermEnd, "2026-11-05") {
		t.Errorf("entry that began 2026-11-05 = %+v, want ended 2026-11-05 (its first day)", prev)
	}
}

// AC-19c: the hand-made association gets its code, the member's entry gets head, no term ends.
func TestApprovalTransfer_AHandMadeAssociationGetsItsCodeAndItsHead(t *testing.T) {
	e := newApprovalEnv(t, nil)
	transferApprovalSeedMember(t, e, transferApprovalHandAssoc, "first@example.test", "member")
	transferApprovalSeedMember(t, e, transferApprovalHandAssoc, "other@example.test", "member")
	f := transferApprovalFile(t, e, "first", "55667788", transferApprovalHandAssoc, "2026-11-05")
	approvalRecordCheck(t, e, f, true, "55667788", "Test User")

	if status, body := approvalApprove(e, t, "operator", f, ""); status != http.StatusOK {
		t.Fatalf("approveApplication (hand-made) = %d, want 200: %s", status, body)
	}
	var code string
	var confirmed *uuid.UUID
	if err := e.pool.QueryRow(context.Background(), `SELECT coalesce(edrpou, ''), confirmed_by_decision_id FROM associations WHERE id = $1`, transferApprovalHandAssoc).Scan(&code, &confirmed); err != nil {
		t.Fatal(err)
	}
	if code != "55667788" || confirmed == nil || confirmed.String() != transferApprovalDecision(t, e, f) {
		t.Errorf("association = code %q confirmed by %v, want 55667788 confirmed by the decision", code, confirmed)
	}
	if got := transferApprovalLatest(t, e, transferApprovalHandAssoc, "first@example.test"); got.Position != "head" || got.Entries != 1 || got.TermEnd != nil {
		t.Errorf("member's entry = %+v, want one open head entry", got)
	}
	if n := approvalScalar[int](t, e, `SELECT count(*) FROM applications WHERE id = $1 AND association_id = $2`, f.app, transferApprovalHandAssoc); n != 1 {
		t.Error("application is not attached to the hand-made association")
	}
	var ended int
	transferApprovalInRoster(t, e, transferApprovalHandAssoc, func(ctx context.Context, tx db.DBTX) {
		_ = tx.QueryRow(ctx, `SELECT count(*) FROM board_member_revisions WHERE association_id = $1 AND term_end IS NOT NULL`, transferApprovalHandAssoc).Scan(&ended)
	})
	if ended != 0 {
		t.Errorf("%d entries ended, want none", ended)
	}
}

// AC-19c, AC-16c: while the голова of the hand-made association is the оператор, it is refused.
func TestApprovalTransfer_AnOperatorHeadOfAHandMadeAssociationCannotBeDecided(t *testing.T) {
	e := newApprovalEnv(t, nil)
	transferApprovalSeedMember(t, e, transferApprovalHandAssoc, guardOperatorEmail, "member")
	f := transferApprovalFile(t, e, "operator", "55667788", transferApprovalHandAssoc, "2026-11-05")
	approvalRecordCheck(t, e, f, true, "55667788", "Test User")
	assocs, decisions := approvalAssociations(t, e), approvalDecisions(t, e)
	status, body := approvalApprove(e, t, "operator", f, "")
	approvalRefused(t, e, f, status, body, http.StatusConflict, "identity.operator_own_application", assocs, decisions)
	if got := approvalScalar[string](t, e, `SELECT coalesce(edrpou, '') FROM associations WHERE id = $1`, transferApprovalHandAssoc); got != "" {
		t.Errorf("association code = %q, want none", got)
	}
}

// AC-18b: no check, or an extract of another code, is refused and the roster stays as it was.
func TestApprovalTransfer_RefusesWithoutAMatchingCheckAndLeavesTheRoster(t *testing.T) {
	e := newApprovalEnv(t, nil)
	assoc, _ := transferApprovalSeedHead(t, e)
	assocs, decisions := approvalAssociations(t, e), approvalDecisions(t, e)
	for _, c := range []struct {
		name, code string
		check      bool
		want       string
	}{
		{"no check", fileAppEdrpou, false, "identity.signature_not_checked"},
		{"another code", fileAppEdrpou, true, "identity.extract_code_mismatch"},
	} {
		t.Run(c.name, func(t *testing.T) {
			f := transferApprovalFile(t, e, "second", c.code, "", "2026-11-05")
			defer func() {
				if _, err := e.pool.Exec(context.Background(), `UPDATE applications SET state = 'rejected' WHERE id = $1`, f.app); err != nil {
					t.Error(err)
				}
			}()
			if c.check {
				approvalRecordCheck(t, e, f, true, "87654321", "Test User")
			}
			status, body := approvalApprove(e, t, "operator", f, "")
			approvalRefused(t, e, f, status, body, http.StatusConflict, c.want, assocs, decisions)
			if _, ok := approvalMember(t, e, assoc, "second@example.test", time.Now()); ok {
				t.Error("applicant was seated by a refused approval")
			}
			if m, ok := approvalMember(t, e, assoc, "first@example.test", time.Now()); !ok || m.Position != "head" {
				t.Errorf("previous head = %+v ok=%v after a refusal, want unchanged", m, ok)
			}
		})
	}
}

// AC-24b: the association stays disabled.
func TestApprovalTransfer_ADisabledAssociationStaysDisabled(t *testing.T) {
	e := newApprovalEnv(t, nil)
	assoc, _ := transferApprovalSeedHead(t, e)
	transferApprovalInRoster(t, e, assoc, func(ctx context.Context, tx db.DBTX) {
		if _, err := tx.Exec(ctx, `UPDATE associations SET state = 'disabled' WHERE id = $1`, assoc); err != nil {
			t.Fatal(err)
		}
	})
	f := transferApprovalFile(t, e, "second", fileAppEdrpou, "", "2026-11-05")
	approvalRecordCheck(t, e, f, true, fileAppEdrpou, "Test User")
	if status, body := approvalApprove(e, t, "operator", f, ""); status != http.StatusOK {
		t.Fatalf("approveApplication (disabled) = %d, want 200: %s", status, body)
	}
	if got := approvalScalar[string](t, e, `SELECT state FROM associations WHERE id = $1`, assoc); got != "disabled" {
		t.Errorf("association state = %q, want disabled", got)
	}
	if m, ok := approvalMember(t, e, assoc, "second@example.test", transferApprovalDay("2026-11-05")); !ok || m.Position != "head" {
		t.Errorf("new head = %+v ok=%v, want head", m, ok)
	}
}

// A failing end-of-term step undoes everything, including the decision.
func TestApprovalTransfer_AFailingRosterStepRollsBackEverything(t *testing.T) {
	failEnd := false
	e := newApprovalEnvWith(t, nil, func(ctx context.Context, assoc, member uuid.UUID, date time.Time, decision uuid.UUID) error {
		if failEnd {
			return errors.New("roster unavailable")
		}
		return errors.New("unexpected end of a term")
	})
	assoc, _ := transferApprovalSeedHead(t, e)
	failEnd = true
	f := transferApprovalFile(t, e, "second", fileAppEdrpou, "", "2026-11-05")
	approvalRecordCheck(t, e, f, true, fileAppEdrpou, "Test User")
	assocs, decisions := approvalAssociations(t, e), approvalDecisions(t, e)

	status, body := approvalApprove(e, t, "operator", f, "")
	approvalRefused(t, e, f, status, body, http.StatusInternalServerError, "", assocs, decisions)
	if _, ok := approvalMember(t, e, assoc, "second@example.test", time.Now()); ok {
		t.Error("applicant stayed seated after the rollback")
	}
	if m, ok := approvalMember(t, e, assoc, "first@example.test", time.Now()); !ok || m.Position != "head" {
		t.Errorf("previous head = %+v ok=%v after the rollback, want unchanged", m, ok)
	}
}
