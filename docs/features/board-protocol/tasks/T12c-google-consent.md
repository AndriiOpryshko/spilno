---
id: T12c
title: "Turn one Google consent into a grant: refresh token, account email and the spilno folder"
layer: "infra"
parent: T12
acs: ["AC-26", "AC-14"]
scope: core
package: services/core/internal/platform/google
test_cmd: "cd services/core && go test -count=1 ./internal/platform/google/"
stub: services/core/internal/platform/google/consent.go
test_file: services/core/internal/platform/google/consent_test.go
estimate: "half a day"
status: "in_progress"
---

# T12c — Turn one Google consent into a grant

Slice of [T12](./T12-google-connection-token-source.md); [ADR-0005](../adr/0005-store-files-in-association-owned-google-drive-folder.md):
spilno is authorized once against the association-owned Google account for `drive.file` and Docs with an offline
refresh token, and keeps the association's files in a `spilno/` folder it created. This slice is the Google side
of that step: the consent link, the exchange of the authorization code, the account's email and the folder. What
comes out is everything [T12b](./T12b-google-connection-store.md)'s `identity.GoogleConnection` needs except the
association and the operator's account.

Not here (T12d): the loopback listener that receives the code and checks `state`; reading the key and the OAuth
client from the environment; the operator's account for `authorized_by_account_id`; `cmd/google-connect` itself
and `EnableGoogle` in `cmd/api`. Not here either: Drive and Docs clients for file work. T13b and T27 build theirs
from `identity.GoogleTokenSource` and decide whether they share one in this package.

## Interface

Package `google` (`services/core/internal/platform/google`), file `consent.go`:

```go
const (
    ScopeDriveFile = "https://www.googleapis.com/auth/drive.file"
    ScopeDocuments = "https://www.googleapis.com/auth/documents"

    // RootFolderName is the Drive folder spilno keeps an association's files in.
    RootFolderName = "spilno"
)

var (
    ErrCodeExchangeFailed = errors.New("google: authorization code exchange failed")
    ErrNoRefreshToken     = errors.New("google: the consent returned no refresh token")
    ErrScopeNotGranted    = errors.New("google: a required scope was not granted")
    ErrDriveUnavailable   = errors.New("google: drive request failed")
)

// Consent is one authorization of spilno against the association-owned Google account.
type Consent struct {
    OAuth    *oauth2.Config // client id and secret, endpoint, redirect URL
    DriveURL string         // Drive API base, "https://www.googleapis.com/drive/v3" in production; no trailing slash
}

// Grant is what a completed consent gives.
type Grant struct {
    AccountEmail      string   // the Google account that gave the consent
    RefreshToken      string
    Scopes            []string // ScopeDriveFile, ScopeDocuments
    DriveRootFolderID string   // the spilno folder
}

// AuthURL returns the link the operator opens to give the consent.
func (c Consent) AuthURL(state string) string

// Complete exchanges the authorization code and prepares the account: reads its email, finds or creates the folder.
func (c Consent) Complete(ctx context.Context, code string) (Grant, error)
```

`oauth2` is `golang.org/x/oauth2`, already in `go.mod`. No new module: the three Drive calls are plain HTTP with
`encoding/json`.

## Rules

- `AuthURL` asks for exactly `ScopeDriveFile` and `ScopeDocuments`, in that order, whatever `OAuth.Scopes` holds,
  and never changes `c.OAuth`. It adds `access_type=offline` and `prompt=consent`: without them Google returns no
  refresh token on a repeated consent.
- `Complete` works in this order and stops at the first failure with the zero `Grant`:
  1. Exchange the code at the token endpoint (`OAuth.Exchange`). Any failure is `ErrCodeExchangeFailed`.
  2. An empty refresh token in the response is `ErrNoRefreshToken`.
  3. The response's `scope` field (scopes separated by one space) must contain both scopes; otherwise
     `ErrScopeNotGranted` (Google lets the account untick a scope). A response without a `scope` field grants what
     was asked (RFC 6749 §5.1). Other granted scopes are ignored.
  4. `GET {DriveURL}/about?fields=user(emailAddress)` gives the account's email (`user.emailAddress`).
  5. `GET {DriveURL}/files` with `q` =
     `name = 'spilno' and mimeType = 'application/vnd.google-apps.folder' and trashed = false`,
     `orderBy=createdTime` and `fields=files(id)` lists the folders spilno created earlier (`drive.file` shows
     the client only its own files). The first one is reused, so a repeated consent keeps the folder.
  6. With no such folder, `POST {DriveURL}/files?fields=id` with the JSON body
     `{"name":"spilno","mimeType":"application/vnd.google-apps.folder"}` creates it; the answer's `id` is the
     folder id.
