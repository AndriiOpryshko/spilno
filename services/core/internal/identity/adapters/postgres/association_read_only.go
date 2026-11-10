package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/domain"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/db"
)

// DisabledState reads, on the transaction in ctx, whether the association is disabled and the
// reason of its latest disabling.
func DisabledState(ctx context.Context, id uuid.UUID) (disabled bool, reason string, err error) {
	tx, err := db.From(ctx)
	if err != nil {
		return false, "", err
	}
	var state string
	if err := tx.QueryRow(ctx, `SELECT state FROM associations WHERE id = $1`, id).Scan(&state); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, "", nil // no row, nothing disabled it
		}
		return false, "", err
	}
	if domain.AssociationState(state) != domain.AssociationDisabled {
		return false, "", nil
	}
	var r *string
	err = tx.QueryRow(ctx, `SELECT reason FROM association_state_changes
WHERE association_id = $1 AND to_state = $2 ORDER BY created_at DESC, id DESC LIMIT 1`, id, state).Scan(&r)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return false, "", err
	}
	if r != nil {
		reason = *r
	}
	return true, reason, nil
}
