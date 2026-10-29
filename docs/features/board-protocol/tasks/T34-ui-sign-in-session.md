---
id: T34
title: "Add the sign-in page, the session guard and 401/403 handling"
layer: "ui"
deps: ["T07"]
acs: ["AC-19", "AC-16"]
files_hint:
  - "apps/web/src/routes/sign-in.tsx"
  - "apps/web/src/routes/__root.tsx"
  - "apps/web/src/api/fetch.ts"
  - "apps/web/src/i18n/uk.ts"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T34 — Add the sign-in page, the session guard and 401/403 handling

## Why

[ADR-0003](../adr/0003-sign-in-with-oidc-and-postgres-sessions.md); AC-19 (existence hidden). Reuses Button, Card and InlineAlert (docs/design/design-system.md).

**Satisfies:** [PRD §5](../PRD.md) AC-19, AC-16 · **Blocked by:** [T07](./T07-merge-contract-identity-skeleton.md)

## What

- Root route loads getSession; sign-in page links startSignIn.
- Shared fetch error handling for 401 → sign-in, 403 → no-access.

## Definition of Done

- [ ] Acceptance: Component tests show an unauthenticated visit redirects to the sign-in page, a 403 `identity.not_current_member` shows a no-access screen that reveals no record, and stories for both states pass the a11y check.
- [ ] Vitest component tests for both redirects
- [ ] stories for sign-in and no-access pass `test:stories`
- [ ] `mise run lint` and `mise run test` green

## Notes

Edits `uk.ts` — every ui task touches it; keep keys under a per-feature namespace to avoid conflicts.
