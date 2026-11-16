---
id: T68
title: "Put the association card in the navigation and the suspended notice in the association layout"
layer: "ui"
deps: ["T67"]
acs: ["AC-04d", "AC-21", "AC-25", "AC-27", "AC-29"]
files_hint:
  - "apps/web/src/components/app-shell.tsx"
  - "apps/web/src/routes/o.$associationId.tsx"
  - "apps/web/src/routes/index.tsx"
  - "apps/web/src/features/identity/shell/association-picker.tsx"
  - "apps/web/src/features/identity/association/suspended-notice.tsx"
  - "apps/web/src/features/governance/cases-screen.tsx"
  - "apps/web/src/i18n/uk.ts"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T68 — Put the association card in the navigation and the suspended notice in the association layout

## Why

[Review of 2026-11-10](../_review/review-2026-11-10.md), F08, F14. [PRD](../PRD.md) AC-21 (the member sees «Доступ призупинено» on their next action), AC-27 (the storage line is shown to every член правління), AC-29 and AC-04d (every page names its association). The card is reachable only by typing its URL, the mobile case list links to a retired route and the mobile top bar names a fixed association.

**Satisfies:** [PRD §5](../PRD.md) AC-04d, AC-21, AC-25, AC-27, AC-29 · **Blocked by:** [T67](./T67-ui-operator-layout-entry-guard.md)

## What

- Sidebar entry «ОСББ» to `/o/$associationId/association`; the layout renders `SuspendedNotice` from the membership's `state` and reason.
- `CaseCard` links to `/o/$associationId/cases/$caseId/protocol`; `TopBar` gets the membership's `association_name`; `uk.app.osbbShort` is removed.
- The retired `/cases/*` routes keep redirecting.

## Definition of Done

- [ ] Acceptance: Component tests show the card linked from the sidebar and the picker, «Доступ призупинено» with its reason on every page under `/o/$associationId` of a вимкнене ОСББ, the mobile case card opening the protocol under the association, and the association's own name in the mobile top bar.
- [ ] test of the layout for an enabled and a disabled membership
- [ ] test of `CaseCard`'s href with an association id
- [ ] stories for the shell with and without the notice pass a11y
- [ ] `mise run lint` and `mise run test` green

## Notes

Reuses `InlineAlert` for the notice. The справи screen leaves its in-memory mock in T80.
