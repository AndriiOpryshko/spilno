---
id: T41
title: "Build the association card with the confirmation chain and the roster marks"
layer: "ui"
deps: ["T35"]
acs: ["AC-19c", "AC-20", "AC-29"]
files_hint:
  - "apps/web/src/routes/o.$associationId.association.tsx"
  - "apps/web/src/features/identity/association/association-card.tsx"
  - "apps/web/src/features/identity/association/association-card.test.tsx"
  - "apps/web/src/features/identity/association/association-card.stories.tsx"
  - "apps/web/src/features/identity/association/roster-decision-mark.tsx"
  - "apps/web/src/features/identity/association/roster-decision-mark.test.tsx"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T41 — Build the association card with the confirmation chain and the roster marks

## Why

[PRD](../PRD.md) US-09, AC-29, AC-20; [sad](../sad.md) §6 Flow 12; [api-sync-report](../contracts/api-sync-report.md) §E «Roster marks».

**Satisfies:** [PRD §5](../PRD.md) AC-19c, AC-20, AC-29 · **Blocked by:** [T35](./T35-ui-shell-sign-in-picker.md)

## What

- `/o/$associationId/association`: `AssociationCard` from `getAssociationCard` — the current голова, the extract (download), its date and signer, the decision date, later запити and disablings with their reasons.
- The hand-made variant shows «підключено вручну, без витягу» and «Подати запит на передачу», which opens the form of T36 with the card's `association_id`.
- `RosterDecisionMark`: «призначено / повноваження завершено за запитом на передачу від <дата>, схвалено оператором <дата>» with the extract link, found by `decision_id` in the card's history.

## Definition of Done

- [ ] Acceptance: Component tests show the card with the голова, the extract, its date and signer, the decision date and the history, the «підключено вручну, без витягу» variant with a link to the запит form, and the appointed and ended marks of a transfer.
- [ ] both card variants render every part named in AC-29
- [ ] the mark renders the appointed and the ended text with both dates
- [ ] stories for both variants pass a11y
- [ ] `mise run lint` and `mise run test` green

## Notes

The roster screen is board-protocol's ui task T35 (settings and roster): it mounts `RosterDecisionMark` on its rows. Until then the mark is proven here and visible in the card's history. Reuses `Card`, `Badge`, `Button`. Visual baseline in T43.
