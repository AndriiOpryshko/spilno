package governance_test

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
)

// The skeleton stubs every new operation with 501 problem+json; the session is
// not enforced yet (T11), so no cookie is sent.
func TestNewOperationsAreStubbed(t *testing.T) {
	srv := newServer(t)

	tests := []struct{ name, path string }{
		{"settings", "/governance/settings"},
		{"protocol", "/governance/protocols/" + uuid.NewString()},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var p problem
			do(t, "GET", srv.URL+tc.path, nil, http.StatusNotImplemented, &p)
			if p.Status != http.StatusNotImplemented {
				t.Errorf("body status = %d, want 501", p.Status)
			}
		})
	}
}

func TestExistingMeetingsStillWorkWithoutCookie(t *testing.T) {
	srv := newServer(t)
	do(t, "GET", srv.URL+"/governance/meetings", nil, http.StatusOK, nil)
}
