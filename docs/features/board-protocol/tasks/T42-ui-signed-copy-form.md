---
id: T42
title: "Build the signed-scan upload with the meeting facts form"
layer: "ui"
deps: ["T41"]
acs: ["AC-14", "AC-34", "AC-15", "AC-40", "AC-41"]
files_hint:
  - "apps/web/src/features/governance/signed-copy-form.tsx"
  - "apps/web/src/features/governance/signed-copy-form.stories.tsx"
  - "apps/web/src/features/governance/protocol-screen.tsx"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T42 — Build the signed-scan upload with the meeting facts form

## Why

[ADR-0018](../adr/0018-enter-meeting-facts-with-the-signed-scan-and-refuse-lock-without-them.md); [sad §6 Flow 5](../sad.md).

**Satisfies:** [PRD §5](../PRD.md) AC-14, AC-34, AC-15, AC-40, AC-41 · **Blocked by:** [T41](./T41-ui-print-withdraw.md)

## What

- Form on lockProtocolWithSignedCopy (multipart); reuse Input, Textarea, Button, InlineAlert.

## Definition of Done

- [ ] Acceptance: Component tests show the form takes the scan, members present, a vote count per agenda item and the corrections flag with a required note, shows every missing fact at once (AC-34), is never pre-filled from the poll (AC-41), and is offered only in готовий до підпису (AC-40).
- [ ] stories for empty, errors, with corrections
- [ ] `mise run lint` and `mise run test` green