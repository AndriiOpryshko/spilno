package server_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/AndriiOpryshko/spilno/mcp/nbu-rates/internal/nbu"
	"github.com/AndriiOpryshko/spilno/mcp/nbu-rates/internal/server"
)

// fakeSource answers like the NBU client would.
type fakeSource struct {
	rate string
	err  error
}

func (f fakeSource) Rate(_ context.Context, currency string, date time.Time) (nbu.Rate, error) {
	if currency == "UAH" || len(currency) != 3 {
		return nbu.Rate{}, nbu.ErrInvalidCurrency
	}
	if f.err != nil {
		return nbu.Rate{}, f.err
	}
	return nbu.Rate{Currency: currency, Date: date, Value: f.rate}, nil
}

// connect runs the server on an in-memory transport and returns a client session.
func connect(t *testing.T, src server.RateSource) *mcp.ClientSession {
	t.Helper()
	ct, st := mcp.NewInMemoryTransports()
	ctx := context.Background()
	ss, err := server.New(src).Connect(ctx, st, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ss.Close() })
	cs, err := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "0"}, nil).Connect(ctx, ct, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cs.Close() })
	return cs
}

func text(t *testing.T, res *mcp.CallToolResult) string {
	t.Helper()
	if len(res.Content) == 0 {
		t.Fatal("result has no content")
	}
	tc, ok := res.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("content is %T, want text", res.Content[0])
	}
	return tc.Text
}

func TestServerListsItsThreePrimitives(t *testing.T) {
	cs := connect(t, fakeSource{rate: "50.5311"})
	ctx := context.Background()

	tools, err := cs.ListTools(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(tools.Tools) != 1 || tools.Tools[0].Name != server.ToolGetRate {
		t.Fatalf("tools = %+v, want only get_rate", tools.Tools)
	}
	tool := tools.Tools[0]
	if tool.Annotations == nil || !tool.Annotations.ReadOnlyHint {
		t.Error("get_rate is not marked read-only")
	}
	if tool.OutputSchema == nil {
		t.Error("get_rate has no output schema")
	}

	resources, err := cs.ListResources(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(resources.Resources) != 1 || resources.Resources[0].URI != server.AboutURI {
		t.Errorf("resources = %+v, want only %s", resources.Resources, server.AboutURI)
	}

	prompts, err := cs.ListPrompts(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(prompts.Prompts) != 1 || prompts.Prompts[0].Name != server.PromptName {
		t.Errorf("prompts = %+v, want only %s", prompts.Prompts, server.PromptName)
	}
}

func TestGetRateReturnsStructuredOutput(t *testing.T) {
	cs := connect(t, fakeSource{rate: "50.5311"})
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      server.ToolGetRate,
		Arguments: map[string]any{"currency": " eur ", "date": "2026-10-20"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		t.Fatalf("tool error: %s", text(t, res))
	}
	raw, err := json.Marshal(res.StructuredContent)
	if err != nil {
		t.Fatal(err)
	}
	var out server.GetRateOutput
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatal(err)
	}
	want := server.GetRateOutput{Currency: "EUR", Date: "2026-10-20", Rate: "50.5311", Source: "NBU"}
	if out != want {
		t.Errorf("output = %+v, want %+v", out, want)
	}
	// A client without structured-content support still gets the same data as text.
	if !strings.Contains(text(t, res), `"rate":"50.5311"`) {
		t.Errorf("text content = %s", text(t, res))
	}
}

func TestGetRateFailsWithAStableCode(t *testing.T) {
	tests := []struct {
		name string
		src  fakeSource
		args map[string]any
		code string
	}{
		{"malformed date", fakeSource{rate: "1"}, map[string]any{"currency": "EUR", "date": "20.10.2026"}, server.CodeInvalidArgument},
		{"hryvnia", fakeSource{rate: "1"}, map[string]any{"currency": "UAH", "date": "2026-10-20"}, server.CodeInvalidArgument},
		{"not published", fakeSource{err: nbu.ErrNotPublished}, map[string]any{"currency": "EUR", "date": "2031-01-01"}, server.CodeRateNotPublished},
		{"upstream down", fakeSource{err: fmt.Errorf("%w: dial tcp 10.0.0.7:443", nbu.ErrUnavailable)}, map[string]any{"currency": "EUR", "date": "2026-10-20"}, server.CodeUpstreamUnavailable},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := connect(t, tt.src)
			res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: server.ToolGetRate, Arguments: tt.args})
			if err != nil {
				t.Fatalf("protocol error %v, want a tool error", err)
			}
			if !res.IsError {
				t.Fatal("IsError = false, want a tool error")
			}
			got := text(t, res)
			if !strings.HasPrefix(got, tt.code+": ") {
				t.Errorf("error text = %q, want prefix %q", got, tt.code+": ")
			}
			if strings.Contains(got, "10.0.0.7") {
				t.Errorf("error text leaks the upstream detail: %q", got)
			}
		})
	}
}

