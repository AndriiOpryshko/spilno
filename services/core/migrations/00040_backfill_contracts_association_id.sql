-- Staged as docs/features/association-onboarding/migrations/NNNNN_backfill_contracts_association_id.sql (order in MANIFEST.md).
-- Promote with `mise run migrate:new backfill_contracts_association_id`; never hand-type the number.
-- Owner: finance. Origin: SAD ADR-0003 (all existing contracts belong to the one association seeded in 00004).
-- Phase 2 of 3 (backfill). PR 2, after PR 1 is deployed.
-- Fills association_id for rows written before the dual write; touches only rows still NULL (resumable).
-- Small table (see backfill-contracts-association-id.md): one transaction, no COMMIT in the loop.

-- +goose Up
UPDATE contracts SET association_id = '00000000-0000-7000-8000-000000000001'::uuid WHERE association_id IS NULL;

-- +goose Down
-- Intentionally a no-op: the expand Down drops the column, which removes the backfilled values.
SELECT 1;
