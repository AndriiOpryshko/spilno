---
id: T66
title: "Point «Увійти» at the API and build every API link from the generated URL helpers"
layer: "ui"
deps: []
acs: ["AC-04d", "AC-05"]
files_hint:
  - "apps/web/src/features/identity/shell/sign-in-page.tsx"
  - "apps/web/src/features/identity/shell/shell.test.tsx"
  - "apps/web/src/features/identity/storage/storage-line.tsx"
  - "apps/web/src/features/identity/association/extract-href.ts"
  - "apps/web/src/features/identity/association/association-card.tsx"
  - "apps/web/src/features/identity/association/roster-decision-mark.tsx"
owner: "Andrii Opryshko"
estimate: "S"
status: "done"
---

# T66 — Point «Увійти» at the API and build every API link from the generated URL helpers

## Why

[Review of 2026-11-10](../_review/review-2026-11-10.md), F02, W5 (hand-built URLs). [PRD](../PRD.md) AC-05 and AC-04d start with «when they sign in». The button links to `/identity/sign-in`, outside the `/api` proxy, so a browser lands on the SPA's Not Found; the unit test asserts that wrong href.

**Satisfies:** [PRD §5](../PRD.md) AC-04d, AC-05 · **Blocked by:** —

## What

- `sign-in-page.tsx` uses `getStartSignInUrl()`; `storage-line.tsx` uses `getStartStorageConnectionUrl()`; `extract-href.ts` uses `getDownloadConfirmationExtractUrl()` or is removed.
- `shell.test.tsx` asserts the generated URL.

## Definition of Done

- [ ] Acceptance: A component test shows the sign-in link equal to `getStartSignInUrl()`, and no file outside `api/gen` spells an API path by hand.
- [ ] `grep` finds no `'/api/` or `"/identity/` literal under `src/` outside `api/gen`
- [ ] the click-through proof is in T81
- [ ] `mise run lint` and `mise run test` green

## Notes

Reuses `Button`.
