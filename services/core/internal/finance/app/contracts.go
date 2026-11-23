// Package app holds the finance context's use cases.
package app

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
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

// RateTimeout bounds the wait for the rate source when a contract is created.
var RateTimeout = 3 * time.Second

type ContractService struct {
	repo  ports.ContractRepository
	rates ports.RateProvider
}

func NewContractService(repo ports.ContractRepository, rates ports.RateProvider) *ContractService {
	return &ContractService{repo: repo, rates: rates}
}

// Create validates the terms and stores them. Broken invariants come back as
// *domain.ValidationError.
func (s *ContractService) Create(ctx context.Context, terms domain.ContractTerms) (domain.Contract, error) {
	if err := terms.Validate(); err != nil {
		return domain.Contract{}, err
	}
	return s.repo.Create(ctx, terms, s.equivalent(ctx, terms))
}

// equivalent asks the rate source for the signing day's rate of a
// foreign-currency contract. A failed lookup is logged and yields nil so the
// contract is still stored.
func (s *ContractService) equivalent(ctx context.Context, terms domain.ContractTerms) *domain.UAHEquivalent {
	if !terms.Currency.Foreign() {
		return nil
	}
	rate, err := s.signingDayRate(ctx, terms.Currency, terms.SignedAt)
	if err != nil {
		return nil
	}
	eq := rate.Convert(terms.AmountCents)
	return &eq
}

// signingDayRate fetches the rate for the calendar day of signedAt within
// RateTimeout. It logs any failure; the caller decides what it means.
func (s *ContractService) signingDayRate(ctx context.Context, currency domain.Currency, signedAt time.Time) (domain.Rate, error) {
	y, m, d := signedAt.Date()
	day := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)

	rctx, cancel := context.WithTimeout(ctx, RateTimeout)
	defer cancel()
	rate, err := s.rates.Rate(rctx, currency, day)
	if err != nil {
		slog.WarnContext(ctx, "uah equivalent not resolved",
			"currency", currency, "day", day.Format(time.DateOnly), "error", err)
	}
	return rate, err
}

// ResolveUAHEquivalent works out a missing equivalent once (Flow 2). A stored
// equivalent is returned as is; a lost race returns what the winner stored.
func (s *ContractService) ResolveUAHEquivalent(ctx context.Context, id uuid.UUID) (domain.Contract, error) {
	c, err := s.repo.Get(ctx, id)
	if err != nil {
		return domain.Contract{}, err
	}
	if c.UAHEquivalent != nil {
		return c, nil
	}
	if !c.Currency.Foreign() {
		return domain.Contract{}, &domain.DomainError{Code: domain.CodeEquivalentNotApplicable}
	}
	rate, err := s.signingDayRate(ctx, c.Currency, c.SignedAt)
	if err != nil {
		if errors.Is(err, domain.ErrRateNotPublished) {
			return domain.Contract{}, &domain.DomainError{Code: domain.CodeRateNotPublished}
		}
		return domain.Contract{}, &domain.DomainError{Code: domain.CodeRateUnavailable}
	}
	updated, ok, err := s.repo.SetUAHEquivalent(ctx, id, rate.Convert(c.AmountCents))
	if err != nil {
		return domain.Contract{}, err
	}
	if !ok {
		return s.repo.Get(ctx, id)
	}
	return updated, nil
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
