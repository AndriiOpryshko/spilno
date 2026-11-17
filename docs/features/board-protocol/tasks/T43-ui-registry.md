---
id: T43
title: "Build the реєстр протоколів search and the signed protocol view"
layer: "ui"
deps: ["T07"]
acs: ["AC-18", "AC-41", "AC-15", "AC-17"]
files_hint:
  - "apps/web/src/routes/registry.index.tsx"
  - "apps/web/src/routes/registry.$protocolId.tsx"
  - "apps/web/src/features/governance/registry/"
owner: "Andrii Opryshko"
estimate: "M"
status: "in_progress"
---

# T43 — Build the реєстр протоколів search and the signed protocol view

## Why

[sad §6 Flow 13](../sad.md).

**Satisfies:** [PRD §5](../PRD.md) AC-18, AC-41, AC-15, AC-17 · **Blocked by:** [T07](./T07-merge-contract-identity-skeleton.md)

## What

- Pages on searchRegistry and getSignedProtocol; reuse Card, Badge, Input, ProtocolStateBadge.
- Signed view explains immutability and links «нове рішення, що посилається на цей протокол» (AC-17, review finding 6).

## Definition of Done

- [ ] Acceptance: Component tests show search by counterparty, amount or range, subject and date period with cursor paging, and the signed view shows the signed copy, meeting facts, cited documents, the correction note beside the frozen text and poll history labelled informal. The signed view has no edit action and states that a locked protocol changes only through a new board decision that references it (AC-17, review finding 6).
- [ ] stories for results, empty, signed view with corrections
- [ ] component test: the signed view renders the AC-17 explanation and no edit, upload or withdraw control
- [ ] `mise run lint` and `mise run test` green

## Status (2026-11-17)

The search and the signed protocol view are built by association-onboarding [T79](../../association-onboarding/tasks/T79-ui-registry-screens.md)
(commits `a9b32bc`, `ae2aba8`) at `/o/$associationId/registry` and `/o/$associationId/registry/$protocolId`, with
component tests and stories (`apps/web/src/features/governance/registry/`): filters, cursor pages, the gap of a
withdrawn number, the files, the poll history labelled informal, the AC-17 explanation and no edit control.

Left here: the correction note is one section under the agenda, not beside the corrected item as AC-15 words
it. Open with the product owner.
