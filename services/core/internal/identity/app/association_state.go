package app

import (
	"context"

	"github.com/google/uuid"

	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/domain"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/ports"
)

// AssociationStates lets the оператор list, disable and enable associations.
type AssociationStates struct {
	repo ports.AssociationStateRepository
}

func NewAssociationStates(repo ports.AssociationStateRepository) *AssociationStates {
	return &AssociationStates{repo: repo}
}

func (s *AssociationStates) List(ctx context.Context, after *ports.AssociationCursor, limit int) ([]ports.OperatorAssociation, error) {
	return s.repo.List(ctx, after, limit)
}

func (s *AssociationStates) Disable(ctx context.Context, id uuid.UUID, byEmail, reason string) (ports.StateChange, error) {
	return s.repo.Change(ctx, id, byEmail, func(a *domain.Association) error { return a.Disable(reason) })
}

func (s *AssociationStates) Enable(ctx context.Context, id uuid.UUID, byEmail string) (ports.StateChange, error) {
	return s.repo.Change(ctx, id, byEmail, func(a *domain.Association) error { return a.Enable() })
}
