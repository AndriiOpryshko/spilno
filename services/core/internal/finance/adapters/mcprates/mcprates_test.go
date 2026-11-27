package mcprates_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
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

// T34: a failed lookup names a fixed error class, never the text net/http or the go-sdk printed.
func TestErrorNamesTheFailureClass(t *testing.T) {
	down := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "down", http.StatusServiceUnavailable)
	}))
	t.Cleanup(down.Close)
	closed := httptest.NewServer(http.NotFoundHandler())
	closed.Close()

	cases := map[string]struct{ url, want string }{
		"source answers 503": {down.URL + "/mcp?token=q-secret", ": connect: endpoint=" + down.URL + "/mcp: http 503"},
		"source is down":     {closed.URL + "/mcp?token=q-secret", ": connect: endpoint=" + closed.URL + "/mcp: connection refused"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := mcprates.New(c.url).Rate(context.Background(), domain.CurrencyEUR, day)
			if !errors.Is(err, domain.ErrRateUnavailable) {
				t.Fatalf("err = %v, want ErrRateUnavailable", err)
			}
			if msg := err.Error(); !strings.HasSuffix(msg, c.want) {
				t.Errorf("error %q, want it to end with %q", msg, c.want)
			}
		})
	}
}

// A failed call must not leak the endpoint's userinfo, query or fragment into
// the error: the address is logged and the query may carry a token.
func TestErrorKeepsCredentialsOutOfTheAddress(t *testing.T) {
	closed := httptest.NewServer(http.NotFoundHandler())
	closedHost := closed.Listener.Addr().String()
	closed.Close()
	// Completes the MCP initialize handshake, then drops the connection on tools/call, so net/http reports
	// `Post "<url>": EOF` on the call step and the error names the configured address.
	mcpSrv := mcp.NewServer(&mcp.Implementation{Name: "stub", Version: "0"}, nil)
	mcp.AddTool(mcpSrv, &mcp.Tool{Name: "get_rate", Description: "stub"},
		func(context.Context, *mcp.CallToolRequest, in) (*mcp.CallToolResult, out, error) {
			return nil, good(), nil
		})
	inner := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return mcpSrv }, nil)
	dropper := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if r.Method == http.MethodPost && bytes.Contains(body, []byte(`"method":"tools/call"`)) {
			if conn, _, err := w.(http.Hijacker).Hijack(); err == nil {
				_ = conn.Close()
			}
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(body))
		inner.ServeHTTP(w, r)
	}))
	t.Cleanup(dropper.Close)

	cases := map[string]struct{ host, step string }{
		"refused connection":         {closedHost, "connect"},
		"connection dropped on call": {dropper.Listener.Addr().String(), "call"},
	}
	for name, c := range cases {
		host := c.host
		t.Run(name, func(t *testing.T) {
			addr := "http://svc-user:pw-secret@" + host + "/mcp?token=q-secret#frag"
			_, err := mcprates.New(addr).Rate(context.Background(), domain.CurrencyEUR, day)
			if !errors.Is(err, domain.ErrRateUnavailable) {
				t.Fatalf("err = %v, want ErrRateUnavailable", err)
			}
			msg := err.Error()
			if !strings.Contains(msg, ": "+c.step+": ") {
				t.Errorf("error %q does not name the %q step", msg, c.step)
			}
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

// hungStub is an MCP server that completes the handshake (initialize, notifications/initialized) and then
// reads every tools/call, notifications/cancelled and session-closing DELETE request and stops answering.
// After an expired call the go-sdk (v1.8.0) sends the DELETE of session.Close on a context of its own; the
// stub hangs a cancel notice too, should a later go-sdk send one. It counts the hung requests that are
// still open and how many tools/call requests it has received.
type hungStub struct {
	url      string
	open     atomic.Int32 // hung requests currently open on the server
	calls    atomic.Int32 // tools/call requests that reached the server
	released chan struct{}
}

func newHungStub(t *testing.T) *hungStub {
	t.Helper()
	h := &hungStub{released: make(chan struct{})}
	mcpSrv := mcp.NewServer(&mcp.Implementation{Name: "stub", Version: "0"}, nil)
	mcp.AddTool(mcpSrv, &mcp.Tool{Name: "get_rate", Description: "stub"},
		func(context.Context, *mcp.CallToolRequest, in) (*mcp.CallToolResult, out, error) {
			return nil, good(), nil
		})
	inner := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return mcpSrv }, nil)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// net/http notices a client hang-up (and cancels r.Context) only once the body has been read.
		body, _ := io.ReadAll(r.Body)
		isCall := bytes.Contains(body, []byte(`"method":"tools/call"`))
		isCancel := bytes.Contains(body, []byte(`"method":"notifications/cancelled"`))
		hang := r.Method == http.MethodDelete || (r.Method == http.MethodPost && (isCall || isCancel))
		if !hang {
			r.Body = io.NopCloser(bytes.NewReader(body))
			inner.ServeHTTP(w, r)
			return
		}
		if isCall {
			h.calls.Add(1)
		}
		h.open.Add(1)
		defer h.open.Add(-1)
		select {
		case <-r.Context().Done():
		case <-h.released:
		case <-time.After(6 * time.Second):
		}
	}))
	t.Cleanup(srv.Close)
	t.Cleanup(func() { close(h.released) }) // runs before srv.Close, which waits for the handlers
	h.url = srv.URL
	return h
}

