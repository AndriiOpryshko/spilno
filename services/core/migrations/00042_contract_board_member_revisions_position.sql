-- Staged as docs/features/association-onboarding/migrations/NNNNN_contract_board_member_revisions_position.sql (order in MANIFEST.md).
-- Promote with `mise run migrate:new contract_board_member_revisions_position`; never hand-type the number.
-- Owner: governance. Origin: SAD ADR-0006.
-- Phase 3 of 3 (contract). PR 3, after PR 2 is deployed and the backfill check returned 0.
-- NOT VALID CHECK -> VALIDATE -> SET NOT NULL (uses the validated CHECK) -> drop the helper.

-- +goose Up
ALTER TABLE board_member_revisions ADD CONSTRAINT board_member_revisions_position_not_null CHECK (position IS NOT NULL) NOT VALID;
ALTER TABLE board_member_revisions VALIDATE CONSTRAINT board_member_revisions_position_not_null;
ALTER TABLE board_member_revisions ALTER COLUMN position SET NOT NULL;
ALTER TABLE board_member_revisions DROP CONSTRAINT board_member_revisions_position_not_null;

-- +goose Down
ALTER TABLE board_member_revisions ALTER COLUMN position DROP NOT NULL;
