package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/adapters/postgres/sqlcgen"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/ports"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/db"
)

// Roster reads the board roster on the transaction in ctx.
type Roster struct{}

var _ ports.RosterReader = Roster{}

func NewRoster() Roster { return Roster{} }

func dateOf(t time.Time) pgtype.Date {
	return pgtype.Date{Time: t, Valid: true}
}

func (Roster) CurrentMember(ctx context.Context, association uuid.UUID, email string, date time.Time) (ports.Membership, bool, error) {
	tx, err := db.From(ctx)
	if err != nil {
		return ports.Membership{}, false, err
	}
	position, err := sqlcgen.New(tx).CurrentMembership(ctx, sqlcgen.CurrentMembershipParams{
		AssociationID: uuidOf(association), Email: email, OnDate: dateOf(date),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return ports.Membership{}, false, nil
	}
	if err != nil {
		return ports.Membership{}, false, err
	}
	return ports.Membership{AssociationID: association, Position: position}, true, nil
}

func (Roster) MembershipsOf(ctx context.Context, email string, date time.Time) ([]ports.Membership, error) {
	tx, err := db.From(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := sqlcgen.New(tx).MembershipsOf(ctx, sqlcgen.MembershipsOfParams{Email: email, OnDate: dateOf(date)})
	if err != nil {
		return nil, err
	}
	out := make([]ports.Membership, 0, len(rows))
	for _, r := range rows {
		out = append(out, ports.Membership{AssociationID: uuid.UUID(r.AssociationID.Bytes), Position: r.Position})
	}
	return out, nil
}

func (Roster) MemberSignInAddresses(ctx context.Context, association uuid.UUID, date time.Time) ([]string, error) {
	tx, err := db.From(ctx)
	if err != nil {
		return nil, err
	}
	return sqlcgen.New(tx).MemberSignInAddresses(ctx, sqlcgen.MemberSignInAddressesParams{
		AssociationID: uuidOf(association), OnDate: dateOf(date),
	})
}
