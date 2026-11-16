package app

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/domain"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/ports"
)

const (
	stateTTL = 10 * time.Minute
	// maxPendingStates caps the in-flight sign-in and consent states held in memory.
	maxPendingStates = 1000
	sessionTTL       = 30 * 24 * time.Hour
)

// ErrNoSession is returned when the token names no live session.
var ErrNoSession = errors.New("identity: no session")

type pending struct {
	verifier, nonce string
	expires         time.Time
}

// SignIn runs the sign-in flow and keeps sessions.
type SignIn struct {
	provider      ports.SignInProvider
	accounts      ports.AccountRepository
	sessions      ports.SessionRepository
	operatorEmail string
	now           func() time.Time

	mu     sync.Mutex
	states map[string]pending // state -> PKCE verifier and nonce, single use
}

func NewSignIn(p ports.SignInProvider, a ports.AccountRepository, s ports.SessionRepository, operatorEmail string) *SignIn {
	return &SignIn{provider: p, accounts: a, sessions: s, operatorEmail: operatorEmail, now: time.Now, states: map[string]pending{}}
}

func randomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func hashToken(token string) []byte {
	h := sha256.Sum256([]byte(token))
	return h[:]
}

// Start remembers a fresh state and returns the provider URL to redirect to.
func (s *SignIn) Start() (string, error) {
	state, err := randomToken()
	if err != nil {
		return "", err
	}
	nonce, err := randomToken()
	if err != nil {
		return "", err
	}
	verifier, err := randomToken()
	if err != nil {
		return "", err
	}
	challenge := sha256.Sum256([]byte(verifier))
	now := s.now()

	s.mu.Lock()
	evictToCap(s.states, now, func(p pending) time.Time { return p.expires })
	s.states[state] = pending{verifier: verifier, nonce: nonce, expires: now.Add(stateTTL)}
	s.mu.Unlock()

	return s.provider.AuthCodeURL(state, nonce, base64.RawURLEncoding.EncodeToString(challenge[:])), nil
}

// Complete exchanges the code and returns a new session token for the verified person.
// domain.ErrSignInRefused means the state is unknown or used, or the address is not verified.
func (s *SignIn) Complete(ctx context.Context, code, state string) (string, error) {
	s.mu.Lock()
	p, ok := s.states[state]
	delete(s.states, state)
	s.mu.Unlock()
	if !ok || s.now().After(p.expires) {
		return "", domain.ErrSignInRefused
	}
	id, err := s.provider.Exchange(ctx, code, p.verifier, p.nonce)
	if err != nil || !id.EmailVerified || id.Email == "" {
		return "", domain.ErrSignInRefused
	}
	accountID, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	acc, err := s.accounts.Upsert(ctx, accountID, id.Issuer, id.Subject, id.Email)
	if err != nil {
		return "", err
	}
	token, err := randomToken()
	if err != nil {
		return "", err
	}
	sessionID, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	if err := s.sessions.Create(ctx, sessionID, acc.ID, hashToken(token), s.now().Add(sessionTTL)); err != nil {
		return "", err
	}
	return token, nil
}

// Principal returns the person behind token, or ErrNoSession.
func (s *SignIn) Principal(ctx context.Context, token string) (domain.Principal, error) {
	if token == "" {
		return domain.Principal{}, ErrNoSession
	}
	acc, err := s.sessions.Account(ctx, hashToken(token))
	if errors.Is(err, ports.ErrSessionNotFound) {
		return domain.Principal{}, ErrNoSession
	}
	if err != nil {
		return domain.Principal{}, err
	}
	return domain.Principal{AccountID: acc.ID, Email: acc.Email, IsOperator: domain.IsOperator(acc.Email, s.operatorEmail)}, nil
}

// SignOut deletes the session behind token, or returns ErrNoSession.
func (s *SignIn) SignOut(ctx context.Context, token string) error {
	if _, err := s.Principal(ctx, token); err != nil {
		return err
	}
	return s.sessions.Delete(ctx, hashToken(token))
}

// evictToCap makes room for one more entry: it drops expired states once the map
// is at maxPendingStates and, if it is still full of live states, forgets the
// one closest to expiry.
func evictToCap[V any](states map[string]V, now time.Time, expires func(V) time.Time) {
	if len(states) < maxPendingStates {
		return
	}
	for k, v := range states {
		if now.After(expires(v)) {
			delete(states, k)
		}
	}
	for len(states) >= maxPendingStates {
		var oldest string
		var at time.Time
		for k, v := range states {
			if oldest == "" || expires(v).Before(at) {
				oldest, at = k, expires(v)
			}
		}
		delete(states, oldest)
	}
}
