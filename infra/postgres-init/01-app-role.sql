-- Local dev only: the login role the API serves as. spilno_app (the group role) is created by migration 00036.
-- The group role may not exist yet on a fresh volume, so create it here too (the migration is idempotent).
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'spilno_app') THEN
        CREATE ROLE spilno_app NOLOGIN NOBYPASSRLS;
    END IF;
END
$$;
CREATE ROLE spilno_app_login LOGIN PASSWORD 'spilno_app_login' NOBYPASSRLS IN ROLE spilno_app;
