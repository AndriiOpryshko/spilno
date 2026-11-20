---
id: T05
title: "Read the official rate through the nbu-rates MCP server behind a finance port"
layer: "infra"
deps: ["T02"]
acs: ["AC-10"]
files_hint:
  - "services/core/internal/finance/ports/rates.go"
  - "services/core/internal/finance/adapters/mcprates/"
  - "services/core/go.mod"
  - "services/core/go.sum"
  - "services/core/.golangci.yml"
owner: "Andrii Opryshko"
estimate: "L"
status: "done"
---

# T05 — Read the official rate through the nbu-rates MCP server behind a finance port

## Why

[ADR-0001](../adr/0001-read-the-nbu-rate-through-an-own-mcp-server-behind-a-finance-port.md); tool contract in [`mcp/nbu-rates/README.md`](../../../../mcp/nbu-rates/README.md) and `mcp/nbu-rates/internal/server/server.go`.

**Satisfies:** [PRD §5](../PRD.md) AC-10 · **Blocked by:** T02

## What

- `ports.RateProvider { Rate(ctx, currency domain.Currency, day time.Time) (domain.Rate, error) }`.
- `mcprates.New(endpoint)` on `github.com/modelcontextprotocol/go-sdk` (same version as `mcp/nbu-rates/go.mod`): per lookup connect with `StreamableClientTransport{DisableStandaloneSSE: true}`, `CallTool(get_rate, {currency, date})`, close.
- Read `StructuredContent` (`currency`, `date`, `rate`, `source`); verify it against the question; `domain.ParseRate`.
- Error text prefix `rate_not_published` → `ErrRateNotPublished`; everything else → `ErrRateUnavailable` (wrap the cause for the log).
- `mcprates.Disabled{}` for an empty URL.
- The adapter sets no timeout of its own; it honours `ctx`.

## Definition of Done

- [ ] Acceptance: An adapter test against an in-process MCP server proves: a good `get_rate` answer becomes a `domain.Rate`; a `rate_not_published` tool error becomes `ErrRateNotPublished`; any other tool error, a refused connection, a wrong currency or date in the answer, a wrong source or a malformed rate become `ErrRateUnavailable`; the disabled provider always answers `ErrRateUnavailable`.
- [ ] integration: stub MCP server (go-sdk, `httptest`) with a scripted `get_rate`
- [ ] integration: a context that expires mid-call returns `ErrRateUnavailable` promptly
- [ ] `mise run lint` and `mise run test` green
