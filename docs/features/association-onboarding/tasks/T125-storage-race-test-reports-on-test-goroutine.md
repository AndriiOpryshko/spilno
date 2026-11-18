---
id: T125
title: "Report the storage race test's failures on the test goroutine"
layer: "tests"
deps: []
acs: ["AC-27"]
files_hint:
  - "services/core/internal/identity/storage_test.go"
owner: "Andrii Opryshko"
estimate: "S"
status: "todo"
---

# T125 — Report the storage race test's failures on the test goroutine

## Why

[Second re-review of 2026-11-17](../_review/review-2026-11-17-2.md), S02. `TestStorageCodeExchangeHoldsNoConnection` calls a helper that `t.Fatal`s from two goroutines, the pattern T101 removed from the decision tests.

**Satisfies:** [PRD §5](../PRD.md) AC-27 · **Blocked by:** —

## What

- The storage test environment gets a variant of `do` that returns its error; the goroutines send it to the test goroutine, which reports it.

## Definition of Done

- [ ] Acceptance: No goroutine in `storage_test.go` calls a helper that can `t.Fatal`, and the test still fails with its message when the exchange holds a connection.
- [ ] the two goroutines use the error-returning helper
- [ ] the test fails as before with the fix of T96 reverted locally (checked once, not committed)
- [ ] `mise run lint` and `mise run test` green
