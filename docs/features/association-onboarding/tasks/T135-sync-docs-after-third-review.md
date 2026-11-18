---
id: T135
title: "Bring the test plan, the trackers and the review record in line with the third follow-up"
layer: "docs"
deps: ["T113", "T115", "T119", "T120", "T123", "T125", "T126", "T128", "T131", "T132", "T133", "T134"]
acs: ["AC-01", "AC-02", "AC-04d", "AC-17", "AC-26", "AC-27"]
files_hint:
  - "docs/features/association-onboarding/test-plan.md"
  - "docs/features/association-onboarding/tasks/tracker.md"
  - "docs/features/association-onboarding/tasks/_epic.md"
  - "docs/features/association-onboarding/_review/review-2026-11-17-2.md"
  - "docs/features/board-protocol/tasks/tracker.md"
  - "CLAUDE.md"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T135 — Bring the test plan, the trackers and the review record in line with the third follow-up

## Why

[Second re-review of 2026-11-17](../_review/review-2026-11-17-2.md), Every follow-up ended with its documents behind the code (F21, N08, R08, R11). This task is the last one and says what each finding got.

**Satisfies:** [PRD §5](../PRD.md) AC-01, AC-02, AC-04d, AC-17, AC-26, AC-27 · **Blocked by:** [T113](./T113-drive-resolver-one-read-cached-token-deadline.md), [T115](./T115-ui-storage-account-refusal.md), [T119](./T119-ui-one-reader-for-memberships.md), [T120](./T120-ui-protocol-screen-one-cache-key.md), [T123](./T123-sign-in-cookie-path-from-callback-address.md), [T125](./T125-storage-race-test-reports-on-test-goroutine.md), [T126](./T126-decision-effect-after-two-approvals.md), [T128](./T128-load-job-gates-on-throughput-and-errors.md), [T131](./T131-reconcile-sync-report-and-design-docs.md), [T132](./T132-record-open-gaps-in-prd.md), [T133](./T133-e2e-third-review-fixes.md), [T134](./T134-refresh-visual-baselines-third-review.md)

## What

- `test-plan.md` maps every AC to the tests that now exist.
- The review record gets a «Follow-up» section: each finding, its task, its commit, and what stayed open.
- The tracker's «Known gaps» and the e2e gotcha of `CLAUDE.md` say what is true after T127 and T133; the board-protocol tracker says what T112 and T113 delivered of its T13.

## Definition of Done

- [ ] Acceptance: `test-plan.md` names an existing test for every AC, the review record of the second re-review lists each of R01–R13 and S01–S12 with the commit that closed it or the PRD §8 line that holds it, and no «Known gap» describes behaviour the code no longer has.
- [ ] every test name cited in `test-plan.md` exists (grep)
- [ ] the follow-up section of the review record, one row per finding
- [ ] tracker and `CLAUDE.md` gotchas reread against the code
- [ ] `mise run lint` and `mise run test` green

## Notes

Bookkeeping by the implementer, not a verdict: the gate stays CHANGES REQUESTED until `/sdlc:review-feature association-onboarding` runs again.
