package mcprates_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/AndriiOpryshko/spilno/services/core/internal/finance/adapters/mcprates"
	"github.com/AndriiOpryshko/spilno/services/core/internal/finance/domain"
	"github.com/AndriiOpryshko/spilno/services/core/internal/finance/ports"
)

var (
	_ ports.RateProvider = (*mcprates.Provider)(nil)
	_ ports.RateProvider = mcprates.Disabled{}
)

type in struct {
	Currency string `json:"currency"`
	Date     string `json:"date"`
}

type out struct {
	Currency string `json:"currency"`
	Date     string `json:"date"`
	Rate     string `json:"rate"`
	Source   string `json:"source"`
}

var day = time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)

// stub starts an in-process MCP server whose get_rate runs h.
func stub(t *testing.T, h func(ctx context.Context, in in) (out, error)) string {
	t.Helper()
	s := mcp.NewServer(&mcp.Implementation{Name: "stub", Version: "0"}, nil)
	mcp.AddTool(s, &mcp.Tool{Name: "get_rate", Description: "stub"},
		func(ctx context.Context, _ *mcp.CallToolRequest, i in) (*mcp.CallToolResult, out, error) {
			o, err := h(ctx, i)
			return nil, o, err
		})
	srv := httptest.NewServer(mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return s }, nil))
	t.Cleanup(srv.Close)
	return srv.URL
}

func answer(o out) func(context.Context, in) (out, error) {
	return func(context.Context, in) (out, error) { return o, nil }
}

func good() out {
	return out{Currency: "EUR", Date: "2026-03-10", Rate: "50.5311", Source: "NBU"}
}

func TestGoodAnswerBecomesRate(t *testing.T) {
	var got in
	url := stub(t, func(_ context.Context, i in) (out, error) { got = i; return good(), nil })
	r, err := mcprates.New(url).Rate(context.Background(), domain.CurrencyEUR, day)
	if err != nil {
		t.Fatalf("Rate: unexpected error %v", err)
	}
	if r.Currency != domain.CurrencyEUR || !r.Day.Equal(day) || r.Micros != 50531100 {
		t.Fatalf("rate = %+v, want EUR 2026-03-10 50531100 micros", r)
	}
	if got.Currency != "EUR" || got.Date != "2026-03-10" {
		t.Fatalf("tool called with %+v, want EUR 2026-03-10", got)
	}
}

func TestNotPublished(t *testing.T) {
	url := stub(t, func(context.Context, in) (out, error) {
		return out{}, errors.New("rate_not_published: the NBU has no official rate")
	})
	_, err := mcprates.New(url).Rate(context.Background(), domain.CurrencyEUR, day)
	if !errors.Is(err, domain.ErrRateNotPublished) {
		t.Fatalf("err = %v, want ErrRateNotPublished", err)
	}
}

func TestUnavailable(t *testing.T) {
	bad := func(mut func(*out)) out { o := good(); mut(&o); return o }
	cases := map[string]func(context.Context, in) (out, error){
		"other tool error": func(context.Context, in) (out, error) {
			return out{}, errors.New("upstream_unavailable: the NBU rate service did not answer")
		},
		"invalid argument": func(context.Context, in) (out, error) {
			return out{}, errors.New("invalid_argument: bad")
		},
		"wrong currency":  answer(bad(func(o *out) { o.Currency = "USD" })),
		"wrong date":      answer(bad(func(o *out) { o.Date = "2026-03-09" })),
		"wrong source":    answer(bad(func(o *out) { o.Source = "ECB" })),
		"malformed rate":  answer(bad(func(o *out) { o.Rate = "abc" })),
		"zero rate":       answer(bad(func(o *out) { o.Rate = "0" })),
		"too many digits": answer(bad(func(o *out) { o.Rate = "1.1234567" })),
	}
	for name, h := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := mcprates.New(stub(t, h)).Rate(context.Background(), domain.CurrencyEUR, day)
			if !errors.Is(err, domain.ErrRateUnavailable) {
				t.Fatalf("err = %v, want ErrRateUnavailable", err)
			}
		})
	}
}

func TestRefusedConnection(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	url := srv.URL
	srv.Close()
	_, err := mcprates.New(url).Rate(context.Background(), domain.CurrencyEUR, day)
	if !errors.Is(err, domain.ErrRateUnavailable) {
		t.Fatalf("err = %v, want ErrRateUnavailable", err)
	}
}

