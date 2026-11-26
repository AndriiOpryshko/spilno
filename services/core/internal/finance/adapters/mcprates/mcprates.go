// Package mcprates reads official rates through the nbu-rates MCP server.
package mcprates

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"
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
	safe     string   // scheme://host/path, fit for errors and logs
	spelled  []string // whole spellings of the endpoint, replaced by safe in an error
	secrets  []string // secret parts of the endpoint (userinfo, query, fragment), masked in an error
}

// New returns a provider for the MCP endpoint URL.
func New(endpoint string) *Provider {
	safe, spelled, secrets := addressForms(endpoint)
	return &Provider{endpoint: endpoint, safe: safe, spelled: spelled, secrets: secrets}
}

// httpClient returns the client for one lookup. It refuses redirects, which would carry the request to an
// address nobody configured, and bounds every request by the caller's deadline: the go-sdk sends some
// requests (the notifications/cancelled that follows an expired call, the DELETE on close) on a context of
// its own, which would otherwise wait on a source that stopped answering for as long as it likes.
func httpClient(ctx context.Context) *http.Client {
	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(fallbackBudget)
	}
	return &http.Client{
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		Transport:     deadlineTransport{base: http.DefaultTransport, deadline: deadline},
	}
}

// deadlineTransport gives every request it carries the same absolute deadline.
type deadlineTransport struct {
	base     http.RoundTripper
	deadline time.Time
}

func (t deadlineTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx, cancel := context.WithDeadline(req.Context(), t.deadline)
	resp, err := t.base.RoundTrip(req.WithContext(ctx))
	if err != nil {
		cancel()
		return nil, err
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

// addressForms returns the form of endpoint that is safe for errors and logs (scheme, host and path only,
// or "<unparsable>"), every spelling of the endpoint that net/http may print in a *url.Error (the raw
// string, the parsed url.String() and, when a password is set, net/http's masked form user:***@), and
// its secret parts on their own, so they are caught inside an address nobody configured too.
func addressForms(endpoint string) (safe string, spelled, secrets []string) {
	safe = "<unparsable>"
	spellings := []string{endpoint}
	if u, err := url.Parse(endpoint); err == nil {
		safe = (&url.URL{Scheme: u.Scheme, Host: u.Host, Path: u.Path}).String()
		spellings = append(spellings, u.String())
		if u.User != nil {
			secrets = append(secrets, u.User.Username())
			if pw, has := u.User.Password(); has {
				spellings = append(spellings, strings.Replace(u.String(), u.User.String()+"@", u.User.Username()+":***@", 1))
				secrets = append(secrets, pw, url.QueryEscape(pw), url.PathEscape(pw))
			}
		}
		secrets = append(secrets, u.RawQuery, u.Fragment, u.EscapedFragment())
		for _, vs := range u.Query() {
			for _, v := range vs {
				secrets = append(secrets, v, url.QueryEscape(v))
			}
		}
	}
	return safe, uniqueLongestFirst(spellings), uniqueLongestFirst(secrets)
}

// uniqueLongestFirst drops empty and repeated strings and orders the rest longest first.
func uniqueLongestFirst(in []string) []string {
	var out []string
	for _, c := range in {
		if c != "" && !slices.Contains(out, c) {
			out = append(out, c)
		}
	}
	slices.SortStableFunc(out, func(a, b string) int { return len(b) - len(a) })
	return out
}

// Rate asks the server for the official rate of currency on day.
func (p *Provider) Rate(ctx context.Context, currency domain.Currency, day time.Time) (domain.Rate, error) {
	date := day.Format(dayLayout)
	client := mcp.NewClient(&mcp.Implementation{Name: "spilno-core", Version: "1"}, nil)
	session, err := client.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: p.endpoint, HTTPClient: httpClient(ctx), DisableStandaloneSSE: true}, nil)
	if err != nil {
		return domain.Rate{}, p.unavailable("connect", err)
	}
	// Close may wait on a server that stopped answering; never make the caller wait for it.
	defer func() { go func() { _ = session.Close() }() }()

	res, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      toolName,
		Arguments: map[string]any{"currency": string(currency), "date": date},
	})
	if err != nil {
		return domain.Rate{}, p.unavailable("call", err)
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

// unavailable wraps err with the step and the log-safe address, with the raw endpoint scrubbed out: a
// whole spelling of it becomes the safe form, and a secret part left anywhere else (an address net/http
// reports after a redirect, say) becomes ***.
func (p *Provider) unavailable(step string, err error) error {
	// The placeholder keeps the safe form out of the secret masking, which could otherwise hit its host.
	const placeholder = "\x00endpoint\x00"
	msg := err.Error()
	for _, s := range p.spelled {
		msg = strings.ReplaceAll(msg, s, placeholder)
	}
	for _, s := range p.secrets {
		msg = strings.ReplaceAll(msg, s, "***")
	}
	msg = strings.ReplaceAll(msg, placeholder, p.safe)
	return fmt.Errorf("%w: %s: endpoint=%s: %s", domain.ErrRateUnavailable, step, p.safe, msg)
}

// Disabled answers domain.ErrRateUnavailable to every lookup.
type Disabled struct{}

// Rate always answers domain.ErrRateUnavailable.
func (Disabled) Rate(context.Context, domain.Currency, time.Time) (domain.Rate, error) {
	return domain.Rate{}, fmt.Errorf("%w: rate source not configured", domain.ErrRateUnavailable)
}
