package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/adapters/postgres/sqlcgen"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/domain"
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

func (Roster) ListEntries(ctx context.Context, association uuid.UUID, on *time.Time, after *ports.EntryKey, limit int) ([]ports.RosterEntry, error) {
	// limit is bounded by app.MaxPageSize, so it fits int32.
	arg := sqlcgen.ListRosterEntriesParams{AssociationID: uuidOf(association), RowLimit: int32(limit)}
	if on != nil {
		arg.OnDate = dateOf(*on)
	}
	if after != nil {
		arg.AfterCreatedAt = pgtype.Timestamptz{Time: after.CreatedAt, Valid: true}
		arg.AfterID = uuidOf(after.ID)
	}
	return listEntries(ctx, arg)
}

func (Roster) Entry(ctx context.Context, association, id uuid.UUID) (ports.RosterEntry, bool, error) {
	rows, err := listEntries(ctx, sqlcgen.ListRosterEntriesParams{AssociationID: uuidOf(association), EntryID: uuidOf(id), RowLimit: 1})
	if err != nil || len(rows) == 0 {
		return ports.RosterEntry{}, false, err
	}
	return rows[0], true, nil
}

func listEntries(ctx context.Context, arg sqlcgen.ListRosterEntriesParams) ([]ports.RosterEntry, error) {
	tx, err := db.From(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := sqlcgen.New(tx).ListRosterEntries(ctx, arg)
	if err != nil {
		return nil, err
	}
	effects, err := sqlcgen.New(tx).RosterDecisionEffects(ctx, arg.AssociationID)
	if err != nil {
		return nil, err
	}
	effectOf := make(map[uuid.UUID]string, len(effects))
	for _, f := range effects {
		effectOf[uuid.UUID(f.BoardMemberID.Bytes)] = f.Effect
	}
	out := make([]ports.RosterEntry, len(rows))
	for i, r := range rows {
		e := ports.RosterEntry{
			ID:             uuid.UUID(r.ID.Bytes),
			CreatedAt:      r.EntryCreatedAt.Time,
			RevisedAt:      r.RevisedAt.Time,
			DecisionEffect: effectOf[uuid.UUID(r.ID.Bytes)],
			Entry: domain.BoardMemberEntry{
				FullName: r.FullName, SignInEmail: r.SignInEmail.String,
				Term:     domain.Term{Start: r.TermStart.Time},
				Position: domain.Position(r.Position),
			},
		}
		if r.TermEnd.Valid {
			e.Entry.Term.End = r.TermEnd.Time
		}
		if r.DecisionID.Valid {
			e.Entry.DecisionID = uuid.UUID(r.DecisionID.Bytes).String()
		}
		if r.RevisedByMemberID.Valid {
			e.RevisedByMemberID = uuid.UUID(r.RevisedByMemberID.Bytes)
		}
		out[i] = e
	}
	return out, nil
}
