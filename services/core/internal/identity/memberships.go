package identity

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/adapters/postgres"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/app"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/db"
)

// RosterSeat is one roster entry of a person, as listMemberships needs it.
type RosterSeat struct {
	AssociationID uuid.UUID
	Position      string // "head" | "member"
}

// MembershipsConfig is what listMemberships needs from the roster (cmd/api passes the governance reads).
type MembershipsConfig struct {
	MembershipsOf   func(ctx context.Context, email string, date time.Time) ([]RosterSeat, error)
	CurrentMemberID func(ctx context.Context, association uuid.UUID, email string, date time.Time) (uuid.UUID, bool, error)
}

// EnableMemberships turns listMemberships on.
func (m *Module) EnableMemberships(cfg MembershipsConfig) {
	m.handler.EnableMemberships(app.NewMemberships(scopedRoster{m.pool, cfg}, postgres.NewCards(m.pool)))
}

// scopedRoster runs the roster reads in a transaction: the person's own entries in every
// association need the actor's address set, one entry's id needs its association.
type scopedRoster struct {
	pool *pgxpool.Pool
	cfg  MembershipsConfig
}

func (s scopedRoster) MembershipsOf(ctx context.Context, email string, date time.Time) (out []app.MembershipSeat, err error) {
	err = db.InAssociation(ctx, s.pool, uuid.Nil, email, func(ctx context.Context) error {
		seats, e := s.cfg.MembershipsOf(ctx, email, date)
		for _, st := range seats {
			out = append(out, app.MembershipSeat{AssociationID: st.AssociationID, Position: st.Position})
		}
		return e
	})
	return out, err
}

func (s scopedRoster) CurrentMemberID(ctx context.Context, association uuid.UUID, email string, date time.Time) (id uuid.UUID, ok bool, err error) {
	err = db.InAssociation(ctx, s.pool, association, "", func(ctx context.Context) (e error) {
		id, ok, e = s.cfg.CurrentMemberID(ctx, association, email, date)
		return e
	})
	return id, ok, err
}