func TestGetRateRejectsArgumentsOutsideTheSchema(t *testing.T) {
	cs := connect(t, fakeSource{rate: "1"})
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      server.ToolGetRate,
		Arguments: map[string]any{"currency": "EUR"},
	})
	if err == nil && !res.IsError {
		t.Fatal("a call without date succeeded")
	}
}

func TestAboutResourceDocumentsTheErrorCodes(t *testing.T) {
	cs := connect(t, fakeSource{})
	res, err := cs.ReadResource(context.Background(), &mcp.ReadResourceParams{URI: server.AboutURI})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Contents) != 1 || res.Contents[0].MIMEType != "text/markdown" {
		t.Fatalf("contents = %+v", res.Contents)
	}
	for _, code := range []string{server.CodeInvalidArgument, server.CodeRateNotPublished, server.CodeUpstreamUnavailable} {
		if !strings.Contains(res.Contents[0].Text, code) {
			t.Errorf("about does not mention %s", code)
		}
	}
}

func TestPromptNamesTheToolAndItsArguments(t *testing.T) {
	cs := connect(t, fakeSource{})
	res, err := cs.GetPrompt(context.Background(), &mcp.GetPromptParams{
		Name:      server.PromptName,
		Arguments: map[string]string{"amount": "1250.00", "currency": "EUR", "date": "2026-10-20"},
	})
	if err != nil {
		t.Fatal(err)
	}
	got := res.Messages[0].Content.(*mcp.TextContent).Text
	for _, want := range []string{"1250.00 EUR", "2026-10-20", server.ToolGetRate} {
		if !strings.Contains(got, want) {
			t.Errorf("prompt %q lacks %q", got, want)
		}
	}
	if _, err := cs.GetPrompt(context.Background(), &mcp.GetPromptParams{Name: server.PromptName, Arguments: map[string]string{"amount": "1"}}); err == nil {
		t.Error("a prompt without currency and date succeeded")
	}
}

// The transport the core API uses: Streamable HTTP, stateless, against a fake NBU.
func TestGetRateOverStreamableHTTP(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`[{"rate":45.051,"cc":"USD"}]`))
	}))
	defer upstream.Close()
	srv := server.New(nbu.NewClient(upstream.URL))
	hs := httptest.NewServer(mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return srv },
		&mcp.StreamableHTTPOptions{Stateless: true, JSONResponse: true}))
	defer hs.Close()

	ctx := context.Background()
	cs, err := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "0"}, nil).
		Connect(ctx, &mcp.StreamableClientTransport{Endpoint: hs.URL, DisableStandaloneSSE: true}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cs.Close() }()
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: server.ToolGetRate, Arguments: map[string]any{"currency": "USD", "date": "2026-10-24"}})
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError || !strings.Contains(text(t, res), `"rate":"45.051"`) {
		t.Errorf("result = %s (isError=%v)", text(t, res), res.IsError)
	}
}
