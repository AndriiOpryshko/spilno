package domain

import (
	"errors"
	"reflect"
	"testing"
)

func draftWith(t *testing.T, items ...*AgendaItem) *Protocol {
	t.Helper()
	p := NewProtocol(actor, t0)
	for _, it := range items {
		if _, err := p.AddAgendaItem(p.Revision(), it, actor, t0); err != nil {
			t.Fatal(err)
		}
	}
	return p
}

func pointers(bs []Blocker) []string {
	out := make([]string, len(bs))
	for i, b := range bs {
		out[i] = b.Pointer
	}
	return out
}

// AC-04: a draft without any agenda item is not ready.
func TestReadiness_AC04_NoAgendaItem(t *testing.T) {
	p := NewProtocol(actor, t0)
	got := p.Readiness()
	if len(got) != 1 || got[0].Code != CodeDraftIncomplete || got[0].Pointer != "/agenda_items" {
		t.Fatalf("blockers = %+v, want one draft_incomplete at /agenda_items", got)
	}
	var ce *CodedFieldsError
	if err := p.EnsureReadyForSigning(); !errors.As(err, &ce) || ce.Code != CodeDraftIncomplete {
		t.Fatalf("EnsureReadyForSigning = %v, want %s", err, CodeDraftIncomplete)
	}
}

// AC-04 + AC-21: every missing part and unconfirmed slot is named at once.
func TestReadiness_AC04_AC21_NamesEveryBlockerAtOnce(t *testing.T) {
	a := NewAgendaItem("case-1", "Ремонт покрівлі")
	a.SetSlot("amount", ProposeSlot(slotAmount, originDocument, uah1500, contrV1))
	a.SetSlot("counterparty", ProposeSlot(slotCounterparty, originDocument, SlotValue{Text: "ТОВ «Тест Буд»"}, contrV1))
	b := NewAgendaItem("case-2", "  ")
	p := draftWith(t, a, b)

	want := []string{"/agenda_items/0/slots/amount", "/agenda_items/0/slots/counterparty", "/agenda_items/1/subject"}
	if got := pointers(p.Readiness()); !reflect.DeepEqual(got, want) {
		t.Fatalf("pointers = %v, want %v", got, want)
	}
	var ce *CodedFieldsError
	if err := p.EnsureReadyForSigning(); !errors.As(err, &ce) || ce.Code != CodeDraftIncomplete || len(ce.Fields) != 3 {
		t.Fatalf("EnsureReadyForSigning = %v, want draft_incomplete with 3 fields", err)
	}
}

// AC-21: an unconfirmed slot alone blocks with slot_unconfirmed; confirming clears it.
func TestReadiness_AC21_UnconfirmedSlotBlocksUntilConfirmed(t *testing.T) {
	a := NewAgendaItem("case-1", "Ремонт покрівлі")
	s := ProposeSlot(slotAmount, originAI, uah1500, contrV1)
	a.SetSlot("amount", s)
	p := draftWith(t, a)

	var ce *CodedFieldsError
	if err := p.EnsureReadyForSigning(); !errors.As(err, &ce) || ce.Code != CodeSlotUnconfirmed {
		t.Fatalf("EnsureReadyForSigning = %v, want %s", err, CodeSlotUnconfirmed)
	}
	s.Confirm(reviewer, t1)
	if err := p.EnsureReadyForSigning(); err != nil {
		t.Fatalf("EnsureReadyForSigning after confirm = %v, want nil", err)
	}
	if got := p.Readiness(); len(got) != 0 {
		t.Fatalf("blockers = %+v, want none", got)
	}
}

// AC-28b: a confirmed slot with a later value beside it needs review and blocks again.
func TestReadiness_AC28b_SlotNeedingReviewBlocks(t *testing.T) {
	a := NewAgendaItem("case-1", "Ремонт покрівлі")
	s := confirmedAmount(t)
	a.SetSlot("amount", s)
	p := draftWith(t, a)
	s.Offer(uah1800, originAI, SlotSource{})

	got := p.Readiness()
	if len(got) != 1 || got[0].Code != CodeSlotUnconfirmed || got[0].Pointer != "/agenda_items/0/slots/amount" {
		t.Fatalf("blockers = %+v, want the amount slot", got)
	}
	if s.Value() != uah1500 {
		t.Fatalf("confirmed value overwritten: %+v", s.Value())
	}
}

