package ports

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/domain"
)

// RosterWriter changes the board roster on the transaction in ctx (db.InAssociation).
type RosterWriter interface {
	// CurrentMemberID finds the entry whose latest revision carries email and covers date.
	CurrentMemberID(ctx context.Context, association uuid.UUID, email string, date time.Time) (uuid.UUID, bool, error)
	// LatestEntry returns the newest revision of an entry; false when the association has no such entry.
	LatestEntry(ctx context.Context, association, boardMember uuid.UUID) (domain.BoardMemberEntry, bool, error)
	// CreateMember stores a new, empty roster entry.
	CreateMember(ctx context.Context, association uuid.UUID) (uuid.UUID, error)
	// WriteRevision appends a revision to an entry.
	WriteRevision(ctx context.Context, association, boardMember uuid.UUID, entry domain.BoardMemberEntry) error
	// WriteRevisionBy is WriteRevision recording the entry of the member who made the change.
	WriteRevisionBy(ctx context.Context, association, boardMember uuid.UUID, entry domain.BoardMemberEntry, by uuid.UUID) error
	// ActorEntryID finds the entry of the signed-in person of this transaction, if they sit on the board on date.
	ActorEntryID(ctx context.Context, association uuid.UUID, date time.Time) (uuid.UUID, bool, error)
}
