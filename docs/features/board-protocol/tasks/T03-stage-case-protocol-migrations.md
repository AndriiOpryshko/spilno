---
id: T03
title: "Promote the case, protocol, agenda-item and decision-slot migrations"
layer: "migration"
deps: ["T02"]
acs: ["AC-26", "AC-03", "AC-03c", "AC-21", "AC-29"]
files_hint:
  - "docs/features/board-protocol/migrations/20261027120012_create_cases.up.sql"
  - "docs/features/board-protocol/migrations/20261027120012_create_cases.down.sql"
  - "docs/features/board-protocol/migrations/20261027120013_create_protocols.up.sql"
  - "docs/features/board-protocol/migrations/20261027120013_create_protocols.down.sql"
  - "docs/features/board-protocol/migrations/20261027120014_create_agenda_items.up.sql"
  - "docs/features/board-protocol/migrations/20261027120014_create_agenda_items.down.sql"
  - "docs/features/board-protocol/migrations/20261027120016_create_decision_slots.up.sql"
  - "docs/features/board-protocol/migrations/20261027120016_create_decision_slots.down.sql"
owner: "Andrii Opryshko"
estimate: "S"
status: "todo"
---

# T03 — Promote the case, protocol, agenda-item and decision-slot migrations

## Why

The protocol aggregate and its case sources ([data-model.md](../data-model.md) §cases and files, §protocol aggregate; [ADR-0009](../adr/0009-track-draft-lifecycle-with-states-and-revision-numbers.md), [ADR-0011](../adr/0011-store-decision-text-as-fragments-with-slot-placeholders.md)).

**Satisfies:** [PRD §5](../PRD.md) AC-26, AC-03, AC-03c, AC-21, AC-29 · **Blocked by:** [T02](./T02-stage-roster-settings-migrations.md)

## What

- Promote staged pairs 12, 13, 14, 16 in that order (15 depends on nothing here and goes with T04).

## Definition of Done

- [ ] Acceptance: files, cases, case_documents, protocols, protocol_transitions, agenda_items(+versions) and decision_slots(+values, confirmations) are live and roundtrip cleanly.
- [ ] the goose files apply and revert cleanly (`/roundtrip`)
- [ ] every FK column has its index in the same or the next file (.claude/rules/migrations.md)
- [ ] `mise run lint` and `mise run test` green

## Notes

Migration lane.
