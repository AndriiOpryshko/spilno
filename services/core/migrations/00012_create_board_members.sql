-- +goose Up
-- Promoted from docs/features/board-protocol/migrations/20261027120005_create_board_members.up.sql.
-- governance: board roster. One board_members row = one person's one term (re-election is a new
-- entry). Every save of the entry is an append-only revision; the latest revision is current.
-- association_id is an id only: no foreign key across contexts (ADR-0015).
CREATE TABLE IF NOT EXISTS board_members (
    id                   UUID        PRIMARY KEY,
    association_id       UUID        NOT NULL,
    created_by_member_id UUID        REFERENCES board_members (id),
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_board_members_created_by ON board_members (created_by_member_id);

CREATE TABLE IF NOT EXISTS board_member_revisions (
    id                   UUID         PRIMARY KEY,
    association_id       UUID         NOT NULL,
    board_member_id      UUID         NOT NULL REFERENCES board_members (id),
    full_name            VARCHAR(200) NOT NULL,
    sign_in_email        VARCHAR(320),
    term_start           DATE         NOT NULL,
    term_end             DATE,
    created_by_member_id UUID         REFERENCES board_members (id),
    created_at           TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_board_member_revisions_member_created
    ON board_member_revisions (board_member_id, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_board_member_revisions_association_email
    ON board_member_revisions (association_id, sign_in_email);
CREATE INDEX IF NOT EXISTS idx_board_member_revisions_created_by
    ON board_member_revisions (created_by_member_id);

-- +goose Down
DROP INDEX IF EXISTS idx_board_member_revisions_created_by;
DROP INDEX IF EXISTS idx_board_member_revisions_association_email;
DROP INDEX IF EXISTS idx_board_member_revisions_member_created;
DROP TABLE IF EXISTS board_member_revisions;
DROP INDEX IF EXISTS idx_board_members_created_by;
DROP TABLE IF EXISTS board_members;
