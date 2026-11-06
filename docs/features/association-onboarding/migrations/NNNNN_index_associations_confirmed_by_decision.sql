-- Staged as docs/features/association-onboarding/migrations/NNNNN_index_associations_confirmed_by_decision.sql (order in MANIFEST.md).
-- Promote with `mise run migrate:new index_associations_confirmed_by_decision`; never hand-type the number.
-- Owner: identity. Origin: FK index rule; AC-29.
-- Serves the FK associations.confirmed_by_decision_id (data-model.md, Query -> index map). Existing table: CONCURRENTLY, one statement.
-- A failed build leaves an INVALID index; Down removes it, then re-run Up.

-- +goose NO TRANSACTION

-- +goose Up
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_associations_confirmed_by_decision ON associations (confirmed_by_decision_id);

-- +goose Down
DROP INDEX CONCURRENTLY IF EXISTS idx_associations_confirmed_by_decision;
