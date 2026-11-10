package app

import (
	"bytes"
	"context"
	"crypto/sha256"
	"time"

	"github.com/google/uuid"

	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/ports"
)

// Cards serves the association card and its confirmation extract.
type Cards struct {
	repo ports.CardRepository
	now  func() time.Time
}

func NewCards(repo ports.CardRepository) *Cards {
	return &Cards{repo: repo, now: time.Now}
}

// Card returns the card of the association.
func (c *Cards) Card(ctx context.Context, id uuid.UUID) (ports.Card, error) {
	return c.repo.Card(ctx, id, c.now().UTC().Truncate(24*time.Hour))
}

// Extract returns an extract of the association after comparing its hash with the stored one.
func (c *Cards) Extract(ctx context.Context, association, extract uuid.UUID) (ports.StoredExtract, error) {
	e, err := c.repo.Extract(ctx, association, extract)
	if err != nil {
		return e, err
	}
	if sum := sha256.Sum256(e.Content); !bytes.Equal(sum[:], e.Sha256) {
		return ports.StoredExtract{}, ErrExtractCorrupted
	}
	return e, nil
}
