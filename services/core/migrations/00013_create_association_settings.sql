-- +goose Up
-- Promoted from docs/features/board-protocol/migrations/20261027120006_create_association_settings.up.sql.
-- governance: association settings (AC-42) as append-only revisions — the latest is current.
-- The register counter lives in its own row (ADR-0017, moved out of the settings revisions):
-- next_number is taken under the row lock at the first print copy and never given back.
CREATE TABLE IF NOT EXISTS association_settings (
    id                    UUID         PRIMARY KEY,
    association_id        UUID         NOT NULL,
    meeting_form          VARCHAR(500) NOT NULL,
    signers               VARCHAR(500) NOT NULL,
    register_start_number INTEGER      NOT NULL,
    created_by_member_id  UUID         NOT NULL REFERENCES board_members (id),
    created_at            TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_association_settings_association_created
    ON association_settings (association_id, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_association_settings_created_by
    ON association_settings (created_by_member_id);

CREATE TABLE IF NOT EXISTS protocol_registers (
    id             UUID        PRIMARY KEY,
    association_id UUID        NOT NULL,
    next_number    INTEGER     NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_protocol_registers_association_id UNIQUE (association_id)
);

-- +goose Down
DROP TABLE IF EXISTS protocol_registers;
DROP INDEX IF EXISTS idx_association_settings_created_by;
DROP INDEX IF EXISTS idx_association_settings_association_created;
DROP TABLE IF EXISTS association_settings;
