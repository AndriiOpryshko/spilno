package domain

import "time"

// SlotKind names what a decision slot holds.
type SlotKind string

// Slot kinds of a decision.
const (
	SlotKindAmount        SlotKind = "amount"
	SlotKindCounterparty  SlotKind = "counterparty"
	SlotKindFundingSource SlotKind = "funding_source"
	SlotKindDocumentRef   SlotKind = "document_ref"
	SlotKindDate          SlotKind = "date"
	SlotKindPerson        SlotKind = "person"
)

// SlotOrigin says where a slot value came from.
type SlotOrigin string

// Origins of a slot value.
const (
	SlotOriginAI       SlotOrigin = "ai"
	SlotOriginDocument SlotOrigin = "document"
	SlotOriginHuman    SlotOrigin = "human"
)

// SlotStatus is the confirmation state of a slot.
type SlotStatus string

// Slot statuses.
const (
	SlotStatusUnconfirmed SlotStatus = "unconfirmed"
	SlotStatusConfirmed   SlotStatus = "confirmed"
	SlotStatusNeedsReview SlotStatus = "needs_review"
)

// SlotValue is the content of a slot: Kopiyky and Currency for an amount,
// Text for every other kind.
type SlotValue struct {
	Kopiyky  int64
	Currency string
	Text     string
}

// SlotSource cites the document fragment a value was taken from; it is zero
// for a value a person entered.
type SlotSource struct {
	DocumentID, VersionID, Fragment string
}

// Slot is one field of a decision that a person must confirm before it counts
// (AC-21); a confirmed value is never overwritten silently (AC-28b, AC-09).
type Slot struct {
	kind        SlotKind
	origin      SlotOrigin
	status      SlotStatus
	value       SlotValue
	source      SlotSource
	alternative *SlotValue
	confirmedBy string
	confirmedAt time.Time
}

// ProposeSlot returns an unconfirmed slot pre-filled with v from src.
func ProposeSlot(kind SlotKind, origin SlotOrigin, v SlotValue, src SlotSource) *Slot {
	return &Slot{kind: kind, origin: origin, status: SlotStatusUnconfirmed, value: v, source: src}
}

// Kind returns what the slot holds.
func (s *Slot) Kind() SlotKind { return s.kind }

// Origin returns where the current value came from.
func (s *Slot) Origin() SlotOrigin { return s.origin }

// Status returns the confirmation state.
func (s *Slot) Status() SlotStatus { return s.status }

// Value returns the current value.
func (s *Slot) Value() SlotValue { return s.value }

// Source returns the citation of the current value.
func (s *Slot) Source() SlotSource { return s.source }

// Alternative returns a later value kept beside a confirmed one, if any.
func (s *Slot) Alternative() (SlotValue, bool) {
	if s.alternative == nil {
		return SlotValue{}, false
	}
	return *s.alternative, true
}

// Confirmed reports whether the slot is confirmed.
func (s *Slot) Confirmed() bool { return s.status == SlotStatusConfirmed }

// ConfirmedBy returns who last confirmed the slot.
func (s *Slot) ConfirmedBy() string { return s.confirmedBy }

// ConfirmedAt returns when the slot was last confirmed.
func (s *Slot) ConfirmedAt() time.Time { return s.confirmedAt }

// Confirm accepts the current value on behalf of actorID at at and ends any review.
func (s *Slot) Confirm(actorID string, at time.Time) {
	s.status = SlotStatusConfirmed
	s.alternative = nil
	s.confirmedBy = actorID
	s.confirmedAt = at
}

// Correct replaces the value with the person's own v and confirms it.
func (s *Slot) Correct(v SlotValue, actorID string, at time.Time) {
	s.value = v
	s.origin = SlotOriginHuman
	s.source = SlotSource{}
	s.Confirm(actorID, at)
}

// Offer proposes a newer value: it replaces an unconfirmed one, and is kept as
// an alternative beside a confirmed or reviewed one, which then needs review.
func (s *Slot) Offer(v SlotValue, origin SlotOrigin, src SlotSource) {
	if s.status == SlotStatusUnconfirmed {
		s.value, s.origin, s.source = v, origin, src
		return
	}
	if v == s.value {
		return
	}
	s.alternative = &v
	s.status = SlotStatusNeedsReview
}

// DocumentRevised sends the slot to review when it cites documentID at a
// version other than versionID, and reports whether it did.
func (s *Slot) DocumentRevised(documentID, versionID string) bool {
	if s.source.DocumentID == "" || s.source.DocumentID != documentID || s.source.VersionID == versionID {
		return false
	}
	s.status = SlotStatusNeedsReview
	return true
}
