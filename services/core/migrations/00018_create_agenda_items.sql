-- +goose Up
-- Promoted from docs/features/board-protocol/migrations/20261027120014_create_agenda_items.up.sql.
-- governance: agenda items of a protocol, one per справа (AC-03, AC-03c). Subject and decision text
-- are append-only versions (AC-29: new wording kept with its author and time); latest is current.
-- fragments is the ADR-0011 sequence [{text} | {slot} | {meeting}] — a polymorphic list, hence JSONB.
CREATE TABLE IF NOT EXISTS agenda_items (
    id                   UUID        PRIMARY KEY,
    association_id       UUID        NOT NULL,
    protocol_id          UUID        NOT NULL REFERENCES protocols (id),
    case_id              UUID        REFERENCES cases (id),
    position             SMALLINT    NOT NULL,
    created_by_member_id UUID        REFERENCES board_members (id),
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_agenda_items_protocol_position UNIQUE (protocol_id, position)
);

CREATE INDEX IF NOT EXISTS idx_agenda_items_case_id ON agenda_items (case_id);
CREATE INDEX IF NOT EXISTS idx_agenda_items_created_by ON agenda_items (created_by_member_id);

CREATE TABLE IF NOT EXISTS agenda_item_versions (
    id                UUID         PRIMARY KEY,
    association_id    UUID         NOT NULL,
    agenda_item_id    UUID         NOT NULL REFERENCES agenda_items (id),
    protocol_revision INTEGER      NOT NULL,
    subject           VARCHAR(500),
    fragments         JSONB        NOT NULL,
    author_member_id  UUID         REFERENCES board_members (id),
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_agenda_item_versions_item_created
    ON agenda_item_versions (agenda_item_id, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_agenda_item_versions_author ON agenda_item_versions (author_member_id);

-- +goose Down
DROP INDEX IF EXISTS idx_agenda_item_versions_author;
DROP INDEX IF EXISTS idx_agenda_item_versions_item_created;
DROP TABLE IF EXISTS agenda_item_versions;
DROP INDEX IF EXISTS idx_agenda_items_created_by;
DROP INDEX IF EXISTS idx_agenda_items_case_id;
DROP TABLE IF EXISTS agenda_items;
