package mcprates

import (
	"errors"
	"strings"
	"testing"

	"github.com/AndriiOpryshko/spilno/services/core/internal/finance/domain"
)

// An address net/http reports for another URL (after a redirect, say) matches no whole spelling of the
// endpoint, so its secret parts must be masked on their own.
func TestUnavailableMasksSecretPartsOfAnotherAddress(t *testing.T) {
	p := New("http://svc-user:pw%2Fsecret@127.0.0.1:8090/mcp?token=q-secret&x=1#frag-secret")
	cause := errors.New(`Post "http://svc-user:***@127.0.0.1:8090/moved?token=q-secret#frag-secret": EOF ` +
		`(pw/secret, pw%2Fsecret)`)

	err := p.unavailable("call", cause)

	if !errors.Is(err, domain.ErrRateUnavailable) {
		t.Fatalf("got %v, want ErrRateUnavailable", err)
	}
	msg := err.Error()
	for _, leak := range []string{"svc-user", "pw/secret", "pw%2Fsecret", "q-secret", "frag-secret"} {
		if strings.Contains(msg, leak) {
			t.Errorf("error %q leaks %q", msg, leak)
		}
	}
	if !strings.Contains(msg, "endpoint=http://127.0.0.1:8090/mcp:") {
		t.Errorf("error %q lost the safe endpoint", msg)
	}
}

// A whole spelling of the endpoint becomes the safe form, which the secret masking must leave intact
// even when a secret is a substring of the host.
func TestUnavailableKeepsTheSafeFormReadable(t *testing.T) {
	p := New("http://127.0.0.1:8090/mcp?n=1")
	cause := errors.New(`Post "http://127.0.0.1:8090/mcp?n=1": EOF`)

	msg := p.unavailable("connect", cause).Error()

	want := `Post "http://127.0.0.1:8090/mcp": EOF`
	if !strings.Contains(msg, want) {
		t.Errorf("error %q, want it to contain %q", msg, want)
	}
}
