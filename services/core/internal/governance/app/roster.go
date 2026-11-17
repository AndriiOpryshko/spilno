package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/domain"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/ports"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/kyivday"
)

// RosterService answers who sits on which board; every call runs on the
// transaction in ctx.
type RosterService struct {
	roster ports.RosterReader
	writer ports.RosterWriter
	now    func() time.Time
}

func NewRosterService(roster ports.RosterReader, writer ports.RosterWriter) *RosterService {
	return &RosterService{roster: roster, writer: writer, now: time.Now}
}

func (s *RosterService) CurrentMember(ctx context.Context, association uuid.UUID, email string, date time.Time) (ports.Membership, bool, error) {
	return s.roster.CurrentMember(ctx, association, email, date)
}

func (s *RosterService) MembershipsOf(ctx context.Context, email string, date time.Time) ([]ports.Membership, error) {
	return s.roster.MembershipsOf(ctx, email, date)
}

// CurrentMemberID returns the roster entry id of the member whose sign-in address is email on date.
func (s *RosterService) CurrentMemberID(ctx context.Context, association uuid.UUID, email string, date time.Time) (uuid.UUID, bool, error) {
	return s.writer.CurrentMemberID(ctx, association, email, date)
}

// MemberSignInAddresses lists the addresses of today's members.
func (s *RosterService) MemberSignInAddresses(ctx context.Context, association uuid.UUID) ([]string, error) {
	return s.roster.MemberSignInAddresses(ctx, association, kyivday.Today(s.now()))
}

// SeatHead makes the person head under the decision. A person with a current entry gets a
// new revision of it, keeping the entry's own term; anyone else gets a new entry with term.
func (s *RosterService) SeatHead(ctx context.Context, association uuid.UUID, fullName, email string, term domain.Term, decisionID uuid.UUID) (uuid.UUID, error) {
	today := kyivday.Today(s.now())
	id, found, err := s.writer.CurrentMemberID(ctx, association, email, today)
	if err != nil {
		return uuid.Nil, err
	}
	var entry domain.BoardMemberEntry
	if found {
		latest, ok, err := s.writer.LatestEntry(ctx, association, id)
		if err != nil {
			return uuid.Nil, err
		}
		if !ok {
			return uuid.Nil, fmt.Errorf("roster entry %s vanished", id)
		}
		entry = domain.Reseat(latest, decisionID.String())
	} else {
		if id, err = s.writer.CreateMember(ctx, association); err != nil {
			return uuid.Nil, err
		}
		entry = domain.Reseat(domain.BoardMemberEntry{FullName: fullName, SignInEmail: email, Term: term}, decisionID.String())
	}
	if err := s.writer.WriteRevision(ctx, association, id, entry); err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

// EndTerm ends the whole entry on date (never before its first day) under the decision.
func (s *RosterService) EndTerm(ctx context.Context, association, memberID uuid.UUID, date time.Time, decisionID uuid.UUID) error {
	entry, ok, err := s.writer.LatestEntry(ctx, association, memberID)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("roster entry %s not found", memberID)
	}
	end := domain.PreviousHeadTermEnd(date.AddDate(0, 0, 1), entry.Term.Start)
	if entry.Term.End.IsZero() || end.Before(entry.Term.End) {
		entry.Term.End = end
	}
	entry.DecisionID = decisionID.String()
	return s.writer.WriteRevision(ctx, association, memberID, entry)
}

// ErrRosterEntryNotFound is returned when the association has no roster entry with the id.
var ErrRosterEntryNotFound = errors.New("roster entry not found")

// RosterPage is one slice of the newest-first roster listing.
type RosterPage struct {
	Items []ports.RosterEntry
	// NextCursor resumes the listing after Items; empty on the last page.
	NextCursor string
}

