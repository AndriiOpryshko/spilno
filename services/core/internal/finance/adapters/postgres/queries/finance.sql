-- name: CreateContract :one
INSERT INTO contracts (counterparty, amount_cents, currency, signed_at, ends_at, association_id)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetContract :one
SELECT * FROM contracts WHERE id = $1 AND association_id = $2;

-- name: ListContracts :many
-- Pages newest first by the (created_at, id) keyset; a NULL after_created_at
-- starts from the top.
SELECT * FROM contracts
WHERE association_id = sqlc.arg(association_id)
  AND (sqlc.narg(after_created_at)::timestamptz IS NULL
   OR (created_at, id) < (sqlc.narg(after_created_at)::timestamptz, sqlc.narg(after_id)::uuid))
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg(row_limit);
