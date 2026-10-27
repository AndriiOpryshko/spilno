-- governance: попереднє голосування (AC-30, AC-31, AC-05d, AC-05e). A round freezes the wording at a
-- revision; frozen is that immutable copy (agenda items, fragments, slot values), read only as a
-- whole, hence JSONB. Votes are append-only; the latest vote of a member on an item counts.
CREATE TABLE IF NOT EXISTS poll_rounds (
    id                  UUID        PRIMARY KEY,
    association_id      UUID        NOT NULL,
    protocol_id         UUID        NOT NULL REFERENCES protocols (id),
    round_number        INTEGER     NOT NULL,
    protocol_revision   INTEGER     NOT NULL,
    frozen              JSONB       NOT NULL,
    opened_by_member_id UUID        NOT NULL REFERENCES board_members (id),
    closed_at           TIMESTAMPTZ,
    closed_by_member_id UUID        REFERENCES board_members (id),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_poll_rounds_protocol_round UNIQUE (protocol_id, round_number)
);

CREATE INDEX IF NOT EXISTS idx_poll_rounds_opened_by ON poll_rounds (opened_by_member_id);
CREATE INDEX IF NOT EXISTS idx_poll_rounds_closed_by ON poll_rounds (closed_by_member_id);

CREATE TABLE IF NOT EXISTS poll_votes (
    id              UUID        PRIMARY KEY,
    association_id  UUID        NOT NULL,
    poll_round_id   UUID        NOT NULL REFERENCES poll_rounds (id),
    agenda_item_id  UUID        NOT NULL REFERENCES agenda_items (id),
    board_member_id UUID        NOT NULL REFERENCES board_members (id),
    choice          VARCHAR(16) NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_poll_votes_round_item_member_created
    ON poll_votes (poll_round_id, agenda_item_id, board_member_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_poll_votes_agenda_item ON poll_votes (agenda_item_id);
CREATE INDEX IF NOT EXISTS idx_poll_votes_board_member ON poll_votes (board_member_id);
