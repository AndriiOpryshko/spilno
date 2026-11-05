package domain

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

// Contract under test (T-B):
//
//	func (p *Protocol) ConfirmSlot(expectedRevision int, item int, key string, actorID string, at time.Time) (Effects, error)
//	func (p *Protocol) CorrectSlot(expectedRevision int, item int, key string, v SlotValue, actorID string, at time.Time) (Effects, error)

// reviewDraftIn builds a protocol in state s whose agenda item 0 holds an
// amount slot that needs review (confirmed 1500, a later 1800 offered beside it).
// The slot is confirmed while the protocol moves to s, because an unconfirmed
// slot blocks готовий до підпису (T-A); the later value is offered afterwards.
func reviewDraftIn(t *testing.T, s ProtocolState) (*Protocol, *Slot) {
	t.Helper()
	slot := confirmedAmount(t)
	a := NewAgendaItem("case-1", "Ремонт покрівлі")
	a.SetSlot("amount", slot)
	p := draftWith(t, a)
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
	slot.Offer(uah1800, originAI, SlotSource{})
	if slot.Status() != slotNeedsReview {
		t.Fatalf("setup status = %q, want needs_review", slot.Status())
	}
	return p, slot
}

type slotSnapshot struct {
	status      SlotStatus
	value       SlotValue
	origin      SlotOrigin
	source      SlotSource
	confirmedBy string
	confirmedAt time.Time
	alternative SlotValue
	hasAlt      bool
}

type protocolSnapshot struct {
	state       ProtocolState
	revision    int
	number      int
	transitions []ProtocolTransition
}

func snapSlot(s *Slot) slotSnapshot {
	alt, hasAlt := s.Alternative()
	return slotSnapshot{s.Status(), s.Value(), s.Origin(), s.Source(), s.ConfirmedBy(), s.ConfirmedAt(), alt, hasAlt}
}

func snapProtocol(p *Protocol) protocolSnapshot {
	return protocolSnapshot{p.State(), p.Revision(), p.Number(), p.Transitions()}
}

// slotEdits runs ConfirmSlot or CorrectSlot under one signature.
var slotEdits = map[string]func(p *Protocol, rev, item int, key string) (Effects, error){
	"confirm": func(p *Protocol, rev, item int, key string) (Effects, error) {
		return p.ConfirmSlot(rev, item, key, reviewer, t1)
	},
	"correct": func(p *Protocol, rev, item int, key string) (Effects, error) {
		return p.CorrectSlot(rev, item, key, SlotValue{Kopiyky: 170_000, Currency: "UAH"}, reviewer, t1)
	},
}

func assertUnchanged(t *testing.T, p *Protocol, s *Slot, wantP protocolSnapshot, wantS slotSnapshot) {
	t.Helper()
	if got := snapProtocol(p); !reflect.DeepEqual(got, wantP) {
		t.Fatalf("protocol changed on a refused call: %+v, want %+v", got, wantP)
	}
	if got := snapSlot(s); !reflect.DeepEqual(got, wantS) {
		t.Fatalf("slot changed on a refused call: %+v, want %+v", got, wantS)
	}
}

// B2: the protocol guard comes first; a refused call changes neither slot nor protocol.
func TestProtocol_TB_SlotEditGuard(t *testing.T) {
	cases := []struct {
		name  string
		state ProtocolState
		delta int // expectedRevision = Revision() + delta
		code  string
	}{
		{"signed is locked", stSigned, 0, codeLocked},
		{"signed is locked even with a stale revision", stSigned, -1, codeLocked},
		{"withdrawn is not allowed", stWithdr, 0, codeNotAllowed},
		{"draft with a stale revision", stDraft, -1, codeConflict},
		{"draft with a future revision", stDraft, +1, codeConflict},
		{"ready for vote with a stale revision", stVote, -1, codeConflict},
		{"ready for signing with a stale revision", stSigning, -1, codeConflict},
	}
	for name, edit := range slotEdits {
		for _, tc := range cases {
			t.Run(name+"/"+tc.name, func(t *testing.T) {
				p, s := reviewDraftIn(t, tc.state)
				wantP, wantS := snapProtocol(p), snapSlot(s)

				eff, err := edit(p, p.Revision()+tc.delta, 0, "amount")
				if got := domainCode(t, err); got != tc.code {
					t.Fatalf("code = %q, want %q", got, tc.code)
				}
				if eff != (Effects{}) {
					t.Fatalf("effects = %+v on a refused call, want none", eff)
				}
				assertUnchanged(t, p, s, wantP, wantS)
			})
		}
	}
}

// B2 + B5: a stale revision leaves a needs-review slot unconfirmed and still blocking.
func TestProtocol_TB_StaleRevisionLeavesNeedsReviewSlotUnconfirmed(t *testing.T) {
	p, s := reviewDraftIn(t, stDraft)
	stale := p.Revision()
	if _, err := p.Edit(stale, actor, t0); err != nil {
		t.Fatal(err)
	}

	_, err := p.ConfirmSlot(stale, 0, "amount", reviewer, t1)
	if got := domainCode(t, err); got != codeConflict {
		t.Fatalf("code = %q, want %q", got, codeConflict)
	}
	if s.Confirmed() || s.Status() != slotNeedsReview || s.ConfirmedBy() != actor {
		t.Fatalf("slot status %q by %q, want needs_review by %q", s.Status(), s.ConfirmedBy(), actor)
	}
	if alt, ok := s.Alternative(); !ok || alt != uah1800 {
		t.Fatalf("alternative = %+v/%v, want 1800 kept", alt, ok)
	}
	if got := pointers(p.Readiness()); !reflect.DeepEqual(got, []string{"/agenda_items/0/slots/amount"}) {
		t.Fatalf("blockers = %v, want the amount slot", got)
	}
}

