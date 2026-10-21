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

// ContractRepository stores contracts.
type ContractRepository interface {
	// Create stores already validated terms and returns the stored contract.
	Create(ctx context.Context, terms domain.ContractTerms) (domain.Contract, error)
	// Get returns domain.ErrContractNotFound when no contract has the id.
	Get(ctx context.Context, id uuid.UUID) (domain.Contract, error)
	// List returns up to limit contracts ordered by created_at, then id, both
	// descending, starting strictly after the key (from the top when nil).
	List(ctx context.Context, after *ContractKey, limit int) ([]domain.Contract, error)
}
