-- identity: the association Google account's drive.file connection (ADR-0005). The refresh token is
-- AES-GCM ciphertext; re-authorization deletes the row and inserts a new one.
CREATE TABLE IF NOT EXISTS google_connections (
    id                       UUID         PRIMARY KEY,
    association_id           UUID         NOT NULL REFERENCES associations (id),
    google_account_email     VARCHAR(320) NOT NULL,
    refresh_token_ciphertext BYTEA        NOT NULL,
    refresh_token_nonce      BYTEA        NOT NULL,
    encryption_key_id        VARCHAR(64)  NOT NULL,
    scopes                   VARCHAR(500) NOT NULL,
    drive_root_folder_id     VARCHAR(128),
    authorized_by_account_id UUID         NOT NULL REFERENCES accounts (id),
    created_at               TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT uq_google_connections_association_id UNIQUE (association_id)
);

CREATE INDEX IF NOT EXISTS idx_google_connections_authorized_by ON google_connections (authorized_by_account_id);
