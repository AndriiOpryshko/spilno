-- name: CountAccounts :one
SELECT count(*) AS total FROM accounts;

-- name: UpsertAccount :one
INSERT INTO accounts (id, oidc_issuer, oidc_subject, email)
VALUES (@id, @oidc_issuer, @oidc_subject, @email)
ON CONFLICT (oidc_issuer, oidc_subject) DO UPDATE SET email = EXCLUDED.email
RETURNING id, email;

-- name: InsertSession :exec
INSERT INTO sessions (id, account_id, token_hash, expires_at)
VALUES (@id, @account_id, @token_hash, @expires_at);

-- name: GetSessionPrincipal :one
SELECT a.id, a.email
FROM sessions s
JOIN accounts a ON a.id = s.account_id
WHERE s.token_hash = @token_hash AND s.expires_at > now();

-- name: DeleteSession :exec
DELETE FROM sessions WHERE token_hash = @token_hash;
