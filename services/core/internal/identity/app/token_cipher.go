// Package app holds the identity context's use cases.
package app

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"unicode/utf8"
)

const (
	tokenKeySize  = 32 // AES-256
	maxKeyIDChars = 64 // google_connections.encryption_key_id is VARCHAR(64)
)

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
type TokenCipher struct {
	keyID string
	aead  cipher.AEAD
}

// NewTokenCipher takes the key id and the AES-256 key.
func NewTokenCipher(keyID string, key []byte) (*TokenCipher, error) {
	if len(key) != tokenKeySize {
		return nil, ErrInvalidKey
	}
	if n := utf8.RuneCountInString(keyID); n < 1 || n > maxKeyIDChars {
		return nil, ErrInvalidKey
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, ErrInvalidKey
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, ErrInvalidKey
	}
	return &TokenCipher{keyID: keyID, aead: aead}, nil
}

// Seal encrypts token under a fresh random nonce.
func (c *TokenCipher) Seal(token string) (SealedToken, error) {
	if token == "" {
		return SealedToken{}, ErrEmptyToken
	}
	nonce := make([]byte, c.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return SealedToken{}, err
	}
	return SealedToken{
		Ciphertext: c.aead.Seal(nil, nonce, []byte(token), nil),
		Nonce:      nonce,
		KeyID:      c.keyID,
	}, nil
}

// Open decrypts a token sealed by a cipher with the same key id and key.
func (c *TokenCipher) Open(s SealedToken) (string, error) {
	if s.KeyID != c.keyID {
		return "", ErrUnknownKey
	}
	// GCM panics on a nonce of the wrong length, so the length is checked here.
	if len(s.Nonce) != c.aead.NonceSize() || len(s.Ciphertext) <= c.aead.Overhead() {
		return "", ErrTokenUnreadable
	}
	token, err := c.aead.Open(nil, s.Nonce, s.Ciphertext, nil)
	if err != nil {
		return "", ErrTokenUnreadable
	}
	return string(token), nil
}
