// Package oidc holds the sign-in providers of the identity context.
package oidc

import (
	"context"
	"errors"
	"net/url"
	"strings"

	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/ports"
)

// Dev is the development and e2e sign-in shim: it sends the browser straight back to
// the callback and treats the code as a verified email address. Only a dev or e2e
// configuration (SPILNO_DEV_SIGN_IN) may turn it on.
type Dev struct {
	// Email is the address the shim signs in.
	Email string
}

var _ ports.SignInProvider = Dev{}

const devIssuer = "https://dev.spilno.invalid"

func (d Dev) AuthCodeURL(state, _, _ string) string {
	return "/identity/sign-in/callback?" + url.Values{"code": {d.Email}, "state": {state}}.Encode()
}

func (d Dev) Exchange(_ context.Context, code, _, _ string) (ports.ProviderIdentity, error) {
	email := strings.ToLower(strings.TrimSpace(code))
	if email == "" {
		return ports.ProviderIdentity{}, errors.New("oidc dev: empty code")
	}
	return ports.ProviderIdentity{Issuer: devIssuer, Subject: email, Email: email, EmailVerified: true}, nil
}
