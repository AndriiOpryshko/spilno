package domain

import (
	"testing"
	"time"
)

// Expected API (to be implemented in slot.go):
//
//	type SlotKind string   // "amount","counterparty","funding_source","document_ref","date","person"
//	type SlotOrigin string // "ai","document","human"
//	type SlotStatus string // "unconfirmed","confirmed","needs_review"
//	type SlotValue struct {
//		Kopiyky  int64  // amount only
//		Currency string // amount only
//		Text     string // every other kind
//	}
//	type SlotSource struct{ DocumentID, VersionID, Fragment string } // zero for a human value
//	func ProposeSlot(kind SlotKind, origin SlotOrigin, v SlotValue, src SlotSource) *Slot // unconfirmed
//	func (s *Slot) Kind() SlotKind
//	func (s *Slot) Origin() SlotOrigin
//	func (s *Slot) Status() SlotStatus
//	func (s *Slot) Value() SlotValue
//	func (s *Slot) Source() SlotSource
//	func (s *Slot) Alternative() (SlotValue, bool)  // a later value kept beside a confirmed one
//	func (s *Slot) Confirmed() bool                 // status == confirmed
//	func (s *Slot) ConfirmedBy() string
//	func (s *Slot) ConfirmedAt() time.Time
//	func (s *Slot) Confirm(actorID string, at time.Time)
//		// keeps Value, drops Alternative, records actor and time
//	func (s *Slot) Correct(v SlotValue, actorID string, at time.Time)
//		// Value = v, Origin = human, Source zero, drops Alternative, confirmed
//	func (s *Slot) Offer(v SlotValue, origin SlotOrigin, src SlotSource)
//		// unconfirmed: replaces Value, Origin and Source, stays unconfirmed
//		// confirmed or needs_review: Value untouched, v kept as Alternative, needs_review
//		// v equal to the confirmed Value: no change
//	func (s *Slot) DocumentRevised(documentID, versionID string) bool
//		// true and needs_review when the slot cites documentID at another version;
//		// a slot without a source or citing another document is untouched (false)

const (
	slotAmount       = SlotKind("amount")
	slotCounterparty = SlotKind("counterparty")

	originAI       = SlotOrigin("ai")
	originDocument = SlotOrigin("document")
	originHuman    = SlotOrigin("human")

	slotUnconfirmed = SlotStatus("unconfirmed")
	slotConfirmed   = SlotStatus("confirmed")
	slotNeedsReview = SlotStatus("needs_review")

	contractDoc = "doc-contract"
)

var (
	uah1500  = SlotValue{Kopiyky: 150_000, Currency: "UAH"}
	uah1800  = SlotValue{Kopiyky: 180_000, Currency: "UAH"}
	contrV1  = SlotSource{DocumentID: contractDoc, VersionID: "v1", Fragment: "Загальна вартість робіт 1 500,00 грн"}
	reviewer = "member-2"
)

// confirmedAmount is an amount slot proposed from the contract and confirmed by actor at t0.
func confirmedAmount(t *testing.T) *Slot {
	t.Helper()
	s := ProposeSlot(slotAmount, originDocument, uah1500, contrV1)
	s.Confirm(actor, t0)
	if !s.Confirmed() {
		t.Fatalf("status after Confirm = %q, want confirmed", s.Status())
	}
	return s
}

// AC-21: a value taken from a document is pre-filled with its fragment and stays unconfirmed.
func TestSlot_ProposedFromDocumentIsUnconfirmedWithItsSource(t *testing.T) {
	s := ProposeSlot(slotAmount, originDocument, uah1500, contrV1)

	if s.Kind() != slotAmount || s.Origin() != originDocument {
		t.Fatalf("kind/origin = %q/%q, want amount/document", s.Kind(), s.Origin())
	}
	if s.Status() != slotUnconfirmed || s.Confirmed() {
		t.Fatalf("status = %q, confirmed = %v; want unconfirmed, false", s.Status(), s.Confirmed())
	}
	if s.Value() != uah1500 || s.Source() != contrV1 {
		t.Fatalf("value/source = %+v/%+v, want %+v/%+v", s.Value(), s.Source(), uah1500, contrV1)
	}
	if _, ok := s.Alternative(); ok {
		t.Fatal("a fresh slot has an alternative")
	}
	if s.ConfirmedBy() != "" || !s.ConfirmedAt().IsZero() {
		t.Fatalf("confirmation = %q at %v, want none", s.ConfirmedBy(), s.ConfirmedAt())
	}
}

