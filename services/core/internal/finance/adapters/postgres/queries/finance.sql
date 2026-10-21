-- name: CreateContract :one
INSERT INTO contracts (counterparty, amount_cents, currency, signed_at, ends_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetContract :one
SELECT * FROM contracts WHERE id = $1;

-- name: ListContracts :many
-- Pages newest first by the (created_at, id) keyset; a NULL after_created_at
-- starts from the top.
SELECT * FROM contracts
WHERE sqlc.narg(after_created_at)::timestamptz IS NULL
   OR (created_at, id) < (sqlc.narg(after_created_at)::timestamptz, sqlc.narg(after_id)::uuid)
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg(row_limit);
