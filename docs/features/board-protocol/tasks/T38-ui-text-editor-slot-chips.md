---
id: T38
title: "Build the decision text editor with slot chips and highlighted typed facts"
layer: "ui"
deps: ["T37"]
acs: ["AC-29", "AC-28"]
files_hint:
  - "apps/web/src/features/governance/decision-text-editor.tsx"
  - "apps/web/src/features/governance/decision-text-editor.stories.tsx"
  - "apps/web/src/features/governance/protocol-screen.tsx"
owner: "Andrii Opryshko"
estimate: "L"
status: "todo"
---

# T38 — Build the decision text editor with slot chips and highlighted typed facts

## Why

[ADR-0011](../adr/0011-store-decision-text-as-fragments-with-slot-placeholders.md) (plain text + slot chips, Google Docs is never an editing surface); AC-28, AC-29.

**Satisfies:** [PRD §5](../PRD.md) AC-29, AC-28 · **Blocked by:** [T37](./T37-ui-protocol-on-hooks.md)

## What

- Editor built on Textarea + Badge chips; new component justified: no existing primitive edits text with inline chips.

## Definition of Done

- [ ] Acceptance: Component tests show the editor renders fragments with slot chips, saving calls saveAgendaItemWording, a typed number is highlighted with an offer to make it a slot (not blocked), and its stories pass the a11y check.
- [ ] component tests for chip rendering, save and highlight
- [ ] stories mirror the Figma variants
- [ ] `mise run lint` and `mise run test` green