-- governance: the nightly verify_protocol_snapshots run (ADR-0007, ADR-0005, Flow 10). One run row
-- written at the end with its counts; one finding per mismatch, missing or restored (untrashed) file.
CREATE TABLE IF NOT EXISTS integrity_runs (
    id               UUID        PRIMARY KEY,
    association_id   UUID        NOT NULL,
    started_at       TIMESTAMPTZ NOT NULL,
    verified_count   INTEGER     NOT NULL,
    mismatched_count INTEGER     NOT NULL,
    restored_count   INTEGER     NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_integrity_runs_association_created
    ON integrity_runs (association_id, created_at DESC);

CREATE TABLE IF NOT EXISTS integrity_findings (
    id               UUID         PRIMARY KEY,
    association_id   UUID         NOT NULL,
    integrity_run_id UUID         NOT NULL REFERENCES integrity_runs (id),
    protocol_id      UUID         NOT NULL REFERENCES protocols (id),
    file_id          UUID         REFERENCES files (id),
    kind             VARCHAR(32)  NOT NULL,
    detail           VARCHAR(500),
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_integrity_findings_run ON integrity_findings (integrity_run_id);
CREATE INDEX IF NOT EXISTS idx_integrity_findings_protocol ON integrity_findings (protocol_id);
CREATE INDEX IF NOT EXISTS idx_integrity_findings_file ON integrity_findings (file_id);
