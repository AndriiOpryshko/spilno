---
id: T67
title: "Give the оператор's section an index, its own navigation and one access guard"
layer: "ui"
deps: ["T66"]
acs: ["AC-16", "AC-16b", "AC-21"]
files_hint:
  - "apps/web/src/routes/operator.tsx"
  - "apps/web/src/routes/operator.index.tsx"
  - "apps/web/src/components/app-shell.tsx"
  - "apps/web/src/features/identity/shell/association-picker.tsx"
  - "apps/web/src/features/identity/shell/shell.test.tsx"
  - "apps/web/src/features/identity/operator/queue.tsx"
  - "apps/web/src/features/identity/operator/associations.tsx"
  - "apps/web/src/features/identity/operator/review-page.tsx"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T67 — Give the оператор's section an index, its own navigation and one access guard

## Why

[Review of 2026-11-10](../_review/review-2026-11-10.md), F03, W1. [ADR-0001](../adr/0001-serve-the-operator-back-office-as-an-operator-only-section-of-the-web-app.md) and [sad](../sad.md) §4: `/operator/` is the оператор's section. The «Оператор» link goes to a route that does not exist, the queue and the associations list do not link to each other, and each page handles 401 and 403 its own way.

**Satisfies:** [PRD §5](../PRD.md) AC-16, AC-16b, AC-21 · **Blocked by:** [T66](./T66-ui-sign-in-link-generated-urls.md)

## What

- `routes/operator.tsx` as a layout mirroring `o.$associationId.tsx`: session check, `isNoAccessProblem` → `/no-access`, navigation to «Заявки» and «ОСББ».
- `routes/operator.index.tsx` redirects to `/operator/applications/`; the shell and the picker link with `<Link>`.
- The three pages drop their own 403 branches.

## Definition of Done

- [ ] Acceptance: Component tests show «Оператор» opening the queue, links between the queue and the associations list, a non-оператор sent to `/no-access` and an expired session sent to `/sign-in` from every operator page.
- [ ] `shell.test.tsx` asserts the exact path, not `toContain('/operator')`
- [ ] `OperatorAssociations` with a 403 shows no empty list
- [ ] the server stays the rule: the guard only routes, `identity.not_operator` still refuses
- [ ] `mise run lint` and `mise run test` green

## Notes

Reuses `NoAccessPage`, `Card`, `Button`. No new primitive.
