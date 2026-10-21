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
	Create(ctx context.Context, fields domain.MeetingFields) (domain.Meeting, error)
	// Get returns domain.ErrMeetingNotFound when no meeting has the id.
	Get(ctx context.Context, id uuid.UUID) (domain.Meeting, error)
	// List returns up to limit meetings ordered by created_at, then id, both
	// descending, starting strictly after the key (from the top when nil).
	List(ctx context.Context, after *MeetingKey, limit int) ([]domain.Meeting, error)
}
