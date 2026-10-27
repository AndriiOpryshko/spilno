-- governance: the meeting facts entered from the signed paper together with the scan (AC-14, AC-34,
-- AC-15, ADR-0018). Never pre-filled from the poll (AC-41). Written once, in the lock transaction.
CREATE TABLE IF NOT EXISTS meeting_facts (
    id                          UUID        PRIMARY KEY,
    association_id              UUID        NOT NULL,
    protocol_id                 UUID        NOT NULL REFERENCES protocols (id),
    print_copy_id               UUID        NOT NULL REFERENCES print_copies (id),
    signed_scan_file_id         UUID        NOT NULL REFERENCES files (id),
    has_handwritten_corrections BOOLEAN     NOT NULL,
    corrections_note            TEXT,
    entered_by_member_id        UUID        NOT NULL REFERENCES board_members (id),
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_meeting_facts_protocol_id UNIQUE (protocol_id),
    CONSTRAINT uq_meeting_facts_signed_scan UNIQUE (signed_scan_file_id)
);

CREATE INDEX IF NOT EXISTS idx_meeting_facts_print_copy ON meeting_facts (print_copy_id);
CREATE INDEX IF NOT EXISTS idx_meeting_facts_entered_by ON meeting_facts (entered_by_member_id);

CREATE TABLE IF NOT EXISTS meeting_attendees (
    meeting_facts_id UUID        NOT NULL REFERENCES meeting_facts (id),
    board_member_id  UUID        NOT NULL REFERENCES board_members (id),
    association_id   UUID        NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (meeting_facts_id, board_member_id)
);

CREATE INDEX IF NOT EXISTS idx_meeting_attendees_board_member ON meeting_attendees (board_member_id);

CREATE TABLE IF NOT EXISTS meeting_item_votes (
    meeting_facts_id UUID        NOT NULL REFERENCES meeting_facts (id),
    agenda_item_id   UUID        NOT NULL REFERENCES agenda_items (id),
    association_id   UUID        NOT NULL,
    votes_for        SMALLINT    NOT NULL,
    votes_against    SMALLINT    NOT NULL,
    votes_abstained  SMALLINT    NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (meeting_facts_id, agenda_item_id)
);

CREATE INDEX IF NOT EXISTS idx_meeting_item_votes_agenda_item ON meeting_item_votes (agenda_item_id);
