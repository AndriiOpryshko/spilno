-- +goose Up
-- Promoted from docs/features/board-protocol/migrations/20261027120002_create_accounts.up.sql.
-- identity: a person signed in through OIDC (ADR-0003). Linked to the roster by verified email at
-- request time through governance.CurrentMember; no foreign key across contexts.
CREATE TABLE IF NOT EXISTS accounts (
    id             UUID         PRIMARY KEY,
    association_id UUID         NOT NULL REFERENCES associations (id),
    oidc_issuer    VARCHAR(255) NOT NULL,
    oidc_subject   VARCHAR(255) NOT NULL,
    email          VARCHAR(320) NOT NULL,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT uq_accounts_issuer_subject UNIQUE (oidc_issuer, oidc_subject)
);

CREATE INDEX IF NOT EXISTS idx_accounts_association_id ON accounts (association_id);

-- +goose Down
DROP INDEX IF EXISTS idx_accounts_association_id;
DROP TABLE IF EXISTS accounts;
