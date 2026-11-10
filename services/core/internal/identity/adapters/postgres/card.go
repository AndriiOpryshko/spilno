package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/adapters/postgres/sqlcgen"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/ports"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/db"
)

// Cards implements ports.CardRepository.
type Cards struct {
	pool *pgxpool.Pool
}

var _ ports.CardRepository = (*Cards)(nil)

func NewCards(pool *pgxpool.Pool) *Cards { return &Cards{pool: pool} }

func optTime(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

// Briefs reads the associations table only; it has no policy.
func (s *Cards) Briefs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]ports.AssociationBrief, error) {
	out := map[uuid.UUID]ports.AssociationBrief{}
	if len(ids) == 0 {
		return out, nil
	}
	pg := make([]pgtype.UUID, len(ids))
	for i, id := range ids {
		pg[i] = pgUUID(id)
	}
	rows, err := sqlcgen.New(s.pool).BriefAssociations(ctx, pg)
	if err != nil {
		return nil, err
	}
	for _, r := range rows {
		out[uuid.UUID(r.ID.Bytes)] = ports.AssociationBrief{Name: r.Name, State: r.State}
	}
	return out, nil
}

// Card reads the card on the transaction in ctx.
func (s *Cards) Card(ctx context.Context, id uuid.UUID, today time.Time) (ports.Card, error) {
	tx, err := db.From(ctx)
	if err != nil {
		return ports.Card{}, err
	}
	q := sqlcgen.New(tx)
	h, err := q.GetCardHead(ctx, pgUUID(id))
	if errors.Is(err, pgx.ErrNoRows) {
		return ports.Card{}, ports.ErrCardNotFound
	}
	if err != nil {
		return ports.Card{}, err
	}
	c := ports.Card{
		ID: id, Name: h.Name, State: h.State, Edrpou: optText(h.Edrpou), Address: optText(h.Address),
		ConnectedByHand: !h.ConfirmedByDecisionID.Valid, History: []ports.CardEvent{},
	}
	name, err := q.GetCardHeadFullName(ctx, sqlcgen.GetCardHeadFullNameParams{
		AssociationID: pgUUID(id), Today: pgtype.Date{Time: today, Valid: true},
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return ports.Card{}, err
	}
	if err == nil {
		c.HeadFullName = &name
	}
	rows, err := q.ListCardHistory(ctx, pgUUID(id))
	if err != nil {
		return ports.Card{}, err
	}
	for _, r := range rows {
		c.History = append(c.History, ports.CardEvent{
			Kind: r.Kind, OccurredAt: r.OccurredAt.Time, DecisionID: optUUID(r.DecisionID), FiledAt: optTime(r.FiledAt),
			Reason: optText(r.Reason), HeadFullName: optText(r.HeadFullName), SignerFullName: optText(r.SignerFullName),
			ExtractID: optUUID(r.ExtractID), ExtractDatedOn: optDate(r.ExtractDatedOn),
		})
	}
	return c, nil
}

// Extract reads an extract an approval of the association relied on.
func (s *Cards) Extract(ctx context.Context, association, extract uuid.UUID) (ports.StoredExtract, error) {
	tx, err := db.From(ctx)
	if err != nil {
		return ports.StoredExtract{}, err
	}
	r, err := sqlcgen.New(tx).GetCardExtract(ctx, sqlcgen.GetCardExtractParams{AssociationID: pgUUID(association), ExtractID: pgUUID(extract)})
	if errors.Is(err, pgx.ErrNoRows) {
		return ports.StoredExtract{}, ports.ErrCardNotFound
	}
	if err != nil {
		return ports.StoredExtract{}, err
	}
	return ports.StoredExtract{Name: r.OriginalName, MediaType: r.MediaType, Content: r.Content, Sha256: r.Sha256}, nil
}
