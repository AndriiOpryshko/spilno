package identity_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

type rejectionRow struct {
	Outcome, Reason, DecidedBy string
	HasAssociation             bool
}

func rejectionBody(filing, reason string) string {
	b, _ := json.Marshal(map[string]string{"filing_id": filing, "reason": reason})
	return string(b)
}

func rejectionDecisions(t *testing.T, e *reviewEnv, app string) []rejectionRow {
	t.Helper()
	rows, err := e.pool.Query(context.Background(),
		`SELECT outcome, coalesce(reason, ''), decided_by_email, association_id IS NOT NULL FROM application_decisions WHERE application_id = $1 ORDER BY created_at`, app)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var out []rejectionRow
	for rows.Next() {
		var r rejectionRow
		if err := rows.Scan(&r.Outcome, &r.Reason, &r.DecidedBy, &r.HasAssociation); err != nil {
			t.Fatal(err)
		}
		out = append(out, r)
	}
	return out
}

func rejectionState(t *testing.T, e *reviewEnv, app string) string {
	t.Helper()
	var s string
	if err := e.pool.QueryRow(context.Background(), `SELECT state FROM applications WHERE id = $1`, app).Scan(&s); err != nil {
		t.Fatal(err)
	}
	return s
}

func rejectionPost(e *reviewEnv, t *testing.T, as, app, filing, reason string) (int, string, []byte) {
	t.Helper()
	status, _, body := e.do(t, as, http.MethodPost, "/identity/operator/applications/"+app+"/rejection", rejectionBody(filing, reason))
	return status, assocStateProblemCode(body), body
}

// AC-14: a rejection records one decision with the reason and the оператор, sets the заявка
// rejected, creates no association, and the applicant's own view carries the reason. A rejection
// may come before any check.
func TestRejection_RecordsDecisionAndTheApplicantSeesTheReason(t *testing.T) {
	e := newReviewEnv(t)
	id := e.fileAs(t, "first", fileAppFields(), fileAppPDF(512))
	filing := e.review(t, id).Filings[0].ID
	assocBefore := count(t, e.pool, "associations")

	status, _, body := rejectionPost(e, t, "operator", id, filing, "Витяг нечитабельний")
	if status != http.StatusOK {
		t.Fatalf("rejectApplication = %d, want 200: %s", status, body)
	}
	var raw map[string]any
	_ = json.Unmarshal(body, &raw)
	if raw["outcome"] != "rejected" || raw["filing_id"] != filing || raw["reason"] != "Витяг нечитабельний" || raw["association_id"] != nil || raw["decided_at"] == nil {
		t.Fatalf("decision body = %s, want rejected with filing, reason, date and no association", body)
	}

	rows := rejectionDecisions(t, e, id)
	if len(rows) != 1 || rows[0].Outcome != "rejected" || rows[0].Reason != "Витяг нечитабельний" || rows[0].DecidedBy != guardOperatorEmail || rows[0].HasAssociation {
		t.Fatalf("decisions = %+v, want one rejected with the reason by %s", rows, guardOperatorEmail)
	}
	if s := rejectionState(t, e, id); s != "rejected" {
		t.Fatalf("state = %q, want rejected", s)
	}
	if n := count(t, e.pool, "associations"); n != assocBefore {
		t.Fatalf("associations = %d, want %d: a rejection creates none", n, assocBefore)
	}

	status, _, mine := e.do(t, "first", http.MethodGet, "/identity/applications/"+id, "")
	if status != http.StatusOK {
		t.Fatalf("getMyApplication = %d: %s", status, mine)
	}
	var detail struct {
		Application struct {
			State           string  `json:"state"`
			RejectionReason *string `json:"rejection_reason"`
		} `json:"application"`
	}
	_ = json.Unmarshal(mine, &detail)
	if detail.Application.State != "rejected" || detail.Application.RejectionReason == nil || *detail.Application.RejectionReason != "Витяг нечитабельний" {
		t.Fatalf("applicant view = %s, want rejected with the reason", mine)
	}
}

