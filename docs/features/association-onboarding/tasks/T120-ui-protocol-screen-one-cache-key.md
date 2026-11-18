---
id: T120
title: "Read and write the protocol draft under one cache key and tighten the reuse test"
layer: "ui"
deps: ["T119"]
acs: ["AC-25"]
files_hint:
  - "apps/web/src/features/governance/protocol-screen.tsx"
  - "apps/web/src/features/governance/slot-row.tsx"
  - "apps/web/src/features/governance/protocol-screen.test.tsx"
  - "apps/web/src/features/identity/reuse.test.tsx"
owner: "Andrii Opryshko"
estimate: "S"
status: "done"
---

# T120 — Read and write the protocol draft under one cache key and tighten the reuse test

## Why

[Second re-review of 2026-11-17](../_review/review-2026-11-17-2.md), S08, S11. Since T103 the protocol screen reads the generated key and writes `['protocol', caseId]`, so a slot save or a transition leaves the old draft on the screen. The reuse test matches only `queryKey: ['` and exempts two whole folders from its alert rule.

**Satisfies:** [PRD §5](../PRD.md) AC-25 · **Blocked by:** [T119](./T119-ui-one-reader-for-memberships.md)

## What

- One key helper for the read and both writes of the draft.
- The reuse test also matches hand-built keys in `setQueryData(` and `invalidateQueries(`, and exempts only `inline-alert.tsx` and `health-card.tsx` from the alert rule.

## Definition of Done

- [ ] Acceptance: A component test saves a slot and makes a transition on the protocol screen and sees the new draft without a reload; the reuse test fails on a hand-built key in `setQueryData` and on an ad-hoc `role="alert"` in a governance screen.
- [ ] `protocol-screen.test.tsx`: save a slot, the slot is no longer dirty; a transition changes the state and the buttons
- [ ] `reuse.test.tsx`: the wider key pattern and the two-file exemption
- [ ] `mise run lint` and `mise run test` green

## Notes

The screen is board-protocol's and still runs on the mock API; this task only repairs what T103 changed. AC-25 is this feature's stake in the screen (no draft opens in a вимкнене ОСББ).
