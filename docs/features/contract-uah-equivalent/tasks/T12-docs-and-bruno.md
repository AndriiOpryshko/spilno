---
id: T12
title: "Document the feature: Bruno request, environment variable, CLAUDE.md gotcha"
layer: "docs"
deps: ["T09"]
acs: ["AC-05"]
files_hint:
  - "services/core/api/bruno/finance/"
  - "README.md"
  - "CLAUDE.md"
  - "mcp/nbu-rates/README.md"
owner: "Andrii Opryshko"
estimate: "S"
status: "todo"
---

# T12 — Document the feature: Bruno request, environment variable, CLAUDE.md gotcha

## Why

The next session and the next person must learn about the new process without reading the SAD.

**Satisfies:** [PRD §5](../PRD.md) AC-05 · **Blocked by:** T09

## What

- Bruno: `Resolve UAH equivalent.bru` in `services/core/api/bruno/finance/`, in the style of its neighbours.
- README: the variable and the new pane of `mise run dev`.
- CLAUDE.md «Gotchas»: one bullet (rate source optional; 503 `finance.rate_unavailable` when it is down; tests never call the NBU).

## Definition of Done

- [ ] Acceptance: A Bruno request for the new operation sits next to the other finance requests; README and CLAUDE.md name `NBU_RATES_MCP_URL`, `mise run mcp:nbu` and what happens when the server is down; the root `.mcp.json` question is noted for the PR that owns that file.
- [ ] none (docs); `mise run lint` stays green
- [ ] `mise run lint` and `mise run test` green
