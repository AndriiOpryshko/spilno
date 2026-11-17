---
id: T101
title: "Report race-test failures on the test goroutine and stop the term-end tests from expiring"
layer: "tests"
deps: ["T86"]
acs: ["AC-08", "AC-15", "AC-19c", "AC-30"]
files_hint:
  - "services/core/internal/identity/approval_transfer_test.go"
  - "services/core/internal/identity/decision_race_test.go"
  - "services/core/internal/identity/approval_test.go"
  - "services/core/internal/identity/review_test.go"
owner: "Andrii Opryshko"
estimate: "S"
status: "todo"
---

# T101 — Report race-test failures on the test goroutine and stop the term-end tests from expiring

## Why

[Review of 2026-11-17](../_review/review-2026-11-17.md), Q05, Q06. Four tests use the term end 2026-12-31 on the real clock and fail from 2027-01-01 with `identity.term_ended`; the race tests call helpers that `t.Fatalf` from goroutines.

**Satisfies:** [PRD §5](../PRD.md) AC-08, AC-15, AC-19c, AC-30 · **Blocked by:** [T86](./T86-decision-entries-by-first-revision.md)

## What

- The term-end tests pin the clock with `Module.SetClock` or derive the dates from it.
- The race closures return errors; the test goroutine reports them.

## Definition of Done

- [ ] Acceptance: The identity tests pass with the module clock set to 2027-06-01, and no goroutine of a race test calls `t.Fatal`.
- [ ] a run with the clock a year ahead
- [ ] `go vet` reports no `t.Fatal` in a goroutine (testinggoroutine)
- [ ] `mise run lint` and `mise run test` green
