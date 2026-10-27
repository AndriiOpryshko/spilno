-- identity: the association every owned row is scoped to (ADR-0015). One row in Stage 1.
CREATE TABLE IF NOT EXISTS associations (
    id         UUID         PRIMARY KEY,
    name       VARCHAR(200) NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT now()
);
