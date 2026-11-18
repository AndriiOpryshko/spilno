---
id: T104
title: "Read every page of the roster and of a person's memberships"
layer: "ui"
deps: ["T88"]
acs: ["AC-04d", "AC-20"]
files_hint:
  - "apps/web/src/features/governance/roster/roster-screen.tsx"
  - "apps/web/src/features/governance/roster/roster-screen.test.tsx"
  - "apps/web/src/routes/o.$associationId.tsx"
  - "apps/web/src/routes/index.tsx"
  - "apps/web/src/features/identity/shell/association-shell.test.tsx"
owner: "Andrii Opryshko"
estimate: "M"
status: "done"
---

# T104 — Read every page of the roster and of a person's memberships

## Why

[Review of 2026-11-17](../_review/review-2026-11-17.md), Q15. The roster returns ended entries too, twenty a page; later entries and their AC-20 marks disappear. The layout decides access from the first page of memberships.

**Satisfies:** [PRD §5](../PRD.md) AC-04d, AC-20 · **Blocked by:** [T88](./T88-ui-roster-marks-by-effect-and-kind.md)

## What

- Roster: an infinite query on the generated key with «Показати ще», as the справи list.
- Memberships: the layout and the entry read pages until `next_cursor` is null.

## Definition of Done

- [ ] Acceptance: Component tests show the roster offering «Показати ще» until `next_cursor` is null, with the marks on entries of later pages, and a person on 21 rosters opening the twenty-first association.
- [ ] `roster-screen.test.tsx`: two pages
- [ ] `association-shell.test.tsx`: 21 memberships
- [ ] `mise run lint` and `mise run test` green
