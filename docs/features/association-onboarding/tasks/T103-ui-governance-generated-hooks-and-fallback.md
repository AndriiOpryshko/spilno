---
id: T103
title: "Use the generated hooks, their query keys and the shared fallback in the governance screens"
layer: "ui"
deps: []
acs: ["AC-19", "AC-26", "AC-28"]
files_hint:
  - "apps/web/src/features/governance/cases-screen.tsx"
  - "apps/web/src/features/governance/registry/registry-screen.tsx"
  - "apps/web/src/features/governance/registry/signed-protocol-screen.tsx"
  - "apps/web/src/features/governance/cases.test.tsx"
  - "apps/web/src/features/governance/registry/registry.test.tsx"
  - "apps/web/src/features/identity/reuse.test.tsx"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T103 — Use the generated hooks, their query keys and the shared fallback in the governance screens

## Why

[Review of 2026-11-17](../_review/review-2026-11-17.md), Q12, Q13. `['memberships']` is not the layout's key, so each screen reads memberships a second time into a cache the per-navigation re-read of AC-19 does not refresh; the screens show their own alert, with «спробувати ще раз» for an expired session.

**Satisfies:** [PRD §5](../PRD.md) AC-19, AC-26, AC-28 · **Blocked by:** —

## What

- `useListMemberships()`, `useGetSignedProtocol()`; the infinite queries build their keys from `getListCasesQueryKey` and `getSearchRegistryQueryKey`.
- Errors without a domain code go through `RequestFallback`; the screen's own `InlineAlert` stays for `governance.search_invalid` and `identity.association_disabled`.
- The reuse test scans `src/features/**`.

## Definition of Done

- [ ] Acceptance: The справи, реєстр and signed-protocol screens read memberships from the layout's cache entry and the cases once, an expired session leads to sign-in from each of them, and the reuse test fails on a hand-made query key anywhere under `src/features`.
- [ ] a test counts one memberships request for a справи visit
- [ ] 401 on each screen leads to `/sign-in`
- [ ] `mise run lint` and `mise run test` green
