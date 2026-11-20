---
id: T09
title: "Return the equivalent on every contract answer and serve the «get it later» operation"
layer: "ports"
deps: ["T04", "T08"]
acs: ["AC-01", "AC-05", "AC-06", "AC-07", "AC-08", "AC-09"]
files_hint:
  - "services/core/internal/finance/adapters/http/handler.go"
  - "services/core/internal/finance/module.go"
  - "services/core/internal/finance/module_test.go"
  - "services/core/internal/finance/guarded_test.go"
  - "services/core/internal/finance/uah_equivalent_test.go"
  - "services/core/internal/platform/isolation_test.go"
owner: "Andrii Opryshko"
estimate: "L"
status: "todo"
---

# T09 — Return the equivalent on every contract answer and serve the «get it later» operation

## Why

[contracts/openapi.yaml](../contracts/openapi.yaml); `.claude/rules/openapi.md` (problem shape, `code`).

**Satisfies:** [PRD §5](../PRD.md) AC-01, AC-05, AC-06, AC-07, AC-08, AC-09 · **Blocked by:** T04, T08

## What

- `toAPI` maps `UAHEquivalent` → `{amount_kopiyky, rate, rate_date = signed_at, source: NBU}`.
- `ResolveContractUahEquivalent` replaces the stub of T04: 200 / 404 / 409 / 503 from the registry of T03.
- `finance.New(pool, finance.Config{Rates: …})` so tests can pass a fake port; `Config{}` means disabled.
- Extend the isolation test's operation table if it lists operations explicitly.

## Definition of Done

- [ ] Acceptance: Module tests through the router on a real database, with a fake rate port, show AC-01 (create, get and list carry the same `uah_equivalent`), AC-05, AC-06 (also two concurrent calls → one stored value), every answer of AC-07 as `application/problem+json` with its `code`, and AC-08 (401, 403, 403 disabled, 404 for another association's id); the isolation test covers the new operation.
- [ ] integration: the scenarios of the DoD in `uah_equivalent_test.go`
- [ ] integration: `isolation_test.go` green with the new operation
- [ ] `mise run lint` and `mise run test` green
