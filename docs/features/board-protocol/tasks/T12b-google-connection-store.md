---
id: T12b
title: "Store the association's Google connection and hand out its token source"
layer: "infra"
parent: T12
acs: ["AC-26", "AC-14"]
scope: identity
package: services/core/internal/identity
test_cmd: "cd services/core && go test -count=1 ./internal/identity/..."
stubs:
  - services/core/internal/identity/google.go
  - services/core/go.mod
  - services/core/go.sum
test_files:
  - services/core/internal/identity/google_test.go
estimate: "half a day"
status: "done"
---

# T12b — Store the association's Google connection and hand out its token source

Slice of [T12](./T12-google-connection-token-source.md); [ADR-0005](../adr/0005-store-files-in-association-owned-google-drive-folder.md),
[ADR-0014](../adr/0014-keep-board-protocol-in-governance-and-add-only-an-identity-context.md) and
[data-model](../data-model.md) `google_connections`: `identity` keeps the association Google account's refresh
token encrypted and hands `governance` a token source. The cipher is [T12a](./T12a-google-token-cipher.md)
(`identity/app.TokenCipher`); this slice puts it on Postgres and behind the module.

Not here: reading the key and the OAuth client from the environment and calling `EnableGoogle` in `cmd/api`
(T12c, together with `cmd/google-connect`, which needs the same configuration); the Drive and Docs clients
(`platform/google`, T12c); a way to read `drive_root_folder_id` back (T13b, when the Drive adapter needs it).

## Interface

Package `identity` (the module package), file `google.go`:

```go
// GoogleConfig is what the context needs to keep and use the association's Google connection.
type GoogleConfig struct {
    KeyID string         // names the token encryption key; stored with every connection
    Key   []byte         // the AES-256 key, 32 bytes
    OAuth *oauth2.Config // the OAuth client: id, secret and token endpoint
}

// GoogleConnection is the consent the association-owned Google account gave once.
type GoogleConnection struct {
    AssociationID     uuid.UUID
    AccountEmail      string    // the association-owned Google account
    RefreshToken      string    // sealed before it is stored
    Scopes            []string
    DriveRootFolderID string    // the spilno/ folder; empty when it is not created yet
    AuthorizedBy      uuid.UUID // the account of whoever gave the consent
}

var (
    ErrGoogleNotConfigured     = errors.New("identity: google connection is not configured")
    ErrGoogleConfigInvalid     = errors.New("identity: invalid google configuration")
    ErrGoogleConnectionInvalid = errors.New("identity: invalid google connection")
    ErrGoogleNotConnected      = errors.New("identity: association has no google connection")
    ErrGoogleTokenUnreadable   = errors.New("identity: stored google token cannot be decrypted")
)

// EnableGoogle turns the Google connection on; without it the two methods below refuse to work.
func (m *Module) EnableGoogle(cfg GoogleConfig) error

// ConnectGoogle stores the connection of c.AssociationID, replacing the one it had.
func (m *Module) ConnectGoogle(ctx context.Context, c GoogleConnection) error

// GoogleTokenSource returns a token source for the association's Google account.
func (m *Module) GoogleTokenSource(ctx context.Context, associationID uuid.UUID) (oauth2.TokenSource, error)
```

`uuid` is `github.com/google/uuid` (already in `go.mod`); `oauth2` is `golang.org/x/oauth2`, the one new module.

## Rules

- `identity.New(pool)` keeps its signature. A module without `EnableGoogle` answers `ErrGoogleNotConfigured` from
  `ConnectGoogle` and `GoogleTokenSource`.
- `EnableGoogle` answers `ErrGoogleConfigInvalid` when the key is not 32 bytes, the key id is empty or longer than
  64 characters, `OAuth` is nil, or its `ClientID`, `ClientSecret` or `Endpoint.TokenURL` is empty. After such an
  error the module is still not configured.
