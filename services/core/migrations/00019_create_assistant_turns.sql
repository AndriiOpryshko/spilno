-- +goose Up
-- Promoted from docs/features/board-protocol/migrations/20261027120015_create_assistant_turns.up.sql.
-- governance: the conversation with the помічник протоколу (AC-26: the request, the conversation and
-- every attached document kept with the draft). proposal is the schema-validated structured answer,
-- kept verbatim for audit — opaque to queries, hence JSONB.
CREATE TABLE IF NOT EXISTS assistant_turns (
    id                UUID        PRIMARY KEY,
    association_id    UUID        NOT NULL,
    protocol_id       UUID        NOT NULL REFERENCES protocols (id),
    role              VARCHAR(16) NOT NULL,
    author_member_id  UUID        REFERENCES board_members (id),
    message           TEXT        NOT NULL,
    proposal          JSONB,
    protocol_revision INTEGER     NOT NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_assistant_turns_protocol_created
    ON assistant_turns (protocol_id, created_at, id);
CREATE INDEX IF NOT EXISTS idx_assistant_turns_author ON assistant_turns (author_member_id);

CREATE TABLE IF NOT EXISTS assistant_turn_documents (
    assistant_turn_id UUID        NOT NULL REFERENCES assistant_turns (id),
    case_document_id  UUID        NOT NULL REFERENCES case_documents (id),
    association_id    UUID        NOT NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (assistant_turn_id, case_document_id)
);

CREATE INDEX IF NOT EXISTS idx_assistant_turn_documents_case_document
    ON assistant_turn_documents (case_document_id);

-- +goose Down
DROP INDEX IF EXISTS idx_assistant_turn_documents_case_document;
DROP TABLE IF EXISTS assistant_turn_documents;
DROP INDEX IF EXISTS idx_assistant_turns_author;
DROP INDEX IF EXISTS idx_assistant_turns_protocol_created;
DROP TABLE IF EXISTS assistant_turns;