// AC-09: a newer version of the cited document marks the citing slots for review and returns the draft to draft.
func TestProtocol_AC09_DocumentRevisedMarksCitingSlotsAndReopens(t *testing.T) {
	a := NewAgendaItem("case-1", "Ремонт покрівлі")
	cited := confirmedAmount(t)
	other := ProposeSlot(slotCounterparty, originDocument, SlotValue{Text: "ТОВ «Тест Буд»"},
		SlotSource{DocumentID: "doc-invoice", VersionID: "v1"})
	other.Confirm(actor, t0)
	a.SetSlot("amount", cited)
	a.SetSlot("counterparty", other)
	p := draftWith(t, a)
	if _, err := p.MarkReadyForSigning(p.Revision(), actor, t0, 7); err != nil {
		t.Fatal(err)
	}
	rev := p.Revision()

	eff, changed := p.DocumentRevised(contractDoc, "v2", actor, t1)
	if !changed || !eff.VoidPrintCopy || p.State() != ProtocolDraft || p.Revision() != rev+1 {
		t.Fatalf("changed %v effects %+v state %q rev %d", changed, eff, p.State(), p.Revision())
	}
	if cited.Status() != slotNeedsReview || !other.Confirmed() {
		t.Fatalf("statuses = %q/%q, want needs_review/confirmed", cited.Status(), other.Status())
	}
	if got := pointers(p.Readiness()); !reflect.DeepEqual(got, []string{"/agenda_items/0/slots/amount"}) {
		t.Fatalf("blockers = %v", got)
	}
}

// AC-09: a revision of a document no slot cites changes nothing.
func TestProtocol_AC09_UncitedDocumentRevisionIsNoChange(t *testing.T) {
	a := NewAgendaItem("case-1", "Ремонт покрівлі")
	a.SetSlot("amount", confirmedAmount(t))
	p := draftWith(t, a)
	rev := p.Revision()

	if _, changed := p.DocumentRevised("doc-other", "v2", actor, t1); changed || p.Revision() != rev {
		t.Fatalf("changed %v revision %d, want no change at %d", changed, p.Revision(), rev)
	}
}

// unreadyDraft holds one agenda item whose amount slot is only proposed.
func unreadyDraft(t *testing.T) (*Protocol, *Slot) {
	t.Helper()
	a := NewAgendaItem("case-1", "Ремонт покрівлі")
	s := ProposeSlot(slotAmount, originAI, uah1500, contrV1)
	a.SetSlot("amount", s)
	return draftWith(t, a), s
}

// AC-21, AC-04, AC-14c: marking готовий до підпису is refused with exactly the
// readiness error while any blocker is left, and the refusal changes nothing.
func TestProtocol_AC21_MarkReadyForSigningRefusesUnreadyDraft(t *testing.T) {
	cases := map[string]struct {
		build    func(t *testing.T) *Protocol
		wantCode string
	}{
		"draft with an unconfirmed slot": {func(t *testing.T) *Protocol { p, _ := unreadyDraft(t); return p }, CodeSlotUnconfirmed},
		"draft without an agenda item":   {func(t *testing.T) *Protocol { return NewProtocol(actor, t0) }, CodeDraftIncomplete},
		"draft with an empty subject":    {func(t *testing.T) *Protocol { return draftWith(t, NewAgendaItem("case-1", " ")) }, CodeDraftIncomplete},
		"ready_for_vote with an unconfirmed slot": {func(t *testing.T) *Protocol {
			p, _ := unreadyDraft(t)
			if _, err := p.MarkReadyForVote(p.Revision(), actor, t0); err != nil {
				t.Fatal(err)
			}
			return p
		}, CodeSlotUnconfirmed},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			p := tc.build(t)
			want := p.EnsureReadyForSigning()
			state, rev, transitions := p.State(), p.Revision(), p.Transitions()

			eff, err := p.MarkReadyForSigning(p.Revision(), actor, t1, 14)

			var ce *CodedFieldsError
			if !errors.As(err, &ce) || ce.Code != tc.wantCode {
				t.Fatalf("err = %v (%T), want *CodedFieldsError %s", err, err, tc.wantCode)
			}
			if !reflect.DeepEqual(err, want) {
				t.Errorf("err = %+v, want exactly EnsureReadyForSigning() = %+v", err, want)
			}
			if eff != (Effects{}) {
				t.Errorf("effects = %+v, want none on a refusal", eff)
			}
			if p.State() != state || p.Revision() != rev || p.Number() != 0 || !reflect.DeepEqual(p.Transitions(), transitions) {
				t.Errorf("refused print changed the protocol: state=%s rev=%d number=%d transitions=%d",
					p.State(), p.Revision(), p.Number(), len(p.Transitions()))
			}
		})
	}
}

