-- governance: file metadata (ADR-0005 — bytes live in the association's Drive), справи and their
-- document versions. A newer version of a document is a new file and a new case_documents row.
CREATE TABLE IF NOT EXISTS files (
    id                    UUID         PRIMARY KEY,
    association_id        UUID         NOT NULL,
    drive_file_id         VARCHAR(128) NOT NULL,
    drive_revision_id     VARCHAR(128) NOT NULL,
    sha256                BYTEA        NOT NULL,
    size_bytes            BIGINT       NOT NULL,
    media_type            VARCHAR(100) NOT NULL,
    original_name         VARCHAR(255) NOT NULL,
    uploaded_by_member_id UUID         REFERENCES board_members (id),
    created_at            TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT uq_files_drive_file_id UNIQUE (drive_file_id)
);

CREATE INDEX IF NOT EXISTS idx_files_uploaded_by ON files (uploaded_by_member_id);

CREATE TABLE IF NOT EXISTS cases (
    id                   UUID         PRIMARY KEY,
    association_id       UUID         NOT NULL,
    title                VARCHAR(200) NOT NULL,
    drive_folder_id      VARCHAR(128),
    created_by_member_id UUID         NOT NULL REFERENCES board_members (id),
    created_at           TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_cases_association_created ON cases (association_id, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_cases_created_by ON cases (created_by_member_id);

CREATE TABLE IF NOT EXISTS case_documents (
    id                   UUID        PRIMARY KEY,
    association_id       UUID        NOT NULL,
    case_id              UUID        NOT NULL REFERENCES cases (id),
    file_id              UUID        NOT NULL REFERENCES files (id),
    previous_version_id  UUID        REFERENCES case_documents (id),
    document_date        DATE,
    created_by_member_id UUID        NOT NULL REFERENCES board_members (id),
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_case_documents_file_id UNIQUE (file_id)
);

CREATE INDEX IF NOT EXISTS idx_case_documents_case_id ON case_documents (case_id);
CREATE INDEX IF NOT EXISTS idx_case_documents_previous_version ON case_documents (previous_version_id);
CREATE INDEX IF NOT EXISTS idx_case_documents_created_by ON case_documents (created_by_member_id);