// B3: an unknown item index or slot key is refused without a panic and changes nothing.
func TestProtocol_TB_UnknownSlotIsRefused(t *testing.T) {
	cases := []struct {
		name string
		item int
		key  string
	}{
		{"negative item", -1, "amount"},
		{"item out of range", 1, "amount"},
		{"far out of range", 99, "amount"},
		{"unknown key", 0, "counterparty"},
		{"empty key", 0, ""},
	}
	for name, edit := range slotEdits {
		for _, tc := range cases {
			t.Run(name+"/"+tc.name, func(t *testing.T) {
				p, s := reviewDraftIn(t, stSigning)
				wantP, wantS := snapProtocol(p), snapSlot(s)

				eff, err := edit(p, p.Revision(), tc.item, tc.key)
				if err == nil {
					t.Fatal("err = nil, want a refusal")
				}
				if !errors.Is(err, ErrSlotNotFound) {
					t.Fatalf("err = %v, want ErrSlotNotFound", err)
				}
				if eff != (Effects{}) {
					t.Fatalf("effects = %+v on a refused call, want none", eff)
				}
				assertUnchanged(t, p, s, wantP, wantS)
			})
		}
	}
}

// B3: a protocol with no agenda item refuses item 0 too.
func TestProtocol_TB_NoAgendaItemIsRefused(t *testing.T) {
	p := NewProtocol(actor, t0)
	want := snapProtocol(p)
	if _, err := p.ConfirmSlot(p.Revision(), 0, "amount", reviewer, t1); !errors.Is(err, ErrSlotNotFound) {
		t.Fatalf("err = %v, want ErrSlotNotFound", err)
	}
	if got := snapProtocol(p); !reflect.DeepEqual(got, want) {
		t.Fatalf("protocol changed: %+v, want %+v", got, want)
	}
}

// B4: success is a content change - revision +1, a ready protocol returns to
// draft with its effects and a recorded transition, a draft stays a draft; the number is kept.
func TestProtocol_TB_SlotEditIsAContentChange(t *testing.T) {
	cases := []struct {
		state      ProtocolState
		wantEff    Effects
		wantNumber int
	}{
		{stDraft, Effects{}, 0},
		{stVote, Effects{ClosePollRound: true}, 0},
		{stSigning, Effects{VoidPrintCopy: true}, 14},
	}
	for name, edit := range slotEdits {
		for _, tc := range cases {
			t.Run(name+"/"+string(tc.state), func(t *testing.T) {
				p, s := reviewDraftIn(t, tc.state)
				rev, before := p.Revision(), len(p.Transitions())

				eff, err := edit(p, rev, 0, "amount")
				if err != nil {
					t.Fatalf("err = %v", err)
				}
				if eff != tc.wantEff {
					t.Fatalf("effects = %+v, want %+v", eff, tc.wantEff)
				}
				if p.State() != stDraft || p.Revision() != rev+1 || p.Number() != tc.wantNumber {
					t.Fatalf("state %q rev %d number %d, want draft rev %d number %d",
						p.State(), p.Revision(), p.Number(), rev+1, tc.wantNumber)
				}
				tr := p.Transitions()
				if tc.state == stDraft {
					if len(tr) != before {
						t.Fatalf("transitions = %+v, want none added for a draft", tr[before:])
					}
				} else {
					if len(tr) != before+1 {
						t.Fatalf("transitions = %d, want %d", len(tr), before+1)
					}
					want := ProtocolTransition{From: tc.state, To: stDraft, Revision: rev + 1, ActorID: reviewer, At: t1}
					if tr[len(tr)-1] != want {
						t.Fatalf("transition = %+v, want %+v", tr[len(tr)-1], want)
					}
				}
				if !s.Confirmed() || s.ConfirmedBy() != reviewer || !s.ConfirmedAt().Equal(t1) {
					t.Fatalf("slot %q by %q at %v, want confirmed by %q at %v", s.Status(), s.ConfirmedBy(), s.ConfirmedAt(), reviewer, t1)
				}
				if len(p.Readiness()) != 0 {
					t.Fatalf("blockers = %+v, want none", p.Readiness())
				}
			})
		}
	}
}

// B1: ConfirmSlot keeps the value; CorrectSlot takes the person's own value.
func TestProtocol_TB_ConfirmKeepsValueCorrectReplacesIt(t *testing.T) {
	p, s := reviewDraftIn(t, stDraft)
	if _, err := p.ConfirmSlot(p.Revision(), 0, "amount", reviewer, t1); err != nil {
		t.Fatal(err)
	}
	if s.Value() != uah1500 || s.Origin() != originDocument || s.Source() != contrV1 {
		t.Fatalf("confirm changed the value: %+v %q %+v", s.Value(), s.Origin(), s.Source())
	}
	if _, ok := s.Alternative(); ok {
		t.Fatal("alternative kept after confirm, want cleared")
	}

	own := SlotValue{Kopiyky: 170_000, Currency: "UAH"}
	if _, err := p.CorrectSlot(p.Revision(), 0, "amount", own, actor, t1); err != nil {
		t.Fatal(err)
	}
	if s.Value() != own || s.Origin() != originHuman || s.Source() != (SlotSource{}) || s.ConfirmedBy() != actor {
		t.Fatalf("correct = %+v %q %+v by %q, want own human value by %q", s.Value(), s.Origin(), s.Source(), s.ConfirmedBy(), actor)
	}
}