// ListEntries returns up to limit entries after cursor, newest first; on, when set, keeps the entries
// whose term covers that date. A limit outside 1..MaxPageSize falls back like the meeting listing.
func (s *RosterService) ListEntries(ctx context.Context, association uuid.UUID, on *time.Time, limit int, cursor string) (RosterPage, error) {
	switch {
	case limit <= 0:
		limit = DefaultPageSize
	case limit > MaxPageSize:
		limit = MaxPageSize
	}
	var after *ports.EntryKey
	if cursor != "" {
		key, err := decodeCursor(cursor)
		if err != nil {
			return RosterPage{}, err
		}
		after = &key
	}
	// One extra row tells whether another page follows.
	items, err := s.roster.ListEntries(ctx, association, on, after, limit+1)
	if err != nil {
		return RosterPage{}, err
	}
	page := RosterPage{Items: items}
	if len(items) > limit {
		page.Items = items[:limit]
		last := page.Items[limit-1]
		page.NextCursor = encodeCursor(ports.EntryKey{CreatedAt: last.CreatedAt, ID: last.ID})
	}
	return page, nil
}

// AddMember stores a new plain member (no decision) authored by the signed-in member and returns the entry.
// Every broken field comes back at once as *domain.CodedFieldsError and nothing is saved.
func (s *RosterService) AddMember(ctx context.Context, association uuid.UUID, in domain.BoardMemberEntry) (ports.RosterEntry, error) {
	in.Position, in.DecisionID = domain.PositionMember, ""
	if err := in.Validate(); err != nil {
		return ports.RosterEntry{}, err
	}
	by, err := s.actor(ctx, association)
	if err != nil {
		return ports.RosterEntry{}, err
	}
	id, err := s.writer.CreateMember(ctx, association)
	if err != nil {
		return ports.RosterEntry{}, err
	}
	if err := s.writer.WriteRevisionBy(ctx, association, id, in, by); err != nil {
		return ports.RosterEntry{}, err
	}
	return s.entry(ctx, association, id)
}

// ChangeMember saves a new revision of the entry. The position and the decision of the latest revision
// are kept: they come from an approval, never from this input. ErrRosterEntryNotFound when the
// association has no such entry.
func (s *RosterService) ChangeMember(ctx context.Context, association, id uuid.UUID, in domain.BoardMemberEntry) (ports.RosterEntry, error) {
	latest, ok, err := s.writer.LatestEntry(ctx, association, id)
	if err != nil {
		return ports.RosterEntry{}, err
	}
	if !ok {
		return ports.RosterEntry{}, ErrRosterEntryNotFound
	}
	in.Position, in.DecisionID = latest.Position, latest.DecisionID
	if err := in.Validate(); err != nil {
		return ports.RosterEntry{}, err
	}
	current, err := s.entry(ctx, association, id)
	if err != nil {
		return ports.RosterEntry{}, err
	}
	if err := refuseReopeningEnded(current, latest.Term.End, in.Term.End); err != nil {
		return ports.RosterEntry{}, err
	}
	by, err := s.actor(ctx, association)
	if err != nil {
		return ports.RosterEntry{}, err
	}
	if err := s.writer.WriteRevisionBy(ctx, association, id, in, by); err != nil {
		return ports.RosterEntry{}, err
	}
	return s.entry(ctx, association, id)
}

// refuseReopeningEnded rejects a change that clears or postpones the term end an approval set on the entry.
func refuseReopeningEnded(current ports.RosterEntry, setEnd, newEnd time.Time) error {
	if current.DecisionEffect != "ended" || setEnd.IsZero() {
		return nil
	}
	if !newEnd.IsZero() && !newEnd.After(setEnd) {
		return nil
	}
	return &domain.CodedFieldsError{
		Code:   domain.CodeRosterEntryInvalid,
		Fields: []domain.FieldIssue{{Pointer: "/term_end", Detail: "an approval ended this entry; its term end cannot be cleared or moved later"}},
	}
}

// actor is the roster entry of the signed-in member; nil when the transaction's person has none.
func (s *RosterService) actor(ctx context.Context, association uuid.UUID) (uuid.UUID, error) {
	id, _, err := s.writer.ActorEntryID(ctx, association, kyivday.Today(s.now()))
	return id, err
}

func (s *RosterService) entry(ctx context.Context, association, id uuid.UUID) (ports.RosterEntry, error) {
	e, ok, err := s.roster.Entry(ctx, association, id)
	if err == nil && !ok {
		err = fmt.Errorf("roster entry %s vanished", id)
	}
	return e, err
}
