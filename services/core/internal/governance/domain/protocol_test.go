package domain

import (
	"errors"
	"testing"
	"time"
)

// Expected API (to be implemented in protocol.go):
//
//	type ProtocolState string // "draft","ready_for_vote","ready_for_signing","signed","withdrawn"
//	type Effects struct{ ClosePollRound, VoidPrintCopy bool }
//	type ProtocolTransition struct{ From, To ProtocolState; Revision int; ActorID string; At time.Time }
//	type DomainError struct{ Code string }  // Error() string; errors.As(err, &*DomainError)
//	func NewProtocol(actorID string, at time.Time) *Protocol           // draft, revision 1, no number
//	func (p *Protocol) State() ProtocolState
//	func (p *Protocol) Revision() int
//	func (p *Protocol) Number() int                                    // 0 = none reserved
//	func (p *Protocol) Transitions() []ProtocolTransition             // state changes only, oldest first
//	func (p *Protocol) Edit(expectedRevision int, actorID string, at time.Time) (Effects, error)
//	func (p *Protocol) ReturnToDraft(expectedRevision int, actorID string, at time.Time) (Effects, error)
//	func (p *Protocol) MarkReadyForVote(expectedRevision int, actorID string, at time.Time) (Effects, error)
//	func (p *Protocol) MarkReadyForSigning(expectedRevision int, actorID string, at time.Time, number int) (Effects, error)
//	func (p *Protocol) Sign(expectedRevision int, actorID string, at time.Time) (Effects, error)
//	func (p *Protocol) Withdraw(expectedRevision int, actorID string, at time.Time, reason string) (Effects, error)
//	func (p *Protocol) EnsurePollOpen() error                          // poll_not_open unless ready_for_vote

const (
	stDraft   = ProtocolState("draft")
	stVote    = ProtocolState("ready_for_vote")
	stSigning = ProtocolState("ready_for_signing")
	stSigned  = ProtocolState("signed")
	stWithdr  = ProtocolState("withdrawn")

	codeConflict   = "governance.draft_revision_conflict"
	codeLocked     = "governance.protocol_locked"
	codeNotAllowed = "governance.transition_not_allowed"
	codeNotReady   = "governance.not_ready_for_signing"
	codeReason     = "governance.reason_required"
	codePollClosed = "governance.poll_not_open"
)

var (
	t0    = time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)
	t1    = t0.Add(time.Hour)
	actor = "member-1"
)

func domainCode(t *testing.T, err error) string {
	t.Helper()
	var de *DomainError
	if !errors.As(err, &de) {
		t.Fatalf("error = %v (%T), want *DomainError", err, err)
	}
	return de.Code
}

// protocolIn builds a protocol in the given state through the public API.
// ready_for_signing and later states carry number 14.
func protocolIn(t *testing.T, s ProtocolState) *Protocol {
	t.Helper()
	p := NewProtocol(actor, t0)
	must := func(_ Effects, err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("setup to %s: %v", s, err)
		}
	}
	switch s {
	case stDraft:
	case stVote:
		must(p.MarkReadyForVote(p.Revision(), actor, t0))
	case stSigning:
		must(p.MarkReadyForSigning(p.Revision(), actor, t0, 14))
	case stSigned:
		must(p.MarkReadyForSigning(p.Revision(), actor, t0, 14))
		must(p.Sign(p.Revision(), actor, t0))
	case stWithdr:
		must(p.Withdraw(p.Revision(), actor, t0, "не підписано"))
	}
	return p
}

