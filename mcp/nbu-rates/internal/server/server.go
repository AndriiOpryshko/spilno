// Package server exposes the NBU official exchange rates as an MCP server:
// one tool, one resource and one prompt.
package server

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/AndriiOpryshko/spilno/mcp/nbu-rates/internal/nbu"
)

const (
	Name    = "nbu-rates"
	Version = "0.1.0"

	ToolGetRate = "get_rate"
	AboutURI    = "nbu://about"
	PromptName  = "uah_equivalent"
)

// Error codes a failed get_rate call starts its text with ("<code>: <detail>"),
// so that a program can tell them apart without parsing prose.
const (
	CodeInvalidArgument     = "invalid_argument"
	CodeRateNotPublished    = "rate_not_published"
	CodeUpstreamUnavailable = "upstream_unavailable"
)

// RateSource is what the server needs from the NBU client.
type RateSource interface {
	Rate(ctx context.Context, currency string, date time.Time) (nbu.Rate, error)
}

type GetRateInput struct {
	Currency string `json:"currency" jsonschema:"ISO 4217 code of the foreign currency, e.g. EUR or USD (not UAH)"`
	Date     string `json:"date" jsonschema:"calendar day the rate is for, YYYY-MM-DD"`
}

type GetRateOutput struct {
	Currency string `json:"currency" jsonschema:"ISO 4217 code the rate is for"`
	Date     string `json:"date" jsonschema:"calendar day the rate is for, YYYY-MM-DD"`
	Rate     string `json:"rate" jsonschema:"hryvnias per one unit of the currency, a decimal string exactly as published, e.g. 50.5311"`
	Source   string `json:"source" jsonschema:"who published the rate; always NBU"`
}

//go:embed about.md
var about string

func New(src RateSource) *mcp.Server {
	s := mcp.NewServer(&mcp.Implementation{Name: Name, Version: Version}, nil)

	notDestructive := false
	mcp.AddTool(s, &mcp.Tool{
		Name:  ToolGetRate,
		Title: "Official NBU exchange rate",
		Description: "Returns the official exchange rate of the National Bank of Ukraine for one foreign currency " +
			"on one calendar day: hryvnias per one unit of the currency. Read-only. On failure the result is an " +
			"error whose text starts with invalid_argument, rate_not_published or upstream_unavailable.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true, DestructiveHint: &notDestructive},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in GetRateInput) (*mcp.CallToolResult, GetRateOutput, error) {
		out, err := getRate(ctx, src, in)
		return nil, out, err
	})

	s.AddResource(&mcp.Resource{
		URI:         AboutURI,
		Name:        "about",
		Title:       "About the NBU official rates",
		Description: "What the rate means, when it is published, how to turn an amount into hryvnias, and the error codes of get_rate.",
		MIMEType:    "text/markdown",
	}, func(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		return &mcp.ReadResourceResult{Contents: []*mcp.ResourceContents{{URI: req.Params.URI, MIMEType: "text/markdown", Text: about}}}, nil
	})

	s.AddPrompt(&mcp.Prompt{
		Name:        PromptName,
		Title:       "Hryvnia equivalent of an amount",
		Description: "Asks for the hryvnia equivalent of a foreign-currency amount at the official NBU rate of a given day.",
		Arguments: []*mcp.PromptArgument{
			{Name: "amount", Description: "amount in the foreign currency, e.g. 1250.00", Required: true},
			{Name: "currency", Description: "ISO 4217 code, e.g. EUR", Required: true},
			{Name: "date", Description: "day of the rate, YYYY-MM-DD", Required: true},
		},
	}, func(_ context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		a := req.Params.Arguments
		for _, k := range []string{"amount", "currency", "date"} {
			if strings.TrimSpace(a[k]) == "" {
				return nil, fmt.Errorf("argument %q is required", k)
			}
		}
		text := fmt.Sprintf("Find the hryvnia equivalent of %s %s at the official NBU rate on %s. "+
			"Call the get_rate tool for that currency and day, multiply the amount by the rate and round half up to the kopiyka. "+
			"Answer with the rate, its date, the source and the result in hryvnias. If get_rate fails, report its error code and do not guess a rate.",
			a["amount"], a["currency"], a["date"])
		return &mcp.GetPromptResult{
			Description: "Hryvnia equivalent at the official NBU rate",
			Messages:    []*mcp.PromptMessage{{Role: "user", Content: &mcp.TextContent{Text: text}}},
		}, nil
	})
	return s
}

func getRate(ctx context.Context, src RateSource, in GetRateInput) (GetRateOutput, error) {
	day, err := time.Parse(time.DateOnly, in.Date)
	if err != nil {
		return GetRateOutput{}, fmt.Errorf("%s: date must be YYYY-MM-DD, got %q", CodeInvalidArgument, in.Date)
	}
	r, err := src.Rate(ctx, strings.ToUpper(strings.TrimSpace(in.Currency)), day)
	switch {
	case errors.Is(err, nbu.ErrInvalidCurrency):
		return GetRateOutput{}, fmt.Errorf("%s: currency must be a three-letter ISO 4217 code other than UAH, got %q", CodeInvalidArgument, in.Currency)
	case errors.Is(err, nbu.ErrNotPublished):
		return GetRateOutput{}, fmt.Errorf("%s: the NBU has no official rate of %s for %s", CodeRateNotPublished, in.Currency, in.Date)
	case err != nil:
		// The upstream detail stays in the server log; the caller gets a stable line.
		return GetRateOutput{}, fmt.Errorf("%s: the NBU rate service did not answer", CodeUpstreamUnavailable)
	}
	return GetRateOutput{Currency: r.Currency, Date: r.Date.Format(time.DateOnly), Rate: r.Value, Source: "NBU"}, nil
}
