---
id: T132
title: "Record every gap left open in PRD §8 with an owner and a due date"
layer: "docs"
deps: []
acs: ["AC-01", "AC-26", "AC-27", "AC-28"]
files_hint:
  - "docs/features/association-onboarding/PRD.md"
  - "docs/features/association-onboarding/tasks/tracker.md"
owner: "Andrii Opryshko"
estimate: "S"
status: "done"
---

# T132 — Record every gap left open in PRD §8 with an owner and a due date

## Why

[Second re-review of 2026-11-17](../_review/review-2026-11-17-2.md), R12. The gaps the follow-ups left open sit only in the tracker's «Known gaps», without owner or due date; the review gate asks for deferrals in PRD §8.

**Satisfies:** [PRD §5](../PRD.md) AC-01, AC-26, AC-27, AC-28 · **Blocked by:** —

## What

- PRD §8 gets one line each, with owner and due date, for: the real sign-in provider's adapter, the per-process sealing key, the run of the storage against real Google, files left after a failed removal, the fresh database the e2e suite needs, the latency the CI load job does not prove, the open findings of the security review.
- The tracker's «Known gaps» points at those lines instead of repeating them.

## Definition of Done

- [ ] Acceptance: Every entry of the tracker's «Known gaps» that this follow-up does not close has a line in PRD §8 with an owner and a due date, and the tracker links to it.
- [ ] PRD §8: seven lines, each with owner and due
- [ ] tracker: no gap without a PRD §8 line or a task that closes it
- [ ] `mise run lint` and `mise run test` green

## Notes

The due dates are the product owner's: the task proposes them and the commit waits for their word. Shares `PRD.md` with [T114](./T114-refuse-another-storage-account-while-files-exist.md) and [T131](./T131-reconcile-sync-report-and-design-docs.md); their `files_hint` serializes them.
