package domain

import (
	"strings"
	"time"
	"unicode/utf8"
)

// Term is a board membership period; a zero End means open, a zero Start is missing.
type Term struct {
	Start, End time.Time
}

// CoversDate reports whether d falls in [Start, End], ignoring time of day.
func (t Term) CoversDate(d time.Time) bool {
	if t.Start.IsZero() {
		return false
	}
	dd := dateKey(d)
	if dd < dateKey(t.Start) {
		return false
	}
	return t.End.IsZero() || dd <= dateKey(t.End)
}

func dateKey(t time.Time) int {
	y, m, d := t.Date()
	return y*10000 + int(m)*100 + d
}

// BoardMemberEntry is one roster row.
type BoardMemberEntry struct {
	FullName    string
	SignInEmail string
	Term        Term
	Position    Position
	DecisionID  string
}

// Position is the seat an entry holds on the board.
type Position string

const (
	PositionHead   Position = "head"
	PositionMember Position = "member"
)

// PreviousHeadTermEnd is the end date of the outgoing head: the day before the new
// term starts, never before the entry's own first day.
func PreviousHeadTermEnd(newTermStart, previousEntryStart time.Time) time.Time {
	end := newTermStart.AddDate(0, 0, -1)
	if dateKey(end) < dateKey(previousEntryStart) {
		return previousEntryStart
	}
	return end
}

// Reseat returns the entry as head under the given decision, keeping its own term.
func Reseat(entry BoardMemberEntry, decisionID string) BoardMemberEntry {
	entry.Position = PositionHead
	entry.DecisionID = decisionID
	return entry
}

// Validate names every broken field at once.
func (e BoardMemberEntry) Validate() error {
	var issues []FieldIssue
	if n := utf8.RuneCountInString(e.FullName); strings.TrimSpace(e.FullName) == "" || n > 200 {
		issues = append(issues, FieldIssue{"/full_name", "must be 1 to 200 characters"})
	}
	if utf8.RuneCountInString(e.SignInEmail) > 320 {
		issues = append(issues, FieldIssue{"/sign_in_email", "must be at most 320 characters"})
	}
	if e.Term.Start.IsZero() {
		issues = append(issues, FieldIssue{"/term_start", "is required"})
	}
	if !e.Term.End.IsZero() && !e.Term.Start.IsZero() && dateKey(e.Term.End) < dateKey(e.Term.Start) {
		issues = append(issues, FieldIssue{"/term_end", "must not be before term_start"})
	}
	switch e.Position {
	case "", PositionHead, PositionMember:
	default:
		issues = append(issues, FieldIssue{"/position", "must be head or member"})
	}
	if len(issues) == 0 {
		return nil
	}
	return &CodedFieldsError{Code: CodeRosterEntryInvalid, Fields: issues}
}
