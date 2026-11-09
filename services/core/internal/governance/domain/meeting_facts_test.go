package domain

import (
	"errors"
	"reflect"
	"testing"
)

// Tests for T19a (docs/features/board-protocol/tasks/T19a-meeting-facts.md), one per AC.

const (
	codeFactsMissing = "governance.meeting_facts_missing"
	codeNoteRequired = "governance.correction_note_required"
)

func completeFacts(items int) MeetingFacts {
	f := MeetingFacts{MembersPresent: 5, Votes: map[int]VoteCount{}}
	for i := 0; i < items; i++ {
		f.Votes[i] = VoteCount{For: 4, Against: 1}
	}
	return f
}

// codedPointers asserts err is a *CodedFieldsError with code and returns its pointers in order.
func codedPointers(t *testing.T, err error, code string) []string {
	t.Helper()
	var ce *CodedFieldsError
	if !errors.As(err, &ce) {
		t.Fatalf("error = %v (%T), want *CodedFieldsError", err, err)
	}
	if ce.Code != code {
		t.Errorf("code = %q, want %q", ce.Code, code)
	}
	return ce.Pointers()
}

// AC-1: 3 items, members present 5, a vote count for every item, no corrections -> nil.
func TestMeetingFacts_AC1_CompleteFactsAreValid(t *testing.T) {
	if err := completeFacts(3).Validate(3); err != nil {
		t.Fatalf("Validate = %v, want nil", err)
	}
}

// AC-2: members present 0 and votes for items 0 and 2 missing -> every missing fact at once, in order.
func TestMeetingFacts_AC2_NamesEveryMissingFactAtOnce(t *testing.T) {
	f := completeFacts(3)
	f.MembersPresent = 0
	delete(f.Votes, 0)
	delete(f.Votes, 2)

	got := codedPointers(t, f.Validate(3), codeFactsMissing)
	want := []string{"/members_present", "/votes/0", "/votes/2"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("pointers = %v, want %v", got, want)
	}
}

// AC-3: corrections reported with an empty or blank note -> correction_note_required; with a note -> nil.
func TestMeetingFacts_AC3_CorrectionsNeedANote(t *testing.T) {
	for _, note := range []string{"", "   \t"} {
		f := completeFacts(2)
		f.HasCorrections = true
		f.CorrectionNote = note
		got := codedPointers(t, f.Validate(2), codeNoteRequired)
		if want := []string{"/correction_note"}; !reflect.DeepEqual(got, want) {
			t.Errorf("note %q: pointers = %v, want %v", note, got, want)
		}
	}
	f := completeFacts(2)
	f.HasCorrections = true
	f.CorrectionNote = "п. 2: сума виправлена від руки"
	if err := f.Validate(2); err != nil {
		t.Errorf("with a note: Validate = %v, want nil", err)
	}
}

// AC-4: members present 0 and corrections without a note -> meeting_facts_missing naming both.
func TestMeetingFacts_AC4_MissingFactsWinOverMissingNote(t *testing.T) {
	f := completeFacts(1)
	f.MembersPresent = 0
	f.HasCorrections = true

	got := codedPointers(t, f.Validate(1), codeFactsMissing)
	if want := []string{"/members_present", "/correction_note"}; !reflect.DeepEqual(got, want) {
		t.Errorf("pointers = %v, want %v", got, want)
	}
}

func printedWithTwoItems(t *testing.T) *Protocol {
	t.Helper()
	p := draftWith(t, NewAgendaItem("case-1", "Ремонт покрівлі"), NewAgendaItem("case-2", "Заміна ліфта"))
	if _, err := p.MarkReadyForSigning(p.Revision(), actor, t0, 14); err != nil {
		t.Fatalf("setup: %v", err)
	}
	return p
}

// AC-5: incomplete facts -> the Validate error, still ready for signing, nothing changed, no facts stored.
func TestProtocol_AC5_SignWithIncompleteFactsChangesNothing(t *testing.T) {
	p := printedWithTwoItems(t)
	rev, transitions := p.Revision(), p.Transitions()
	f := completeFacts(2)
	delete(f.Votes, 1)

	_, err := p.SignWithFacts(rev, f, actor, t1)

	got := codedPointers(t, err, codeFactsMissing)
	if want := []string{"/votes/1"}; !reflect.DeepEqual(got, want) {
		t.Errorf("pointers = %v, want %v", got, want)
	}
	if p.State() != stSigning || p.Revision() != rev || !reflect.DeepEqual(p.Transitions(), transitions) {
		t.Errorf("after refusal: state %q rev %d transitions %d, want %q %d %d",
			p.State(), p.Revision(), len(p.Transitions()), stSigning, rev, len(transitions))
	}
	if _, _, _, ok := p.MeetingFacts(); ok {
		t.Error("MeetingFacts ok = true after a refused lock")
	}
}

// AC-6: complete facts -> signed with the facts, who and when; the guard comes first.
func TestProtocol_AC6_SignWithFactsLocksAndKeepsThem(t *testing.T) {
	p := printedWithTwoItems(t)
	f := completeFacts(2)
	f.HasCorrections = true
	f.CorrectionNote = "п. 1: дата виправлена"

	if _, err := p.SignWithFacts(p.Revision(), f, "member-2", t1); err != nil {
		t.Fatalf("SignWithFacts = %v", err)
	}
	if p.State() != stSigned {
		t.Errorf("state = %q, want %q", p.State(), stSigned)
	}
	got, by, at, ok := p.MeetingFacts()
	if !ok || !reflect.DeepEqual(got, f) || by != "member-2" || !at.Equal(t1) {
		t.Errorf("MeetingFacts = %+v %q %v %v, want %+v member-2 %v true", got, by, at, ok, f, t1)
	}

	stale := printedWithTwoItems(t)
	if _, err := stale.SignWithFacts(stale.Revision()-1, completeFacts(2), actor, t1); domainCode(t, err) != codeConflict {
		t.Errorf("stale revision: code = %q, want %q", domainCode(t, err), codeConflict)
	}
	draft := draftWith(t, NewAgendaItem("case-1", "Ремонт покрівлі"))
	if _, err := draft.SignWithFacts(draft.Revision(), completeFacts(1), actor, t1); domainCode(t, err) != codeNotReady {
		t.Errorf("draft: code = %q, want %q", domainCode(t, err), codeNotReady)
	}
}
