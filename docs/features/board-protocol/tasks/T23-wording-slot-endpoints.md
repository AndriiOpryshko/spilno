---
id: T23
title: "Edit wording and слоти рішення: save wording, add, confirm and correct slots"
layer: "ports"
deps: ["T17", "T22"]
acs: ["AC-29", "AC-21", "AC-28b", "AC-09", "AC-05d", "AC-14b"]
files_hint:
  - "services/core/internal/governance/app/editing.go"
  - "services/core/internal/governance/adapters/http/wording.go"
  - "services/core/internal/governance/adapters/http/slots.go"
owner: "Andrii Opryshko"
estimate: "L"
status: "todo"
---

# T23 — Edit wording and слоти рішення: save wording, add, confirm and correct slots

## Why

[sad §6](../sad.md) Flows 1 and 3; [ADR-0009](../adr/0009-track-draft-lifecycle-with-states-and-revision-numbers.md), [ADR-0011](../adr/0011-store-decision-text-as-fragments-with-slot-placeholders.md).

**Satisfies:** [PRD §5](../PRD.md) AC-29, AC-21, AC-28b, AC-09, AC-05d, AC-14b · **Blocked by:** [T17](./T17-text-fragments-proposal-rules.md), [T22](./T22-open-draft-endpoints.md)

## What

- Handlers for saveAgendaItemWording, addDecisionSlot, confirmDecisionSlot, correctDecisionSlot with `protocol_revision`.
- Subscriber to the T21 document-version event → needs review + return to draft.

## Definition of Done

- [ ] Acceptance: Handler tests show saveAgendaItemWording keeps author and time for any current член правління (AC-29), confirm/correctDecisionSlot record person and time, any edit on готовий до голосування or готовий до підпису returns the draft to чернетка (AC-05d, AC-14b), and a newer case document marks citing slots needs_review (AC-09).
- [ ] test: a person-typed number in wording is returned as a highlight, not an error
- [ ] `mise run lint` and `mise run test` green