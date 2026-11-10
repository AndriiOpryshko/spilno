package identity

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/oauth2"

	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/adapters/postgres"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/app"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/ports"
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

// googleState is what EnableGoogle sets up.
type googleState struct {
	connections *app.GoogleConnections
	oauth       *oauth2.Config
}

// EnableGoogle turns the Google connection on; without it the two methods below refuse to work.
func (m *Module) EnableGoogle(cfg GoogleConfig) error {
	if !oauthClientComplete(cfg.OAuth) {
		return ErrGoogleConfigInvalid
	}
	cipher, err := app.NewTokenCipher(cfg.KeyID, cfg.Key)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrGoogleConfigInvalid, err)
	}
	m.google = &googleState{
		connections: app.NewGoogleConnections(postgres.NewGoogleConnections(m.pool), cipher),
		oauth:       cfg.OAuth,
	}
	return nil
}

// oauthClientComplete reports whether c has what a refresh token exchange needs:
// the client id, the client secret and the token endpoint.
func oauthClientComplete(c *oauth2.Config) bool {
	return c != nil && c.ClientID != "" && c.ClientSecret != "" && c.Endpoint.TokenURL != ""
}

// ConnectGoogle stores the connection of c.AssociationID, replacing the one it had.
func (m *Module) ConnectGoogle(ctx context.Context, c GoogleConnection) error {
	if m.google == nil {
		return ErrGoogleNotConfigured
	}
	err := m.google.connections.Connect(ctx, app.NewGoogleConnection{
		AssociationID:     c.AssociationID,
		AccountEmail:      c.AccountEmail,
		RefreshToken:      c.RefreshToken,
		Scopes:            c.Scopes,
		DriveRootFolderID: c.DriveRootFolderID,
		AuthorizedBy:      c.AuthorizedBy,
	})
	if errors.Is(err, app.ErrInvalidGoogleConnection) {
		return ErrGoogleConnectionInvalid
	}
	return err
}

// GoogleTokenSource returns a token source for the association's Google account.
func (m *Module) GoogleTokenSource(ctx context.Context, associationID uuid.UUID) (oauth2.TokenSource, error) {
	if m.google == nil {
		return nil, ErrGoogleNotConfigured
	}
	refreshToken, err := m.google.connections.RefreshToken(ctx, associationID)
	switch {
	case errors.Is(err, ports.ErrGoogleConnectionNotFound):
		return nil, ErrGoogleNotConnected
	case errors.Is(err, app.ErrUnknownKey), errors.Is(err, app.ErrTokenUnreadable):
		return nil, fmt.Errorf("%w: %w", ErrGoogleTokenUnreadable, err)
	case err != nil:
		return nil, err
	}
	return m.google.oauth.TokenSource(ctx, &oauth2.Token{RefreshToken: refreshToken}), nil
}
