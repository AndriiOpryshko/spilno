package identity

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/adapters/postgres"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/app"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/ports"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/db"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/google"
)

// StorageConnectionConfig is what the web flow that connects the association's document storage needs.
type StorageConnectionConfig struct {
	Consent google.Consent
	Members app.StorageMembers
}

// EnableStorageConnection turns the storage connection operations on. It needs EnableGoogle first;
// without it the operations stay unavailable.
func (m *Module) EnableStorageConnection(cfg StorageConnectionConfig) {
	if m.google == nil {
		return
	}
	m.handler.EnableStorageConnections(app.NewStorageConnections(
		googleConsent{cfg.Consent}, scopedMembers{m.pool, cfg.Members}, m.google.connections, postgres.NewGoogleConnections(m.pool)))
}

// googleConsent adapts google.Consent to the storage consent port.
type googleConsent struct{ c google.Consent }

func (g googleConsent) AuthURL(state string) string { return g.c.AuthURL(state) }

func (g googleConsent) Complete(ctx context.Context, code string) (ports.StorageGrant, error) {
	grant, err := g.c.Complete(ctx, code)
	switch {
	case errors.Is(err, google.ErrCodeExchangeFailed), errors.Is(err, google.ErrNoRefreshToken), errors.Is(err, google.ErrScopeNotGranted):
		return ports.StorageGrant{}, ports.ErrStorageCodeRefused
	case err != nil:
		return ports.StorageGrant{}, err
	}
	return ports.StorageGrant{AccountEmail: grant.AccountEmail, RefreshToken: grant.RefreshToken, Scopes: grant.Scopes, DriveRootFolderID: grant.DriveRootFolderID}, nil
}

// scopedMembers runs the roster reads inside the association's transaction (a pass-through when
// the request already is in it).
type scopedMembers struct {
	pool *pgxpool.Pool
	app.StorageMembers
}

func (s scopedMembers) MemberSignInAddresses(ctx context.Context, association uuid.UUID) (out []string, err error) {
	err = db.InAssociation(ctx, s.pool, association, "", func(ctx context.Context) (e error) {
		out, e = s.StorageMembers.MemberSignInAddresses(ctx, association)
		return e
	})
	return out, err
}

func (s scopedMembers) CurrentMemberID(ctx context.Context, association uuid.UUID, email string, date time.Time) (id uuid.UUID, ok bool, err error) {
	err = db.InAssociation(ctx, s.pool, association, "", func(ctx context.Context) (e error) {
		id, ok, e = s.StorageMembers.CurrentMemberID(ctx, association, email, date)
		return e
	})
	return id, ok, err
}
