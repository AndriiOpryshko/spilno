---
id: T02
title: "Model the official rate and compute the hryvnia equivalent with half-up rounding"
layer: "domain"
deps: []
acs: ["AC-02", "AC-09"]
files_hint:
  - "services/core/internal/finance/domain/rate.go"
  - "services/core/internal/finance/domain/rate_test.go"
  - "services/core/internal/finance/domain/contract.go"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T02 — Model the official rate and compute the hryvnia equivalent with half-up rounding

## Why

[PRD](../PRD.md) AC-02 fixes the arithmetic; [SAD](../sad.md) §8 «Money» fixes the representation (micros, `math/big`).

**Satisfies:** [PRD §5](../PRD.md) AC-02, AC-09 · **Blocked by:** —

## What

- `domain.Rate`: currency, day and `Micros int64`; `ParseRate(currency, day, decimal string)`; `Rate.Decimal()` string.
- `domain.UAHEquivalent{AmountKopiyky, Rate}` and `Rate.Convert(amountCents) UAHEquivalent`.
- `ErrRateNotPublished`, `ErrRateUnavailable` sentinels.
- `Contract` gains `UAHEquivalent *UAHEquivalent`; `Currency.Foreign()`.
- stdlib only (repo ADR 0002).

## Definition of Done

- [ ] Acceptance: Unit tests prove the rounding table of AC-02 with exact integers, refuse a rate that is not a positive decimal with at most six fraction digits, and format a rate back to the published decimal with trailing zeros dropped.
- [ ] unit: every row of the AC-02 table, including the 9e15 no-overflow case
- [ ] unit: malformed rates (`""`, `"-1"`, `"0"`, `"1e2"`, `"1.2345678"`, `"1,5"`) are refused
- [ ] unit: `Decimal()` round-trips `50.5311`, `45.051`, `45`
- [ ] `mise run lint` and `mise run test` green
