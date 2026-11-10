package ports

import (
	"context"

	"github.com/google/uuid"

	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/domain"
)

// NewRejection is a rejection the оператор records.
type NewRejection struct {
	ApplicationID     uuid.UUID
	FilingID          uuid.UUID
	OperatorAccountID uuid.UUID
	OperatorEmail     string
	Reason            string
}

// RejectionRepository stores a rejection.
type RejectionRepository interface {
	// Reject locks the application (ErrApplicationNotFound for an unknown id or a filing that is
	// not of it), lets decide judge it, then stores the decision and the new state in one
	// transaction. An error from decide is returned and nothing is stored.
	Reject(ctx context.Context, r NewRejection, decide func(domain.Application) (domain.Application, error)) (ReviewDecision, error)
}
