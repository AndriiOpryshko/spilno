package identity

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

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

	mu     sync.Mutex
	tokens map[uuid.UUID]cachedToken // per association, dropped when the connection row changes
}

// cachedToken is the token source of one connection row, keyed by the row's id.
type cachedToken struct {
	connection uuid.UUID
	source     oauth2.TokenSource
}

// tokenEndpointTimeout bounds each call to the token endpoint.
const tokenEndpointTimeout = 15 * time.Second

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
		tokens:      map[uuid.UUID]cachedToken{},
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

// HasStorageConnection reports whether the association has connected its Google storage.
func (m *Module) HasStorageConnection(ctx context.Context, associationID uuid.UUID) (bool, error) {
	_, err := postgres.NewGoogleConnections(m.pool).Get(ctx, associationID)
	if errors.Is(err, ports.ErrGoogleConnectionNotFound) {
		return false, nil
	}
	return err == nil, err
}

// GoogleStorage returns the token source and the spilno folder of the association's connection from one
// read; the folder is "" while it is not created. The token source is reused until the connection row
// changes (a reconnect writes a new row id). ErrGoogleNotConnected when there is no connection.
func (m *Module) GoogleStorage(ctx context.Context, associationID uuid.UUID) (oauth2.TokenSource, string, error) {
	if m.google == nil {
		return nil, "", ErrGoogleNotConfigured
	}
	g := m.google
	c, err := g.connections.Get(ctx, associationID)
	if errors.Is(err, ports.ErrGoogleConnectionNotFound) {
		return nil, "", ErrGoogleNotConnected
	}
	if err != nil {
		return nil, "", err
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if cached, ok := g.tokens[associationID]; ok && cached.connection == c.ID {
		return cached.source, c.DriveRootFolderID, nil
	}
	refreshToken, err := g.connections.OpenToken(c)
	if err != nil {
		return nil, "", fmt.Errorf("%w: %w", ErrGoogleTokenUnreadable, err)
	}
	// The source outlives the request, so it gets its own context and a bounded token client.
	tokenCtx := context.WithValue(context.Background(), oauth2.HTTPClient, &http.Client{Timeout: tokenEndpointTimeout})
	ts := g.oauth.TokenSource(tokenCtx, &oauth2.Token{RefreshToken: refreshToken})
	g.tokens[associationID] = cachedToken{connection: c.ID, source: ts}
	return ts, c.DriveRootFolderID, nil
}
