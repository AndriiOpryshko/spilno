-- Staged as docs/features/association-onboarding/migrations/NNNNN_index_board_member_revisions_lower_sign_in_email.sql (order in MANIFEST.md).
-- Promote with `mise run migrate:new index_board_member_revisions_lower_sign_in_email`; never hand-type the number.
-- Owner: governance. Origin: AC-04d; review of 2026-11-10, F20 (T59).
-- Serves MembershipsOf and the own-entry clause of the row-level security policy, which both compare
-- lower(sign_in_email); the plain index of idx_board_member_revisions_sign_in_email cannot serve them
-- (data-model.md, Query -> index map). Existing table: CONCURRENTLY, one statement.
-- A failed build leaves an INVALID index; Down removes it, then re-run Up.

-- +goose NO TRANSACTION

-- +goose Up
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_board_member_revisions_lower_sign_in_email ON board_member_revisions (lower(sign_in_email));

-- +goose Down
DROP INDEX CONCURRENTLY IF EXISTS idx_board_member_revisions_lower_sign_in_email;
