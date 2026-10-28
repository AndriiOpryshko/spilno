---
id: T35
title: "Build the association settings and roster screens"
layer: "ui"
deps: ["T07"]
acs: ["AC-42", "AC-02"]
files_hint:
  - "apps/web/src/routes/settings.association.tsx"
  - "apps/web/src/routes/settings.roster.tsx"
  - "apps/web/src/features/governance/settings/"
owner: "Andrii Opryshko"
estimate: "L"
status: "todo"
---

# T35 — Build the association settings and roster screens

## Why

[sad §6 Flow 12](../sad.md); AC-42 links here from the draft refusal.

**Satisfies:** [PRD §5](../PRD.md) AC-42, AC-02 · **Blocked by:** [T07](./T07-merge-contract-identity-skeleton.md)

## What

- Forms on generated hooks (getAssociationSettings, saveAssociationSettings, board-member operations).
- Reuse Input, Textarea, Button, Card, InlineAlert; no new primitives.

## Definition of Done

- [ ] Acceptance: Component tests show the settings form and the roster form render every 422 `errors[]` entry next to its field at once (AC-02), the register start field is read-only after `register_start_locked`, and stories pass the a11y check.
- [ ] component tests for multi-field errors and the locked start
- [ ] stories for empty, filled, error states
- [ ] `mise run lint` and `mise run test` green