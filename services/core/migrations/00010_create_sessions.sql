-- +goose Up
-- Promoted from docs/features/board-protocol/migrations/20261027120003_create_sessions.up.sql.
-- identity: server-side session behind the HttpOnly cookie (ADR-0003). The cookie carries a random
-- token; only its SHA-256 is stored. Sign-out and revocation hard-delete the row.
CREATE TABLE IF NOT EXISTS sessions (
    id           UUID        PRIMARY KEY,
    account_id   UUID        NOT NULL REFERENCES accounts (id),
    token_hash   BYTEA       NOT NULL,
    expires_at   TIMESTAMPTZ NOT NULL,
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_sessions_token_hash UNIQUE (token_hash)
);

CREATE INDEX IF NOT EXISTS idx_sessions_account_id ON sessions (account_id);

-- +goose Down
DROP INDEX IF EXISTS idx_sessions_account_id;
DROP TABLE IF EXISTS sessions;
