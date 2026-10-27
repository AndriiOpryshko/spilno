-- governance: signed and withdrawn snapshots (ADR-0007). snapshot is the canonical, versioned
-- document that the fingerprint covers — read and hashed as a whole, hence JSONB. One per protocol.
-- registry_entries is a rebuildable projection of signed snapshots, one row per decision, for the
-- реєстр протоколів search (AC-18).
CREATE TABLE IF NOT EXISTS protocol_snapshots (
    id                   UUID        PRIMARY KEY,
    association_id       UUID        NOT NULL,
    protocol_id          UUID        NOT NULL REFERENCES protocols (id),
    kind                 VARCHAR(16) NOT NULL,
    number               INTEGER,
    meeting_date         DATE        NOT NULL,
    schema_version       SMALLINT    NOT NULL,
    snapshot             JSONB       NOT NULL,
    fingerprint          BYTEA       NOT NULL,
    created_by_member_id UUID        NOT NULL REFERENCES board_members (id),
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_protocol_snapshots_protocol_id UNIQUE (protocol_id)
);

CREATE INDEX IF NOT EXISTS idx_protocol_snapshots_association_kind_number
    ON protocol_snapshots (association_id, kind, number);
CREATE INDEX IF NOT EXISTS idx_protocol_snapshots_created_by ON protocol_snapshots (created_by_member_id);

CREATE TABLE IF NOT EXISTS registry_entries (
    id             UUID         PRIMARY KEY,
    association_id UUID         NOT NULL,
    snapshot_id    UUID         NOT NULL REFERENCES protocol_snapshots (id),
    agenda_item_id UUID         NOT NULL REFERENCES agenda_items (id),
    number         INTEGER      NOT NULL,
    meeting_date   DATE         NOT NULL,
    subject        VARCHAR(500) NOT NULL,
    counterparty   VARCHAR(500),
    amount_kopiyky BIGINT,
    currency       VARCHAR(3),
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT uq_registry_entries_snapshot_item UNIQUE (snapshot_id, agenda_item_id)
);

CREATE INDEX IF NOT EXISTS idx_registry_entries_association_meeting_date
    ON registry_entries (association_id, meeting_date DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_registry_entries_agenda_item ON registry_entries (agenda_item_id);
