---
id: T04
title: "Merge the contract delta and stub the new finance operation"
layer: "wiring"
deps: []
acs: ["AC-05", "AC-09"]
files_hint:
  - "services/core/api/openapi.yaml"
  - "services/core/internal/finance/adapters/http/gen/"
  - "services/core/internal/finance/adapters/http/handler.go"
  - "services/core/internal/finance/module_test.go"
  - "apps/web/src/api/gen/"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T04 — Merge the contract delta and stub the new finance operation

## Why

Repo ADR 0003: handlers follow the contract. The delta is [contracts/openapi.yaml](../contracts/openapi.yaml).

**Satisfies:** [PRD §5](../PRD.md) AC-05, AC-09 · **Blocked by:** —

## What

- Merge the delta into `services/core/api/openapi.yaml`: the path, the two responses, `UahEquivalent`, the example, `uah_equivalent` on `Contract`.
- `mise run gen` (Go server + TS client).
- Stub `ResolveContractUahEquivalent` on the handler (answers `503 finance.rate_unavailable` shape for now); `toAPI` emits `uah_equivalent: null`.

## Definition of Done

- [ ] Acceptance: The live contract holds `resolveContractUahEquivalent`, `UahEquivalent` and the changed `Contract`; `mise run gen` is clean; every contract answer carries `uah_equivalent: null`; the new operation compiles as a stub; lint and tests are green.
- [ ] integration: create/get/list answers contain the key `uah_equivalent` with value null
- [ ] `npx @stoplight/spectral-cli lint` of the delta has no errors (manual, see rules)
- [ ] `mise run lint` and `mise run test` green
