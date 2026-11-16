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
)

// Registry reads the реєстр протоколів on the transaction in ctx.
type Registry struct{}

var _ ports.RegistryReader = Registry{}

func NewRegistry() Registry { return Registry{} }

func int8Of(v *int64) pgtype.Int8 {
	if v == nil {
		return pgtype.Int8{}
	}
	return pgtype.Int8{Int64: *v, Valid: true}
}

func optDateOf(v *time.Time) pgtype.Date {
	if v == nil {
		return pgtype.Date{}
	}
	return dateOf(*v)
}

func (Registry) Search(ctx context.Context, association uuid.UUID, f ports.RegistryFilter, after *ports.RegistryKey, limit int) ([]ports.RegistryDecision, error) {
	q, err := queries(ctx)
	if err != nil {
		return nil, err
	}
	// limit is bounded by app.MaxPageSize, so it fits int32.
	arg := sqlcgen.SearchRegistryDecisionsParams{
		AssociationID: uuidOf(association),
		Counterparty:  textOf(f.Counterparty), Subject: textOf(f.Subject),
		Amount: int8Of(f.Amount), AmountFrom: int8Of(f.AmountFrom), AmountTo: int8Of(f.AmountTo),
		DateFrom: optDateOf(f.DateFrom), DateTo: optDateOf(f.DateTo),
		RowLimit: int32(limit),
	}
	if after != nil {
		arg.AfterDate, arg.AfterID = dateOf(after.MeetingDate), uuidOf(after.ID)
	}
	rows, err := q.SearchRegistryDecisions(ctx, arg)
	if err != nil {
		return nil, err
	}
	out := make([]ports.RegistryDecision, len(rows))
	for i, r := range rows {
		d := ports.RegistryDecision{
			Key:        ports.RegistryKey{MeetingDate: r.MeetingDate.Time, ID: uuid.UUID(r.ID.Bytes)},
			ProtocolID: uuid.UUID(r.ProtocolID.Bytes), AgendaItemID: uuid.UUID(r.AgendaItemID.Bytes),
			Number: int(r.Number), Subject: r.Subject,
			Counterparty: textPtr(r.Counterparty), Currency: textPtr(r.Currency),
		}
		if r.AmountKopiyky.Valid {
			d.AmountKopiyky = &r.AmountKopiyky.Int64
		}
		out[i] = d
	}
	return out, nil
}

func (Registry) WithdrawnNumbers(ctx context.Context, association uuid.UUID, lo, hi int) ([]int, error) {
	q, err := queries(ctx)
	if err != nil {
		return nil, err
	}
	// Numbers come from stored int4 values, so the bounds fit int32.
	rows, err := q.WithdrawnNumbersBetween(ctx, sqlcgen.WithdrawnNumbersBetweenParams{
		AssociationID: uuidOf(association), Lo: int32(lo), Hi: int32(hi),
	})
	if err != nil {
		return nil, err
	}
	out := make([]int, len(rows))
	for i, n := range rows {
		out[i] = int(n)
	}
	return out, nil
}

func (Registry) SignedSnapshot(ctx context.Context, association, protocol uuid.UUID) (ports.SignedSnapshot, bool, error) {
	q, err := queries(ctx)
	if err != nil {
		return ports.SignedSnapshot{}, false, err
	}
	row, err := q.GetSignedSnapshot(ctx, sqlcgen.GetSignedSnapshotParams{AssociationID: uuidOf(association), ProtocolID: uuidOf(protocol)})
	if errors.Is(err, pgx.ErrNoRows) {
		return ports.SignedSnapshot{}, false, nil
	}
	if err != nil {
		return ports.SignedSnapshot{}, false, err
	}
	return ports.SignedSnapshot{
		ProtocolID: uuid.UUID(row.ProtocolID.Bytes), Number: int(row.Number), MeetingDate: row.MeetingDate.Time, Payload: row.Snapshot,
	}, true, nil
}
