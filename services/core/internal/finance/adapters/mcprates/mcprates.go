// Package mcprates reads official rates through the nbu-rates MCP server.
package mcprates

import (
	"context"
	"fmt"
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
type Provider struct{ endpoint string }

// New returns a provider for the MCP endpoint URL.
func New(endpoint string) *Provider { return &Provider{endpoint: endpoint} }

// Rate asks the server for the official rate of currency on day.
func (p *Provider) Rate(ctx context.Context, currency domain.Currency, day time.Time) (domain.Rate, error) {
	date := day.Format(dayLayout)
	client := mcp.NewClient(&mcp.Implementation{Name: "spilno-core", Version: "1"}, nil)
	session, err := client.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: p.endpoint, DisableStandaloneSSE: true}, nil)
	if err != nil {
		return domain.Rate{}, unavailable("connect", err)
	}
	// Close may wait on a server that stopped answering; never make the caller wait for it.
	defer func() { go func() { _ = session.Close() }() }()

	res, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      toolName,
		Arguments: map[string]any{"currency": string(currency), "date": date},
	})
	if err != nil {
		return domain.Rate{}, unavailable("call", err)
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

func unavailable(step string, err error) error {
	return fmt.Errorf("%w: %s: %v", domain.ErrRateUnavailable, step, err)
}

// Disabled answers domain.ErrRateUnavailable to every lookup.
type Disabled struct{}

// Rate always answers domain.ErrRateUnavailable.
func (Disabled) Rate(context.Context, domain.Currency, time.Time) (domain.Rate, error) {
	return domain.Rate{}, fmt.Errorf("%w: rate source not configured", domain.ErrRateUnavailable)
}
