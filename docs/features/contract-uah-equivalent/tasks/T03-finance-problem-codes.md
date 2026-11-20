---
id: T03
title: "Register the finance problem codes of PRD §5.1"
layer: "domain"
deps: []
acs: ["AC-07"]
files_hint:
  - "services/core/internal/finance/domain/errors.go"
  - "services/core/internal/finance/domain/errors_test.go"
owner: "Andrii Opryshko"
estimate: "S"
status: "done"
---

# T03 — Register the finance problem codes of PRD §5.1

## Why

`.claude/rules/openapi.md`: one error `code` in the contract = one definition in the repo's registry. The finance context has no registry yet.

**Satisfies:** [PRD §5](../PRD.md) AC-07 · **Blocked by:** —

## What

- `finance/domain/errors.go` in the shape of `governance/domain/errors.go`: `Code…` constants, `ErrorDef{Code, Status}`, `ErrorRegistry`, `DomainError`.
- `finance.equivalent_not_applicable` 409, `finance.rate_not_published` 409, `finance.rate_unavailable` 503.

## Definition of Done

- [ ] Acceptance: Each of the three codes of PRD §5.1 has exactly one constant whose string equals the catalogue, with its HTTP status in a registry, and a test fails when a code is missing or misspelt.
- [ ] unit: the set of constants equals the `Code` column of PRD §5.1
- [ ] unit: the statuses match `x-problem-types` of the feature contract
- [ ] `mise run lint` and `mise run test` green
