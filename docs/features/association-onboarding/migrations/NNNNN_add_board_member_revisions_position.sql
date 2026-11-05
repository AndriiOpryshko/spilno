-- Staged as docs/features/association-onboarding/migrations/NNNNN_add_board_member_revisions_position.sql (order in MANIFEST.md).
-- Promote with `mise run migrate:new add_board_member_revisions_position`; never hand-type the number.
-- Owner: governance. Origin: PRD §9, AC-08, AC-19, AC-20; SAD ADR-0006.
-- Phase 1 of 3 (expand) for position. PR 1. decision_id stays nullable (entries made by the board have none).
-- decision_id: id of identity.application_decisions, no FK (ADR 0002). Not queried by itself: read
-- with the revision, so no index.
-- App code in this PR writes position on every new revision.

-- +goose Up
ALTER TABLE board_member_revisions ADD COLUMN IF NOT EXISTS position VARCHAR(32);
ALTER TABLE board_member_revisions ADD COLUMN IF NOT EXISTS decision_id UUID;

-- +goose Down
ALTER TABLE board_member_revisions DROP COLUMN IF EXISTS decision_id;
ALTER TABLE board_member_revisions DROP COLUMN IF EXISTS position;
