---
id: T08
title: "Add the use case that works out a missing equivalent later, once"
layer: "app"
deps: ["T03", "T07"]
acs: ["AC-05", "AC-06", "AC-07"]
files_hint:
  - "services/core/internal/finance/app/contracts.go"
  - "services/core/internal/finance/app/contracts_test.go"
owner: "Andrii Opryshko"
estimate: "M"
status: "done"
---

# T08 — Add the use case that works out a missing equivalent later, once

## Why

[SAD](../sad.md) Flow 2; codes from T03.

**Satisfies:** [PRD §5](../PRD.md) AC-05, AC-06, AC-07 · **Blocked by:** T03, T07

## What

- `ContractService.ResolveUAHEquivalent(ctx, id) (domain.Contract, error)` following Flow 2 step by step.
- Errors are `*domain.DomainError` with the codes of T03; not found stays `domain.ErrContractNotFound`.
- Same 3 s budget as T07.

## Definition of Done

- [ ] Acceptance: Use-case tests show: a contract without an equivalent gets one; a contract with one is returned unchanged and the port is not called; a UAH contract, a not-published rate and an unavailable source return errors carrying `finance.equivalent_not_applicable`, `finance.rate_not_published` and `finance.rate_unavailable`; a lost race (`SetUAHEquivalent` updated nothing) returns the stored contract.
- [ ] unit: the six cases of the DoD with a fake port and repo
- [ ] `mise run lint` and `mise run test` green
