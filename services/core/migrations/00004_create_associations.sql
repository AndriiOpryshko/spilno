-- +goose Up
-- identity: the association every owned row is scoped to (ADR-0015). One row in Stage 1,
-- with a deterministic id that the meetings backfill (00006) uses.
CREATE TABLE associations (
    id uuid PRIMARY KEY,
    name text NOT NULL CHECK (char_length(name) BETWEEN 1 AND 200),
    created_at timestamptz NOT NULL DEFAULT now()
);

INSERT INTO associations (id, name)
VALUES ('00000000-0000-7000-8000-000000000001', 'Test Association')
ON CONFLICT DO NOTHING;

-- +goose Down
DROP TABLE associations;