- `ConnectGoogle` answers `ErrGoogleConnectionInvalid` and writes nothing when the refresh token or the account
  email is empty or there is no scope. Otherwise it seals the refresh token with the `TokenCipher` and, in one
  transaction, deletes the association's row and inserts a new one with a new UUID v7 `id` (re-authorization
  replaces the row, so `created_at` is the authorization time). `scopes` is the scopes joined by one space; an
  empty `DriveRootFolderID` is stored as `NULL`.
- The refresh token is never stored, logged or put into an error message in plain text.
- `GoogleTokenSource` answers `ErrGoogleNotConnected` when the association has no row, and
  `ErrGoogleTokenUnreadable` when the row was sealed under another key id or another key. Otherwise it returns
  `cfg.OAuth.TokenSource(ctx, &oauth2.Token{RefreshToken: <opened token>})`, which exchanges the refresh token at
  the token endpoint on the first `Token()` call.
- Errors are matched with `errors.Is`.
- Layout (CLAUDE.md): the repository interface in `identity/ports`, the use case in `identity/app`, its Postgres
  implementation in `identity/adapters/postgres` on sqlc. GREEN adds the queries in
  `identity/adapters/postgres/queries/google_connections.sql` and regenerates with `mise run gen`; generated
  files are never edited by hand. `module.go` may change to hold the new state.
- Tests are integration tests in package `identity_test` on `testdb.New(t)`. The association
  `00000000-0000-7000-8000-000000000001` exists after the migrations; a test inserts its own `accounts` row
  (`operator@example.test`) and reads `google_connections` with plain SQL. The token endpoint is an
  `httptest.Server` that records the request form and answers
  `{"access_token":"at-1","token_type":"Bearer","expires_in":3600}`; the OAuth client is `client-1` / `secret-1`
  with `AuthStyle: oauth2.AuthStyleInParams`, so the credentials arrive in the form.

## Acceptance criteria

- **AC-1** Given a module enabled with key id `k1` and a 32-byte key, When `ConnectGoogle` stores the account
  `osbb@example.test` with the refresh token `1//refresh-token-example` and the scopes
  `https://www.googleapis.com/auth/drive.file` and `https://www.googleapis.com/auth/documents`, Then the
  association has exactly one `google_connections` row with that email, the two scopes joined by one space,
  `encryption_key_id` `k1`, a 12-byte nonce, a ciphertext that does not contain the token's bytes, a `NULL`
  `drive_root_folder_id` and the operator's account in `authorized_by_account_id`.
- **AC-2** Given that connection, When `GoogleTokenSource` is asked for the association and `Token()` is called,
  Then the token endpoint received `grant_type=refresh_token`, `refresh_token=1//refresh-token-example`,
  `client_id=client-1` and `client_secret=secret-1`, and the returned access token is `at-1`.
- **AC-3** Given that connection, When `ConnectGoogle` is called again with the refresh token `1//second-token`
  and the folder id `folder-1`, Then the association still has exactly one row, its `id` differs from the first
  one, `drive_root_folder_id` is `folder-1`, and a new token source sends `refresh_token=1//second-token`.
- **AC-4** Given an enabled module and an association with no connection, Then `GoogleTokenSource` returns
  `ErrGoogleNotConnected`.
- **AC-5** Given a connection stored under key id `k1`, When a second module on the same database is enabled
  with key id `k1` and a different key, or with key id `k2` and the same key, Then `GoogleTokenSource` returns
  `ErrGoogleTokenUnreadable` in both cases.
- **AC-6** Given a module without `EnableGoogle`, Then `ConnectGoogle` and `GoogleTokenSource` return
  `ErrGoogleNotConfigured`; and `EnableGoogle` with a 16-byte key, an empty key id, a nil `OAuth`, an empty
  `ClientID` or an empty `Endpoint.TokenURL` returns `ErrGoogleConfigInvalid` and leaves the module not
  configured.
- **AC-7** Given an enabled module, When `ConnectGoogle` gets an empty refresh token, an empty account email or
  no scopes, Then it returns `ErrGoogleConnectionInvalid` each time and the association has no row.
