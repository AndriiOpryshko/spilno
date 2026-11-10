package ports

import (
	"context"
	"errors"
)

// ErrStorageCodeRefused is returned when the provider does not accept the authorization code.
var ErrStorageCodeRefused = errors.New("identity: storage consent code refused")

// StorageGrant is what a completed storage consent gives.
type StorageGrant struct {
	AccountEmail      string
	RefreshToken      string
	Scopes            []string
	DriveRootFolderID string
}

// StorageConsent is the provider side of connecting the association's document storage.
type StorageConsent interface {
	AuthURL(state string) string
	// Complete returns ErrStorageCodeRefused when the provider refuses the code.
	Complete(ctx context.Context, code string) (StorageGrant, error)
}
