package config

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"
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
	names := []string{"GOOGLE_OAUTH_CLIENT_ID", "GOOGLE_OAUTH_CLIENT_SECRET", "GOOGLE_TOKEN_KEY_ID", "GOOGLE_TOKEN_KEY"}
	values := make(map[string]string, len(names))
	var missing []string
	for _, name := range names {
		values[name] = os.Getenv(name)
		if values[name] == "" {
			missing = append(missing, name)
		}
	}
	if len(missing) == len(names) {
		return nil, nil
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("%w: missing %s", ErrGoogleIncomplete, strings.Join(missing, ", "))
	}
	// The decoding error is dropped: nothing derived from the key reaches a message.
	key, err := base64.StdEncoding.DecodeString(values["GOOGLE_TOKEN_KEY"])
	if err != nil || len(key) != 32 {
		return nil, ErrGoogleTokenKey
	}
	return &Google{
		ClientID:     values["GOOGLE_OAUTH_CLIENT_ID"],
		ClientSecret: values["GOOGLE_OAUTH_CLIENT_SECRET"],
		TokenKeyID:   values["GOOGLE_TOKEN_KEY_ID"],
		TokenKey:     key,
	}, nil
}

// OAuth returns the OAuth client against Google's endpoints. redirectURL is empty for a client that only
// refreshes tokens.
func (g Google) OAuth(redirectURL string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     g.ClientID,
		ClientSecret: g.ClientSecret,
		RedirectURL:  redirectURL,
		Endpoint:     endpoints.Google,
	}
}