func TestProtocol_StateActionMatrix(t *testing.T) {
	type action struct {
		name string
		do   func(p *Protocol) (Effects, error)
	}
	rev := func(p *Protocol) int { return p.Revision() }
	actions := []action{
		{"edit", func(p *Protocol) (Effects, error) { return p.Edit(rev(p), actor, t1) }},
		{"return_to_draft", func(p *Protocol) (Effects, error) { return p.ReturnToDraft(rev(p), actor, t1) }},
		{"ready_for_vote", func(p *Protocol) (Effects, error) { return p.MarkReadyForVote(rev(p), actor, t1) }},
		{"ready_for_signing", func(p *Protocol) (Effects, error) { return p.MarkReadyForSigning(rev(p), actor, t1, 15) }},
		{"sign", func(p *Protocol) (Effects, error) { return p.Sign(rev(p), actor, t1) }},
		{"withdraw", func(p *Protocol) (Effects, error) { return p.Withdraw(rev(p), actor, t1, "причина") }},
	}

	// want: next state, or "" with an error code.
	type outcome struct {
		next ProtocolState
		code string
	}
	matrix := map[ProtocolState]map[string]outcome{
		stDraft: {
			"edit":              {next: stDraft},
			"return_to_draft":   {code: codeNotAllowed},
			"ready_for_vote":    {next: stVote},
			"ready_for_signing": {next: stSigning},
			"sign":              {code: codeNotReady}, // AC-40
			"withdraw":          {next: stWithdr},     // AC-35
		},
		stVote: {
			"edit":              {next: stDraft}, // AC-05d
			"return_to_draft":   {next: stDraft}, // AC-05d
			"ready_for_vote":    {code: codeNotAllowed},
			"ready_for_signing": {next: stSigning}, // AC-33
			"sign":              {code: codeNotReady},
			"withdraw":          {next: stWithdr}, // AC-36
		},
		stSigning: {
			"edit":              {next: stDraft}, // AC-14b
			"return_to_draft":   {next: stDraft}, // AC-14b
			"ready_for_vote":    {code: codeNotAllowed},
			"ready_for_signing": {code: codeNotAllowed},
			"sign":              {next: stSigned}, // AC-14
			"withdraw":          {next: stWithdr}, // AC-37
		},
		stSigned: { // AC-17, AC-38
			"edit":              {code: codeLocked},
			"return_to_draft":   {code: codeLocked},
			"ready_for_vote":    {code: codeLocked},
			"ready_for_signing": {code: codeLocked},
			"sign":              {code: codeLocked},
			"withdraw":          {code: codeLocked},
		},
		stWithdr: { // AC-39
			"edit":              {code: codeNotAllowed},
			"return_to_draft":   {code: codeNotAllowed},
			"ready_for_vote":    {code: codeNotAllowed},
			"ready_for_signing": {code: codeNotAllowed},
			"sign":              {code: codeNotAllowed},
			"withdraw":          {code: codeNotAllowed},
		},
	}

	for from, row := range matrix {
		for _, a := range actions {
			want, ok := row[a.name]
			if !ok {
				t.Fatalf("matrix misses %s x %s", from, a.name)
			}
			t.Run(string(from)+"/"+a.name, func(t *testing.T) {
				p := protocolIn(t, from)
				revBefore, nBefore := p.Revision(), len(p.Transitions())
				_, err := a.do(p)
				if want.code != "" {
					if err == nil {
						t.Fatalf("err = nil, want %s", want.code)
					}
					if got := domainCode(t, err); got != want.code {
						t.Errorf("code = %q, want %q", got, want.code)
					}
					if p.State() != from || p.Revision() != revBefore || len(p.Transitions()) != nBefore {
						t.Errorf("refused action changed the protocol: state=%s rev=%d transitions=%d", p.State(), p.Revision(), len(p.Transitions()))
					}
					return
				}
				if err != nil {
					t.Fatalf("err = %v, want next state %s", err, want.next)
				}
				if p.State() != want.next {
					t.Errorf("state = %s, want %s", p.State(), want.next)
				}
			})
		}
	}
}

