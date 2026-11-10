-- Staged as docs/features/association-onboarding/migrations/NNNNN_enable_rls_identity_state_changes.sql (order in MANIFEST.md).
-- Promote with `mise run migrate:new enable_rls_identity_state_changes`; never hand-type the number.
-- Owner: identity. Origin: SAD ADR-0003, ADR-0007.
-- Row-level security (ADR-0003) for the one identity table that always names its association and was
-- left out of NNNNN_enable_rls_identity.sql: rows are visible to spilno_app only for the association set
-- by platform/db.InAssociation; unset -> no rows (fail closed). The оператор disables and enables an
-- association, and reads its reason, inside InAssociation for that association, as an approval does.
-- `applications` and `application_decisions` also carry an association_id but stay without a policy:
-- it is NULL before approval and they are reached only by account or by the оператор (ADR-0003).

-- +goose Up
ALTER TABLE association_state_changes ENABLE ROW LEVEL SECURITY;
ALTER TABLE association_state_changes FORCE ROW LEVEL SECURITY;

CREATE POLICY association_isolation ON association_state_changes
    USING (association_id = NULLIF(current_setting('app.association_id', true), '')::uuid)
    WITH CHECK (association_id = NULLIF(current_setting('app.association_id', true), '')::uuid);

-- +goose Down
DROP POLICY IF EXISTS association_isolation ON association_state_changes;
ALTER TABLE association_state_changes NO FORCE ROW LEVEL SECURITY;
ALTER TABLE association_state_changes DISABLE ROW LEVEL SECURITY;
