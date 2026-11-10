-- Staged as docs/features/association-onboarding/migrations/NNNNN_enable_rls_finance.sql (order in MANIFEST.md).
-- Promote with `mise run migrate:new enable_rls_finance`; never hand-type the number.
-- Owner: finance. Origin: SAD ADR-0003.
-- Row-level security (ADR-0003): rows are visible to spilno_app only for the association set by
-- platform/db.InAssociation (set_config('app.association_id', ..., true)); unset -> no rows (fail
-- closed). FORCE applies it to the table owner too; the migration/owner role therefore needs
-- BYPASSRLS (a superuser in compose and testdb has it) so goose and backfills still see every row.

-- +goose Up
ALTER TABLE contracts ENABLE ROW LEVEL SECURITY;
ALTER TABLE contracts FORCE ROW LEVEL SECURITY;

CREATE POLICY association_isolation ON contracts
    USING (association_id = NULLIF(current_setting('app.association_id', true), '')::uuid)
    WITH CHECK (association_id = NULLIF(current_setting('app.association_id', true), '')::uuid);

-- +goose Down
DROP POLICY IF EXISTS association_isolation ON contracts;
ALTER TABLE contracts NO FORCE ROW LEVEL SECURITY;
ALTER TABLE contracts DISABLE ROW LEVEL SECURITY;