func TestProtocol_EditAndReturnBumpRevisionAndReportEffects(t *testing.T) {
	tests := []struct {
		name string
		from ProtocolState
		act  func(p *Protocol) (Effects, error)
		want Effects
	}{
		{"edit in draft", stDraft, func(p *Protocol) (Effects, error) { return p.Edit(p.Revision(), actor, t1) }, Effects{}},
		{"edit in ready_for_vote closes round (AC-05d)", stVote, func(p *Protocol) (Effects, error) { return p.Edit(p.Revision(), actor, t1) }, Effects{ClosePollRound: true}},
		{"return from ready_for_vote closes round (AC-05d)", stVote, func(p *Protocol) (Effects, error) { return p.ReturnToDraft(p.Revision(), actor, t1) }, Effects{ClosePollRound: true}},
		{"edit in ready_for_signing voids print copy (AC-14b)", stSigning, func(p *Protocol) (Effects, error) { return p.Edit(p.Revision(), actor, t1) }, Effects{VoidPrintCopy: true}},
		{"return from ready_for_signing voids print copy (AC-14b)", stSigning, func(p *Protocol) (Effects, error) { return p.ReturnToDraft(p.Revision(), actor, t1) }, Effects{VoidPrintCopy: true}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := protocolIn(t, tc.from)
			before := p.Revision()
			got, err := tc.act(p)
			if err != nil {
				t.Fatalf("err = %v", err)
			}
			if got != tc.want {
				t.Errorf("effects = %+v, want %+v", got, tc.want)
			}
			if p.Revision() != before+1 {
				t.Errorf("revision = %d, want %d", p.Revision(), before+1)
			}
			if p.State() != stDraft {
				t.Errorf("state = %s, want draft", p.State())
			}
		})
	}
}

func TestProtocol_StateChangesWithoutContentChangeKeepRevision(t *testing.T) {
	p := NewProtocol(actor, t0)
	before := p.Revision()
	if _, err := p.MarkReadyForVote(before, actor, t1); err != nil {
		t.Fatal(err)
	}
	if p.Revision() != before {
		t.Errorf("revision after MarkReadyForVote = %d, want %d", p.Revision(), before)
	}
}

func TestProtocol_MarkReadyForSigningClosesOpenRound(t *testing.T) { // AC-33
	p := protocolIn(t, stVote)
	eff, err := p.MarkReadyForSigning(p.Revision(), actor, t1, 14)
	if err != nil {
		t.Fatal(err)
	}
	if !eff.ClosePollRound || eff.VoidPrintCopy {
		t.Errorf("effects = %+v, want ClosePollRound only", eff)
	}
}

func TestProtocol_StaleRevisionIsAConflict(t *testing.T) {
	stale := func(p *Protocol) int { return p.Revision() - 1 }
	mutations := map[string]struct {
		from ProtocolState
		do   func(p *Protocol) (Effects, error)
	}{
		"edit":              {stDraft, func(p *Protocol) (Effects, error) { return p.Edit(stale(p), actor, t1) }},
		"return_to_draft":   {stVote, func(p *Protocol) (Effects, error) { return p.ReturnToDraft(stale(p), actor, t1) }},
		"ready_for_vote":    {stDraft, func(p *Protocol) (Effects, error) { return p.MarkReadyForVote(stale(p), actor, t1) }},
		"ready_for_signing": {stDraft, func(p *Protocol) (Effects, error) { return p.MarkReadyForSigning(stale(p), actor, t1, 14) }},
		"sign":              {stSigning, func(p *Protocol) (Effects, error) { return p.Sign(stale(p), actor, t1) }},
		"withdraw":          {stDraft, func(p *Protocol) (Effects, error) { return p.Withdraw(stale(p), actor, t1, "причина") }},
	}
	for name, m := range mutations {
		t.Run(name, func(t *testing.T) {
			// build at revision 2 so that revision-1 is a real earlier revision
			p := NewProtocol(actor, t0)
			if _, err := p.Edit(p.Revision(), actor, t0); err != nil {
				t.Fatal(err)
			}
			switch m.from {
			case stVote:
				if _, err := p.MarkReadyForVote(p.Revision(), actor, t0); err != nil {
					t.Fatal(err)
				}
			case stSigning:
				if _, err := p.MarkReadyForSigning(p.Revision(), actor, t0, 14); err != nil {
					t.Fatal(err)
				}
			}
			state, rev, n := p.State(), p.Revision(), len(p.Transitions())
			_, err := m.do(p)
			if err == nil {
				t.Fatal("err = nil, want conflict")
			}
			if got := domainCode(t, err); got != codeConflict {
				t.Errorf("code = %q, want %q", got, codeConflict)
			}
			if p.State() != state || p.Revision() != rev || len(p.Transitions()) != n {
				t.Errorf("conflicting write changed the protocol")
			}
		})
	}
}

