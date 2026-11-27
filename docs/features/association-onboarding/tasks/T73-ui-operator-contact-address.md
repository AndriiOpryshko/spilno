---
id: T73
title: "Show the оператор's contact address in every environment"
layer: "ui"
deps: []
acs: ["AC-05"]
files_hint:
  - "apps/web/src/features/identity/applications/status-line.tsx"
  - "apps/web/src/features/identity/applications/status-line.test.tsx"
  - "apps/web/vite.config.ts"
  - "apps/web/.env.example"
  - "mprocs.yaml"
  - "README.md"
owner: "Andrii Opryshko"
estimate: "S"
status: "done"
---

# T73 — Show the оператор's contact address in every environment

## Why

[Review of 2026-11-10](../_review/review-2026-11-10.md), F13. [PRD](../PRD.md) AC-05: the applicant sees how to reach the оператор. `VITE_OPERATOR_CONTACT_EMAIL` is set only by the Playwright config and the line disappears when it is unset.

**Satisfies:** [PRD §5](../PRD.md) AC-05 · **Blocked by:** —

## What

- Set the variable for the dev server and in `apps/web/.env.example`; README already names it.
- The production build fails when it is missing; the status line never renders an empty contact silently.

## Definition of Done

- [ ] Acceptance: `mise run dev` shows the contact line on a pending заявка, and a production build without the address fails with a message that names the variable.
- [ ] `status-line.test.tsx`: the line is present with the address
- [ ] a build without the variable exits non-zero
- [ ] `mise run lint` and `mise run test` green
