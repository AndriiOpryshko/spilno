---
id: T111
title: "Bring the test plan, the trackers and the review record in line with the second follow-up"
layer: "docs"
deps: ["T96", "T97", "T98", "T101", "T103", "T107", "T108", "T109", "T110"]
acs: ["AC-01", "AC-19", "AC-20", "AC-26", "AC-28"]
files_hint:
  - "docs/features/association-onboarding/test-plan.md"
  - "docs/features/association-onboarding/tasks/tracker.md"
  - "docs/features/association-onboarding/_review/review-2026-11-17.md"
  - "docs/features/board-protocol/tasks/tracker.md"
  - "docs/features/board-protocol/tasks/"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T111 — Bring the test plan, the trackers and the review record in line with the second follow-up

## Why

[Review of 2026-11-17](../_review/review-2026-11-17.md). The record of 2026-11-17 asks for a re-review of the changed surface; it needs the map from findings to commits and a test plan that names the new tests.

**Satisfies:** [PRD §5](../PRD.md) AC-01, AC-19, AC-20, AC-26, AC-28 · **Blocked by:** [T96](./T96-code-exchange-outside-transaction.md), [T97](./T97-wire-drive-storage-per-association.md), [T98](./T98-oversize-body-413-everywhere.md), [T101](./T101-race-tests-and-term-dates-that-last.md), [T103](./T103-ui-governance-generated-hooks-and-fallback.md), [T107](./T107-load-smoke-in-ci.md), [T108](./T108-e2e-second-review-fixes.md), [T109](./T109-refresh-visual-baselines-second-review.md), [T110](./T110-contract-texts-and-reconcile.md)

## What

- A «Follow-up» section in the review record, marked as the implementer's bookkeeping.
- The tracker's known gaps: the ones closed here removed, the ones left reworded.
- board-protocol: the wiring of the Drive adapter and `Files.Remove` are done here; the run against real Google and the `remove_orphan_file` job stay there.

## Definition of Done

- [ ] Acceptance: `test-plan.md` maps every AC to the tests that now exist, the board-protocol tracker and its T13, T13b and T21 files say what T92, T94 and T97 delivered and what is left there, and the review record of 2026-11-17 lists each finding with the commit that closed it.
- [ ] every finding of the record names a task and a commit
- [ ] no known gap of the tracker describes something a task of this follow-up closed
- [ ] `mise run lint` and `mise run test` green
