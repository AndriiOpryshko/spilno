---
id: T35
title: "Build the signed-in shell: sign-in page, association picker, routes under /o/$associationId and the оператор's entry"
layer: "ui"
deps: ["T02"]
acs: ["AC-04d", "AC-07", "AC-16", "AC-16b"]
files_hint:
  - "apps/web/src/routes/__root.tsx"
  - "apps/web/src/routes/index.tsx"
  - "apps/web/src/routes/sign-in.tsx"
  - "apps/web/src/routes/no-access.tsx"
  - "apps/web/src/routes/o.$associationId.tsx"
  - "apps/web/src/routes/o.$associationId.cases.index.tsx"
  - "apps/web/src/routes/o.$associationId.cases.$caseId.protocol.tsx"
  - "apps/web/src/routes/cases.index.tsx"
  - "apps/web/src/routes/cases.$caseId.protocol.tsx"
  - "apps/web/src/components/app-shell.tsx"
  - "apps/web/src/features/identity/shell/"
  - "apps/web/src/i18n/uk.ts"
owner: "Andrii Opryshko"
estimate: "L"
status: "todo"
---

# T35 — Build the signed-in shell: sign-in page, association picker, routes under /o/$associationId and the оператор's entry

## Why

[ADR-0001](../adr/0001-serve-the-operator-back-office-as-an-operator-only-section-of-the-web-app.md), [ADR-0002](../adr/0002-put-the-association-id-in-every-association-scoped-path.md), [ADR-0009](../adr/0009-treat-the-operator-role-and-a-roster-membership-as-independent-grants-on-one-account.md); [sad](../sad.md) §4 pillar 5: association routes under `/o/{association_id}/`, a picker after sign-in, `/applications` for an applicant, `/operator/` for the оператор. The choice is in the address, not in the session (AC-04d: two windows, two associations).

**Satisfies:** [PRD §5](../PRD.md) AC-04d, AC-07, AC-16, AC-16b · **Blocked by:** [T02](./T02-merge-contract-delta.md)

## What

- The root route loads `getSession`; `/sign-in` links `startSignIn`; `/` lists `listMemberships` (one membership goes straight in, none goes to `/applications`); the cases and protocol routes move under `/o/$associationId/`.
- `Sidebar` shows the association's name instead of the fixed one and an «Оператор» link only when `is_operator`; `/no-access` for 403 `identity.not_current_member` and `identity.not_operator`; 401 goes to sign-in.
- Reuses `Sidebar`, `TopBar`, `BrandMark`, `Button`, `Card`, `Skeleton`, `Badge`, `InlineAlert` (`docs/design/design-system.md`). No new primitive.
- `uk.ts` gets the namespaces `applications`, `operator`, `association`, `storage` with the messages of [PRD](../PRD.md) §5.1 verbatim.

## Definition of Done

- [ ] Acceptance: Component tests show that a visitor without a session lands on sign-in, a person on two rosters picks an association whose id is in every link, a person on none lands on their applications, only the оператор sees the operator entry, and a 403 shows a no-access page with no record.
- [ ] Vitest: no session → sign-in; no membership and not the оператор → `/applications`; two memberships → both listed, each link carries its id
- [ ] the operator link is absent for a person who is not the оператор and present next to the picker for an оператор who is also a member
- [ ] a 403 renders the no-access page and none of the requested record
- [ ] stories for the picker and the no-access page pass `test:stories`; `routeTree.gen.ts` is regenerated, not edited
- [ ] `mise run lint` and `mise run test` green

## Notes

Supersedes [board-protocol T34](../../board-protocol/tasks/T34-ui-sign-in-session.md). board-protocol ui tasks that touch the same route files (T36, T37 there) build under `/o/$associationId/` after this. Later ui tasks add keys only inside their own namespace of `uk.ts`.
