---
id: T119
title: "Read memberships through one hook everywhere and stop the page loop on a repeated cursor"
layer: "ui"
deps: []
acs: ["AC-04d", "AC-19"]
files_hint:
  - "apps/web/src/features/identity/shell/use-all-memberships.ts"
  - "apps/web/src/features/governance/protocol-screen.tsx"
  - "apps/web/src/features/identity/storage/storage-line.tsx"
  - "apps/web/src/features/governance/cases-screen.tsx"
  - "apps/web/src/features/governance/registry/registry-screen.tsx"
  - "apps/web/src/features/governance/registry/signed-protocol-screen.tsx"
  - "apps/web/src/features/governance/roster/roster-screen.tsx"
  - "apps/web/src/features/identity/shell/association-shell.test.tsx"
  - "apps/web/src/features/identity/storage/storage-line.test.tsx"
  - "apps/web/src/features/identity/reuse.test.tsx"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T119 — Read memberships through one hook everywhere and stop the page loop on a repeated cursor

## Why

[Second re-review of 2026-11-17](../_review/review-2026-11-17-2.md), R06, S12. `useAllMemberships` reads every page into the generated list key; `ProtocolScreen` and `CardLinks` read the same key with the first-page query function, so the entry can fall back to one page and the layout then refuses a member of a 21st association. The page loop has no guard against a cursor that repeats.

**Satisfies:** [PRD §5](../PRD.md) AC-04d, AC-19 · **Blocked by:** —

## What

- Every reader of memberships goes through `useAllMemberships`; the reuse test refuses a plain `useListMemberships(` outside it.
- The loop stops with an error when a cursor repeats or a page cap is passed, so the layout shows its fallback.

## Definition of Done

- [ ] Acceptance: With 21 memberships, reloading the protocol page of the 21st association and following the storage callback's link both end inside that association, not on «no access»; a server that repeats a cursor ends in the shared fallback after a bounded number of requests.
- [ ] `association-shell.test.tsx`: the 21-association case on the protocol route and after the storage callback
- [ ] test: a repeated cursor ends the loop with an error
- [ ] `reuse.test.tsx`: no `useListMemberships(` outside `use-all-memberships.ts`
- [ ] `mise run lint` and `mise run test` green

## Notes

Starts by reproducing R06 in the test: the review verified it by reading only. If it does not reproduce, the review record says so and the task keeps the single reader and the guard.
