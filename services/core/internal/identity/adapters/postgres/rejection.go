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
)

// Rejections implements ports.RejectionRepository.
type Rejections struct {
	pool *pgxpool.Pool
}

var _ ports.RejectionRepository = (*Rejections)(nil)

func NewRejections(pool *pgxpool.Pool) *Rejections { return &Rejections{pool: pool} }

func (s *Rejections) Reject(ctx context.Context, r ports.NewRejection, decide func(domain.Application) (domain.Application, error)) (ports.ReviewDecision, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return ports.ReviewDecision{}, err
	}
	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()
	q := sqlcgen.New(tx)
	row, err := q.LockApplicationForRejection(ctx, pgUUID(r.ApplicationID))
	if errors.Is(err, pgx.ErrNoRows) {
		return ports.ReviewDecision{}, ports.ErrApplicationNotFound
	}
	if err != nil {
		return ports.ReviewDecision{}, err
	}
	if _, err := q.FilingBelongsToApplication(ctx, sqlcgen.FilingBelongsToApplicationParams{FilingID: pgUUID(r.FilingID), ApplicationID: pgUUID(r.ApplicationID)}); errors.Is(err, pgx.ErrNoRows) {
		return ports.ReviewDecision{}, ports.ErrApplicationNotFound
	} else if err != nil {
		return ports.ReviewDecision{}, err
	}
	next, err := decide(domain.Application{
		Kind: domain.ApplicationKind(row.Kind), State: domain.ApplicationState(row.State),
		AccountID: uuid.UUID(row.AccountID.Bytes).String(), EDRPOU: row.Edrpou,
	})
	if err != nil {
		return ports.ReviewDecision{}, err
	}
	id, err := uuid.NewV7()
	if err != nil {
		return ports.ReviewDecision{}, err
	}
	at, err := q.CreateRejectionDecision(ctx, sqlcgen.CreateRejectionDecisionParams{
		ID: pgUUID(id), ApplicationID: pgUUID(r.ApplicationID), FilingID: pgUUID(r.FilingID),
		Reason: pgText(r.Reason), DecidedByEmail: r.OperatorEmail,
	})
	if err != nil {
		return ports.ReviewDecision{}, err
	}
	if err := q.UpdateApplicationState(ctx, sqlcgen.UpdateApplicationStateParams{ID: pgUUID(r.ApplicationID), State: string(next.State)}); err != nil {
		return ports.ReviewDecision{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return ports.ReviewDecision{}, err
	}
	reason := r.Reason
	return ports.ReviewDecision{ID: id, FilingID: r.FilingID, Outcome: "rejected", Reason: &reason, DecidedAt: at.Time}, nil
}

func pgText(s string) pgtype.Text { return pgtype.Text{String: s, Valid: true} }
