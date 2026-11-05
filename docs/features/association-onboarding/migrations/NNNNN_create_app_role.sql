-- Staged as docs/features/association-onboarding/migrations/NNNNN_create_app_role.sql (order in MANIFEST.md).
-- Promote with `mise run migrate:new create_app_role`; never hand-type the number.
-- Owner: platform. Origin: SAD ADR-0003, ADR-0005, §7.
-- spilno_app: the NOLOGIN group role the API and the worker act as. It is not the owner and has no
-- BYPASSRLS, so row-level security applies to it. The login role per environment (compose, testdb,
-- production) is granted membership in spilno_app outside this file; no password lives here.
-- Insert-only tables get SELECT, INSERT only (ADR-0005, ADR-0007) instead of a trigger.
-- Roles are cluster-wide: Down revokes this database's grants and keeps the role (other test
-- databases of the same cluster may still use it).

-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'spilno_app') THEN
        CREATE ROLE spilno_app NOLOGIN NOBYPASSRLS;
    END IF;
END
$$;
-- +goose StatementEnd

GRANT USAGE ON SCHEMA public TO spilno_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO spilno_app;
GRANT USAGE, SELECT, UPDATE ON ALL SEQUENCES IN SCHEMA public TO spilno_app;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO spilno_app;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT USAGE, SELECT, UPDATE ON SEQUENCES TO spilno_app;
REVOKE UPDATE, DELETE ON application_extracts, application_filings, application_checks,
    application_decisions, association_state_changes FROM spilno_app;

-- +goose Down
ALTER DEFAULT PRIVILEGES IN SCHEMA public REVOKE USAGE, SELECT, UPDATE ON SEQUENCES FROM spilno_app;
ALTER DEFAULT PRIVILEGES IN SCHEMA public REVOKE SELECT, INSERT, UPDATE, DELETE ON TABLES FROM spilno_app;
REVOKE ALL ON ALL SEQUENCES IN SCHEMA public FROM spilno_app;
REVOKE ALL ON ALL TABLES IN SCHEMA public FROM spilno_app;
REVOKE USAGE ON SCHEMA public FROM spilno_app;
