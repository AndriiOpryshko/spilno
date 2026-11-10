package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/ports"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/db"
)

// Cards implements ports.CardRepository.
type Cards struct {
	pool *pgxpool.Pool
}

var _ ports.CardRepository = (*Cards)(nil)

func NewCards(pool *pgxpool.Pool) *Cards { return &Cards{pool: pool} }

// Briefs reads the associations table only; it has no policy.
func (s *Cards) Briefs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]ports.AssociationBrief, error) {
	out := map[uuid.UUID]ports.AssociationBrief{}
	if len(ids) == 0 {
		return out, nil
	}
	rows, err := s.pool.Query(ctx, `SELECT id, name, state FROM associations WHERE id = ANY($1)`, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id uuid.UUID
		var b ports.AssociationBrief
		if err := rows.Scan(&id, &b.Name, &b.State); err != nil {
			return nil, err
		}
		out[id] = b
	}
	return out, rows.Err()
}

const cardHistorySQL = `
SELECT CASE WHEN a.kind = 'connection' THEN 'connected' ELSE 'transfer_approved' END AS kind,
       d.created_at AS occurred_at, d.id AS decision_id, f.created_at AS filed_at, NULL::text AS reason,
       k.head_full_name, k.signer_full_name, f.extract_id, k.extract_dated_on, d.id AS tie
FROM application_decisions d
JOIN applications a ON a.id = d.application_id
JOIN application_filings f ON f.id = d.filing_id
LEFT JOIN application_checks k ON k.id = d.check_id
WHERE d.association_id = $1 AND d.outcome = 'approved'
UNION ALL
SELECT CASE WHEN c.to_state = 'disabled' THEN 'disabled' ELSE 'enabled' END,
       c.created_at, NULL::uuid, NULL::timestamptz, c.reason, NULL::text, NULL::text, NULL::uuid, NULL::date, c.id
FROM association_state_changes c
WHERE c.association_id = $1
ORDER BY occurred_at DESC, tie DESC`

const cardHeadSQL = `
SELECT full_name FROM (
    SELECT DISTINCT ON (board_member_id) full_name, position, term_start, term_end
    FROM board_member_revisions
    WHERE association_id = $1
    ORDER BY board_member_id, created_at DESC, id DESC
) r
WHERE position = 'head' AND term_start <= $2 AND (term_end IS NULL OR term_end >= $2)
ORDER BY term_start DESC LIMIT 1`

// Card reads the card on the transaction in ctx.
func (s *Cards) Card(ctx context.Context, id uuid.UUID, today time.Time) (ports.Card, error) {
	tx, err := db.From(ctx)
	if err != nil {
		return ports.Card{}, err
	}
	c := ports.Card{ID: id, History: []ports.CardEvent{}}
	var confirmed *uuid.UUID
	err = tx.QueryRow(ctx, `SELECT name, state, edrpou, address, confirmed_by_decision_id FROM associations WHERE id = $1`, id).
		Scan(&c.Name, &c.State, &c.Edrpou, &c.Address, &confirmed)
	if errors.Is(err, pgx.ErrNoRows) {
		return ports.Card{}, ports.ErrCardNotFound
	}
	if err != nil {
		return ports.Card{}, err
	}
	c.ConnectedByHand = confirmed == nil
	err = tx.QueryRow(ctx, cardHeadSQL, id, today).Scan(&c.HeadFullName)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return ports.Card{}, err
	}
	rows, err := tx.Query(ctx, cardHistorySQL, id)
	if err != nil {
		return ports.Card{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var e ports.CardEvent
		var tie uuid.UUID
		if err := rows.Scan(&e.Kind, &e.OccurredAt, &e.DecisionID, &e.FiledAt, &e.Reason, &e.HeadFullName,
			&e.SignerFullName, &e.ExtractID, &e.ExtractDatedOn, &tie); err != nil {
			return ports.Card{}, err
		}
		c.History = append(c.History, e)
	}
	return c, rows.Err()
}

// Extract reads an extract an approval of the association relied on.
func (s *Cards) Extract(ctx context.Context, association, extract uuid.UUID) (ports.StoredExtract, error) {
	tx, err := db.From(ctx)
	if err != nil {
		return ports.StoredExtract{}, err
	}
	var e ports.StoredExtract
	err = tx.QueryRow(ctx, `SELECT x.original_name, x.media_type, x.content, x.sha256
FROM application_decisions d
JOIN application_filings f ON f.id = d.filing_id
JOIN application_extracts x ON x.id = f.extract_id
WHERE d.association_id = $1 AND d.outcome = 'approved' AND x.id = $2
LIMIT 1`, association, extract).Scan(&e.Name, &e.MediaType, &e.Content, &e.Sha256)
	if errors.Is(err, pgx.ErrNoRows) {
		return ports.StoredExtract{}, ports.ErrCardNotFound
	}
	return e, err
}
