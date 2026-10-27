-- governance: PDF print copies (AC-14c, AC-14b, ADR-0013). Written by render_print_copy once the PDF
-- is in Drive; (protocol_id, protocol_revision) is the job's idempotency key. A copy is voided by any
-- edit and kept in the history.
CREATE TABLE IF NOT EXISTS print_copies (
    id                     UUID        PRIMARY KEY,
    association_id         UUID        NOT NULL,
    protocol_id            UUID        NOT NULL REFERENCES protocols (id),
    protocol_revision      INTEGER     NOT NULL,
    number                 INTEGER     NOT NULL,
    file_id                UUID        NOT NULL REFERENCES files (id),
    requested_by_member_id UUID        NOT NULL REFERENCES board_members (id),
    voided_at              TIMESTAMPTZ,
    created_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_print_copies_protocol_revision UNIQUE (protocol_id, protocol_revision),
    CONSTRAINT uq_print_copies_file_id UNIQUE (file_id)
);

CREATE INDEX IF NOT EXISTS idx_print_copies_requested_by ON print_copies (requested_by_member_id);
