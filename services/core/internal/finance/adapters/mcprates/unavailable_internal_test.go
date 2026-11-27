package mcprates

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"strings"
	"syscall"
	"testing"

	"github.com/AndriiOpryshko/spilno/services/core/internal/finance/domain"
)

// T32: an endpoint net/http would print escaped (%q) is refused, and neither the error nor the safe form
// carries any part of it beyond scheme, host and path.
func TestCheckRefusesAnEndpointThatNetHTTPPrintsEscaped(t *testing.T) {
	tests := []struct {
		name, endpoint, wantSafe string
		wantErr                  bool
		leaks                    []string
	}{
		{"quote in query", `http://127.0.0.1:1/mcp?token="quoted-secret"`, "http://127.0.0.1:1/mcp", true, []string{"quoted-secret", "token"}},
		{"quote in password", `http://u:pa"ss-secret@127.0.0.1:1/mcp`, "http://127.0.0.1:1/mcp", true, []string{"ss-secret", `pa"ss`, "u:"}},
		{"plain", "http://127.0.0.1:1/mcp?token=ok-secret", "http://127.0.0.1:1/mcp", false, nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			safe, err := Check(tc.endpoint)

			if (err != nil) != tc.wantErr {
				t.Fatalf("Check(%q) error = %v, want error: %v", tc.endpoint, err, tc.wantErr)
			}
			if tc.wantErr && safe != "" && safe != tc.wantSafe && safe != "<unparsable>" {
				t.Errorf("safe = %q, want %q or <unparsable>", safe, tc.wantSafe)
			}
			if !tc.wantErr && safe != tc.wantSafe {
				t.Errorf("safe = %q, want %q", safe, tc.wantSafe)
			}
			text := safe
			if err != nil {
				text += " " + err.Error()
			}
			for _, leak := range tc.leaks {
				if strings.Contains(text, leak) {
					t.Errorf("error/safe %q leaks %q", text, leak)
				}
			}
		})
	}
}

// T34 (AC-10): the error names the step, the safe endpoint and a fixed error class, and nothing of the
// cause's text: whatever net/http or the go-sdk print (the raw address, a redirect target, a quoted secret)
// never reaches the log, so there is nothing left to redact.
func TestUnavailableNamesTheErrorClassOnly(t *testing.T) {
	const secretURL = `http://svc-user:pw-secret@127.0.0.1:1/mcp?token="q-secret"#frag-secret`
	urlErr := func(cause error) error { return &url.Error{Op: "Post", URL: secretURL, Err: cause} }
	refused := &net.OpError{Op: "dial", Net: "tcp", Err: os.NewSyscallError("connect", syscall.ECONNREFUSED)}
	reset := &net.OpError{Op: "read", Net: "tcp", Err: os.NewSyscallError("read", syscall.ECONNRESET)}
	tests := []struct {
		name   string
		cause  error
		status int
		class  string
	}{
		{"refused", fmt.Errorf("calling %q: %w", "initialize", urlErr(refused)), 0, "connection refused"},
		{"deadline", urlErr(context.DeadlineExceeded), 0, "timeout"},
		{"canceled", urlErr(context.Canceled), 0, "canceled"},
		{"dropped", urlErr(io.EOF), 0, "connection closed"},
		{"unexpected eof", urlErr(io.ErrUnexpectedEOF), 0, "connection closed"},
		{"other network error", urlErr(reset), 0, "network error"},
		{"http status", fmt.Errorf("broken session: %s", secretURL), 503, "http 503"},
		{"anything else", errors.New(`bad answer from ` + secretURL), 0, "protocol error"},
	}
	p := New(secretURL)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := p.unavailable("connect", tc.cause, tc.status)

			if !errors.Is(err, domain.ErrRateUnavailable) {
				t.Fatalf("got %v, want ErrRateUnavailable", err)
			}
			want := "rate unavailable: connect: endpoint=http://127.0.0.1:1/mcp: " + tc.class
			if got := err.Error(); got != want {
				t.Errorf("error = %q, want %q", got, want)
			}
		})
	}
}
