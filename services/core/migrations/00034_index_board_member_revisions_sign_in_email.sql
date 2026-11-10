-- Staged as docs/features/association-onboarding/migrations/NNNNN_index_board_member_revisions_sign_in_email.sql (order in MANIFEST.md).
-- Promote with `mise run migrate:new index_board_member_revisions_sign_in_email`; never hand-type the number.
-- Owner: governance. Origin: AC-04d; SAD ADR-0004 MembershipsOf.
-- Serves MembershipsOf (WHERE sign_in_email = $1 across associations) (data-model.md, Query -> index map). Existing table: CONCURRENTLY, one statement.
-- A failed build leaves an INVALID index; Down removes it, then re-run Up.

-- +goose NO TRANSACTION

-- +goose Up
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_board_member_revisions_sign_in_email ON board_member_revisions (sign_in_email);

-- +goose Down
DROP INDEX CONCURRENTLY IF EXISTS idx_board_member_revisions_sign_in_email;
