---
id: T83
title: "Test the decisions and the filings under concurrency"
layer: "tests"
deps: ["T52", "T54"]
acs: ["AC-04c", "AC-15"]
files_hint:
  - "services/core/internal/identity/approval_test.go"
  - "services/core/internal/identity/rejection_test.go"
  - "services/core/internal/identity/applications_file_test.go"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T83 — Test the decisions and the filings under concurrency

## Why

[Review of 2026-11-10](../_review/review-2026-11-10.md), B9. [PRD](../PRD.md) AC-15 and AC-04c hold «whatever the order». The only race test is two competing applications.

**Satisfies:** [PRD §5](../PRD.md) AC-04c, AC-15 · **Blocked by:** [T52](./T52-refuse-stale-filing-and-foreign-code.md), [T54](./T54-commit-before-response.md)

## What

- Three race tests on `testdb.New(t)` with goroutines released together.
- The other cases named in B9 live with their fixes: stale filing and colliding card transfer in T52, the oversized body in T53, the failed commit in T54, Kyiv midnight in T50.

## Definition of Done

- [ ] Acceptance: Integration tests run one application approved twice at once, approved and rejected at once, and two filings racing into the pending unique key, and each ends with exactly one recorded outcome and a 409 for the loser.
- [ ] each test passes 20 times in a row with `-race -count=20`
- [ ] `mise run lint` and `mise run test` green
