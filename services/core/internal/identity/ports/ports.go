// Package ports declares what the identity context needs from the outside world.
package ports

import (
	"context"
	"errors"

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
}

// GoogleConnectionRepository stores the one Google connection an association has.
type GoogleConnectionRepository interface {
	// Replace deletes the association's connection and stores c, in one transaction.
	Replace(ctx context.Context, c GoogleConnection) error
	// Get returns ErrGoogleConnectionNotFound when the association has no connection.
	Get(ctx context.Context, associationID uuid.UUID) (GoogleConnection, error)
}
