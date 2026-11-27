---
id: T07
title: "Work out the equivalent when a foreign-currency contract is created, and never fail the creation over it"
layer: "app"
deps: ["T05", "T06"]
acs: ["AC-01", "AC-03", "AC-04"]
files_hint:
  - "services/core/internal/finance/app/contracts.go"
  - "services/core/internal/finance/app/contracts_test.go"
  - "services/core/internal/finance/module.go"
owner: "Andrii Opryshko"
estimate: "M"
status: "done"
---

# T07 — Work out the equivalent when a foreign-currency contract is created, and never fail the creation over it

## Why

[SAD](../sad.md) Flow 1; [PRD](../PRD.md) AC-04 (3 s).

**Satisfies:** [PRD §5](../PRD.md) AC-01, AC-03, AC-04 · **Blocked by:** T05, T06

## What

- `NewContractService(repo, rates)`; `RateTimeout = 3 * time.Second` (overridable in tests).
- `Create`: validate → if `Currency.Foreign()` ask the port under `context.WithTimeout` → `rate.Convert(amount)` → `repo.Create(terms, eq)`.
- A failed lookup is logged with `slog.WarnContext` (currency, day, error) and swallowed.
- `module.go` passes `mcprates.Disabled{}` for now (T10 wires the real one).

## Definition of Done

- [ ] Acceptance: Use-case tests with a fake rate port show: a EUR contract is stored with the equivalent of AC-01; a UAH contract is stored without one and the port is not called; when the port answers not-published, unavailable or blocks past 3 s the contract is still stored without an equivalent and the call returns within the budget.
- [ ] unit: fake port + in-memory repo, the five cases of the DoD
- [ ] unit: validation failure asks neither the port nor the repo
- [ ] `mise run lint` and `mise run test` green
