// Package app holds the finance context's use cases.
package app

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/AndriiOpryshko/spilno/services/core/internal/finance/domain"
	"github.com/AndriiOpryshko/spilno/services/core/internal/finance/ports"
)

const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// ErrInvalidCursor is returned when a list cursor was not issued by List.
var ErrInvalidCursor = errors.New("invalid cursor")

// Page is one slice of the newest-first contract listing.
type Page struct {
	Items []domain.Contract
	// NextCursor resumes the listing after Items; empty on the last page.
	NextCursor string
}

type ContractService struct {
	repo ports.ContractRepository
}

func NewContractService(repo ports.ContractRepository) *ContractService {
	return &ContractService{repo: repo}
}

// Create validates the terms and stores them. Broken invariants come back as
// *domain.ValidationError.
func (s *ContractService) Create(ctx context.Context, terms domain.ContractTerms) (domain.Contract, error) {
	if err := terms.Validate(); err != nil {
		return domain.Contract{}, err
	}
	return s.repo.Create(ctx, terms)
}

// Get returns domain.ErrContractNotFound when no contract has the id.
func (s *ContractService) Get(ctx context.Context, id uuid.UUID) (domain.Contract, error) {
	return s.repo.Get(ctx, id)
}

// List returns up to limit contracts after cursor, newest first. A limit
// outside 1..MaxPageSize falls back to DefaultPageSize or MaxPageSize; an
// empty cursor starts from the top.
func (s *ContractService) List(ctx context.Context, limit int, cursor string) (Page, error) {
	switch {
	case limit <= 0:
		limit = DefaultPageSize
	case limit > MaxPageSize:
		limit = MaxPageSize
	}
	var after *ports.ContractKey
	if cursor != "" {
		key, err := decodeCursor(cursor)
		if err != nil {
			return Page{}, err
		}
		after = &key
	}

	// One extra row tells whether another page follows.
	items, err := s.repo.List(ctx, after, limit+1)
	if err != nil {
		return Page{}, err
	}
	page := Page{Items: items}
	if len(items) > limit {
		page.Items = items[:limit]
		last := page.Items[limit-1]
		page.NextCursor = encodeCursor(ports.ContractKey{CreatedAt: last.CreatedAt, ID: last.ID})
	}
	return page, nil
}

// A cursor is "<created_at unix microseconds>:<id>" in unpadded base64url.
// Microseconds match the precision Postgres stores for timestamptz.
func encodeCursor(k ports.ContractKey) string {
	raw := strconv.FormatInt(k.CreatedAt.UnixMicro(), 10) + ":" + k.ID.String()
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func decodeCursor(s string) (ports.ContractKey, error) {
	raw, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return ports.ContractKey{}, fmt.Errorf("%w: %v", ErrInvalidCursor, err)
	}
	micros, id, ok := strings.Cut(string(raw), ":")
	if !ok {
		return ports.ContractKey{}, ErrInvalidCursor
	}
	us, err := strconv.ParseInt(micros, 10, 64)
	if err != nil {
		return ports.ContractKey{}, fmt.Errorf("%w: %v", ErrInvalidCursor, err)
	}
	uid, err := uuid.Parse(id)
	if err != nil {
		return ports.ContractKey{}, fmt.Errorf("%w: %v", ErrInvalidCursor, err)
	}
	return ports.ContractKey{CreatedAt: time.UnixMicro(us), ID: uid}, nil
}
