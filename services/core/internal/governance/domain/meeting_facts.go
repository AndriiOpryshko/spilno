package domain

import (
	"strconv"
	"strings"
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

// enteredFacts are the meeting facts stored at lock with who entered them and when.
type enteredFacts struct {
	facts MeetingFacts
	by    string
	at    time.Time
}

// Validate reports every missing fact at once for a protocol with agendaItems items.
// Quorum, majority and signatories are not checked (Stage 1, AC-14).
func (f MeetingFacts) Validate(agendaItems int) error {
	var fields []FieldIssue
	factsMissing := false
	if f.MembersPresent <= 0 {
		fields = append(fields, FieldIssue{Pointer: "/members_present", Detail: "members present is required"})
		factsMissing = true
	}
	for i := 0; i < agendaItems; i++ {
		if _, ok := f.Votes[i]; !ok {
			fields = append(fields, FieldIssue{Pointer: "/votes/" + strconv.Itoa(i), Detail: "vote count is required"})
			factsMissing = true
		}
	}
	if f.HasCorrections && strings.TrimSpace(f.CorrectionNote) == "" {
		fields = append(fields, FieldIssue{Pointer: "/correction_note", Detail: "correction note is required"})
	}
	switch {
	case factsMissing:
		return &CodedFieldsError{Code: CodeFactsMissing, Fields: fields}
	case len(fields) > 0:
		return &CodedFieldsError{Code: CodeNoteRequired, Fields: fields}
	}
	return nil
}

// SignWithFacts locks a printed protocol with the meeting facts entered by actorID at at.
// The revision guard and the printed state are checked first; incomplete facts change nothing.
func (p *Protocol) SignWithFacts(expectedRevision int, f MeetingFacts, actorID string, at time.Time) (Effects, error) {
	if err := p.guard(expectedRevision); err != nil {
		return Effects{}, err
	}
	if p.state != ProtocolReadyForSigning {
		return Effects{}, &DomainError{Code: CodeNotReadyForSigning}
	}
	if err := f.Validate(len(p.agenda)); err != nil {
		return Effects{}, err
	}
	stored := f
	stored.Votes = make(map[int]VoteCount, len(f.Votes))
	for i, v := range f.Votes {
		stored.Votes[i] = v
	}
	p.facts = &enteredFacts{facts: stored, by: actorID, at: at}
	p.move(ProtocolSigned, actorID, at)
	return Effects{}, nil
}

// MeetingFacts returns the facts stored at lock, who entered them and when; ok is false before lock.
func (p *Protocol) MeetingFacts() (f MeetingFacts, enteredBy string, enteredAt time.Time, ok bool) {
	if p.facts == nil {
		return MeetingFacts{}, "", time.Time{}, false
	}
	return p.facts.facts, p.facts.by, p.facts.at, true
}
