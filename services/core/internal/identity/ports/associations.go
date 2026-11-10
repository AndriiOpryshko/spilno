package ports

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/domain"
)

// ErrAssociationNotFound is returned when no association has the id.
var ErrAssociationNotFound = errors.New("identity: association not found")

// OperatorAssociation is one row of the оператор's list; it never reads an association-owned table.
type OperatorAssociation struct {
	ID              uuid.UUID
	Name            string
	Edrpou          string // empty when unknown
	State           domain.AssociationState
	ConnectedByHand bool
	CreatedAt       time.Time
}

// AssociationCursor is the position after which the next page starts.
type AssociationCursor struct {
	CreatedAt time.Time
	ID        uuid.UUID
}

// StateChange is one recorded disabling or enabling.
type StateChange struct {
	AssociationID uuid.UUID
	State         domain.AssociationState
	Reason        string // empty for an enabling
	ChangedAt     time.Time
}

// AssociationStateRepository stores the state of associations and its insert-only history.
type AssociationStateRepository interface {
	// List returns up to limit associations, newest first, after the cursor when given.
	List(ctx context.Context, after *AssociationCursor, limit int) ([]OperatorAssociation, error)
	// Change runs mutate on the association's state in one transaction; when it succeeds the new
	// state and its history row are stored, when it fails nothing is. ErrAssociationNotFound if absent.
	Change(ctx context.Context, id uuid.UUID, byEmail string, mutate func(*domain.Association) error) (StateChange, error)
}
