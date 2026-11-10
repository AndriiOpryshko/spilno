-- Staged as docs/features/association-onboarding/migrations/NNNNN_index_associations_edrpou.sql (order in MANIFEST.md).
-- Promote with `mise run migrate:new index_associations_edrpou`; never hand-type the number.
-- Owner: identity. Origin: PRD §5.1 identity.association_already_connected, AC-04.
-- Serves GetAssociationByEdrpou (data-model.md, Query -> index map). Existing table: CONCURRENTLY, one statement.
-- A failed build leaves an INVALID index; Down removes it, then re-run Up.

-- +goose NO TRANSACTION

-- +goose Up
CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS uq_associations_edrpou ON associations (edrpou) WHERE edrpou IS NOT NULL;

-- +goose Down
DROP INDEX CONCURRENTLY IF EXISTS uq_associations_edrpou;
