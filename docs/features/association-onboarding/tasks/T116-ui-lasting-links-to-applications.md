---
id: T116
title: "Give every signed-in person a lasting way to their заявки and to the заявка на підключення form"
layer: "ui"
deps: []
acs: ["AC-01", "AC-04d", "AC-05", "AC-06", "AC-14"]
files_hint:
  - "apps/web/src/components/app-shell.tsx"
  - "apps/web/src/components/app-shell.stories.tsx"
  - "apps/web/src/features/identity/shell/association-picker.tsx"
  - "apps/web/src/features/identity/shell/association-picker.stories.tsx"
  - "apps/web/src/features/identity/shell/shell.test.tsx"
  - "apps/web/src/features/identity/shell/association-shell.test.tsx"
  - "apps/web/src/i18n/uk.ts"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T116 — Give every signed-in person a lasting way to their заявки and to the заявка на підключення form

## Why

[Second re-review of 2026-11-17](../_review/review-2026-11-17-2.md), R03. `/` sends a person on exactly one roster straight into the association, and the association sidebar has no link to `/applications` or to the form. Such a person cannot file for a second association (AC-01, the Given of AC-04d), follow a заявка (AC-05), read a rejection (AC-14) or refile (AC-06). The picker lacks the link to the list too.

**Satisfies:** [PRD §5](../PRD.md) AC-01, AC-04d, AC-05, AC-06, AC-14 · **Blocked by:** —

## What

- The association sidebar gets «Мої заявки» and «Подати заявку на підключення».
- The picker gets «Мої заявки» next to the link T89 added.
- Router `Link` in the existing navigation markup, no new component.

## Definition of Done

- [ ] Acceptance: Component tests show «Мої заявки» and «Подати заявку на підключення» in the association sidebar and «Мої заявки» in the picker, each leading to its route, for a person on one roster and on several.
- [ ] `association-shell.test.tsx`: both links in the sidebar, their targets
- [ ] `shell.test.tsx`: «Мої заявки» in the picker
- [ ] stories of the sidebar and the picker with the links
- [ ] `mise run lint` and `mise run test` green

## Notes

The e2e half is in [T133](./T133-e2e-third-review-fixes.md). The sidebar stories have no visual baseline; the card's baseline changes in [T117](./T117-ui-transfer-request-link-for-every-member.md).
