-- Staged as docs/features/association-onboarding/migrations/NNNNN_backfill_board_member_revisions_position.sql (order in MANIFEST.md).
-- Promote with `mise run migrate:new backfill_board_member_revisions_position`; never hand-type the number.
-- Owner: governance. Origin: PRD §9 (existing entries are члени правління; nobody is голова until AC-19c).
-- Phase 2 of 3 (backfill). PR 2, after PR 1 is deployed.
-- Fills position for rows written before the dual write; touches only rows still NULL (resumable).
-- Small table (see backfill-board-member-revisions-position.md): one transaction, no COMMIT in the loop.

-- +goose Up
UPDATE board_member_revisions SET position = 'member' WHERE position IS NULL;

-- +goose Down
-- Intentionally a no-op: the expand Down drops the column, which removes the backfilled values.
SELECT 1;
