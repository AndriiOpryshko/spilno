package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/adapters/postgres/sqlcgen"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/domain"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/ports"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/db"
)

// Associations implements ports.AssociationStateRepository.
type Associations struct {
	pool *pgxpool.Pool
}

var _ ports.AssociationStateRepository = (*Associations)(nil)

func NewAssociations(pool *pgxpool.Pool) *Associations { return &Associations{pool: pool} }

// List reads the associations table only; it has no policy and the list never touches an owned table.
func (s *Associations) List(ctx context.Context, after *ports.AssociationCursor, limit int) ([]ports.OperatorAssociation, error) {
	p := sqlcgen.ListAssociationsParams{PageSize: int32(limit)} //nolint:gosec // limit is 1..101
	if after != nil {
		p.CursorCreatedAt = pgtype.Timestamptz{Time: after.CreatedAt, Valid: true}
		p.CursorID = pgtype.UUID{Bytes: after.ID, Valid: true}
	}
	rows, err := sqlcgen.New(s.pool).ListAssociations(ctx, p)
	if err != nil {
		return nil, err
	}
	out := make([]ports.OperatorAssociation, 0, len(rows))
	for _, r := range rows {
		out = append(out, ports.OperatorAssociation{
			ID:              r.ID.Bytes,
			Name:            r.Name,
			Edrpou:          r.Edrpou.String,
			State:           domain.AssociationState(r.State),
			ConnectedByHand: !r.ConfirmedByDecisionID.Valid,
			CreatedAt:       r.CreatedAt.Time,
		})
	}
	return out, nil
}

func (s *Associations) Change(ctx context.Context, id uuid.UUID, byEmail string, mutate func(*domain.Association) error) (ports.StateChange, error) {
	var out ports.StateChange
	err := db.InAssociation(ctx, s.pool, id, byEmail, func(ctx context.Context) error {
		tx, err := db.From(ctx)
		if err != nil {
			return err
		}
		q := sqlcgen.New(tx)
		pgID := pgtype.UUID{Bytes: id, Valid: true}
		state, err := q.GetAssociationStateForUpdate(ctx, pgID)
		if errors.Is(err, pgx.ErrNoRows) {
			return ports.ErrAssociationNotFound
		}
		if err != nil {
			return err
		}
		a := domain.Association{State: domain.AssociationState(state)}
		if err := mutate(&a); err != nil {
			return err
		}
		if err := q.UpdateAssociationState(ctx, sqlcgen.UpdateAssociationStateParams{State: string(a.State), ID: pgID}); err != nil {
			return err
		}
		changeID, err := uuid.NewV7()
		if err != nil {
			return err
		}
		at, err := q.CreateStateChange(ctx, sqlcgen.CreateStateChangeParams{
			ID:             pgtype.UUID{Bytes: changeID, Valid: true},
			AssociationID:  pgID,
			ToState:        string(a.State),
			Reason:         pgtype.Text{String: a.DisabledReason, Valid: a.DisabledReason != ""},
			ChangedByEmail: byEmail,
		})
		if err != nil {
			return err
		}
		out = ports.StateChange{AssociationID: id, State: a.State, Reason: a.DisabledReason, ChangedAt: at.Time}
		return nil
	})
	return out, err
}
