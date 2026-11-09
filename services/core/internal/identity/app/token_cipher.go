// Package app holds the identity context's use cases.
package app

import "errors"

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
type TokenCipher struct{}

// NewTokenCipher takes the key id and the AES-256 key.
func NewTokenCipher(keyID string, key []byte) (*TokenCipher, error) {
	return nil, errors.New("not implemented")
}

// Seal encrypts token under a fresh random nonce.
func (c *TokenCipher) Seal(token string) (SealedToken, error) {
	return SealedToken{}, errors.New("not implemented")
}

// Open decrypts a token sealed by a cipher with the same key id and key.
func (c *TokenCipher) Open(s SealedToken) (string, error) {
	return "", errors.New("not implemented")
}
