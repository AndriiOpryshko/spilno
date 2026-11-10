package app

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/domain"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/ports"
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

// MemberSignInAddresses lists the addresses of today's members.
func (s *RosterService) MemberSignInAddresses(ctx context.Context, association uuid.UUID) ([]string, error) {
	return s.roster.MemberSignInAddresses(ctx, association, s.now().UTC())
}

// SeatHead makes the person head under the decision. A person with a current entry gets a
// new revision of it, keeping the entry's own term; anyone else gets a new entry with term.
func (s *RosterService) SeatHead(ctx context.Context, association uuid.UUID, fullName, email string, term domain.Term, decisionID uuid.UUID) (uuid.UUID, error) {
	today := s.now().UTC()
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
	entry.Term.End = domain.PreviousHeadTermEnd(date.AddDate(0, 0, 1), entry.Term.Start)
	entry.DecisionID = decisionID.String()
	return s.writer.WriteRevision(ctx, association, memberID, entry)
}
