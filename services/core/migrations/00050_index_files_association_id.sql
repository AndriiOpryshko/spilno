-- Staged as docs/features/association-onboarding/migrations/NNNNN_index_files_association_id.sql (order in MANIFEST.md).
-- Promote with `mise run migrate:new index_files_association_id`; never hand-type the number.
-- Owner: governance. Origin: AC-27b; second re-review of 2026-11-17, R07 (T114).
-- Serves AssociationHasFiles (`association_id = $1`), which completing a storage consent asks before it
-- replaces the connection; no index of `files` led with association_id (data-model.md, Query -> index map).
-- Existing table: CONCURRENTLY, one statement. A failed build leaves an INVALID index; Down removes it, then re-run Up.

-- +goose NO TRANSACTION

-- +goose Up
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_files_association_id ON files (association_id);

-- +goose Down
DROP INDEX CONCURRENTLY IF EXISTS idx_files_association_id;
