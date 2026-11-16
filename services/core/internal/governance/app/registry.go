package app

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/domain"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/ports"
)

// ErrSignedProtocolNotFound is returned when the association has no signed protocol with the id; a
// withdrawn protocol is not one (ADR-0007).
var ErrSignedProtocolNotFound = errors.New("signed protocol not found")

const (
	codeSearchInvalid = "governance.search_invalid"
	dateLayout        = "2006-01-02"
)

// RegistryService serves the реєстр протоколів, which is the same read for an enabled and for a
// вимкнене ОСББ; every call runs on the transaction in ctx.
type RegistryService struct {
	registry ports.RegistryReader
}

func NewRegistryService(registry ports.RegistryReader) *RegistryService {
	return &RegistryService{registry: registry}
}

// RegistryPage is one slice of the newest-meeting-first registry.
type RegistryPage struct {
	Items []ports.RegistryDecision
	// WithdrawnNumbers are the numbers of withdrawn protocols inside the page's number range.
	WithdrawnNumbers []int
	// NextCursor resumes the search after Items; empty on the last page.
	NextCursor string
}

// Search returns up to limit decisions matching the filter after cursor. A broken filter comes back as
// *domain.CodedFieldsError naming every broken filter at once; a cursor this service never issued as
// ErrInvalidCursor. A limit outside 1..MaxPageSize falls back like the meeting listing.
func (s *RegistryService) Search(ctx context.Context, association uuid.UUID, f ports.RegistryFilter, limit int, cursor string) (RegistryPage, error) {
	if err := validateFilter(f); err != nil {
		return RegistryPage{}, err
	}
	switch {
	case limit <= 0:
		limit = DefaultPageSize
	case limit > MaxPageSize:
		limit = MaxPageSize
	}
	var after *ports.RegistryKey
	if cursor != "" {
		key, err := decodeRegistryCursor(cursor)
		if err != nil {
			return RegistryPage{}, err
		}
		after = &key
	}
	// One extra row tells whether another page follows.
	items, err := s.registry.Search(ctx, association, f, after, limit+1)
	if err != nil {
		return RegistryPage{}, err
	}
	page := RegistryPage{Items: items, WithdrawnNumbers: []int{}}
	if len(items) > limit {
		page.Items = items[:limit]
		page.NextCursor = encodeRegistryCursor(page.Items[limit-1].Key)
	}
	if len(page.Items) > 0 {
		lo, hi := page.Items[0].Number, page.Items[0].Number
		for _, d := range page.Items {
			lo, hi = min(lo, d.Number), max(hi, d.Number)
		}
		if page.WithdrawnNumbers, err = s.registry.WithdrawnNumbers(ctx, association, lo, hi); err != nil {
			return RegistryPage{}, err
		}
	}
	return page, nil
}

// SignedProtocol returns the signed snapshot of the protocol, or ErrSignedProtocolNotFound.
func (s *RegistryService) SignedProtocol(ctx context.Context, association, protocol uuid.UUID) (ports.SignedSnapshot, error) {
	snap, ok, err := s.registry.SignedSnapshot(ctx, association, protocol)
	if err == nil && !ok {
		err = ErrSignedProtocolNotFound
	}
	return snap, err
}

// validateFilter names every broken filter at once: an amount range or a period that runs backwards.
func validateFilter(f ports.RegistryFilter) error {
	var issues []domain.FieldIssue
	if f.AmountFrom != nil && f.AmountTo != nil && *f.AmountFrom > *f.AmountTo {
		issues = append(issues, domain.FieldIssue{Pointer: "/query/amount_from_kopiyky", Detail: "above amount_to_kopiyky"})
	}
	if f.DateFrom != nil && f.DateTo != nil && f.DateFrom.After(*f.DateTo) {
		issues = append(issues, domain.FieldIssue{Pointer: "/query/date_from", Detail: "after date_to"})
	}
	if len(issues) == 0 {
		return nil
	}
	return &domain.CodedFieldsError{Code: codeSearchInvalid, Fields: issues}
}

// A registry cursor is "<meeting date>:<entry id>" in unpadded base64url.
func encodeRegistryCursor(k ports.RegistryKey) string {
	return base64.RawURLEncoding.EncodeToString([]byte(k.MeetingDate.Format(dateLayout) + ":" + k.ID.String()))
}

func decodeRegistryCursor(s string) (ports.RegistryKey, error) {
	raw, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return ports.RegistryKey{}, fmt.Errorf("%w: %v", ErrInvalidCursor, err)
	}
	day, id, ok := strings.Cut(string(raw), ":")
	if !ok {
		return ports.RegistryKey{}, ErrInvalidCursor
	}
	date, err := time.Parse(dateLayout, day)
	if err != nil {
		return ports.RegistryKey{}, fmt.Errorf("%w: %v", ErrInvalidCursor, err)
	}
	uid, err := uuid.Parse(id)
	if err != nil {
		return ports.RegistryKey{}, fmt.Errorf("%w: %v", ErrInvalidCursor, err)
	}
	return ports.RegistryKey{MeetingDate: date, ID: uid}, nil
}
