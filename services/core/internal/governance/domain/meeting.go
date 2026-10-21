// Package domain holds the governance context's model. It imports only the standard library and uuid.
package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

// ErrMeetingNotFound is returned when no meeting has the requested ID.
var ErrMeetingNotFound = errors.New("meeting not found")

type MeetingKind string

const (
	MeetingKindRegular       MeetingKind = "REGULAR"
	MeetingKindExtraordinary MeetingKind = "EXTRAORDINARY"
)

func (v MeetingKind) Valid() bool {
	switch v {
	case MeetingKindRegular, MeetingKindExtraordinary:
		return true
	}
	return false
}

const (
	MaxTitleLen    = 200
	MaxLocationLen = 200
)

// Field names reported in a ValidationError.
const (
	FieldTitle    = "title"
	FieldKind     = "kind"
	FieldEndsAt   = "ends_at"
	FieldLocation = "location"
)

// MeetingFields is everything a caller provides to create a meeting.
type MeetingFields struct {
	Title          string
	Kind           MeetingKind
	HeldAt         time.Time
	EndsAt         *time.Time
	Location       *string
	QuorumRequired bool
}

// Validate checks every invariant and reports all broken ones at once as a
// *ValidationError.
func (t MeetingFields) Validate() error {
	var v ValidationError
	if n := utf8.RuneCountInString(t.Title); n < 1 || n > MaxTitleLen {
		v.add(FieldTitle, fmt.Sprintf("must be 1 to %d characters", MaxTitleLen))
	}
	if !t.Kind.Valid() {
		v.add(FieldKind, "must be one of REGULAR, EXTRAORDINARY")
	}
	if t.EndsAt != nil && !t.EndsAt.After(t.HeldAt) {
		v.add(FieldEndsAt, "must be after held_at")
	}
	if t.Location != nil {
		if n := utf8.RuneCountInString(*t.Location); n > MaxLocationLen {
			v.add(FieldLocation, fmt.Sprintf("must be 0 to %d characters", MaxLocationLen))
		}
	}
	if len(v.Fields) > 0 {
		return &v
	}
	return nil
}

// Meeting is a stored meeting: its fields plus the identity and timestamps the store assigns.
type Meeting struct {
	ID uuid.UUID
	MeetingFields
	CreatedAt time.Time
	UpdatedAt time.Time
}

type FieldError struct {
	Field  string
	Detail string
}

// ValidationError lists the invariants a MeetingFields breaks.
type ValidationError struct {
	Fields []FieldError
}

func (e *ValidationError) Error() string {
	parts := make([]string, len(e.Fields))
	for i, f := range e.Fields {
		parts[i] = f.Field + ": " + f.Detail
	}
	return "invalid meeting: " + strings.Join(parts, "; ")
}

func (e *ValidationError) add(field, detail string) {
	e.Fields = append(e.Fields, FieldError{Field: field, Detail: detail})
}
