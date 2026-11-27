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

// Short secret parts (a one-letter user, "v=1") are ordinary word fragments outside a URL: masking them as
// substrings would garble the host and the prose around the address. Inside a URL every secret goes, because
// the URL is printed as scheme + host + path only.
func TestUnavailableShortSecretsDoNotGarbleTheMessage(t *testing.T) {
	p := New("http://a@127.0.0.1:33283/mcp?v=1&x=end")
	cause := errors.New(`calling "initialize": Post "http://a@127.0.0.1:33283/mcp?v=1&x=end": ` +
		`dial tcp 127.0.0.1:33283: connect: connection refused (transport endpoint)`)

	err := p.unavailable("initialize", cause)

	if !errors.Is(err, domain.ErrRateUnavailable) {
		t.Fatalf("got %v, want ErrRateUnavailable", err)
	}
	msg := err.Error()
	if !strings.Contains(msg, ": initialize: endpoint=http://127.0.0.1:33283/mcp: ") {
		t.Errorf("error %q lost the step or the safe endpoint", msg)
	}
	for _, want := range []string{
		`Post "http://127.0.0.1:33283/mcp":`,
		"dial tcp 127.0.0.1:33283: connect",
		`calling "initialize"`,
		"transport",
		"endpoint)",
	} {
		if !strings.Contains(msg, want) {
			t.Errorf("error %q, want it to contain %q", msg, want)
		}
	}
	for _, bad := range []string{"\x00", "v=1", "x=end", "a@", "***"} {
		if strings.Contains(msg, bad) {
			t.Errorf("error %q must not contain %q", msg, bad)
		}
	}
}

// Any other address in the message is reduced to scheme + host + path, whatever the length of its secrets.
func TestUnavailableReducesEveryURLToSchemeHostPath(t *testing.T) {
	tests := []struct {
		name  string
		cause string
		want  string
		leaks []string
	}{
		{
			name:  "short secrets",
			cause: `Get "http://b@10.0.0.1/moved?v=1#f": EOF`,
			want:  `Get "http://10.0.0.1/moved": EOF`,
			leaks: []string{"b@", "v=1", "#f"},
		},
		{
			name:  "long secrets of an unrelated address",
			cause: `Get "https://other-user:other-pass-word@10.0.0.2:9/p?key=other-long-key#other-frag": EOF`,
			want:  `Get "https://10.0.0.2:9/p": EOF`,
			leaks: []string{"other-user", "other-pass-word", "other-long-key", "other-frag"},
		},
		{
			name:  "unterminated IPv6 literal",
			cause: `Get "http://[::1": bad`,
			want:  `Get "<unparsable>": bad`,
			leaks: []string{"[::1"},
		},
		{
			name:  "bad port keeps no secret",
			cause: `Get "http://user-secret:pw-secret@h:badport/x?token=tok-secret-1": bad`,
			want:  `Get "<unparsable>": bad`,
			leaks: []string{"user-secret", "pw-secret", "tok-secret-1", "badport"},
		},
		{
			name:  "URL ended by whitespace",
			cause: `redirected to http://c@10.0.0.3/next?v=1 and stopped`,
			want:  `redirected to http://10.0.0.3/next and stopped`,
			leaks: []string{"c@", "v=1"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := New("http://svc@127.0.0.1:8090/mcp?v=1")

			err := p.unavailable("call", errors.New(tc.cause))

			if !errors.Is(err, domain.ErrRateUnavailable) {
				t.Fatalf("got %v, want ErrRateUnavailable", err)
			}
			msg := err.Error()
			if !strings.Contains(msg, tc.want) {
				t.Errorf("error %q, want it to contain %q", msg, tc.want)
			}
			if !strings.Contains(msg, "endpoint=http://127.0.0.1:8090/mcp: ") {
				t.Errorf("error %q lost the safe endpoint", msg)
			}
			for _, leak := range append(tc.leaks, "\x00") {
				if strings.Contains(msg, leak) {
					t.Errorf("error %q leaks %q", msg, leak)
				}
			}
		})
	}
}

// Outside a URL a secret part is masked only when it is long enough to be a credential (6 characters).
func TestUnavailableMasksLongSecretsOutsideURLsOnly(t *testing.T) {
	p := New("http://svc-user:abcdef@127.0.0.1:8090/mcp?k=abc&token=tok-long-1")
	cause := errors.New(`auth failed for svc-user with abcdef and tok-long-1; k=abc stays`)

	msg := p.unavailable("call", cause).Error()

	for _, leak := range []string{"svc-user", "abcdef", "tok-long-1"} {
		if strings.Contains(msg, leak) {
			t.Errorf("error %q leaks %q", msg, leak)
		}
	}
	if !strings.Contains(msg, "k=abc stays") {
		t.Errorf("error %q, want the short value left alone", msg)
	}
	if strings.Contains(msg, "\x00") {
		t.Errorf("error %q contains a NUL byte", msg)
	}
}
