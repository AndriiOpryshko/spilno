// Package mcprates reads official rates through the nbu-rates MCP server.
package mcprates

import (
	"context"
	"fmt"
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

// Provider is a ports.RateProvider backed by an MCP server.
type Provider struct {
	endpoint string
	safe     string   // scheme://host/path, fit for errors and logs
	leaks    []string // spellings of the endpoint that must never reach an error
}

// New returns a provider for the MCP endpoint URL.
func New(endpoint string) *Provider {
	safe, leaks := addressForms(endpoint)
	return &Provider{endpoint: endpoint, safe: safe, leaks: leaks}
}

// addressForms returns the form of endpoint that is safe for errors and logs (scheme, host and path only,
// or "<unparsable>") and every spelling of the endpoint that net/http may print in a *url.Error: the raw
// string, the parsed url.String() and, when a password is set, net/http's masked form (user:***@).
func addressForms(endpoint string) (safe string, leaks []string) {
	safe = "<unparsable>"
	candidates := []string{endpoint}
	if u, err := url.Parse(endpoint); err == nil {
		safe = (&url.URL{Scheme: u.Scheme, Host: u.Host, Path: u.Path}).String()
		candidates = append(candidates, u.String())
		if u.User != nil {
			if _, has := u.User.Password(); has {
				candidates = append(candidates,
					strings.Replace(u.String(), u.User.String()+"@", u.User.Username()+":***@", 1))
			}
		}
	}
	for _, c := range candidates {
		if c != "" && !slices.Contains(leaks, c) {
			leaks = append(leaks, c)
		}
	}
	return safe, leaks
}

// Rate asks the server for the official rate of currency on day.
func (p *Provider) Rate(ctx context.Context, currency domain.Currency, day time.Time) (domain.Rate, error) {
	date := day.Format(dayLayout)
	client := mcp.NewClient(&mcp.Implementation{Name: "spilno-core", Version: "1"}, nil)
	session, err := client.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: p.endpoint, DisableStandaloneSSE: true}, nil)
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

// unavailable wraps err with the step and the log-safe address, with the raw endpoint scrubbed out.
func (p *Provider) unavailable(step string, err error) error {
	msg := err.Error()
	for _, l := range p.leaks {
		msg = strings.ReplaceAll(msg, l, p.safe)
	}
	return fmt.Errorf("%w: %s: endpoint=%s: %s", domain.ErrRateUnavailable, step, p.safe, msg)
}

// Disabled answers domain.ErrRateUnavailable to every lookup.
type Disabled struct{}

// Rate always answers domain.ErrRateUnavailable.
func (Disabled) Rate(context.Context, domain.Currency, time.Time) (domain.Rate, error) {
	return domain.Rate{}, fmt.Errorf("%w: rate source not configured", domain.ErrRateUnavailable)
}
