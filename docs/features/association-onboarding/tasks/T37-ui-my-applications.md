---
id: T37
title: "Show my applications: status, the оператор's contact, the reason and refiling"
layer: "ui"
deps: ["T36"]
acs: ["AC-05", "AC-06", "AC-14"]
files_hint:
  - "apps/web/src/routes/applications.index.tsx"
  - "apps/web/src/routes/applications.$applicationId.tsx"
  - "apps/web/src/features/identity/applications/status-line.tsx"
  - "apps/web/src/features/identity/applications/status-line.test.tsx"
  - "apps/web/src/features/identity/applications/status-line.stories.tsx"
  - "apps/web/src/features/identity/applications/application-detail.tsx"
  - "apps/web/src/features/identity/applications/application-detail.test.tsx"
  - "apps/web/src/vite-env.d.ts"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T37 — Show my applications: status, the оператор's contact, the reason and refiling

## Why

[PRD](../PRD.md) US-02; [sad](../sad.md) §6 Flow 5; [api-sync-report](../contracts/api-sync-report.md) §E «Operator contact»: the address is a build setting of the web app, named here.

**Satisfies:** [PRD §5](../PRD.md) AC-05, AC-06, AC-14 · **Blocked by:** [T36](./T36-ui-filing-forms.md)

## What

- `/applications` and `/applications/$applicationId`: `StatusLine` («очікує розгляду з 10.11.2026», схвалена, відхилена with the reason), the оператор's contact from `VITE_OPERATOR_CONTACT_EMAIL`, the history of filings, the extract link.
- «Виправити і подати знову» opens the form of T36 filled from the latest filing and posts `refileApplication`; an approved заявка links into its association.

## Definition of Done

- [ ] Acceptance: Component tests show the status line for pending, approved and rejected, the оператор's contact address, the rejection reason and history, and a refile that returns the application to pending.
- [ ] the three states render as specified, dates as dd.mm.yyyy; the contact address is shown for a pending one
- [ ] a refile shows the new status and keeps the rejected filing with its reason in the history
- [ ] nothing of the association is rendered for a pending заявка; the story passes a11y
- [ ] `mise run lint` and `mise run test` green

## Notes

Reuses `Card`, `Badge`, `Button`, `InlineAlert` and the form of T36.