// AC-21: a person confirms the proposed value; the confirmation keeps who and when.
func TestSlot_ConfirmRecordsPersonAndTime(t *testing.T) {
	s := ProposeSlot(slotCounterparty, originAI, SlotValue{Text: "ТОВ «Тест Буд»"}, contrV1)
	s.Confirm(reviewer, t1)

	if !s.Confirmed() || s.Status() != slotConfirmed {
		t.Fatalf("status = %q, want confirmed", s.Status())
	}
	if s.ConfirmedBy() != reviewer || !s.ConfirmedAt().Equal(t1) {
		t.Fatalf("confirmed by %q at %v, want %q at %v", s.ConfirmedBy(), s.ConfirmedAt(), reviewer, t1)
	}
	if s.Value().Text != "ТОВ «Тест Буд»" || s.Origin() != originAI {
		t.Fatalf("confirm changed value or origin: %+v, %q", s.Value(), s.Origin())
	}
}

// AC-21: correcting is the other way to confirm; the value becomes the person's own.
func TestSlot_CorrectConfirmsTheHumanValue(t *testing.T) {
	s := ProposeSlot(slotAmount, originDocument, uah1500, contrV1)
	s.Correct(uah1800, reviewer, t1)

	if !s.Confirmed() || s.Value() != uah1800 {
		t.Fatalf("after Correct: status %q value %+v, want confirmed %+v", s.Status(), s.Value(), uah1800)
	}
	if s.Origin() != originHuman || s.Source() != (SlotSource{}) {
		t.Fatalf("after Correct: origin %q source %+v, want human and no source", s.Origin(), s.Source())
	}
	if s.ConfirmedBy() != reviewer || !s.ConfirmedAt().Equal(t1) {
		t.Fatalf("confirmed by %q at %v, want %q at %v", s.ConfirmedBy(), s.ConfirmedAt(), reviewer, t1)
	}
}

// AC-21: until someone confirms, a newer proposal simply replaces the older one.
func TestSlot_OfferReplacesAnUnconfirmedValue(t *testing.T) {
	s := ProposeSlot(slotAmount, originAI, uah1500, contrV1)
	v2 := SlotSource{DocumentID: "doc-invoice", VersionID: "v1", Fragment: "До сплати 1 800,00 грн"}
	s.Offer(uah1800, originDocument, v2)

	if s.Status() != slotUnconfirmed {
		t.Fatalf("status = %q, want unconfirmed", s.Status())
	}
	if s.Value() != uah1800 || s.Origin() != originDocument || s.Source() != v2 {
		t.Fatalf("after Offer: %+v %q %+v, want %+v document %+v", s.Value(), s.Origin(), s.Source(), uah1800, v2)
	}
	if _, ok := s.Alternative(); ok {
		t.Fatal("an unconfirmed slot keeps an alternative")
	}
}

// AC-28b: a later value never overwrites a confirmed one; it waits beside it.
func TestSlot_OfferNeverOverwritesAConfirmedValue(t *testing.T) {
	s := confirmedAmount(t)
	s.Offer(uah1800, originAI, SlotSource{})

	if s.Value() != uah1500 || s.Source() != contrV1 || s.Origin() != originDocument {
		t.Fatalf("confirmed value overwritten: %+v %+v %q", s.Value(), s.Source(), s.Origin())
	}
	alt, ok := s.Alternative()
	if !ok || alt != uah1800 {
		t.Fatalf("alternative = %+v, %v; want %+v, true", alt, ok, uah1800)
	}
	if s.Status() != slotNeedsReview || s.Confirmed() {
		t.Fatalf("status = %q, confirmed = %v; want needs_review, false", s.Status(), s.Confirmed())
	}
}