// AC-13: an empty (or blank) reason is 422 identity.reason_required and the заявка stays pending.
func TestRejection_EmptyReasonIsRefused(t *testing.T) {
	e := newReviewEnv(t)
	id := e.fileAs(t, "first", fileAppFields(), fileAppPDF(512))
	filing := e.review(t, id).Filings[0].ID
	for _, reason := range []string{"", "   "} {
		status, code, body := rejectionPost(e, t, "operator", id, filing, reason)
		if status != http.StatusUnprocessableEntity || code != "identity.reason_required" {
			t.Fatalf("reason %q = %d %s, want 422 identity.reason_required", reason, status, body)
		}
	}
	if s := rejectionState(t, e, id); s != "pending" {
		t.Fatalf("state = %q, want pending", s)
	}
	if n := len(rejectionDecisions(t, e, id)); n != 0 {
		t.Fatalf("decisions = %d, want 0", n)
	}
}

// AC-15: a second decision is 409 identity.application_not_pending and the stored one is unchanged.
func TestRejection_SecondDecisionIsRefusedAndTheFirstStays(t *testing.T) {
	e := newReviewEnv(t)
	id := e.fileAs(t, "first", fileAppFields(), fileAppPDF(512))
	filing := e.review(t, id).Filings[0].ID
	if status, _, body := rejectionPost(e, t, "operator", id, filing, "Перша причина"); status != http.StatusOK {
		t.Fatalf("first rejection = %d: %s", status, body)
	}
	status, code, body := rejectionPost(e, t, "operator", id, filing, "Друга причина")
	if status != http.StatusConflict || code != "identity.application_not_pending" {
		t.Fatalf("second rejection = %d %s, want 409 identity.application_not_pending", status, body)
	}
	rows := rejectionDecisions(t, e, id)
	if len(rows) != 1 || rows[0].Reason != "Перша причина" {
		t.Fatalf("decisions = %+v, want the first one unchanged", rows)
	}
}

// AC-16c: the оператор's own filing, of both kinds, is 409 identity.operator_own_application and
// stays pending.
func TestRejection_OperatorsOwnFilingIsRefusedForBothKinds(t *testing.T) {
	for _, kind := range []string{"connect", "transfer"} {
		t.Run(kind, func(t *testing.T) {
			e := newReviewEnv(t)
			id := e.fileAs(t, "operator", fileAppFields(), fileAppPDF(512))
			if kind == "transfer" {
				if _, err := e.pool.Exec(context.Background(), `UPDATE applications SET kind = 'transfer' WHERE id = $1`, id); err != nil {
					t.Fatal(err)
				}
			}
			filing := e.review(t, id).Filings[0].ID
			status, code, body := rejectionPost(e, t, "operator", id, filing, "Причина")
			if status != http.StatusConflict || code != "identity.operator_own_application" {
				t.Fatalf("own %s rejection = %d %s, want 409 identity.operator_own_application", kind, status, body)
			}
			if s := rejectionState(t, e, id); s != "pending" {
				t.Fatalf("state = %q, want pending", s)
			}
			if n := len(rejectionDecisions(t, e, id)); n != 0 {
				t.Fatalf("decisions = %d, want 0", n)
			}
		})
	}
}

// AC-16: only the оператор may reject; nothing changes for anyone else.
func TestRejection_NonOperatorsAreRefused(t *testing.T) {
	e := newReviewEnv(t)
	id := e.fileAs(t, "first", fileAppFields(), fileAppPDF(512))
	filing := e.review(t, id).Filings[0].ID
	for _, who := range []string{"head", "first", "second"} {
		status, code, body := rejectionPost(e, t, who, id, filing, "Причина")
		if status != http.StatusForbidden || code != "identity.not_operator" || strings.Contains(string(body), id) {
			t.Fatalf("%s = %d %s, want 403 identity.not_operator", who, status, body)
		}
	}
	if s := rejectionState(t, e, id); s != "pending" {
		t.Fatalf("state = %q, want pending", s)
	}
}
