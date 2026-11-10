-- Staged as docs/features/association-onboarding/migrations/NNNNN_enable_rls_identity.sql (order in MANIFEST.md).
-- Promote with `mise run migrate:new enable_rls_identity`; never hand-type the number.
-- Owner: identity. Origin: SAD ADR-0003, ADR-0008.
-- Row-level security (ADR-0003): rows are visible to spilno_app only for the association set by
-- platform/db.InAssociation (set_config('app.association_id', ..., true)); unset -> no rows (fail
-- closed). FORCE applies it to the table owner too; the migration/owner role therefore needs
-- BYPASSRLS (a superuser in compose and testdb has it) so goose and backfills still see every row.

-- +goose Up
ALTER TABLE google_connections ENABLE ROW LEVEL SECURITY;
ALTER TABLE google_connections FORCE ROW LEVEL SECURITY;

CREATE POLICY association_isolation ON google_connections
    USING (association_id = NULLIF(current_setting('app.association_id', true), '')::uuid)
    WITH CHECK (association_id = NULLIF(current_setting('app.association_id', true), '')::uuid);

-- +goose Down
DROP POLICY IF EXISTS association_isolation ON google_connections;
ALTER TABLE google_connections NO FORCE ROW LEVEL SECURITY;
ALTER TABLE google_connections DISABLE ROW LEVEL SECURITY;