// ADR-0009: the revision and lock guards and the already-printed check answer
// before the readiness gate; the readiness gate answers before the number check.
func TestProtocol_MarkReadyForSigningRefusalOrder(t *testing.T) {
	t.Run("stale revision on an unready draft is a conflict", func(t *testing.T) {
		p, _ := unreadyDraft(t)
		_, err := p.MarkReadyForSigning(p.Revision()-1, actor, t1, 14)
		if got := domainCode(t, err); got != codeConflict {
			t.Fatalf("code = %q, want %q", got, codeConflict)
		}
	})
	t.Run("withdrawn unready draft is not allowed", func(t *testing.T) {
		p, _ := unreadyDraft(t)
		if _, err := p.Withdraw(p.Revision(), actor, t0, "причина"); err != nil {
			t.Fatal(err)
		}
		_, err := p.MarkReadyForSigning(p.Revision(), actor, t1, 14)
		if got := domainCode(t, err); got != codeNotAllowed {
			t.Fatalf("code = %q, want %q", got, codeNotAllowed)
		}
	})
	printedThenUnready := func(t *testing.T) *Protocol {
		t.Helper()
		a := NewAgendaItem("case-1", "Ремонт покрівлі")
		s := confirmedAmount(t)
		a.SetSlot("amount", s)
		p := draftWith(t, a)
		if _, err := p.MarkReadyForSigning(p.Revision(), actor, t0, 14); err != nil {
			t.Fatal(err)
		}
		s.Offer(uah1800, originAI, SlotSource{}) // needs review, protocol not reopened
		if len(p.Readiness()) == 0 {
			t.Fatal("setup: want a blocker after the offer")
		}
		return p
	}
	t.Run("already printed protocol is not allowed before readiness", func(t *testing.T) {
		p := printedThenUnready(t)
		_, err := p.MarkReadyForSigning(p.Revision(), actor, t1, 15)
		if got := domainCode(t, err); got != codeNotAllowed {
			t.Fatalf("code = %q, want %q", got, codeNotAllowed)
		}
	})
	t.Run("signed protocol is locked before readiness", func(t *testing.T) {
		p := printedThenUnready(t)
		if _, err := p.Sign(p.Revision(), actor, t1); err != nil {
			t.Fatal(err)
		}
		_, err := p.MarkReadyForSigning(p.Revision(), actor, t1, 15)
		if got := domainCode(t, err); got != codeLocked {
			t.Fatalf("code = %q, want %q", got, codeLocked)
		}
	})
	t.Run("readiness before a missing number", func(t *testing.T) {
		p, _ := unreadyDraft(t)
		_, err := p.MarkReadyForSigning(p.Revision(), actor, t1, 0)
		var ce *CodedFieldsError
		if !errors.As(err, &ce) || ce.Code != CodeSlotUnconfirmed {
			t.Fatalf("err = %v (%T), want *CodedFieldsError %s", err, err, CodeSlotUnconfirmed)
		}
	})
}

// AC-21 + AC-33: once every slot is confirmed the draft prints as before, from
// чернетка and from готовий до голосування (closing the round); the number is
// taken on the first print only.
func TestProtocol_AC21_ConfirmedDraftPrints(t *testing.T) {
	for _, from := range []ProtocolState{stDraft, stVote} {
		t.Run(string(from), func(t *testing.T) {
			p, s := unreadyDraft(t)
			s.Confirm(reviewer, t0)
			if from == stVote {
				if _, err := p.MarkReadyForVote(p.Revision(), actor, t0); err != nil {
					t.Fatal(err)
				}
			}
			eff, err := p.MarkReadyForSigning(p.Revision(), actor, t1, 14)
			if err != nil {
				t.Fatalf("err = %v, want nil", err)
			}
			if want := (Effects{ClosePollRound: from == stVote}); eff != want {
				t.Errorf("effects = %+v, want %+v", eff, want)
			}
			if p.State() != stSigning || p.Number() != 14 {
				t.Fatalf("state %s number %d, want ready_for_signing 14", p.State(), p.Number())
			}
			if _, err := p.ReturnToDraft(p.Revision(), actor, t1); err != nil {
				t.Fatal(err)
			}
			if _, err := p.MarkReadyForSigning(p.Revision(), actor, t1, 99); err != nil || p.Number() != 14 {
				t.Fatalf("re-print err %v number %d, want nil and 14", err, p.Number())
			}
		})
	}
}
