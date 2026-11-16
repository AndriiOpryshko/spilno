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
