---
id: T117
title: "Offer the запит на передачу on the association card to every member"
layer: "ui"
deps: []
acs: ["AC-17"]
files_hint:
  - "apps/web/src/features/identity/association/association-card.tsx"
  - "apps/web/src/features/identity/association/association-card.test.tsx"
  - "apps/web/src/features/identity/association/association-card.stories.tsx"
  - "apps/web/src/i18n/uk.ts"
owner: "Andrii Opryshko"
estimate: "S"
status: "done"
---

# T117 — Offer the запит на передачу on the association card to every member

## Why

[Second re-review of 2026-11-17](../_review/review-2026-11-17-2.md), R04. AC-17: a newly elected голова files a запит на передачу «whether or not they are already on its roster». The link on the card renders only for an association connected by hand; a член правління of an ordinary association has no way to the form.

**Satisfies:** [PRD §5](../PRD.md) AC-17 · **Blocked by:** —

## What

- The card shows «Подати запит на передачу» with `association_id` for every member; the «connected by hand» badge stays where it is.
- The existing link and string are reused.

## Definition of Done

- [ ] Acceptance: A component test shows the «Подати запит на передачу» link on the card of an association connected by a заявка and of one connected by hand, each leading to `/applications/new?kind=transfer&association_id=<id>`.
- [ ] `association-card.test.tsx`: the link in both cases
- [ ] story updated; the baseline is refreshed in T134
- [ ] `mise run lint` and `mise run test` green

## Notes

The card is a baselined story: [T134](./T134-refresh-visual-baselines-third-review.md) refreshes it. The e2e half is in [T133](./T133-e2e-third-review-fixes.md).
