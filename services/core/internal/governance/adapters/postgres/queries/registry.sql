-- name: SearchRegistryDecisions :many
-- Decisions of signed protocols, newest meeting first by the (meeting_date, id) keyset. A NULL filter is
-- off; counterparty and subject match as case-insensitive "contains" (strpos, so % and _ stay literal).
SELECT r.id, r.agenda_item_id, s.protocol_id, r.number, r.meeting_date, r.subject, r.counterparty, r.amount_kopiyky, r.currency
FROM registry_entries r
JOIN protocol_snapshots s ON s.id = r.snapshot_id
WHERE r.association_id = sqlc.arg(association_id)
  AND (sqlc.narg(counterparty)::text IS NULL OR strpos(lower(r.counterparty), lower(sqlc.narg(counterparty)::text)) > 0)
  AND (sqlc.narg(subject)::text IS NULL OR strpos(lower(r.subject), lower(sqlc.narg(subject)::text)) > 0)
  AND (sqlc.narg(amount)::bigint IS NULL OR r.amount_kopiyky = sqlc.narg(amount)::bigint)
  AND (sqlc.narg(amount_from)::bigint IS NULL OR r.amount_kopiyky >= sqlc.narg(amount_from)::bigint)
  AND (sqlc.narg(amount_to)::bigint IS NULL OR r.amount_kopiyky <= sqlc.narg(amount_to)::bigint)
  AND (sqlc.narg(date_from)::date IS NULL OR r.meeting_date >= sqlc.narg(date_from)::date)
  AND (sqlc.narg(date_to)::date IS NULL OR r.meeting_date <= sqlc.narg(date_to)::date)
  AND (sqlc.narg(after_date)::date IS NULL OR (r.meeting_date, r.id) < (sqlc.narg(after_date)::date, sqlc.narg(after_id)::uuid))
ORDER BY r.meeting_date DESC, r.id DESC
LIMIT sqlc.arg(row_limit);

-- name: WithdrawnNumbersBetween :many
-- Numbers held by withdrawn protocols inside lo..hi (inclusive), the gaps the registry marks «знято з розгляду».
SELECT number::int4 AS number
FROM protocol_snapshots
WHERE association_id = sqlc.arg(association_id)
  AND kind = 'withdrawn'
  AND number BETWEEN sqlc.arg(lo)::int4 AND sqlc.arg(hi)::int4
ORDER BY number;

-- name: GetSignedSnapshot :one
-- The signed snapshot of one protocol; a withdrawn protocol has none of this kind.
SELECT protocol_id, number::int4 AS number, meeting_date, snapshot
FROM protocol_snapshots
WHERE association_id = sqlc.arg(association_id)
  AND protocol_id = sqlc.arg(protocol_id)
  AND kind = 'signed';
