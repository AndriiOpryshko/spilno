package domain

import (
	"errors"
	"time"
)

// VoteCount is the tally written on the paper for one agenda item.
type VoteCount struct{ For, Against, Abstained int }

// MeetingFacts are the facts written on the signed paper (AC-14).
type MeetingFacts struct {
	MembersPresent int
	Votes          map[int]VoteCount // keyed by agenda item index, 0-based
	HasCorrections bool
	CorrectionNote string
}

var errNotImplemented = errors.New("not implemented")

// Validate reports every missing fact at once for a protocol with agendaItems items.
func (f MeetingFacts) Validate(agendaItems int) error { return errNotImplemented }

// SignWithFacts locks a printed protocol with the meeting facts entered by actorID at at.
func (p *Protocol) SignWithFacts(expectedRevision int, f MeetingFacts, actorID string, at time.Time) (Effects, error) {
	return Effects{}, errNotImplemented
}

// MeetingFacts returns the facts stored at lock, who entered them and when; ok is false before lock.
func (p *Protocol) MeetingFacts() (f MeetingFacts, enteredBy string, enteredAt time.Time, ok bool) {
	return MeetingFacts{}, "", time.Time{}, false
}
