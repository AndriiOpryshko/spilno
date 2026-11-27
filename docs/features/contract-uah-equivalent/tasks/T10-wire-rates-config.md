---
id: T10
title: "Wire NBU_RATES_MCP_URL from the environment to the finance module"
layer: "wiring"
deps: ["T09"]
acs: ["AC-10", "AC-04"]
files_hint:
  - "services/core/internal/platform/config/config.go"
  - "services/core/internal/platform/config/config_test.go"
  - "services/core/cmd/api/main.go"
  - "services/core/internal/finance/module.go"
  - "mise.toml"
owner: "Andrii Opryshko"
estimate: "S"
status: "done"
---

# T10 — Wire NBU_RATES_MCP_URL from the environment to the finance module

## Why

[SAD](../sad.md) §7; [ADR-0001](../adr/0001-read-the-nbu-rate-through-an-own-mcp-server-behind-a-finance-port.md) «Migration».

**Satisfies:** [PRD §5](../PRD.md) AC-10, AC-04 · **Blocked by:** T09

## What

- `config.Config.NBURatesMCPURL` from `NBU_RATES_MCP_URL` (no default in code).
- `finance.Config{RatesMCPURL}` → `mcprates.New(url)` or `mcprates.Disabled{}`; log once at start which one is active.
- `mise.toml` `[env]`: `NBU_RATES_MCP_URL = "http://127.0.0.1:3336/mcp"`.

## Definition of Done

- [ ] Acceptance: A config test reads `NBU_RATES_MCP_URL`; `cmd/api` builds the finance module with the MCP provider when it is set and the disabled one when it is not; the API starts in both cases; `mise.toml` sets the local default.
- [ ] unit: config test for the variable set and unset
- [ ] integration: `finance.New(pool, finance.Config{})` serves contracts and answers 503 `finance.rate_unavailable` on the new operation
- [ ] `mise run lint` and `mise run test` green
