-- Staged as docs/features/association-onboarding/migrations/NNNNN_contract_contracts_association_id.sql (order in MANIFEST.md).
-- Promote with `mise run migrate:new contract_contracts_association_id`; never hand-type the number.
-- Owner: finance. Origin: SAD ADR-0003.
-- Phase 3 of 3 (contract). PR 3, after PR 2 is deployed and the backfill check returned 0.
-- NOT VALID CHECK -> VALIDATE -> SET NOT NULL (uses the validated CHECK) -> drop the helper.

-- +goose Up
ALTER TABLE contracts ADD CONSTRAINT contracts_association_id_not_null CHECK (association_id IS NOT NULL) NOT VALID;
ALTER TABLE contracts VALIDATE CONSTRAINT contracts_association_id_not_null;
ALTER TABLE contracts ALTER COLUMN association_id SET NOT NULL;
ALTER TABLE contracts DROP CONSTRAINT contracts_association_id_not_null;

-- +goose Down
ALTER TABLE contracts ALTER COLUMN association_id DROP NOT NULL;
