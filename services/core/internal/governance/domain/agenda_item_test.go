package domain

import (
	"errors"
	"testing"
)

// AC-03c: a second справа of the same meeting becomes a new agenda item on the existing draft.
func TestProtocol_AC03c_AddsAgendaItemsWhileDraft(t *testing.T) {
	p := NewProtocol(actor, t0)
	for i, c := range []string{"case-1", "case-2"} {
		if _, err := p.AddAgendaItem(p.Revision(), NewAgendaItem(c, "Ремонт покрівлі"), actor, t1); err != nil {
			t.Fatalf("add item %d: %v", i, err)
		}
	}
	items := p.AgendaItems()
	if len(items) != 2 || items[0].CaseID() != "case-1" || items[1].CaseID() != "case-2" {
		t.Fatalf("agenda = %+v, want case-1 then case-2", items)
	}
	if p.State() != ProtocolDraft || p.Revision() != 3 {
		t.Fatalf("state/revision = %q/%d, want draft/3", p.State(), p.Revision())
	}
}

// AC-03c: adding an item to a ready draft is a change that returns it to draft.
func TestProtocol_AC03c_AddingToAReadyDraftReturnsItToDraft(t *testing.T) {
	p := NewProtocol(actor, t0)
	if _, err := p.MarkReadyForVote(p.Revision(), actor, t0); err != nil {
		t.Fatal(err)
	}
	eff, err := p.AddAgendaItem(p.Revision(), NewAgendaItem("case-2", "Прибирання"), actor, t1)
	if err != nil {
		t.Fatalf("add item: %v", err)
	}
	if p.State() != ProtocolDraft || !eff.ClosePollRound {
		t.Fatalf("state %q effects %+v, want draft and the poll round closed", p.State(), eff)
	}
}

// AC-03c: a stale revision or a locked protocol refuses a new item.
func TestProtocol_AC03c_AddAgendaItemGuards(t *testing.T) {
	p := NewProtocol(actor, t0)
	_, err := p.AddAgendaItem(p.Revision()+1, NewAgendaItem("c", "s"), actor, t1)
	var de *DomainError
	if !errors.As(err, &de) || de.Code != CodeRevisionConflict {
		t.Fatalf("err = %v, want %s", err, CodeRevisionConflict)
	}
	if len(p.AgendaItems()) != 0 {
		t.Fatal("item added despite the conflict")
	}
}

// AC-29: any member rewords; the wording keeps author and time, slots keep value and source.
func TestAgendaItem_AC29_RewordKeepsAuthorTimeAndSlots(t *testing.T) {
	item := NewAgendaItem("case-1", "Ремонт покрівлі")
	s := ProposeSlot(slotAmount, originDocument, uah1500, contrV1)
	s.Confirm(actor, t0)
	item.SetSlot("amount", s)

	item.Reword("Затвердити кошторис", actor, t0)
	item.Reword("Затвердити кошторис робіт", reviewer, t1)

	w, ok := item.Wording()
	if !ok || w.Text != "Затвердити кошторис робіт" || w.AuthorID != reviewer || !w.At.Equal(t1) {
		t.Fatalf("wording = %+v, %v; want the reviewer's at t1", w, ok)
	}
	if hist := item.Wordings(); len(hist) != 2 || hist[0].AuthorID != actor {
		t.Fatalf("history = %+v, want both wordings, oldest first", hist)
	}
	got, _ := item.Slot("amount")
	if !got.Confirmed() || got.Value() != uah1500 || got.Source() != contrV1 {
		t.Fatalf("slot after reword: %q %+v %+v", got.Status(), got.Value(), got.Source())
	}
}

// AC-29: a fresh item has no wording yet.
func TestAgendaItem_AC29_NoWordingUntilWritten(t *testing.T) {
	if _, ok := NewAgendaItem("c", "s").Wording(); ok {
		t.Fatal("fresh item has a wording")
	}
}
