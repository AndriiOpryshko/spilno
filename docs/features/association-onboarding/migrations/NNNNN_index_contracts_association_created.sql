-- Staged as docs/features/association-onboarding/migrations/NNNNN_index_contracts_association_created.sql (order in MANIFEST.md).
-- Promote with `mise run migrate:new index_contracts_association_created`; never hand-type the number.
-- Owner: finance. Origin: SAD ADR-0002, ADR-0003.
-- Serves ListContracts (WHERE association_id = $1 ORDER BY created_at DESC, id DESC) (data-model.md, Query -> index map). Existing table: CONCURRENTLY, one statement.
-- A failed build leaves an INVALID index; Down removes it, then re-run Up.

-- +goose NO TRANSACTION

-- +goose Up
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_contracts_association_created ON contracts (association_id, created_at DESC, id DESC);

-- +goose Down
DROP INDEX CONCURRENTLY IF EXISTS idx_contracts_association_created;
