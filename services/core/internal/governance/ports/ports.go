// Package ports declares what the governance context needs from the outside world.
package ports

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/domain"
)

// MeetingKey is a meeting's position in the newest-first listing order.
type MeetingKey struct {
	CreatedAt time.Time
	ID        uuid.UUID
}

// MeetingRepository stores meetings.
type MeetingRepository interface {
	// Create stores already validated fields and returns the stored meeting.
	Create(ctx context.Context, associationID uuid.UUID, fields domain.MeetingFields) (domain.Meeting, error)
	// Get returns domain.ErrMeetingNotFound when the association has no meeting with the id.
	Get(ctx context.Context, associationID, id uuid.UUID) (domain.Meeting, error)
	// List returns up to limit meetings ordered by created_at, then id, both
	// descending, starting strictly after the key (from the top when nil).
	List(ctx context.Context, associationID uuid.UUID, after *MeetingKey, limit int) ([]domain.Meeting, error)
}

// Membership is a person's current seat on one association's board.
type Membership struct {
	AssociationID uuid.UUID
	Position      string // "head" or "member"
}

// RosterReader reads the board roster on the transaction in ctx (db.InAssociation).
type RosterReader interface {
	// CurrentMember reports the association's member whose sign-in address is email
	// and whose term covers date.
	CurrentMember(ctx context.Context, association uuid.UUID, email string, date time.Time) (Membership, bool, error)
	// MembershipsOf returns every association whose current board holds email on date.
	MembershipsOf(ctx context.Context, email string, date time.Time) ([]Membership, error)
	// MemberSignInAddresses returns the sign-in addresses of the association's members whose term covers date.
	MemberSignInAddresses(ctx context.Context, association uuid.UUID, date time.Time) ([]string, error)
}