// AC-28b: offering the value that is already confirmed is not a change.
func TestSlot_OfferOfTheSameValueKeepsItConfirmed(t *testing.T) {
	s := confirmedAmount(t)
	s.Offer(uah1500, originAI, SlotSource{})

	if !s.Confirmed() {
		t.Fatalf("status = %q, want confirmed", s.Status())
	}
	if _, ok := s.Alternative(); ok {
		t.Fatal("the same value was stored as an alternative")
	}
}

// AC-28b: a member resolves the review by keeping the confirmed value or taking the new one.
func TestSlot_ReviewEndsWhenAMemberChoosesOneValue(t *testing.T) {
	t.Run("keep confirmed", func(t *testing.T) {
		s := confirmedAmount(t)
		s.Offer(uah1800, originAI, SlotSource{})
		s.Confirm(reviewer, t1)

		if !s.Confirmed() || s.Value() != uah1500 {
			t.Fatalf("status %q value %+v, want confirmed %+v", s.Status(), s.Value(), uah1500)
		}
		if _, ok := s.Alternative(); ok {
			t.Fatal("alternative kept after the review ended")
		}
		if s.ConfirmedBy() != reviewer || !s.ConfirmedAt().Equal(t1) {
			t.Fatalf("confirmed by %q at %v, want %q at %v", s.ConfirmedBy(), s.ConfirmedAt(), reviewer, t1)
		}
	})
	t.Run("take the new one", func(t *testing.T) {
		s := confirmedAmount(t)
		s.Offer(uah1800, originAI, SlotSource{})
		alt, _ := s.Alternative()
		s.Correct(alt, reviewer, t1)

		if !s.Confirmed() || s.Value() != uah1800 {
			t.Fatalf("status %q value %+v, want confirmed %+v", s.Status(), s.Value(), uah1800)
		}
		if _, ok := s.Alternative(); ok {
			t.Fatal("alternative kept after the review ended")
		}
	})
}

// AC-09: a newer version of the cited document sends a confirmed slot back to review.
func TestSlot_NewerVersionOfTheCitedDocumentNeedsReview(t *testing.T) {
	s := confirmedAmount(t)

	if !s.DocumentRevised(contractDoc, "v2") {
		t.Fatal("DocumentRevised = false for the cited document, want true")
	}
	if s.Status() != slotNeedsReview || s.Confirmed() {
		t.Fatalf("status = %q, want needs_review", s.Status())
	}
	if s.Value() != uah1500 {
		t.Fatalf("value = %+v, want the confirmed %+v kept", s.Value(), uah1500)
	}
}

// AC-09: only slots that cite the changed document are affected.
func TestSlot_DocumentRevisedLeavesOtherSlotsAlone(t *testing.T) {
	cases := map[string]struct {
		slot       func(t *testing.T) *Slot
		doc, ver   string
		wantStatus SlotStatus
	}{
		"another document": {confirmedAmount, "doc-invoice", "v2", slotConfirmed},
		"same version":     {confirmedAmount, contractDoc, "v1", slotConfirmed},
		"human value without source": {func(t *testing.T) *Slot {
			s := ProposeSlot(slotAmount, originDocument, uah1500, contrV1)
			s.Correct(uah1800, actor, t0)
			return s
		}, contractDoc, "v2", slotConfirmed},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			s := tc.slot(t)
			if s.DocumentRevised(tc.doc, tc.ver) {
				t.Fatal("DocumentRevised = true, want false")
			}
			if s.Status() != tc.wantStatus {
				t.Fatalf("status = %q, want %q", s.Status(), tc.wantStatus)
			}
		})
	}
}

// AC-09 + AC-21: confirming again after a document revision closes the review.
func TestSlot_ConfirmAfterDocumentRevisionClosesTheReview(t *testing.T) {
	s := confirmedAmount(t)
	s.DocumentRevised(contractDoc, "v2")
	at := t1.Add(time.Hour)
	s.Confirm(reviewer, at)

	if !s.Confirmed() || s.ConfirmedBy() != reviewer || !s.ConfirmedAt().Equal(at) {
		t.Fatalf("status %q by %q at %v, want confirmed by %q at %v", s.Status(), s.ConfirmedBy(), s.ConfirmedAt(), reviewer, at)
	}
}
