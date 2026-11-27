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
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/kyivday"
)

// ErrBadStorageState is returned when the state is unknown, foreign, expired or already used.
var ErrBadStorageState = errors.New("identity: storage consent state is not valid")

const storageStateTTL = 15 * time.Minute

// StorageMembers is the roster read the storage connection needs.
type StorageMembers interface {
	MemberSignInAddresses(ctx context.Context, association uuid.UUID) ([]string, error)
	CurrentMemberID(ctx context.Context, association uuid.UUID, email string, date time.Time) (uuid.UUID, bool, error)
	// HasFiles reports whether the association has at least one stored file.
	HasFiles(ctx context.Context, association uuid.UUID) (bool, error)
}

type storageState struct {
	state       string
	association uuid.UUID
	expires     time.Time
}

// StorageConnections connects the association's document storage on behalf of a member.
type StorageConnections struct {
	consent     ports.StorageConsent
	members     StorageMembers
	connections *GoogleConnections
	repo        ports.GoogleConnectionRepository

	now    func() time.Time
	inTx   func(ctx context.Context, fn func(ctx context.Context) error) error
	mu     sync.Mutex
	states map[uuid.UUID]storageState // account -> its one pending consent
}

func NewStorageConnections(consent ports.StorageConsent, members StorageMembers, connections *GoogleConnections, repo ports.GoogleConnectionRepository) *StorageConnections {
	return &StorageConnections{consent: consent, members: members, connections: connections, repo: repo, states: map[uuid.UUID]storageState{}, now: time.Now,
		inTx: func(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) }}
}

// SetClock replaces the clock «today» is taken from.
func (s *StorageConnections) SetClock(now func() time.Time) { s.now = now }

// SetTransactor replaces how the write of a completed connection is wrapped in one transaction.
func (s *StorageConnections) SetTransactor(inTx func(ctx context.Context, fn func(ctx context.Context) error) error) {
	s.inTx = inTx
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
	if _, held := s.states[account]; !held {
		evictToCap(s.states, now, func(v storageState) time.Time { return v.expires })
	}
	s.states[account] = storageState{state: state, association: association, expires: now.Add(storageStateTTL)}
	s.mu.Unlock()
	return s.consent.AuthURL(state), nil
}

func (s *StorageConnections) take(state string, account, association uuid.UUID) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	st, ok := s.states[account]
	if !ok || st.state != state {
		return false
	}
	delete(s.states, account)
	return st.association == association && time.Now().Before(st.expires)
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
	var conn ports.GoogleConnection
	err = s.inTx(ctx, func(ctx context.Context) error {
		if err := s.refuseAnotherAccountWithFiles(ctx, association, grant.AccountEmail); err != nil {
			return err
		}
		if err := s.connections.Connect(ctx, NewGoogleConnection{
			AssociationID: association, AccountEmail: grant.AccountEmail, RefreshToken: grant.RefreshToken,
			Scopes: grant.Scopes, DriveRootFolderID: grant.DriveRootFolderID, AuthorizedBy: account,
		}); err != nil {
			return err
		}
		var e error
		conn, e = s.repo.Get(ctx, association)
		return e
	})
	return conn, err
}

// refuseAnotherAccountWithFiles answers domain.CodeStorageHasFiles when email is not the account of the current
// connection and the association has a file: the files stay on the old account and would answer 404. It runs
// in the transaction of the write, before it.
func (s *StorageConnections) refuseAnotherAccountWithFiles(ctx context.Context, association uuid.UUID, email string) error {
	current, err := s.repo.Get(ctx, association)
	if errors.Is(err, ports.ErrGoogleConnectionNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if strings.EqualFold(strings.TrimSpace(current.AccountEmail), strings.TrimSpace(email)) {
		return nil
	}
	has, err := s.members.HasFiles(ctx, association)
	if err != nil {
		return err
	}
	if has {
		return &domain.CodedError{Code: domain.CodeStorageHasFiles}
	}
	return nil
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
	id, ok, err := s.members.CurrentMemberID(ctx, c.AssociationID, c.AuthorizedByEmail, kyivday.Today(s.now()))
	if err != nil || !ok {
		return nil, err
	}
	return &id, nil
}