// A failed call must not leak the endpoint's userinfo, query or fragment into
// the error: the address is logged and the query may carry a token.
func TestErrorKeepsCredentialsOutOfTheAddress(t *testing.T) {
	closed := httptest.NewServer(http.NotFoundHandler())
	closedHost := closed.Listener.Addr().String()
	closed.Close()
	live := httptest.NewServer(http.NotFoundHandler()) // answers, but not MCP
	t.Cleanup(live.Close)

	cases := map[string]string{
		"refused connection": closedHost,
		"non-MCP answer":     live.Listener.Addr().String(),
	}
	for name, host := range cases {
		t.Run(name, func(t *testing.T) {
			addr := "http://svc-user:pw-secret@" + host + "/mcp?token=q-secret#frag"
			_, err := mcprates.New(addr).Rate(context.Background(), domain.CurrencyEUR, day)
			if !errors.Is(err, domain.ErrRateUnavailable) {
				t.Fatalf("err = %v, want ErrRateUnavailable", err)
			}
			msg := err.Error()
			for _, leak := range []string{"q-secret", "token=", "pw-secret", "frag", "svc-user"} {
				if strings.Contains(msg, leak) {
					t.Errorf("error %q leaks %q", msg, leak)
				}
			}
			for _, keep := range []string{host, "/mcp"} {
				if !strings.Contains(msg, keep) {
					t.Errorf("error %q lacks %q, the log would not say which endpoint failed", msg, keep)
				}
			}
		})
	}
}

// The rate source must not be able to move the client: a redirect is refused, and whatever address the
// transport ends up naming (another path, the query carried over) stays free of every secret part of the
// configured one.
func TestRedirectIsRefusedAndKeepsSecretsOutOfTheError(t *testing.T) {
	cases := map[string]struct{ path, target string }{
		"trailing slash": {"/mcp", "/mcp/"},
		"another path":   {"/mcp", "/moved/elsewhere?token=q-secret"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			followed := false
			mux := http.NewServeMux()
			mux.HandleFunc(c.path, func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, c.target, http.StatusTemporaryRedirect)
			})
			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				followed = true // drop the connection: the client reports a *url.Error naming the followed address
				if conn, _, err := w.(http.Hijacker).Hijack(); err == nil {
					_ = conn.Close()
				}
			})
			srv := httptest.NewServer(mux)
			t.Cleanup(srv.Close)

			addr := "http://svc-user:pw-secret@" + srv.Listener.Addr().String() + c.path + "?token=q-secret#frag"
			_, err := mcprates.New(addr).Rate(context.Background(), domain.CurrencyEUR, day)
			if !errors.Is(err, domain.ErrRateUnavailable) {
				t.Fatalf("err = %v, want ErrRateUnavailable", err)
			}
			if followed {
				t.Errorf("the redirect was followed, want it refused")
			}
			msg := err.Error()
			for _, leak := range []string{"q-secret", "token=", "svc-user", "pw-secret", "frag"} {
				if strings.Contains(msg, leak) {
					t.Errorf("error %q leaks %q", msg, leak)
				}
			}
		})
	}
}

func TestExpiringContextReturnsPromptly(t *testing.T) {
	release := make(chan struct{})
	t.Cleanup(func() { close(release) })
	url := stub(t, func(ctx context.Context, _ in) (out, error) {
		select {
		case <-release:
		case <-ctx.Done():
		}
		return good(), nil
	})
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	start := time.Now()
	_, err := mcprates.New(url).Rate(ctx, domain.CurrencyEUR, day)
	if !errors.Is(err, domain.ErrRateUnavailable) {
		t.Fatalf("err = %v, want ErrRateUnavailable", err)
	}
	if d := time.Since(start); d > 2*time.Second {
		t.Fatalf("returned after %v, want promptly", d)
	}
}

func TestDisabled(t *testing.T) {
	_, err := mcprates.Disabled{}.Rate(context.Background(), domain.CurrencyEUR, day)
	if !errors.Is(err, domain.ErrRateUnavailable) {
		t.Fatalf("err = %v, want ErrRateUnavailable", err)
	}
}
