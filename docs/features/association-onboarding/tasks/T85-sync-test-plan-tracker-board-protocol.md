---
id: T85
title: "Bring the test plan, the trackers and the board-protocol tasks in line with what was built"
layer: "docs"
deps: ["T61", "T81", "T82", "T83", "T84"]
acs: ["AC-09", "AC-20", "AC-26", "AC-28"]
files_hint:
  - "docs/features/association-onboarding/test-plan.md"
  - "docs/features/association-onboarding/tasks/tracker.md"
  - "docs/features/association-onboarding/_review/review-2026-11-10.md"
  - "docs/features/board-protocol/tasks/tracker.md"
  - "docs/features/board-protocol/tasks/"
owner: "Andrii Opryshko"
estimate: "M"
status: "done"
---

# T85 — Bring the test plan, the trackers and the board-protocol tasks in line with what was built

## Why

[Review of 2026-11-10](../_review/review-2026-11-10.md), F15 (bookkeeping), F21. The test plan maps ACs to tests that changed, the tracker's known gaps name things now closed, and six board-protocol tasks were carried in part or whole by this follow-up.

**Satisfies:** [PRD §5](../PRD.md) AC-09, AC-20, AC-26, AC-28 · **Blocked by:** [T61](./T61-record-single-release-migrations.md), [T81](./T81-e2e-review-fixes-through-clicks.md), [T82](./T82-e2e-roster-registry-documents.md), [T83](./T83-concurrency-and-failure-tests.md), [T84](./T84-refresh-visual-baselines.md)

## What

- Test plan rows for the new and changed tests; no AC row points at a `fixme`.
- Board-protocol T10, T21, T31, T35, T36 and T43: status and a note of the remaining scope (settings, the needs-review event, the upload metric).
- Tracker «Known gaps» and the review record updated; then `/sdlc:review-feature association-onboarding` over the follow-up diff.

## Definition of Done

- [ ] Acceptance: `test-plan.md` maps every AC to the tests that now exist, the board-protocol tracker and task files say what T62–T65 and T78–T80 delivered and what is left there, and the review record lists each finding with the commit that closed it.
- [ ] every link in the edited files resolves
- [ ] `/ao-sdlc-audit association-onboarding` reports no stale artifact
- [ ] `mise run lint` and `mise run test` green