func TestProtocol_FutureRevisionIsAlsoAConflict(t *testing.T) {
	p := NewProtocol(actor, t0)
	_, err := p.Edit(p.Revision()+5, actor, t1)
	if err == nil || domainCode(t, err) != codeConflict {
		t.Fatalf("err = %v, want %s", err, codeConflict)
	}
}

func TestProtocol_WithdrawNeedsAReason(t *testing.T) { // AC-35..37
	for _, from := range []ProtocolState{stDraft, stVote, stSigning} {
		for _, reason := range []string{"", "   ", "\t\n"} {
			t.Run(string(from)+"/"+reason, func(t *testing.T) {
				p := protocolIn(t, from)
				_, err := p.Withdraw(p.Revision(), actor, t1, reason)
				if err == nil {
					t.Fatal("err = nil, want reason_required")
				}
				if got := domainCode(t, err); got != codeReason {
					t.Errorf("code = %q, want %q", got, codeReason)
				}
				if p.State() != from {
					t.Errorf("state = %s, want unchanged %s", p.State(), from)
				}
			})
		}
	}
}

func TestProtocol_WithdrawFromSignedIsLockedEvenWithoutReason(t *testing.T) { // AC-38
	p := protocolIn(t, stSigned)
	_, err := p.Withdraw(p.Revision(), actor, t1, "")
	if err == nil || domainCode(t, err) != codeLocked {
		t.Fatalf("err = %v, want %s", err, codeLocked)
	}
}

func TestProtocol_WithdrawKeepsPrintCopy(t *testing.T) { // AC-37
	p := protocolIn(t, stSigning)
	eff, err := p.Withdraw(p.Revision(), actor, t1, "не підписано на засіданні")
	if err != nil {
		t.Fatal(err)
	}
	if eff.VoidPrintCopy {
		t.Error("withdrawal must keep the print copy (as never signed), VoidPrintCopy = true")
	}
}

func TestProtocol_NumberIsReservedOnceAndSurvives(t *testing.T) {
	t.Run("none before first print", func(t *testing.T) {
		p := NewProtocol(actor, t0)
		if p.Number() != 0 {
			t.Errorf("Number = %d, want 0", p.Number())
		}
		if _, err := p.MarkReadyForVote(p.Revision(), actor, t0); err != nil {
			t.Fatal(err)
		}
		if p.Number() != 0 {
			t.Errorf("Number after ready_for_vote = %d, want 0", p.Number())
		}
	})
	t.Run("taken at ready_for_signing", func(t *testing.T) {
		p := protocolIn(t, stSigning)
		if p.Number() != 14 {
			t.Errorf("Number = %d, want 14", p.Number())
		}
	})
	t.Run("survives return to draft and is not replaced on re-print (AC-14b)", func(t *testing.T) {
		p := protocolIn(t, stSigning)
		if _, err := p.ReturnToDraft(p.Revision(), actor, t1); err != nil {
			t.Fatal(err)
		}
		if p.Number() != 14 {
			t.Errorf("Number after return = %d, want 14", p.Number())
		}
		if _, err := p.MarkReadyForSigning(p.Revision(), actor, t1, 99); err != nil {
			t.Fatal(err)
		}
		if p.Number() != 14 {
			t.Errorf("Number after re-print = %d, want 14 (a new number must not be taken)", p.Number())
		}
	})
	t.Run("survives edit", func(t *testing.T) {
		p := protocolIn(t, stSigning)
		if _, err := p.Edit(p.Revision(), actor, t1); err != nil {
			t.Fatal(err)
		}
		if p.Number() != 14 {
			t.Errorf("Number = %d, want 14", p.Number())
		}
	})
	t.Run("survives withdrawal after print (AC-37)", func(t *testing.T) {
		p := protocolIn(t, stSigning)
		if _, err := p.Withdraw(p.Revision(), actor, t1, "не підписано"); err != nil {
			t.Fatal(err)
		}
		if p.Number() != 14 {
			t.Errorf("Number = %d, want 14", p.Number())
		}
	})
	t.Run("withdrawn draft takes none (AC-35, AC-36)", func(t *testing.T) {
		for _, from := range []ProtocolState{stDraft, stVote} {
			p := protocolIn(t, from)
			if _, err := p.Withdraw(p.Revision(), actor, t1, "причина"); err != nil {
				t.Fatal(err)
			}
			if p.Number() != 0 {
				t.Errorf("%s: Number = %d, want 0", from, p.Number())
			}
		}
	})
	t.Run("signing without a number is refused", func(t *testing.T) {
		p := NewProtocol(actor, t0)
		_, err := p.MarkReadyForSigning(p.Revision(), actor, t0, 0)
		if err == nil {
			t.Error("err = nil, want an error for number 0 on first print")
		}
	})
}

