-- governance: the protocol aggregate root (ADR-0009). state, revision and number are the only
-- columns that change; every change is also an append-only protocol_transitions row with its author.
-- number is reserved at the first print copy and never cleared (ADR-0017).
CREATE TABLE IF NOT EXISTS protocols (
    id                      UUID        PRIMARY KEY,
    association_id          UUID        NOT NULL,
    meeting_id              UUID        NOT NULL REFERENCES meetings (id),
    state                   VARCHAR(32) NOT NULL,
    revision                INTEGER     NOT NULL,
    number                  INTEGER,
    copied_from_protocol_id UUID        REFERENCES protocols (id),
    drive_folder_id         VARCHAR(128),
    created_by_member_id    UUID        NOT NULL REFERENCES board_members (id),
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_protocols_association_number UNIQUE (association_id, number)
);

-- AC-03c: one open-or-signed protocol per meeting; a withdrawn one does not block a new draft.
CREATE UNIQUE INDEX IF NOT EXISTS uq_protocols_meeting_not_withdrawn
    ON protocols (meeting_id) WHERE state <> 'withdrawn';
CREATE INDEX IF NOT EXISTS idx_protocols_meeting_id ON protocols (meeting_id);
CREATE INDEX IF NOT EXISTS idx_protocols_copied_from ON protocols (copied_from_protocol_id);
CREATE INDEX IF NOT EXISTS idx_protocols_created_by ON protocols (created_by_member_id);
CREATE INDEX IF NOT EXISTS idx_protocols_association_state_created
    ON protocols (association_id, state, created_at DESC, id DESC);

CREATE TABLE IF NOT EXISTS protocol_transitions (
    id               UUID        PRIMARY KEY,
    association_id   UUID        NOT NULL,
    protocol_id      UUID        NOT NULL REFERENCES protocols (id),
    from_state       VARCHAR(32),
    to_state         VARCHAR(32) NOT NULL,
    revision         INTEGER     NOT NULL,
    actor_member_id  UUID        REFERENCES board_members (id),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_protocol_transitions_protocol_created
    ON protocol_transitions (protocol_id, created_at, id);
CREATE INDEX IF NOT EXISTS idx_protocol_transitions_actor ON protocol_transitions (actor_member_id);
