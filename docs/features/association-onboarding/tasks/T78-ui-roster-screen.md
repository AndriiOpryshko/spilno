---
id: T78
title: "Build the roster screen: list, add and change члени правління, with the decision mark on each голова"
layer: "ui"
deps: ["T62", "T68"]
acs: ["AC-09", "AC-20"]
files_hint:
  - "apps/web/src/routes/o.$associationId.roster.tsx"
  - "apps/web/src/features/governance/roster/"
  - "apps/web/src/features/identity/association/roster-decision-mark.tsx"
  - "apps/web/src/components/app-shell.tsx"
  - "apps/web/src/i18n/uk.ts"
owner: "Andrii Opryshko"
estimate: "L"
status: "todo"
---

# T78 — Build the roster screen: list, add and change члени правління, with the decision mark on each голова

## Why

[Review of 2026-11-10](../_review/review-2026-11-10.md), F15. [PRD](../PRD.md) AC-09 and AC-20. `RosterDecisionMark` exists and is mounted nowhere; no screen shows or changes the roster.

**Satisfies:** [PRD §5](../PRD.md) AC-09, AC-20 · **Blocked by:** [T62](./T62-serve-roster-endpoints.md), [T68](./T68-ui-association-shell-card-notice.md)

## What

- Route `/o/$associationId/roster`, linked from the sidebar; the roster half of [board-protocol T35](../../board-protocol/tasks/T35-ui-settings-roster.md).
- Mount `RosterDecisionMark` on entries that carry a `decision_id`.
- The form is hidden behind the suspended notice in a вимкнене ОСББ.

## Definition of Done

- [ ] Acceptance: Component tests show the roster listed from `listBoardMembers`, a член правління added and changed with every 422 field named, and the entries of the new and the previous голова carrying «призначено / повноваження завершено за запитом на передачу від …, схвалено оператором …» with a link to the extract.
- [ ] stories for empty, filled and error states pass a11y
- [ ] test: an entry without a decision shows no mark
- [ ] generated hooks and keys only
- [ ] `mise run lint` and `mise run test` green

## Notes

Reuses `Input`, `DateField`, `Button`, `Card`, `InlineAlert`, `Dialog`. The settings screen stays in board-protocol T35.
