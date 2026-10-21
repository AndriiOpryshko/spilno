// Package postgres implements the governance context's ports on Postgres.
package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/adapters/postgres/sqlcgen"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/domain"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/ports"
)

type Meetings struct {
	q *sqlcgen.Queries
}

var _ ports.MeetingRepository = (*Meetings)(nil)

func NewMeetings(pool *pgxpool.Pool) *Meetings {
	return &Meetings{q: sqlcgen.New(pool)}
}

func (s *Meetings) Create(ctx context.Context, f domain.MeetingFields) (domain.Meeting, error) {
	row, err := s.q.CreateMeeting(ctx, sqlcgen.CreateMeetingParams{
		Title:          f.Title,
		Kind:           string(f.Kind),
		HeldAt:         pgtype.Timestamptz{Time: f.HeldAt, Valid: true},
		EndsAt:         timestamptzOf(f.EndsAt),
		Location:       textOf(f.Location),
		QuorumRequired: f.QuorumRequired,
	})
	if err != nil {
		return domain.Meeting{}, err
	}
	return toDomain(row), nil
}

func (s *Meetings) Get(ctx context.Context, id uuid.UUID) (domain.Meeting, error) {
	row, err := s.q.GetMeeting(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Meeting{}, domain.ErrMeetingNotFound
	}
	if err != nil {
		return domain.Meeting{}, err
	}
	return toDomain(row), nil
}

func (s *Meetings) List(ctx context.Context, after *ports.MeetingKey, limit int) ([]domain.Meeting, error) {
	// limit is bounded by app.MaxPageSize, so it fits int32.
	arg := sqlcgen.ListMeetingsParams{RowLimit: int32(limit)}
	if after != nil {
		arg.AfterCreatedAt = pgtype.Timestamptz{Time: after.CreatedAt, Valid: true}
		arg.AfterID = pgtype.UUID{Bytes: after.ID, Valid: true}
	}
	rows, err := s.q.ListMeetings(ctx, arg)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Meeting, len(rows))
	for i, r := range rows {
		out[i] = toDomain(r)
	}
	return out, nil
}

func toDomain(r sqlcgen.Meeting) domain.Meeting {
	return domain.Meeting{
		ID: r.ID.Bytes,
		MeetingFields: domain.MeetingFields{
			Title:          r.Title,
			Kind:           domain.MeetingKind(r.Kind),
			HeldAt:         r.HeldAt.Time,
			EndsAt:         timestamptzPtr(r.EndsAt),
			Location:       textPtr(r.Location),
			QuorumRequired: r.QuorumRequired,
		},
		CreatedAt: r.CreatedAt.Time,
		UpdatedAt: r.UpdatedAt.Time,
	}
}

func timestamptzOf(v *time.Time) pgtype.Timestamptz {
	if v == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: *v, Valid: true}
}

func timestamptzPtr(v pgtype.Timestamptz) *time.Time {
	if !v.Valid {
		return nil
	}
	out := v.Time
	return &out
}

func textOf(v *string) pgtype.Text {
	if v == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *v, Valid: true}
}

func textPtr(v pgtype.Text) *string {
	if !v.Valid {
		return nil
	}
	out := v.String
	return &out
}
