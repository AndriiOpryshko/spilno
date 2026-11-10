package identity

import (
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/adapters/oidc"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/adapters/postgres"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/app"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/ports"
)

// ProviderIdentity is what the sign-in provider vouches for.
type ProviderIdentity = ports.ProviderIdentity

// SignInProvider is the OIDC provider behind sign-in.
type SignInProvider = ports.SignInProvider

// SignInConfig is what the context needs to sign people in.
type SignInConfig struct {
	Provider      SignInProvider
	OperatorEmail string // SPILNO_OPERATOR_EMAIL; compared ignoring letter case
}

// EnableSignIn turns startSignIn, completeSignIn, signOut and getSession on.
func (m *Module) EnableSignIn(cfg SignInConfig) {
	repo := postgres.NewSessions(m.pool)
	m.signIn = app.NewSignIn(cfg.Provider, repo, repo, cfg.OperatorEmail)
	m.handler.EnableSignIn(m.signIn)
}

// DevSignInProvider is the development and e2e shim; never wire it in production.
func DevSignInProvider(email string) SignInProvider { return oidc.Dev{Email: email} }