- Every Drive request carries `Authorization: Bearer <access token of the exchange>`.
- A Drive request that fails to send, answers outside 2xx, answers a body that is not the expected JSON, or
  answers an empty email or an empty folder id is `ErrDriveUnavailable`. Its message names the HTTP method, the
  path and the status, nothing from the tokens.
- `Grant.Scopes` is always `ScopeDriveFile`, `ScopeDocuments` in that order.
- No error message and no log line contains the refresh token, the access token or the client secret. Errors are
  matched with `errors.Is`.
- Tests are in package `google` and use `httptest.Server` for both sides. The OAuth client is `client-1` /
  `secret-1`, redirect URL `http://127.0.0.1:8085/callback`, auth URL `https://accounts.example.test/auth`,
  `AuthStyle: oauth2.AuthStyleInParams`. The token endpoint records the request form and by default answers
  `{"access_token":"at-1","refresh_token":"1//refresh-token-example","token_type":"Bearer","expires_in":3600,"scope":"https://www.googleapis.com/auth/drive.file https://www.googleapis.com/auth/documents"}`.
  The Drive fake serves under `/drive/v3`, records every request (method, path, query, `Authorization`, body) and
  by default answers `{"user":{"emailAddress":"osbb@example.test"}}` to the about call, `{"files":[]}` to the
  list call and `{"id":"folder-1"}` to the create call.

## Acceptance criteria

- **AC-1** Given that consent, When `AuthURL("state-1")` is called with `OAuth.Scopes` holding some other scope,
  Then the link starts with `https://accounts.example.test/auth?` and its query has `response_type=code`,
  `client_id=client-1`, `redirect_uri=http://127.0.0.1:8085/callback`, `state=state-1`, `access_type=offline`,
  `prompt=consent` and `scope` equal to the two scopes joined by one space; and `OAuth.Scopes` is unchanged.
- **AC-2** Given the default answers, When `Complete` gets the code `code-1`, Then the grant is the email
  `osbb@example.test`, the refresh token `1//refresh-token-example`, the two scopes and the folder `folder-1`;
  the token endpoint received `grant_type=authorization_code`, `code=code-1`,
  `redirect_uri=http://127.0.0.1:8085/callback`, `client_id=client-1` and `client_secret=secret-1`; Drive
  received, in this order, `GET /drive/v3/about` with `fields=user(emailAddress)`, `GET /drive/v3/files` with the
  `q` above and `POST /drive/v3/files` whose JSON body has `name` `spilno` and `mimeType`
  `application/vnd.google-apps.folder`; and each of the three carried `Authorization: Bearer at-1`.
- **AC-3** Given Drive lists `{"files":[{"id":"folder-0"},{"id":"folder-9"}]}`, When `Complete` runs, Then the
  grant's folder is `folder-0` and Drive received no `POST`.
- **AC-4** Given the token endpoint answers without `refresh_token`, Then `Complete` returns `ErrNoRefreshToken`,
  the zero grant, and Drive received no request.
- **AC-5** Given the token endpoint answers with `scope` equal to `https://www.googleapis.com/auth/drive.file`
  only, Then `Complete` returns `ErrScopeNotGranted` and Drive received no request; and given it answers with no
  `scope` field at all, Then `Complete` succeeds with both scopes in the grant.
- **AC-6** Given the token endpoint answers `400` with `{"error":"invalid_grant"}`, Then `Complete` returns
  `ErrCodeExchangeFailed`, the zero grant, and Drive received no request.
- **AC-7** Given the about call answers `503`, or it answers `{"user":{}}`, or the create call answers `403`, Then
  `Complete` returns `ErrDriveUnavailable` and the zero grant each time, and the error's text contains neither
  `1//refresh-token-example` nor `at-1`.
