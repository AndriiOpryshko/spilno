---
id: T88
title: "Mark roster entries by the decision's effect and word the mark by the kind of the decision"
layer: "ui"
deps: ["T87"]
acs: ["AC-20", "AC-29"]
files_hint:
  - "apps/web/src/features/identity/association/roster-decision-mark.tsx"
  - "apps/web/src/features/identity/association/roster-decision-mark.test.tsx"
  - "apps/web/src/features/governance/roster/roster-screen.tsx"
  - "apps/web/src/features/governance/roster/roster-form.tsx"
  - "apps/web/src/features/governance/roster/roster-screen.test.tsx"
  - "apps/web/src/features/governance/roster/roster-screen.stories.tsx"
  - "apps/web/src/features/identity/association/association-card.tsx"
  - "apps/web/src/i18n/uk.ts"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T88 — Mark roster entries by the decision's effect and word the mark by the kind of the decision

## Why

[Review of 2026-11-17](../_review/review-2026-11-17.md), N01, N02. [PRD](../PRD.md) AC-20 words the mark for a запит на передачу only; the mark is printed for every decision with a filing date, so the first голова of a connected association is said to be appointed by a запит that never existed.

**Satisfies:** [PRD §5](../PRD.md) AC-20, AC-29 · **Blocked by:** [T87](./T87-roster-decision-effect-and-no-reopen.md)

## What

- `RosterScreen` takes `ended` / `appointed` from `decision_effect`; the match by `ended_head_full_name` goes.
- `RosterDecisionMark` words the mark from `event.kind`: the AC-20 text for `transfer_approved`, «за заявкою на підключення» for `connected`. Reuses `formatDate` and the generated extract URL helper.
- `RosterForm` of an entry with `decision_effect` `ended` shows the term as text with the hint «Щоб повернути людину до правління, додайте запис знову».
- The card history uses the same wording helper.

## Definition of Done

- [ ] Acceptance: Component tests show the ended mark on the entry with `decision_effect` `ended` when both голови share a name, «призначено за заявкою на підключення від …» for the голова of a connection, and the term fields of an ended entry replaced by a hint to add the person again.
- [ ] `roster-decision-mark.test.tsx`: a `connected` and a `transfer_approved` event
- [ ] `roster-screen.test.tsx`: two голови with one full name; the form of an ended entry
- [ ] story for the roster of a newly connected association
- [ ] `mise run lint` and `mise run test` green

## Notes

No new component: `RosterDecisionMark`, `InlineHint` and the existing form fields.
