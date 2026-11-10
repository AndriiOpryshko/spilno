package app

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/ports"
)

// RosterService answers who sits on which board; every call runs on the
// transaction in ctx.
type RosterService struct {
	roster ports.RosterReader
	now    func() time.Time
}

func NewRosterService(roster ports.RosterReader) *RosterService {
	return &RosterService{roster: roster, now: time.Now}
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
