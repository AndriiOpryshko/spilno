package config

import (
	"errors"

	"golang.org/x/oauth2"
)

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
func LoadGoogle() (*Google, error) {
	return nil, errors.New("not implemented")
}

// OAuth returns the OAuth client against Google's endpoints. redirectURL is empty for a client that only
// refreshes tokens.
func (g Google) OAuth(redirectURL string) *oauth2.Config {
	return nil
}
