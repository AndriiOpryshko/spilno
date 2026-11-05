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
