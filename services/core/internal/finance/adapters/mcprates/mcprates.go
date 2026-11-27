// Package mcprates reads official rates through the nbu-rates MCP server.
package mcprates

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"strconv"
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

// minMaskedSecret is the shortest secret part of the endpoint that is masked in the text of an error outside
// URLs. A shorter value is not a credential worth hiding there, and masking it would garble ordinary words.
const minMaskedSecret = 6

// urlToken finds an address in an error text: a scheme, "://", then everything up to a quote or whitespace.
var urlToken = regexp.MustCompile(`[A-Za-z][A-Za-z0-9+.-]*://[^\s"]+`)

// fallbackBudget bounds the HTTP requests of a lookup whose caller set no deadline. It matches
// app.RateTimeout (3 s), to which it is tied by hand because an adapter must not import app.
const fallbackBudget = 3 * time.Second

// Provider is a ports.RateProvider backed by an MCP server.
type Provider struct {
	endpoint string
	safe     string   // scheme://host/path, fit for errors and logs
	spelled  []string // whole spellings of the endpoint, replaced by safe in an error
	secrets  []string // secret parts of the endpoint (≥ minMaskedSecret bytes), masked in an error outside URLs
}

// New returns a provider for the MCP endpoint URL.
func New(endpoint string) *Provider {
	safe, spelled, secrets := addressForms(endpoint)
	return &Provider{endpoint: endpoint, safe: safe, spelled: spelled, secrets: secrets}
}

// httpClient returns the client for one lookup. It refuses redirects, which would carry the request to an
// address nobody configured, and bounds every request by the caller's deadline: the go-sdk sends some
// requests (the DELETE of session.Close, which follows an expired call too) on a context of its own, which
// would otherwise wait on a source that stopped answering for as long as it likes.
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
// its secret parts of at least minMaskedSecret bytes on their own, so they are caught in the text around an
// address too (shorter ones are dropped: see minMaskedSecret, and so are those the safe form holds).
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
	// A part the safe form already holds is no secret: endpoint= prints it, and masking it garbles the host.
	secrets = slices.DeleteFunc(secrets, func(v string) bool { return len(v) < minMaskedSecret || strings.Contains(safe, v) })
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

// unavailable wraps err with the step and the log-safe address. In the message, a whole spelling of the
// configured endpoint becomes the safe form; then every URL token (the configured endpoint or any other
// address, a redirect target say) is reduced to scheme, host and path, or "<unparsable>"; and in the text
// outside URL tokens only, the configured endpoint's secret parts become ***.
func (p *Provider) unavailable(step string, err error) error {
	msg := err.Error()
	for _, s := range p.spelled {
		msg = strings.ReplaceAll(msg, s, p.safe)
	}
	var out strings.Builder
	last := 0
	for _, loc := range urlToken.FindAllStringIndex(msg, -1) {
		out.WriteString(p.maskSecrets(msg[last:loc[0]]))
		out.WriteString(redactURL(msg[loc[0]:loc[1]]))
		last = loc[1]
	}
	out.WriteString(p.maskSecrets(msg[last:]))
	return fmt.Errorf("%w: %s: endpoint=%s: %s", domain.ErrRateUnavailable, step, p.safe, out.String())
}

// maskSecrets replaces the endpoint's secret parts in text that holds no URL.
func (p *Provider) maskSecrets(text string) string {
	for _, s := range p.secrets {
		text = strings.ReplaceAll(text, s, "***")
	}
	return text
}

// redactURL reduces a URL to scheme, host and path, or "<unparsable>".
func redactURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return "<unparsable>"
	}
	return (&url.URL{Scheme: u.Scheme, Host: u.Host, Path: u.Path}).String()
}

// Check reports whether endpoint can be used without a leak, and returns its log-safe form (scheme, host and
// path, or "<unparsable>"). net/http prints the request URL in an error with %q; an endpoint that %q escapes
// (one holding a quote or a backslash, say) is printed in a spelling the redaction cannot match, and one that
// does not parse has no safe form to replace it with, so both are refused. The error names the reason only:
// url.Parse's own error is not wrapped because it quotes the raw input.
func Check(endpoint string) (safe string, err error) {
	safe, _, _ = addressForms(endpoint)
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
