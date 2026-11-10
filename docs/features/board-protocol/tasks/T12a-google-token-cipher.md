---
id: T12a
title: "Seal and open the Google refresh token with AES-256-GCM"
layer: "infra"
parent: T12
acs: ["AC-26", "AC-14"]
scope: identity
package: services/core/internal/identity/app
test_cmd: "cd services/core && go test -count=1 ./internal/identity/app/"
stub: services/core/internal/identity/app/token_cipher.go
test_file: services/core/internal/identity/app/token_cipher_test.go
estimate: "60-90 min"
status: "done"
---

# T12a — Seal and open the Google refresh token with AES-256-GCM

Slice of [T12](./T12-google-connection-token-source.md); [ADR-0005](../adr/0005-store-files-in-association-owned-google-drive-folder.md)
and [sad §8](../sad.md) Secrets: the association Google account's refresh token is stored in `google_connections`
encrypted with AES-GCM, with a random nonce per write, a key from the environment and a key id kept for rotation.
This slice is the cipher only. The repository and `GoogleTokenSource` are T12b; reading the key from the
environment is wired there.

## Interface

```go
// SealedToken is an encrypted refresh token, shaped like the google_connections columns.
type SealedToken struct {
    Ciphertext []byte // refresh_token_ciphertext: the encrypted token followed by the 16-byte GCM tag
    Nonce      []byte // refresh_token_nonce: 12 random bytes, new for every Seal
    KeyID      string // encryption_key_id: names the key that sealed the token
}

var (
    ErrInvalidKey      = errors.New("identity: invalid token encryption key")
    ErrEmptyToken      = errors.New("identity: empty token")
    ErrUnknownKey      = errors.New("identity: token sealed with another key id")
    ErrTokenUnreadable = errors.New("identity: token cannot be decrypted")
)

// TokenCipher seals and opens the association's Google refresh token under one key.
type TokenCipher struct{ /* unexported */ }

// NewTokenCipher takes the key id and the AES-256 key.
func NewTokenCipher(keyID string, key []byte) (*TokenCipher, error)

// Seal encrypts token under a fresh random nonce.
func (c *TokenCipher) Seal(token string) (SealedToken, error)

// Open decrypts a token sealed by a cipher with the same key id and key.
func (c *TokenCipher) Open(s SealedToken) (string, error)
```

## Rules

- AES-256-GCM from the standard library only (`crypto/aes`, `crypto/cipher`, `crypto/rand`); no new dependency.
- The key is exactly 32 bytes. The key id is 1 to 64 characters (the `encryption_key_id` column is `VARCHAR(64)`).
  Anything else is `ErrInvalidKey`.
- `Seal` draws a new 12-byte nonce from `crypto/rand` on every call and refuses an empty token with
  `ErrEmptyToken`.
- `Open` checks the key id first: a different key id is `ErrUnknownKey`, so the caller can tell "rotate" from
  "corrupt". Every other failure (a different key, a changed ciphertext or nonce, a nonce that is not 12 bytes, an
  empty ciphertext) is `ErrTokenUnreadable`. `Open` never panics on malformed input.
- On any error the returned token is `""` and the returned `SealedToken` is the zero value. Errors are matched with
  `errors.Is`. No error message contains the token or the key.

## Acceptance criteria

- **AC-1** Given a cipher with key id `k1` and a 32-byte key, When the token `1//refresh-token-example` is sealed,
  Then the result has key id `k1`, a 12-byte nonce and a ciphertext 16 bytes longer than the token that does not
  contain the token's bytes, and `Open` returns `1//refresh-token-example`.
- **AC-2** Given the same cipher, When the same token is sealed twice, Then the two nonces differ, the two
  ciphertexts differ, and both open to the token.
- **AC-3** Given a token sealed by a cipher with key id `k1`, When a cipher with key id `k1` and a different
  32-byte key opens it, Then the error is `ErrTokenUnreadable` and the token is `""`.
- **AC-4** Given a sealed token, When one bit of the ciphertext is flipped, one bit of the nonce is flipped, the
  nonce is cut to 11 bytes, the nonce is empty or the ciphertext is empty, Then `Open` returns
  `ErrTokenUnreadable` each time and does not panic.
- **AC-5** Given a token sealed by a cipher with key id `k1`, When a cipher with key id `k2` and the same key bytes
  opens it, Then the error is `ErrUnknownKey`.
- **AC-6** Given a key of 0, 16, 31 or 33 bytes, or a key id that is empty or 65 characters long, Then
  `NewTokenCipher` returns `ErrInvalidKey` and a nil cipher; a 64-character key id with a 32-byte key is accepted;
  and `Seal("")` on a valid cipher returns `ErrEmptyToken`.
