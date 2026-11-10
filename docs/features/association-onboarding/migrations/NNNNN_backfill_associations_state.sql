-- Staged as docs/features/association-onboarding/migrations/NNNNN_backfill_associations_state.sql (order in MANIFEST.md).
-- Promote with `mise run migrate:new backfill_associations_state`; never hand-type the number.
-- Owner: identity. Origin: PRD §9 (the hand-made association becomes підключене ОСББ).
-- Phase 2 of 3 (backfill). PR 2, after PR 1 is deployed.
-- Fills state for rows written before the dual write; touches only rows still NULL (resumable).
-- Small table (see backfill-associations-state.md): one transaction, no COMMIT in the loop.

-- +goose Up
UPDATE associations SET state = 'connected' WHERE state IS NULL;

-- +goose Down
-- Intentionally a no-op: the expand Down drops the column, which removes the backfilled values.
SELECT 1;
