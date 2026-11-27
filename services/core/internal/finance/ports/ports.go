// Package ports declares what the finance context needs from the outside world.
package ports

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/AndriiOpryshko/spilno/services/core/internal/finance/domain"
)

// ContractKey is a contract's position in the newest-first listing order.
type ContractKey struct {
	CreatedAt time.Time
	ID        uuid.UUID
}

// ContractFilter narrows a contract listing; the zero value lists every contract.
type ContractFilter struct {
	// MissingUAHEquivalent keeps only the foreign-currency contracts that have
	// no hryvnia equivalent yet.
	MissingUAHEquivalent bool
}

// ContractRepository stores contracts.
type ContractRepository interface {
	// Create stores already validated terms and returns the stored contract.
	Create(ctx context.Context, terms domain.ContractTerms, eq *domain.UAHEquivalent) (domain.Contract, error)
	// SetUAHEquivalent stores the equivalent only on a contract that has none; false when nothing was updated.
	SetUAHEquivalent(ctx context.Context, id uuid.UUID, eq domain.UAHEquivalent) (domain.Contract, bool, error)
	// Get returns domain.ErrContractNotFound when no contract has the id.
	Get(ctx context.Context, id uuid.UUID) (domain.Contract, error)
	// List returns up to limit contracts that pass filter, ordered by
	// created_at, then id, both descending, starting strictly after the key
	// (from the top when nil).
	List(ctx context.Context, filter ContractFilter, after *ContractKey, limit int) ([]domain.Contract, error)
}

type associationKey struct{}

// WithAssociation scopes the repository calls made with the returned context
// to one association.
func WithAssociation(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, associationKey{}, id)
}

// AssociationFrom returns the association set by WithAssociation.
func AssociationFrom(ctx context.Context) uuid.UUID {
	id, _ := ctx.Value(associationKey{}).(uuid.UUID)
	return id
}
