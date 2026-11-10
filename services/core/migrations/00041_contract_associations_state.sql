-- Staged as docs/features/association-onboarding/migrations/NNNNN_contract_associations_state.sql (order in MANIFEST.md).
-- Promote with `mise run migrate:new contract_associations_state`; never hand-type the number.
-- Owner: identity. Origin: SAD ADR-0007.
-- Phase 3 of 3 (contract). PR 3, after PR 2 is deployed and the backfill check returned 0.
-- NOT VALID CHECK -> VALIDATE -> SET NOT NULL (uses the validated CHECK) -> drop the helper.

-- +goose Up
ALTER TABLE associations ADD CONSTRAINT associations_state_not_null CHECK (state IS NOT NULL) NOT VALID;
ALTER TABLE associations VALIDATE CONSTRAINT associations_state_not_null;
ALTER TABLE associations ALTER COLUMN state SET NOT NULL;
ALTER TABLE associations DROP CONSTRAINT associations_state_not_null;

-- +goose Down
ALTER TABLE associations ALTER COLUMN state DROP NOT NULL;
