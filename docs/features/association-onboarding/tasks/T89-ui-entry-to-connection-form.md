---
id: T89
title: "Lead a signed-in person to the заявка на підключення form"
layer: "ui"
deps: []
acs: ["AC-01", "AC-04d"]
files_hint:
  - "apps/web/src/routes/applications.index.tsx"
  - "apps/web/src/features/identity/shell/association-picker.tsx"
  - "apps/web/src/features/identity/shell/association-picker.stories.tsx"
  - "apps/web/src/features/identity/shell/shell.test.tsx"
  - "apps/web/src/features/identity/shell/page-states.test.tsx"
  - "apps/web/src/i18n/uk.ts"
owner: "Andrii Opryshko"
estimate: "S"
status: "todo"
---

# T89 — Lead a signed-in person to the заявка на підключення form

## Why

[Review of 2026-11-17](../_review/review-2026-11-17.md), N03. [PRD](../PRD.md) AC-01 starts with a signed-in person who files a заявка; no screen offers the form, and every e2e types its address. AC-04d: a член правління of one association files for a second one.

**Satisfies:** [PRD §5](../PRD.md) AC-01, AC-04d · **Blocked by:** —

## What

- The applications list shows the link above the list and next to «У вас ще немає заявок».
- The association picker shows the same link under the associations, for a member who connects another one.
- Router `Link`, no new component.

## Definition of Done

- [ ] Acceptance: Component tests show a «Подати заявку на підключення» link on the applications list, empty and not, and in the association picker, each leading to `/applications/new?kind=connection`.
- [ ] `shell.test.tsx` or a route test: the link on the empty list, on a list with a заявка, in the picker
- [ ] story of the picker with the link
- [ ] `mise run lint` and `mise run test` green

## Notes

The e2e half is in [T108](./T108-e2e-second-review-fixes.md).
