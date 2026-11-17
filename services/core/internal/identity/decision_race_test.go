package identity_test

import (
	"net/http"
	"sync"
	"testing"
)

type raceReply struct {
	status int
	code   string
	body   []byte
	err    error
}

// raceAll releases every call at once and returns the replies in call order. The calls run on
// goroutines and must not touch t: they return their failure in raceReply.err, which raceAll
// reports from the test goroutine once all of them are done.
func raceAll(t *testing.T, calls ...func() raceReply) []raceReply {
	t.Helper()
	out := make([]raceReply, len(calls))
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i, c := range calls {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			out[i] = c()
		}()
	}
	close(start)
	wg.Wait()
	for i, r := range out {
		if r.err != nil {
			t.Fatalf("racing call %d: %v", i, r.err)
		}
	}
	return out
}

func raceApprove(e *reviewEnv, t *testing.T, f approvalFiled) func() raceReply {
	return func() raceReply {
		s, b, err := approvalApproveErr(e, "operator", f, "")
		return raceReply{s, assocStateProblemCode(b), b, err}
	}
}

func raceReject(e *reviewEnv, t *testing.T, f approvalFiled) func() raceReply {
	return func() raceReply {
		s, c, b, err := rejectionPostErr(e, "operator", f.app, f.filing, "Причина")
		return raceReply{s, c, b, err}
	}
}

// wantOneWinner asserts one success status and one 409 with the given code.
func wantOneWinner(t *testing.T, replies []raceReply, okStatus int, loserCode string) {
	t.Helper()
	ok, lost := 0, 0
	for _, r := range replies {
		switch {
		case r.status == okStatus:
			ok++
		case r.status == http.StatusConflict && r.code == loserCode:
			lost++
		default:
			t.Errorf("racing reply = %d %s, want %d or 409 %s", r.status, r.body, okStatus, loserCode)
		}
	}
	if ok != 1 || lost != 1 {
		t.Fatalf("replies: %d won, %d lost with 409 %s, want 1 and 1", ok, lost, loserCode)
	}
}

// AC-15: one заявка approved twice at once records one decision and one association.
func TestDecisionRace_ApprovedTwiceAtOnce(t *testing.T) {
	e := newApprovalEnv(t, nil)
	f := approvalFile(t, e, "first")
	approvalGoodCheck(t, e, f)
	assocs, decisions := approvalAssociations(t, e), approvalDecisions(t, e)

	wantOneWinner(t, raceAll(t, raceApprove(e, t, f), raceApprove(e, t, f)), http.StatusOK, "identity.application_not_pending")

	if n := approvalAssociations(t, e); n != assocs+1 {
		t.Fatalf("associations = %d, want %d", n, assocs+1)
	}
	if n := approvalDecisions(t, e); n != decisions+1 {
		t.Fatalf("decisions = %d, want %d", n, decisions+1)
	}
	if s := approvalState(t, e, f.app); s != "approved" {
		t.Fatalf("state = %q, want approved", s)
	}
}

// AC-15: approved and rejected at once leaves exactly one decision, and the state matches it.
func TestDecisionRace_ApprovedAndRejectedAtOnce(t *testing.T) {
	e := newApprovalEnv(t, nil)
	f := approvalFile(t, e, "first")
	approvalGoodCheck(t, e, f)
	assocs := approvalAssociations(t, e)

	replies := raceAll(t, raceApprove(e, t, f), raceReject(e, t, f))
	won, lost := 0, 0
	for _, r := range replies {
		switch {
		case r.status == http.StatusOK:
			won++
		case r.status == http.StatusConflict && r.code == "identity.application_not_pending":
			lost++
		default:
			t.Errorf("racing reply = %d %s, want 200 or 409 identity.application_not_pending", r.status, r.body)
		}
	}
	if won != 1 || lost != 1 {
		t.Fatalf("replies: %d won, %d lost, want 1 and 1", won, lost)
	}
	rows := rejectionDecisions(t, e, f.app)
	if len(rows) != 1 {
		t.Fatalf("decisions = %+v, want exactly one", rows)
	}
	state := approvalState(t, e, f.app)
	switch rows[0].Outcome {
	case "approved":
		if state != "approved" || approvalAssociations(t, e) != assocs+1 {
			t.Fatalf("approved decision but state %q, associations %d", state, approvalAssociations(t, e))
		}
	default:
		if state != "rejected" || approvalAssociations(t, e) != assocs {
			t.Fatalf("rejected decision (%s) but state %q, associations %d", rows[0].Outcome, state, approvalAssociations(t, e))
		}
	}
}

// AC-04c: two filings by one person for one code racing into the pending unique key leave one row.
func TestDecisionRace_TwoFilingsRaceIntoThePendingKey(t *testing.T) {
	e := newFileAppEnv(t)
	body, ct := fileAppMultipart(t, fileAppFields(), fileAppPDF(64))
	file := func() raceReply {
		r, err := e.fileBody("first", body, ct)
		return raceReply{r.status, r.code(), r.body, err}
	}
	wantOneWinner(t, raceAll(t, file, file), http.StatusCreated, "identity.application_already_open")
	if c := e.counts(t); c != [3]int{1, 1, 1} {
		t.Fatalf("rows = %v, want [1 1 1]", c)
	}
}
