// Package mcprates reads official rates through the nbu-rates MCP server.
package mcprates

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/AndriiOpryshko/spilno/services/core/internal/finance/domain"
)

const (
	toolName     = "get_rate"
	source       = "NBU"
	dayLayout    = "2006-01-02"
	notPublished = "rate_not_published"
)

// fallbackBudget bounds the HTTP requests of a lookup whose caller set no deadline. It matches
// app.RateTimeout (3 s), to which it is tied by hand because an adapter must not import app.
const fallbackBudget = 3 * time.Second

// Provider is a ports.RateProvider backed by an MCP server.
type Provider struct {
	endpoint string
	safe     string // scheme://host/path, the only form of the endpoint an error carries
}

// New returns a provider for the MCP endpoint URL.
func New(endpoint string) *Provider {
	return &Provider{endpoint: endpoint, safe: safeForm(endpoint)}
}

// httpClient returns the client for one lookup. It refuses redirects, which would carry the request to an
// address nobody configured, and bounds every request by the caller's deadline: the go-sdk sends some
// requests (the DELETE of session.Close, which follows an expired call too) on a context of its own, which
// would otherwise wait on a source that stopped answering for as long as it likes.
func httpClient(ctx context.Context, status *atomic.Int32) *http.Client {
	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(fallbackBudget)
	}
	return &http.Client{
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		Transport:     deadlineTransport{base: http.DefaultTransport, deadline: deadline, status: status},
	}
}

// deadlineTransport gives every request it carries the same absolute deadline, and records the status of
// the last answer that was not a success, so a failed lookup can name it without quoting the go-sdk's text.
type deadlineTransport struct {
	base     http.RoundTripper
	deadline time.Time
	status   *atomic.Int32
}

func (t deadlineTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx, cancel := context.WithDeadline(req.Context(), t.deadline)
	resp, err := t.base.RoundTrip(req.WithContext(ctx))
	if err != nil {
		cancel()
		return nil, err
	}
	if resp.StatusCode >= http.StatusMultipleChoices {
		t.status.Store(int32(resp.StatusCode))
	}
	resp.Body = cancelOnClose{ReadCloser: resp.Body, cancel: cancel}
	return resp, nil
}

// cancelOnClose releases the request's deadline context once the response body is closed.
type cancelOnClose struct {
	io.ReadCloser
	cancel context.CancelFunc
}

func (b cancelOnClose) Close() error {
	err := b.ReadCloser.Close()
	b.cancel()
	return err
}

// safeForm returns endpoint cut to scheme, host and path, or "<unparsable>".
func safeForm(endpoint string) string {
	u, err := url.Parse(endpoint)
	if err != nil {
		return "<unparsable>"
	}
	return (&url.URL{Scheme: u.Scheme, Host: u.Host, Path: u.Path}).String()
}

// Rate asks the server for the official rate of currency on day.
func (p *Provider) Rate(ctx context.Context, currency domain.Currency, day time.Time) (domain.Rate, error) {
	date := day.Format(dayLayout)
	var status atomic.Int32
	client := mcp.NewClient(&mcp.Implementation{Name: "spilno-core", Version: "1"}, nil)
	session, err := client.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: p.endpoint, HTTPClient: httpClient(ctx, &status), DisableStandaloneSSE: true}, nil)
	if err != nil {
		return domain.Rate{}, p.unavailable("connect", err, int(status.Load()))
	}
	// Close may wait on a server that stopped answering; never make the caller wait for it.
	defer func() { go func() { _ = session.Close() }() }()

	res, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      toolName,
		Arguments: map[string]any{"currency": string(currency), "date": date},
	})
	if err != nil {
		return domain.Rate{}, p.unavailable("call", err, int(status.Load()))
	}
	if res.IsError {
		text := toolText(res)
		if strings.HasPrefix(text, notPublished) {
			return domain.Rate{}, fmt.Errorf("%w: %s", domain.ErrRateNotPublished, text)
		}
		return domain.Rate{}, fmt.Errorf("%w: tool error: %s", domain.ErrRateUnavailable, text)
	}
	sc, ok := res.StructuredContent.(map[string]any)
	if !ok {
		return domain.Rate{}, fmt.Errorf("%w: no structured content", domain.ErrRateUnavailable)
	}
	str := func(k string) string { s, _ := sc[k].(string); return s }
	if str("currency") != string(currency) || str("date") != date || str("source") != source {
		return domain.Rate{}, fmt.Errorf("%w: answer does not match the question", domain.ErrRateUnavailable)
	}
	return domain.ParseRate(currency, day, str("rate"))
}

func toolText(res *mcp.CallToolResult) string {
	for _, c := range res.Content {
		if t, ok := c.(*mcp.TextContent); ok {
			return t.Text
		}
	}
	return ""
}

// unavailable wraps domain.ErrRateUnavailable with the step, the safe endpoint and the class of err. The text
// of err is dropped: net/http and the go-sdk print the request URL in many spellings (quoted, with the
// userinfo masked or not, a redirect target), and none of them can be redacted reliably.
func (p *Provider) unavailable(step string, err error, status int) error {
	return fmt.Errorf("%w: %s: endpoint=%s: %s", domain.ErrRateUnavailable, step, p.safe, failureClass(err, status))
}

// failureClass names what went wrong in a fixed word or two: the HTTP status of the last answer that was not
// a success, else the kind of transport failure.
func failureClass(err error, status int) string {
	var netErr net.Error
	switch {
	case status != 0:
		return fmt.Sprintf("http %d", status)
	case errors.Is(err, context.DeadlineExceeded), errors.As(err, &netErr) && netErr.Timeout():
		return "timeout"
	case errors.Is(err, context.Canceled):
		return "canceled"
	case errors.Is(err, syscall.ECONNREFUSED):
		return "connection refused"
	case errors.Is(err, io.EOF), errors.Is(err, io.ErrUnexpectedEOF):
		return "connection closed"
	case errors.As(err, &netErr):
		return "network error"
	default:
		return "protocol error"
	}
}

// Check reports whether endpoint is a usable address, and returns its safe form (scheme, host and path, or
// "<unparsable>"). An address that does not parse, or that holds a character %q escapes (a quote or a
// backslash, never part of a well-formed URL), is a misconfiguration and is refused at start-up. The error
// names the reason only: url.Parse's own error is not wrapped because it quotes the raw input.
func Check(endpoint string) (safe string, err error) {
	safe = safeForm(endpoint)
	if _, perr := url.Parse(endpoint); perr != nil {
		return safe, errors.New("endpoint is not a valid URL")
	}
	if strconv.Quote(endpoint) != `"`+endpoint+`"` {
		return safe, errors.New("endpoint has a character that net/http prints escaped")
	}
	return safe, nil
}

// Disabled answers domain.ErrRateUnavailable to every lookup.
type Disabled struct{}

// Rate always answers domain.ErrRateUnavailable.
func (Disabled) Rate(context.Context, domain.Currency, time.Time) (domain.Rate, error) {
	return domain.Rate{}, fmt.Errorf("%w: rate source not configured", domain.ErrRateUnavailable)
}
