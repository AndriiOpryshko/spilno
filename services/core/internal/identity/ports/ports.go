// Package ports declares what the identity context needs from the outside world.
package ports

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

// ErrGoogleConnectionNotFound is returned when the association has no Google connection.
var ErrGoogleConnectionNotFound = errors.New("identity: google connection not found")

// GoogleConnection is one google_connections row; the refresh token is already sealed.
type GoogleConnection struct {
	ID                     uuid.UUID
	AssociationID          uuid.UUID
	AccountEmail           string
	RefreshTokenCiphertext []byte
	RefreshTokenNonce      []byte
	EncryptionKeyID        string
	Scopes                 string // the scopes joined by one space
	DriveRootFolderID      string // empty when the folder is not created yet
	AuthorizedBy           uuid.UUID
	AuthorizedByEmail      string    // the sign-in address of that account; empty when unknown
	CreatedAt              time.Time // when the connection was stored
}

// GoogleConnectionRepository stores the one Google connection an association has.
type GoogleConnectionRepository interface {
	// Replace deletes the association's connection and stores c, in one transaction.
	Replace(ctx context.Context, c GoogleConnection) error
	// Get returns ErrGoogleConnectionNotFound when the association has no connection.
	Get(ctx context.Context, associationID uuid.UUID) (GoogleConnection, error)
}

// ProviderIdentity is what the sign-in provider vouches for.
type ProviderIdentity struct {
	Issuer        string
	Subject       string
	Email         string
	EmailVerified bool
}

// SignInProvider is the OIDC identity provider (authorization code + PKCE).
type SignInProvider interface {
	AuthCodeURL(state, nonce, codeChallenge string) string
	Exchange(ctx context.Context, code, verifier, nonce string) (ProviderIdentity, error)
}

// ErrSessionNotFound is returned when no live session matches the token hash.
var ErrSessionNotFound = errors.New("identity: session not found")

// Account is one signed-in person.
type Account struct {
	ID    uuid.UUID
	Email string
}

// AccountRepository stores accounts.
type AccountRepository interface {
	// Upsert returns the account of (issuer, subject), creating it with id when new.
	Upsert(ctx context.Context, id uuid.UUID, issuer, subject, email string) (Account, error)
}

// SessionRepository stores server-side sessions keyed by the hash of the cookie token.
type SessionRepository interface {
	Create(ctx context.Context, id, accountID uuid.UUID, tokenHash []byte, expiresAt time.Time) error
	// Account returns ErrSessionNotFound for an unknown or expired token hash.
	Account(ctx context.Context, tokenHash []byte) (Account, error)
	Delete(ctx context.Context, tokenHash []byte) error
}