// A source that completes the handshake and then stops answering must not hold Rate past the caller's
// deadline, nor keep a request open after Rate has given up. The session-closing DELETE that follows an
// expired call carries no deadline of its own, so the transport bounds every request by the caller's
// deadline.
func TestHungSourceIsBoundedByTheCallersDeadline(t *testing.T) {
	h := newHungStub(t)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	start := time.Now()
	_, err := mcprates.New(h.url).Rate(ctx, domain.CurrencyEUR, day)
	took := time.Since(start)
	if !errors.Is(err, domain.ErrRateUnavailable) {
		t.Fatalf("err = %v, want ErrRateUnavailable", err)
	}
	if !strings.Contains(err.Error(), ": call: ") {
		t.Fatalf("err = %v, want the failure on the call step (the handshake must have completed)", err)
	}
	if took > time.Second {
		t.Errorf("Rate returned after %v with a 100ms deadline, want it bounded by the deadline", took)
	}
	if h.calls.Load() == 0 {
		t.Fatalf("the tools/call never reached the server, the test exercises nothing")
	}

	// The go-sdk sends its follow-up requests (cancellation, session close) off Rate's return path, so
	// watch the server for a while instead of looking once: none of them may be open 300ms after Rate
	// returned, nor at any moment up to 800ms.
	time.Sleep(300 * time.Millisecond)
	for watch := time.Now(); time.Since(watch) < 500*time.Millisecond; time.Sleep(10 * time.Millisecond) {
		if n := h.open.Load(); n > 0 {
			t.Fatalf("%d hung request(s) still open on the server %v after Rate returned, want them cancelled", n, time.Since(watch)+300*time.Millisecond)
		}
	}
}

// With no deadline from the caller the lookup is still bounded: by fallbackBudget (3 s, the app's rate timeout).
func TestHungSourceIsBoundedWithoutACallerDeadline(t *testing.T) {
	h := newHungStub(t)

	start := time.Now()
	_, err := mcprates.New(h.url).Rate(context.Background(), domain.CurrencyEUR, day)
	took := time.Since(start)
	if !errors.Is(err, domain.ErrRateUnavailable) {
		t.Fatalf("err = %v, want ErrRateUnavailable", err)
	}
	if h.calls.Load() == 0 {
		t.Fatalf("the tools/call never reached the server, the test exercises nothing")
	}
	if took > 4500*time.Millisecond {
		t.Errorf("Rate returned after %v with no caller deadline, want about 3s", took)
	}
}

func TestDisabled(t *testing.T) {
	_, err := mcprates.Disabled{}.Rate(context.Background(), domain.CurrencyEUR, day)
	if !errors.Is(err, domain.ErrRateUnavailable) {
		t.Fatalf("err = %v, want ErrRateUnavailable", err)
	}
}
