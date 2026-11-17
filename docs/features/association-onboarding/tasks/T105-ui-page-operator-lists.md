---
id: T105
title: "Read every page of the оператор's queue and list of associations"
layer: "ui"
deps: ["T102"]
acs: ["AC-04c", "AC-21"]
files_hint:
  - "apps/web/src/features/identity/operator/queue.tsx"
  - "apps/web/src/features/identity/operator/associations.tsx"
  - "apps/web/src/features/identity/operator/operator-section.test.tsx"
  - "apps/web/src/features/identity/operator/associations.test.tsx"
  - "CLAUDE.md"
  - "docs/features/association-onboarding/tasks/tracker.md"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T105 — Read every page of the оператор's queue and list of associations

## Why

[Review of 2026-11-17](../_review/review-2026-11-17.md), Q15. With more than twenty pending заявки a new one is not on the оператор's screen; the e2e suite needs a fresh database because of it.

**Satisfies:** [PRD §5](../PRD.md) AC-04c, AC-21 · **Blocked by:** [T102](./T102-ui-operator-fallback-and-focus.md)

## What

- Infinite queries on the generated keys with «Показати ще».
- The CLAUDE.md gotcha and the tracker's known gap about the first twenty rows are removed once the second e2e run passes.

## Definition of Done

- [ ] Acceptance: Component tests show both lists offering «Показати ще» until `next_cursor` is null, and the e2e suite passes twice in a row on one database.
- [ ] tests of two pages for each list
- [ ] two e2e runs on one throwaway database
- [ ] `mise run lint` and `mise run test` green
