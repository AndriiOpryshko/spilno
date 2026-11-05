-- Staged as docs/features/association-onboarding/migrations/NNNNN_create_applications.sql (order in MANIFEST.md).
-- Promote with `mise run migrate:new create_applications`; never hand-type the number.
-- Owner: identity. Origin: PRD US-01..US-05, AC-01..AC-20, §9; SAD ADR-0005.
-- identity: заявка на підключення and запит на передачу share one state machine (PRD §10), so one
-- table with a kind. state is the only column that changes; filings, the оператор's checks and
-- decisions are insert-only rows (spilno_app has SELECT, INSERT only — ADR-0005, no trigger).
-- The extract file is kept as uploaded, never parsed (ADR-0005).

-- +goose Up
CREATE TABLE IF NOT EXISTS application_extracts (
    id            UUID         PRIMARY KEY,
    account_id    UUID         NOT NULL REFERENCES accounts (id),
    content       BYTEA        NOT NULL,
    sha256        BYTEA        NOT NULL,
    size_bytes    BIGINT       NOT NULL,
    media_type    VARCHAR(100) NOT NULL,
    original_name VARCHAR(255) NOT NULL,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_application_extracts_account_id ON application_extracts (account_id);

CREATE TABLE IF NOT EXISTS applications (
    id             UUID        PRIMARY KEY,
    kind           VARCHAR(32) NOT NULL,
    account_id     UUID        NOT NULL REFERENCES accounts (id),
    edrpou         VARCHAR(8)  NOT NULL,
    association_id UUID        REFERENCES associations (id),
    state          VARCHAR(32) NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- AC-04b: one open заявка of a person per code; another person's pending заявка is allowed (AC-04c).
CREATE UNIQUE INDEX IF NOT EXISTS uq_applications_account_edrpou_pending
    ON applications (account_id, edrpou) WHERE state = 'pending';
CREATE INDEX IF NOT EXISTS idx_applications_account_created
    ON applications (account_id, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_applications_state_created ON applications (state, created_at, id);
CREATE INDEX IF NOT EXISTS idx_applications_edrpou_state ON applications (edrpou, state);
CREATE INDEX IF NOT EXISTS idx_applications_association_id ON applications (association_id);

CREATE TABLE IF NOT EXISTS application_filings (
    id                  UUID         PRIMARY KEY,
    application_id      UUID         NOT NULL REFERENCES applications (id),
    extract_id          UUID         NOT NULL REFERENCES application_extracts (id),
    association_name    VARCHAR(200) NOT NULL,
    association_address VARCHAR(500) NOT NULL,
    applicant_full_name VARCHAR(200) NOT NULL,
    position            VARCHAR(32)  NOT NULL,
    term_start          DATE         NOT NULL,
    term_end            DATE,
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_application_filings_application_created
    ON application_filings (application_id, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_application_filings_extract_id ON application_filings (extract_id);

CREATE TABLE IF NOT EXISTS application_checks (
    id                UUID         PRIMARY KEY,
    filing_id         UUID         NOT NULL REFERENCES application_filings (id),
    signature_valid   BOOLEAN      NOT NULL,
    signer_full_name  VARCHAR(200) NOT NULL,
    signed_on         DATE         NOT NULL,
    head_full_name    VARCHAR(200) NOT NULL,
    extract_edrpou    VARCHAR(8)   NOT NULL,
    extract_dated_on  DATE         NOT NULL,
    checked_by_email  VARCHAR(320) NOT NULL,
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_application_checks_filing_created
    ON application_checks (filing_id, created_at DESC, id DESC);

CREATE TABLE IF NOT EXISTS application_decisions (
    id                 UUID          PRIMARY KEY,
    application_id     UUID          NOT NULL REFERENCES applications (id),
    filing_id          UUID          NOT NULL REFERENCES application_filings (id),
    check_id           UUID          REFERENCES application_checks (id),
    association_id     UUID          REFERENCES associations (id),
    outcome            VARCHAR(32)   NOT NULL,
    reason             VARCHAR(1000),
    relied_on_document VARCHAR(500),
    decided_by_email   VARCHAR(320)  NOT NULL,
    created_at         TIMESTAMPTZ   NOT NULL DEFAULT now(),
    -- AC-15: one decision per filing; a refile is a new filing.
    CONSTRAINT uq_application_decisions_filing_id UNIQUE (filing_id)
);

CREATE INDEX IF NOT EXISTS idx_application_decisions_application_id ON application_decisions (application_id);
CREATE INDEX IF NOT EXISTS idx_application_decisions_check_id ON application_decisions (check_id);
CREATE INDEX IF NOT EXISTS idx_application_decisions_association_created
    ON application_decisions (association_id, created_at DESC, id DESC);

-- +goose Down
DROP TABLE IF EXISTS application_decisions;
DROP TABLE IF EXISTS application_checks;
DROP TABLE IF EXISTS application_filings;
DROP TABLE IF EXISTS applications;
DROP TABLE IF EXISTS application_extracts;
