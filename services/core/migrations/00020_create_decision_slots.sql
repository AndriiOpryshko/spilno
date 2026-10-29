-- +goose Up
-- Promoted from docs/features/board-protocol/migrations/20261027120016_create_decision_slots.up.sql.
-- governance: слоти рішення. decision_slots is the slot's identity and current status (the AC-21
-- print gate reads it); every value is an append-only slot_values row, every confirmation an
-- append-only slot_confirmations row (AC-21, AC-22, AC-28b, AC-09). Typed value columns, not JSONB:
-- only the columns of the slot's kind are set (enforced in domain, not in the DB).
CREATE TABLE IF NOT EXISTS decision_slots (
    id             UUID        PRIMARY KEY,
    association_id UUID        NOT NULL,
    agenda_item_id UUID        NOT NULL REFERENCES agenda_items (id),
    kind           VARCHAR(32) NOT NULL,
    status         VARCHAR(16) NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_decision_slots_agenda_item_id ON decision_slots (agenda_item_id);

CREATE TABLE IF NOT EXISTS slot_values (
    id                     UUID         PRIMARY KEY,
    association_id         UUID         NOT NULL,
    slot_id                UUID         NOT NULL REFERENCES decision_slots (id),
    origin                 VARCHAR(16)  NOT NULL,
    amount_kopiyky         BIGINT,
    currency               VARCHAR(3),
    text_value             VARCHAR(500),
    date_value             DATE,
    date_end_value         DATE,
    person_member_id       UUID         REFERENCES board_members (id),
    source_document_id     UUID         REFERENCES case_documents (id),
    source_fragment        TEXT,
    assistant_turn_id      UUID         REFERENCES assistant_turns (id),
    author_member_id       UUID         REFERENCES board_members (id),
    created_at             TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_slot_values_slot_created ON slot_values (slot_id, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_slot_values_source_document ON slot_values (source_document_id);
CREATE INDEX IF NOT EXISTS idx_slot_values_person ON slot_values (person_member_id);
CREATE INDEX IF NOT EXISTS idx_slot_values_assistant_turn ON slot_values (assistant_turn_id);
CREATE INDEX IF NOT EXISTS idx_slot_values_author ON slot_values (author_member_id);

CREATE TABLE IF NOT EXISTS slot_confirmations (
    id                     UUID        PRIMARY KEY,
    association_id         UUID        NOT NULL,
    slot_id                UUID        NOT NULL REFERENCES decision_slots (id),
    slot_value_id          UUID        NOT NULL REFERENCES slot_values (id),
    protocol_revision      INTEGER     NOT NULL,
    confirmed_by_member_id UUID        NOT NULL REFERENCES board_members (id),
    created_at             TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_slot_confirmations_slot_created
    ON slot_confirmations (slot_id, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_slot_confirmations_value ON slot_confirmations (slot_value_id);
CREATE INDEX IF NOT EXISTS idx_slot_confirmations_confirmed_by ON slot_confirmations (confirmed_by_member_id);

-- +goose Down
DROP INDEX IF EXISTS idx_slot_confirmations_confirmed_by;
DROP INDEX IF EXISTS idx_slot_confirmations_value;
DROP INDEX IF EXISTS idx_slot_confirmations_slot_created;
DROP TABLE IF EXISTS slot_confirmations;
DROP INDEX IF EXISTS idx_slot_values_author;
DROP INDEX IF EXISTS idx_slot_values_assistant_turn;
DROP INDEX IF EXISTS idx_slot_values_person;
DROP INDEX IF EXISTS idx_slot_values_source_document;
DROP INDEX IF EXISTS idx_slot_values_slot_created;
DROP TABLE IF EXISTS slot_values;
DROP INDEX IF EXISTS idx_decision_slots_agenda_item_id;
DROP TABLE IF EXISTS decision_slots;
