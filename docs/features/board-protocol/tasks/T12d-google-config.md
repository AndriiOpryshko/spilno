---
id: T12d
title: "Read the Google configuration from the environment and enable the connection in the API"
layer: "infra"
parent: T12
acs: ["AC-26", "AC-14"]
scope: core
package: services/core/internal/platform/config
test_cmd: "cd services/core && go test -count=1 ./internal/platform/config/"
stub: services/core/internal/platform/config/google.go
test_file: services/core/internal/platform/config/google_test.go
estimate: "2 h"
status: "done"
---

# T12d — Read the Google configuration from the environment

Slice of [T12](./T12-google-connection-token-source.md); [sad §8](../sad.md) Secrets: the key that encrypts the
association Google account's refresh token and the OAuth client come from the environment (an env file on the
VM). [T12b](./T12b-google-connection-store.md) left `identity.EnableGoogle` uncalled; this slice reads the four
variables and calls it in `cmd/api`, so `GoogleTokenSource` works in the running API.

Not here (T12e): `cmd/google-connect`, which reads the same configuration, receives the authorization code on a
loopback listener and stores the connection with the operator's account.

## Interface

Package `config` (`services/core/internal/platform/config`), new file `google.go`:

```go
// Google is what spilno needs to use the association's Google account (ADR-0005).
type Google struct {
    ClientID     string // GOOGLE_OAUTH_CLIENT_ID
    ClientSecret string // GOOGLE_OAUTH_CLIENT_SECRET
    TokenKeyID   string // GOOGLE_TOKEN_KEY_ID, names the key below
    TokenKey     []byte // GOOGLE_TOKEN_KEY decoded: the AES-256 key, 32 bytes
}

var (
    ErrGoogleIncomplete = errors.New("config: google configuration is incomplete")
    ErrGoogleTokenKey   = errors.New("config: GOOGLE_TOKEN_KEY is not a base64-encoded 32-byte key")
)

// LoadGoogle reads the GOOGLE_* variables; it returns nil and no error when none of them is set.
func LoadGoogle() (*Google, error)

// OAuth returns the OAuth client against Google's endpoints. redirectURL is empty for a client that only
// refreshes tokens.
func (g Google) OAuth(redirectURL string) *oauth2.Config
```

`oauth2` is `golang.org/x/oauth2`, already in `go.mod`; Google's endpoints are `endpoints.Google` from
`golang.org/x/oauth2/endpoints` (the same module, no new dependency).

## Rules

- A variable that is unset and a variable set to the empty string are the same thing.
- None of the four set: `LoadGoogle` returns `nil, nil`. The API then runs without Google and `identity` answers
  `ErrGoogleNotConfigured`.
- Some but not all set: `ErrGoogleIncomplete` and a nil configuration. The message lists the names of the missing
  variables and no value.
- `GOOGLE_TOKEN_KEY` is standard base64 (`base64.StdEncoding`, what `openssl rand -base64 32` prints) of exactly
  32 bytes. Anything else is `ErrGoogleTokenKey` and a nil configuration; the message holds no part of the value.
- The key id is not validated here: `identity.EnableGoogle` owns that rule.
- `OAuth` sets `ClientID`, `ClientSecret`, `RedirectURL` and `Endpoint: endpoints.Google`; no scopes.
- No error message and no log line contains the client secret or the key. Errors are matched with `errors.Is`.
- GREEN also wires `services/core/cmd/api/main.go` (no test; `cmd/api` has none): after `config.Load()` it calls
  `config.LoadGoogle()` and returns its error; with a configuration it calls `EnableGoogle` on the identity
  module with `identity.GoogleConfig{KeyID: g.TokenKeyID, Key: g.TokenKey, OAuth: g.OAuth("")}` and returns its
  error; without one it logs `google connection is not configured` at info level and goes on.
- Tests are in package `config` and set the variables with `t.Setenv`, each test setting all four (empty for
  "not set") so the developer's environment does not leak in.

## Acceptance criteria

- **AC-1** Given none of the four variables is set, Then `LoadGoogle` returns a nil configuration and a nil error.
- **AC-2** Given `GOOGLE_OAUTH_CLIENT_ID=client-1`, `GOOGLE_OAUTH_CLIENT_SECRET=secret-1`,
  `GOOGLE_TOKEN_KEY_ID=k1` and `GOOGLE_TOKEN_KEY=AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8=`, Then `LoadGoogle`
  returns those three strings and a 32-byte key whose bytes are `0x00` to `0x1f` in order.
- **AC-3** Given the variables of AC-2 with exactly one of them empty (four cases), Then `LoadGoogle` returns
  `ErrGoogleIncomplete` and a nil configuration each time; the error's text contains the name of the empty
  variable and contains neither `secret-1` nor the key string.
- **AC-4** Given the variables of AC-2 with `GOOGLE_TOKEN_KEY` set to `not-base64!`, to
  `AAECAwQFBgcICQoLDA0ODw==` (16 bytes) or to `AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8g` (33 bytes), Then
  `LoadGoogle` returns `ErrGoogleTokenKey` and a nil configuration each time, and the error's text does not
  contain the value.
- **AC-5** Given the configuration of AC-2, When `OAuth("http://127.0.0.1:8085/callback")` is called, Then the
  client has `ClientID` `client-1`, `ClientSecret` `secret-1`, that redirect URL, the auth URL
  `https://accounts.google.com/o/oauth2/auth` and the token URL `https://oauth2.googleapis.com/token`; and
  `OAuth("")` has an empty redirect URL.
