-- name: DeleteGoogleConnection :exec
DELETE FROM google_connections WHERE association_id = @association_id;

-- name: InsertGoogleConnection :exec
INSERT INTO google_connections (
    id, association_id, google_account_email, refresh_token_ciphertext, refresh_token_nonce,
    encryption_key_id, scopes, drive_root_folder_id, authorized_by_account_id
) VALUES (
    @id, @association_id, @google_account_email, @refresh_token_ciphertext, @refresh_token_nonce,
    @encryption_key_id, @scopes, @drive_root_folder_id, @authorized_by_account_id
);

-- name: GetAccountEmail :one
SELECT email FROM accounts WHERE id = @id;

-- name: GetGoogleConnection :one
SELECT * FROM google_connections WHERE association_id = @association_id;
