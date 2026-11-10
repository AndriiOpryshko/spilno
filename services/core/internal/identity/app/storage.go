package app

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/domain"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/ports"
)

// ErrBadStorageState is returned when the state is unknown, foreign, expired or already used.
var ErrBadStorageState = errors.New("identity: storage consent state is not valid")

const storageStateTTL = 15 * time.Minute

// StorageMembers is the roster read the storage connection needs.
type StorageMembers interface {
	MemberSignInAddresses(ctx context.Context, association uuid.UUID) ([]string, error)
	CurrentMemberID(ctx context.Context, association uuid.UUID, email string, date time.Time) (uuid.UUID, bool, error)
}

type storageState struct {
	account, association uuid.UUID
	expires              time.Time
}

// StorageConnections connects the association's document storage on behalf of a member.
type StorageConnections struct {
	consent     ports.StorageConsent
	members     StorageMembers
	connections *GoogleConnections
	repo        ports.GoogleConnectionRepository

	mu     sync.Mutex
	states map[string]storageState
}

func NewStorageConnections(consent ports.StorageConsent, members StorageMembers, connections *GoogleConnections, repo ports.GoogleConnectionRepository) *StorageConnections {
	return &StorageConnections{consent: consent, members: members, connections: connections, repo: repo, states: map[string]storageState{}}
}

// Start returns the consent link; its single-use state is bound to the account and the association.
func (s *StorageConnections) Start(account, association uuid.UUID) (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	state := base64.RawURLEncoding.EncodeToString(raw)
	now := time.Now()
	s.mu.Lock()
	for k, v := range s.states {
		if now.After(v.expires) {
			delete(s.states, k)
		}
	}
	s.states[state] = storageState{account: account, association: association, expires: now.Add(storageStateTTL)}
	s.mu.Unlock()
	return s.consent.AuthURL(state), nil
}

func (s *StorageConnections) take(state string, account, association uuid.UUID) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	st, ok := s.states[state]
	if !ok {
		return false
	}
	delete(s.states, state)
	return st.account == account && st.association == association && time.Now().Before(st.expires)
}

// Complete finishes the consent and stores the connection. It returns ErrBadStorageState,
// ports.ErrStorageCodeRefused or a *domain.CodedError for a personal account.
func (s *StorageConnections) Complete(ctx context.Context, account, association uuid.UUID, code, state string) (ports.GoogleConnection, error) {
	if !s.take(state, account, association) {
		return ports.GoogleConnection{}, ErrBadStorageState
	}
	grant, err := s.consent.Complete(ctx, code)
	if err != nil {
		return ports.GoogleConnection{}, err
	}
	addrs, err := s.members.MemberSignInAddresses(ctx, association)
	if err != nil {
		return ports.GoogleConnection{}, err
	}
	for _, a := range addrs {
		if strings.EqualFold(strings.TrimSpace(a), strings.TrimSpace(grant.AccountEmail)) {
			return ports.GoogleConnection{}, &domain.CodedError{Code: domain.CodeStorageAccountPersonal}
		}
	}
	err = s.connections.Connect(ctx, NewGoogleConnection{
		AssociationID: association, AccountEmail: grant.AccountEmail, RefreshToken: grant.RefreshToken,
		Scopes: grant.Scopes, DriveRootFolderID: grant.DriveRootFolderID, AuthorizedBy: account,
	})
	if err != nil {
		return ports.GoogleConnection{}, err
	}
	return s.repo.Get(ctx, association)
}

// Get returns the stored connection or ports.ErrGoogleConnectionNotFound.
func (s *StorageConnections) Get(ctx context.Context, association uuid.UUID) (ports.GoogleConnection, error) {
	return s.repo.Get(ctx, association)
}

// ConnectedBy returns the board member who connected the storage, or nil when that person
// is no longer on the roster.
func (s *StorageConnections) ConnectedBy(ctx context.Context, c ports.GoogleConnection) (*uuid.UUID, error) {
	if c.AuthorizedByEmail == "" {
		return nil, nil
	}
	id, ok, err := s.members.CurrentMemberID(ctx, c.AssociationID, c.AuthorizedByEmail, time.Now().UTC())
	if err != nil || !ok {
		return nil, err
	}
	return &id, nil
}
