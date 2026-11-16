---
id: T79
title: "Build the реєстр протоколів: search and the signed protocol with its download"
layer: "ui"
deps: ["T63", "T64", "T68"]
acs: ["AC-09", "AC-26"]
files_hint:
  - "apps/web/src/routes/o.$associationId.registry.index.tsx"
  - "apps/web/src/routes/o.$associationId.registry.$protocolId.tsx"
  - "apps/web/src/features/governance/registry/"
  - "apps/web/src/components/app-shell.tsx"
  - "apps/web/src/i18n/uk.ts"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T79 — Build the реєстр протоколів: search and the signed protocol with its download

## Why

[Review of 2026-11-10](../_review/review-2026-11-10.md), F15. [PRD](../PRD.md) AC-26 and AC-09: a member opens the registry and downloads a signed protocol, in a вимкнене ОСББ too. No screen opens the registry.

**Satisfies:** [PRD §5](../PRD.md) AC-09, AC-26 · **Blocked by:** [T63](./T63-serve-registry-read.md), [T64](./T64-serve-file-download-read-only-check.md), [T68](./T68-ui-association-shell-card-notice.md)

## What

- Routes `/o/$associationId/registry` and `/o/$associationId/registry/$protocolId`, linked from the sidebar, on `searchRegistry` and `getSignedProtocol`.
- The download link comes from the generated URL helper of `downloadFile`.
- The scope of [board-protocol T43](../../board-protocol/tasks/T43-ui-registry.md) is kept, including the note that a locked protocol changes only through a new decision.

## Definition of Done

- [ ] Acceptance: Component tests show the registry searched with cursor pages, an empty registry explained, and the signed protocol view with a download link and no edit, upload or withdraw control; the same screens render under the suspended notice.
- [ ] stories for results, empty and the signed view pass a11y
- [ ] test: `search_invalid` is shown next to the filter
- [ ] `mise run lint` and `mise run test` green

## Notes

Reuses `Card`, `Badge`, `Input`, `ProtocolStateBadge`. Carries board-protocol T43.
