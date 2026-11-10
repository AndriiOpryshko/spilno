-- Staged as docs/features/association-onboarding/migrations/NNNNN_create_association_state_changes.sql (order in MANIFEST.md).
-- Promote with `mise run migrate:new create_association_state_changes`; never hand-type the number.
-- Owner: identity. Origin: PRD US-06, AC-21..AC-24; SAD ADR-0007.
-- identity: insert-only history of disabling and enabling; associations.state holds the current one.

-- +goose Up
CREATE TABLE IF NOT EXISTS association_state_changes (
    id               UUID          PRIMARY KEY,
    association_id   UUID          NOT NULL REFERENCES associations (id),
    to_state         VARCHAR(32)   NOT NULL,
    reason           VARCHAR(1000),
    changed_by_email VARCHAR(320)  NOT NULL,
    created_at       TIMESTAMPTZ   NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_association_state_changes_association_created
    ON association_state_changes (association_id, created_at DESC, id DESC);

-- +goose Down
DROP TABLE IF EXISTS association_state_changes;
