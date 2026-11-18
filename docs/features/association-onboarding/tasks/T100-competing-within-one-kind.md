---
id: T100
title: "Mark as competing only pending applications of the same kind for the same code"
layer: "infra"
deps: []
acs: ["AC-04c", "AC-19b"]
files_hint:
  - "services/core/internal/identity/adapters/postgres/queries/review.sql"
  - "services/core/internal/identity/review_queue_competing_test.go"
  - "services/core/internal/identity/review_transfer_test.go"
owner: "Andrii Opryshko"
estimate: "S"
status: "done"
---

# T100 — Mark as competing only pending applications of the same kind for the same code

## Why

[Review of 2026-11-17](../_review/review-2026-11-17.md), Q07. [PRD](../PRD.md) AC-04c and AC-19b define competition between two заявки and between two запити. The queries compare the code only, so a заявка left pending after the code was connected is listed as a competitor of a later запит.

**Satisfies:** [PRD §5](../PRD.md) AC-04c, AC-19b · **Blocked by:** —

## What

- `ListPendingApplications` and `ListCompetingApplications` add the kind to the comparison; `mise run gen`.

## Definition of Done

- [ ] Acceptance: An integration test shows a pending заявка на підключення and a pending запит на передачу for one ЄДРПОУ not marking each other, while two заявки, and two запити, still do.
- [ ] `review_queue_competing_test.go`: the mixed pair, the two same-kind pairs
- [ ] `mise run lint` and `mise run test` green
