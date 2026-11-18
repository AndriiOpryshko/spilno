package app

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/url"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/domain"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/ports"
)

const (
	stateTTL = 10 * time.Minute
	// maxPendingStates caps the pending storage consents and the used sign-in states held in memory.
	maxPendingStates = 1000
	sessionTTL       = 30 * 24 * time.Hour
)

// ErrNoSession is returned when the token names no live session.
var ErrNoSession = errors.New("identity: no session")

// pending is what a sign-in in progress needs; it travels sealed in the browser's cookie.
type pending struct {
	State    string    `json:"s"`
	Verifier string    `json:"v"`
	Nonce    string    `json:"n"`
	Expires  time.Time `json:"e"`
}

// apiCallbackPath is the API's own sign-in callback path.
const apiCallbackPath = "/identity/sign-in/callback"

// SignIn runs the sign-in flow and keeps sessions.
type SignIn struct {
	provider      ports.SignInProvider
	accounts      ports.AccountRepository
	sessions      ports.SessionRepository
	operatorEmail string
	now           func() time.Time
	aead          cipher.AEAD // random per process, seals the pending sign-in

	mu   sync.Mutex
	used map[string]time.Time // states already completed -> expiry; only completed sign-ins are kept
}

func NewSignIn(p ports.SignInProvider, a ports.AccountRepository, s ports.SessionRepository, operatorEmail string) *SignIn {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		panic(err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		panic(err)
	}
	return &SignIn{provider: p, accounts: a, sessions: s, operatorEmail: operatorEmail, now: time.Now, aead: aead, used: map[string]time.Time{}}
}

// SetClock replaces the clock the pending sign-in expiry is read from.
func (s *SignIn) SetClock(now func() time.Time) { s.now = now }

// CallbackPath is the path the pending sign-in cookie is scoped to: the path of the provider's redirect address.
// A provider that does not state its address gets the API's own callback path.
func (s *SignIn) CallbackPath() string {
	r, ok := s.provider.(ports.RedirectAddresser)
	if !ok {
		return apiCallbackPath
	}
	u, err := url.Parse(r.RedirectURL())
	if err != nil || u.Path == "" {
		return apiCallbackPath
	}
	return u.Path
}

// StateMaxAge is how long the pending sign-in cookie lives, in seconds.
func StateMaxAge() int { return int(stateTTL / time.Second) }

func (s *SignIn) seal(p pending) (string, error) {
	plain, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, s.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(s.aead.Seal(nonce, nonce, plain, nil)), nil
}

func (s *SignIn) open(sealed string) (pending, bool) {
	raw, err := base64.RawURLEncoding.DecodeString(sealed)
	if err != nil || len(raw) < s.aead.NonceSize() {
		return pending{}, false
	}
	plain, err := s.aead.Open(nil, raw[:s.aead.NonceSize()], raw[s.aead.NonceSize():], nil)
	if err != nil {
		return pending{}, false
	}
	var p pending
	if json.Unmarshal(plain, &p) != nil {
		return pending{}, false
	}
	return p, true
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

// Start returns the provider URL to redirect to and the sealed pending sign-in the browser
// must carry back to the callback. The server keeps nothing.
func (s *SignIn) Start() (redirect, sealed string, err error) {
	state, err := randomToken()
	if err != nil {
		return "", "", err
	}
	nonce, err := randomToken()
	if err != nil {
		return "", "", err
	}
	verifier, err := randomToken()
	if err != nil {
		return "", "", err
	}
	challenge := sha256.Sum256([]byte(verifier))
	sealed, err = s.seal(pending{State: state, Verifier: verifier, Nonce: nonce, Expires: s.now().Add(stateTTL)})
	if err != nil {
		return "", "", err
	}
	return s.provider.AuthCodeURL(state, nonce, base64.RawURLEncoding.EncodeToString(challenge[:])), sealed, nil
}

// Complete exchanges the code and returns a new session token for the verified person.
// sealed is the cookie Start produced. domain.ErrSignInRefused means the cookie is missing, forged,
// expired or of another state, or the address is not verified.
func (s *SignIn) Complete(ctx context.Context, code, state, sealed string) (string, error) {
	p, ok := s.open(sealed)
	if !ok || subtle.ConstantTimeCompare([]byte(p.State), []byte(state)) != 1 || s.now().After(p.Expires) {
		return "", domain.ErrSignInRefused
	}
	if !s.markUsed(p) {
		return "", domain.ErrSignInRefused
	}
	id, err := s.provider.Exchange(ctx, code, p.Verifier, p.Nonce)
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

// markUsed records the state as used and reports false for a replay. States are kept only until
// the cookie that carried them would have expired, and at most maxPendingStates at a time: past
// the cap expired states are swept and, if the set is still full, the one closest to expiry is
// forgotten. Losing an old entry reopens only the replay of a code the provider has already consumed.
func (s *SignIn) markUsed(p pending) bool {
	now := s.now()
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.used[p.State]; ok {
		return false
	}
	evictToCap(s.used, now, func(exp time.Time) time.Time { return exp })
	s.used[p.State] = p.Expires
	return true
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
func evictToCap[K comparable, V any](states map[K]V, now time.Time, expires func(V) time.Time) {
	if len(states) < maxPendingStates {
		return
	}
	for k, v := range states {
		if now.After(expires(v)) {
			delete(states, k)
		}
	}
	for len(states) >= maxPendingStates {
		var oldest K
		var at time.Time
		first := true
		for k, v := range states {
			if first || expires(v).Before(at) {
				oldest, at, first = k, expires(v), false
			}
		}
		delete(states, oldest)
	}
}
