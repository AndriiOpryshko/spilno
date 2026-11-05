-- Staged as docs/features/association-onboarding/migrations/NNNNN_drop_accounts_association_id.sql (order in MANIFEST.md).
-- Promote with `mise run migrate:new drop_accounts_association_id`; never hand-type the number.
-- Owner: identity. Origin: SAD ADR-0004; PRD §9.
-- Contract of a dropped column. PR 4, after the release that stops reading accounts.association_id.
-- Down re-adds the column exactly (UUID NOT NULL REFERENCES associations, its index), filled from
-- the one association seeded in 00004 — valid only while that is the only association an account
-- could have belonged to (Stage 1); after onboarding a second association a rollback past this
-- file needs the roster to pick the association per account (audit).

-- +goose Up
DROP INDEX IF EXISTS idx_accounts_association_id;
ALTER TABLE accounts DROP COLUMN IF EXISTS association_id;

-- +goose Down
ALTER TABLE accounts ADD COLUMN IF NOT EXISTS association_id UUID REFERENCES associations (id);
UPDATE accounts SET association_id = '00000000-0000-7000-8000-000000000001'::uuid WHERE association_id IS NULL;
ALTER TABLE accounts ALTER COLUMN association_id SET NOT NULL;
CREATE INDEX IF NOT EXISTS idx_accounts_association_id ON accounts (association_id);
