package identity

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

// GoogleConfig is what the context needs to keep and use the association's Google connection.
type GoogleConfig struct {
	KeyID string         // names the token encryption key; stored with every connection
	Key   []byte         // the AES-256 key, 32 bytes
	OAuth *oauth2.Config // the OAuth client: id, secret and token endpoint
}

// GoogleConnection is the consent the association-owned Google account gave once.
type GoogleConnection struct {
	AssociationID     uuid.UUID
	AccountEmail      string // the association-owned Google account
	RefreshToken      string // sealed before it is stored
	Scopes            []string
	DriveRootFolderID string    // the spilno/ folder; empty when it is not created yet
	AuthorizedBy      uuid.UUID // the account of whoever gave the consent
}

var (
	ErrGoogleNotConfigured     = errors.New("identity: google connection is not configured")
	ErrGoogleConfigInvalid     = errors.New("identity: invalid google configuration")
	ErrGoogleConnectionInvalid = errors.New("identity: invalid google connection")
	ErrGoogleNotConnected      = errors.New("identity: association has no google connection")
	ErrGoogleTokenUnreadable   = errors.New("identity: stored google token cannot be decrypted")
)

// EnableGoogle turns the Google connection on; without it the two methods below refuse to work.
func (m *Module) EnableGoogle(cfg GoogleConfig) error {
	return errors.New("not implemented")
}

// ConnectGoogle stores the connection of c.AssociationID, replacing the one it had.
func (m *Module) ConnectGoogle(ctx context.Context, c GoogleConnection) error {
	return errors.New("not implemented")
}

// GoogleTokenSource returns a token source for the association's Google account.
func (m *Module) GoogleTokenSource(ctx context.Context, associationID uuid.UUID) (oauth2.TokenSource, error) {
	return nil, errors.New("not implemented")
}
