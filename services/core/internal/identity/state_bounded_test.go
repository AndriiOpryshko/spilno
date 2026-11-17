package identity_test

import (
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"
)

// stateStarts is how many times an anonymous caller (sign-in) or a member (storage consent) starts a flow.
const stateStarts = 10000

// stateBudget is the heap the starts may leave behind. A server that keeps one entry per start
// (state, verifier, nonce, expiry) holds a few MB after stateStarts; a bounded one holds none.
const stateBudget = 1 << 20

func settledHeap() uint64 {
	var m runtime.MemStats
	runtime.GC()
	runtime.GC()
	runtime.ReadMemStats(&m)
	return m.HeapAlloc
}

// heapGrowthOf runs start stateStarts times and returns how much live heap it left behind.
func heapGrowthOf(t *testing.T, start func(i int)) int64 {
	t.Helper()
	start(-1) // warm up lazily built state
	before := settledHeap()
	for i := 0; i < stateStarts; i++ {
		start(i)
	}
	after := settledHeap()
	return int64(after) - int64(before)
}

// AC-05: starting sign-in any number of times does not grow server memory.
func TestStartingSignInManyTimesLeavesServerMemoryBounded(t *testing.T) {
	srv, _ := newSignInServer(t)
	h := srv.Config.Handler
	growth := heapGrowthOf(t, func(int) {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest("GET", "/identity/sign-in", nil))
		if rec.Code != http.StatusFound {
			t.Fatalf("startSignIn status = %d, want 302", rec.Code)
		}
	})
	if growth > stateBudget {
		t.Errorf("heap grew by %d bytes after %d sign-in starts, want at most %d: the state must not be kept per start in server memory", growth, stateStarts, stateBudget)
	}
}

// AC-27: starting the storage consent any number of times does not grow server memory.
func TestStartingStorageConsentManyTimesLeavesServerMemoryBounded(t *testing.T) {
	env := newStorageT33Env(t)
	h := env.srv.Config.Handler
	growth := heapGrowthOf(t, func(int) {
		req := httptest.NewRequest("GET", storageT33Path(env.a)+"/authorization", nil)
		req.AddCookie(env.cookies["member"])
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusFound {
			t.Fatalf("startStorageConnection status = %d, want 302; body %s", rec.Code, rec.Body)
		}
	})
	if growth > stateBudget {
		t.Errorf("heap grew by %d bytes after %d consent starts, want at most %d: the state must not be kept per start in server memory", growth, stateStarts, stateBudget)
	}
}

// AC-05: a flood of starts does not push out a sign-in already in progress.
func TestSignInFloodLeavesAnEarlierSignInAbleToFinish(t *testing.T) {
	srv, _ := newSignInServer(t)
	state, cookies := startSignIn(t, srv)
	for i := 0; i < 2*1000+500; i++ {
		startSignIn(t, srv)
	}
	cb := callbackWith(t, srv, state, cookies)
	if cb.StatusCode != http.StatusFound || sessionCookieOf(cb) == nil {
		t.Errorf("callback of an earlier start after a flood = %d (session cookie %v), want 302 with a session", cb.StatusCode, sessionCookieOf(cb) != nil)
	}
}

// AC-27: a flood of consent starts by one member does not push out another member's pending consent.
func TestStorageConsentFloodLeavesAnotherMembersConsentAbleToFinish(t *testing.T) {
	env := newStorageT33Env(t)
	state := env.start(t, "member", env.a)
	for i := 0; i < 2500; i++ {
		env.start(t, "memberB", env.b)
	}
	if status, body := env.complete(t, "member", env.a, "code-1", state); status != 201 {
		t.Errorf("complete after a flood = %d, want 201; body %s", status, body)
	}
}

// AC-27: one account holds at most one pending storage consent: a later start replaces the earlier one.
func TestStorageConsentHoldsOnePendingEntryPerAccount(t *testing.T) {
	env := newStorageT33Env(t)
	first := env.start(t, "member", env.a)
	second := env.start(t, "member", env.a)
	if status, _ := env.complete(t, "member", env.a, "code-1", first); status != 422 {
		t.Errorf("earlier state after a later start = %d, want 422 (replaced)", status)
	}
	if status, body := env.complete(t, "member", env.a, "code-1", second); status != 201 {
		t.Errorf("latest state = %d, want 201; body %s", status, body)
	}
}
