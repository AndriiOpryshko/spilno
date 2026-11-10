package app

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"

	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/ports"
)

// ErrInvalidGoogleConnection is returned by Connect when the refresh token or
// the account email is empty or there is no scope.
var ErrInvalidGoogleConnection = errors.New("identity: invalid google connection")

// NewGoogleConnection is the consent to store; the refresh token is still in plain text.
type NewGoogleConnection struct {
	AssociationID     uuid.UUID
	AccountEmail      string
	RefreshToken      string
	Scopes            []string
	DriveRootFolderID string
	AuthorizedBy      uuid.UUID
}

// GoogleConnections keeps the association's Google connection with its refresh token sealed.
type GoogleConnections struct {
	repo   ports.GoogleConnectionRepository
	cipher *TokenCipher
}

func NewGoogleConnections(repo ports.GoogleConnectionRepository, cipher *TokenCipher) *GoogleConnections {
	return &GoogleConnections{repo: repo, cipher: cipher}
}

// Connect seals the refresh token and stores the connection, replacing the one the association had.
func (s *GoogleConnections) Connect(ctx context.Context, c NewGoogleConnection) error {
	if c.RefreshToken == "" || c.AccountEmail == "" || len(c.Scopes) == 0 {
		return ErrInvalidGoogleConnection
	}
	sealed, err := s.cipher.Seal(c.RefreshToken)
	if err != nil {
		return err
	}
	id, err := uuid.NewV7()
	if err != nil {
		return err
	}
	return s.repo.Replace(ctx, ports.GoogleConnection{
		ID:                     id,
		AssociationID:          c.AssociationID,
		AccountEmail:           c.AccountEmail,
		RefreshTokenCiphertext: sealed.Ciphertext,
		RefreshTokenNonce:      sealed.Nonce,
		EncryptionKeyID:        sealed.KeyID,
		Scopes:                 strings.Join(c.Scopes, " "),
		DriveRootFolderID:      c.DriveRootFolderID,
		AuthorizedBy:           c.AuthorizedBy,
	})
}

// RefreshToken returns the association's refresh token in plain text. It answers
// ports.ErrGoogleConnectionNotFound when there is no connection, and ErrUnknownKey
// or ErrTokenUnreadable when the stored token cannot be opened with this key.
func (s *GoogleConnections) RefreshToken(ctx context.Context, associationID uuid.UUID) (string, error) {
	c, err := s.repo.Get(ctx, associationID)
	if err != nil {
		return "", err
	}
	return s.cipher.Open(SealedToken{
		Ciphertext: c.RefreshTokenCiphertext,
		Nonce:      c.RefreshTokenNonce,
		KeyID:      c.EncryptionKeyID,
	})
}
