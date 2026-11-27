-- name: CreateContract :one
INSERT INTO contracts (counterparty, amount_cents, currency, signed_at, ends_at, association_id, uah_amount_kopiyky, uah_rate_micros)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: SetContractUAHEquivalent :one
-- Sets the equivalent only on a contract that has none.
UPDATE contracts SET uah_amount_kopiyky = $3, uah_rate_micros = $4
WHERE id = $1 AND association_id = $2 AND uah_amount_kopiyky IS NULL
RETURNING *;

-- name: GetContract :one
SELECT * FROM contracts WHERE id = $1 AND association_id = $2;

-- name: ListContracts :many
-- Pages newest first by the (created_at, id) keyset; a NULL after_created_at
-- starts from the top. missing_uah_equivalent keeps only the foreign-currency
-- contracts without a hryvnia equivalent (a UAH contract never has one).
SELECT * FROM contracts
WHERE association_id = sqlc.arg(association_id)
  AND (NOT sqlc.arg(missing_uah_equivalent)::boolean
   OR (uah_amount_kopiyky IS NULL AND currency <> 'UAH'))
  AND (sqlc.narg(after_created_at)::timestamptz IS NULL
   OR (created_at, id) < (sqlc.narg(after_created_at)::timestamptz, sqlc.narg(after_id)::uuid))
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg(row_limit);