func TestProtocol_TransitionsRecordActorTimeAndRevision(t *testing.T) {
	other := "member-2"
	p := NewProtocol(actor, t0)
	steps := []struct {
		do   func() error
		want ProtocolTransition
	}{
		{func() error { _, e := p.MarkReadyForVote(1, actor, t0); return e },
			ProtocolTransition{From: stDraft, To: stVote, Revision: 1, ActorID: actor, At: t0}},
		{func() error { _, e := p.Edit(1, other, t1); return e },
			ProtocolTransition{From: stVote, To: stDraft, Revision: 2, ActorID: other, At: t1}},
		{func() error { _, e := p.MarkReadyForSigning(2, actor, t1, 14); return e },
			ProtocolTransition{From: stDraft, To: stSigning, Revision: 2, ActorID: actor, At: t1}},
		{func() error { _, e := p.Sign(2, other, t1.Add(time.Hour)); return e },
			ProtocolTransition{From: stSigning, To: stSigned, Revision: 2, ActorID: other, At: t1.Add(time.Hour)}},
	}
	for i, s := range steps {
		if err := s.do(); err != nil {
			t.Fatalf("step %d: %v", i, err)
		}
		got := p.Transitions()
		if len(got) != i+1 {
			t.Fatalf("step %d: %d transitions, want %d", i, len(got), i+1)
		}
		if got[i] != s.want {
			t.Errorf("step %d: transition = %+v, want %+v", i, got[i], s.want)
		}
	}
}

func TestProtocol_WithdrawRecordsTransition(t *testing.T) {
	p := protocolIn(t, stVote)
	if _, err := p.Withdraw(p.Revision(), "member-3", t1, "не набрано підтримки"); err != nil {
		t.Fatal(err)
	}
	ts := p.Transitions()
	last := ts[len(ts)-1]
	want := ProtocolTransition{From: stVote, To: stWithdr, Revision: p.Revision(), ActorID: "member-3", At: t1}
	if last != want {
		t.Errorf("last transition = %+v, want %+v", last, want)
	}
}

func TestProtocol_PollIsOpenOnlyInReadyForVote(t *testing.T) { // AC-32
	for _, s := range []ProtocolState{stDraft, stVote, stSigning, stSigned, stWithdr} {
		t.Run(string(s), func(t *testing.T) {
			err := protocolIn(t, s).EnsurePollOpen()
			if s == stVote {
				if err != nil {
					t.Fatalf("err = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatal("err = nil, want poll_not_open")
			}
			if got := domainCode(t, err); got != codePollClosed {
				t.Errorf("code = %q, want %q", got, codePollClosed)
			}
		})
	}
}
