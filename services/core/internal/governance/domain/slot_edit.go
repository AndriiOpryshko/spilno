package domain

import (
	"errors"
	"time"
)

// ErrSlotNotFound refuses a slot edit addressed to an agenda item index or a
// slot key the draft does not hold; nothing is changed.
var ErrSlotNotFound = errors.New("governance: slot not found")

// ConfirmSlot confirms the slot key of agenda item item (0-based) on behalf of
// actorID (AC-21). Like any content change it passes the revision guard
// (ADR-0009), bumps the revision and returns a ready protocol to draft
// (AC-05d, AC-14b).
func (p *Protocol) ConfirmSlot(expectedRevision int, item int, key string, actorID string, at time.Time) (Effects, error) {
	s, err := p.editableSlot(expectedRevision, item, key)
	if err != nil {
		return Effects{}, err
	}
	s.Confirm(actorID, at)
	return p.reopen(actorID, at), nil
}

// CorrectSlot replaces the slot key of agenda item item (0-based) with the
// person's own v and confirms it, under the same rules as ConfirmSlot.
func (p *Protocol) CorrectSlot(expectedRevision int, item int, key string, v SlotValue, actorID string, at time.Time) (Effects, error) {
	s, err := p.editableSlot(expectedRevision, item, key)
	if err != nil {
		return Effects{}, err
	}
	s.Correct(v, actorID, at)
	return p.reopen(actorID, at), nil
}

// editableSlot runs the protocol guard first, then resolves the slot.
func (p *Protocol) editableSlot(expectedRevision int, item int, key string) (*Slot, error) {
	if err := p.guard(expectedRevision); err != nil {
		return nil, err
	}
	if item < 0 || item >= len(p.agenda) {
		return nil, ErrSlotNotFound
	}
	s, ok := p.agenda[item].Slot(key)
	if !ok || s == nil {
		return nil, ErrSlotNotFound
	}
	return s, nil
}
