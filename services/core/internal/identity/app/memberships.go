package app

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/ports"
)

// MembershipSeat is one roster entry of a person.
type MembershipSeat struct {
	AssociationID uuid.UUID
	Position      string
}

// RosterReader is what listMemberships needs from the roster.
type RosterReader interface {
	MembershipsOf(ctx context.Context, email string, date time.Time) ([]MembershipSeat, error)
	CurrentMemberID(ctx context.Context, association uuid.UUID, email string, date time.Time) (uuid.UUID, bool, error)
}

// Membership is an association of the person with their position there.
type Membership struct {
	AssociationID   uuid.UUID
	AssociationName string
	State           string
	BoardMemberID   uuid.UUID
	Position        string
}

// Memberships lists the associations whose roster has the person today.
type Memberships struct {
	roster RosterReader
	repo   ports.CardRepository
	now    func() time.Time
}

func NewMemberships(roster RosterReader, repo ports.CardRepository) *Memberships {
	return &Memberships{roster: roster, repo: repo, now: time.Now}
}

func (m *Memberships) List(ctx context.Context, email string) ([]Membership, error) {
	today := m.now().UTC().Truncate(24 * time.Hour)
	seats, err := m.roster.MembershipsOf(ctx, email, today)
	if err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, 0, len(seats))
	for _, s := range seats {
		ids = append(ids, s.AssociationID)
	}
	briefs, err := m.repo.Briefs(ctx, ids)
	if err != nil {
		return nil, err
	}
	out := make([]Membership, 0, len(seats))
	for _, s := range seats {
		b, ok := briefs[s.AssociationID]
		if !ok {
			continue
		}
		id, ok, err := m.roster.CurrentMemberID(ctx, s.AssociationID, email, today)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		out = append(out, Membership{AssociationID: s.AssociationID, AssociationName: b.Name, State: b.State, BoardMemberID: id, Position: s.Position})
	}
	return out, nil
}
