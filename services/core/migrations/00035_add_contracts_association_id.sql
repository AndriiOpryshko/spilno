-- Staged as docs/features/association-onboarding/migrations/NNNNN_add_contracts_association_id.sql (order in MANIFEST.md).
-- Promote with `mise run migrate:new add_contracts_association_id`; never hand-type the number.
-- Owner: finance. Origin: SAD ADR-0003 (finance.contracts had no association).
-- Phase 1 of 3 (expand). PR 1.
-- id of identity.associations, no FK (ADR 0002). App code in this PR writes it on every insert.

-- +goose Up
ALTER TABLE contracts ADD COLUMN IF NOT EXISTS association_id UUID;

-- +goose Down
ALTER TABLE contracts DROP COLUMN IF EXISTS association_id;
