// Package app holds the governance context's use cases.
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

	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/domain"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/ports"
)

const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// ErrInvalidCursor is returned when a list cursor was not issued by List.
var ErrInvalidCursor = errors.New("invalid cursor")

// MeetingPage is one slice of the newest-first meeting listing.
type MeetingPage struct {
	Items []domain.Meeting
	// NextCursor resumes the listing after Items; empty on the last page.
	NextCursor string
}

type MeetingService struct {
	repo ports.MeetingRepository
}

func NewMeetingService(repo ports.MeetingRepository) *MeetingService {
	return &MeetingService{repo: repo}
}

// Create validates the fields and stores them. Broken invariants come back as
// *domain.ValidationError.
func (s *MeetingService) Create(ctx context.Context, fields domain.MeetingFields) (domain.Meeting, error) {
	if err := fields.Validate(); err != nil {
		return domain.Meeting{}, err
	}
	return s.repo.Create(ctx, fields)
}

// Get returns domain.ErrMeetingNotFound when no meeting has the id.
func (s *MeetingService) Get(ctx context.Context, id uuid.UUID) (domain.Meeting, error) {
	return s.repo.Get(ctx, id)
}

// List returns up to limit meetings after cursor, newest first. A limit
// outside 1..MaxPageSize falls back to DefaultPageSize or MaxPageSize; an
// empty cursor starts from the top.
func (s *MeetingService) List(ctx context.Context, limit int, cursor string) (MeetingPage, error) {
	switch {
	case limit <= 0:
		limit = DefaultPageSize
	case limit > MaxPageSize:
		limit = MaxPageSize
	}
	var after *ports.MeetingKey
	if cursor != "" {
		key, err := decodeCursor(cursor)
		if err != nil {
			return MeetingPage{}, err
		}
		after = &key
	}

	// One extra row tells whether another page follows.
	items, err := s.repo.List(ctx, after, limit+1)
	if err != nil {
		return MeetingPage{}, err
	}
	page := MeetingPage{Items: items}
	if len(items) > limit {
		page.Items = items[:limit]
		last := page.Items[limit-1]
		page.NextCursor = encodeCursor(ports.MeetingKey{CreatedAt: last.CreatedAt, ID: last.ID})
	}
	return page, nil
}

// A cursor is "<created_at unix microseconds>:<id>" in unpadded base64url.
// Microseconds match the precision Postgres stores for timestamptz.
func encodeCursor(k ports.MeetingKey) string {
	raw := strconv.FormatInt(k.CreatedAt.UnixMicro(), 10) + ":" + k.ID.String()
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func decodeCursor(s string) (ports.MeetingKey, error) {
	raw, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return ports.MeetingKey{}, fmt.Errorf("%w: %v", ErrInvalidCursor, err)
	}
	micros, id, ok := strings.Cut(string(raw), ":")
	if !ok {
		return ports.MeetingKey{}, ErrInvalidCursor
	}
	us, err := strconv.ParseInt(micros, 10, 64)
	if err != nil {
		return ports.MeetingKey{}, fmt.Errorf("%w: %v", ErrInvalidCursor, err)
	}
	uid, err := uuid.Parse(id)
	if err != nil {
		return ports.MeetingKey{}, fmt.Errorf("%w: %v", ErrInvalidCursor, err)
	}
	return ports.MeetingKey{CreatedAt: time.UnixMicro(us), ID: uid}, nil
}
