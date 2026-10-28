---
id: T02
title: "Promote the roster, settings and meetings-chair migrations"
layer: "migration"
deps: ["T01"]
acs: ["AC-42", "AC-02"]
files_hint:
  - "docs/features/board-protocol/migrations/20261027120005_create_board_members.up.sql"
  - "docs/features/board-protocol/migrations/20261027120005_create_board_members.down.sql"
  - "docs/features/board-protocol/migrations/20261027120006_create_association_settings.up.sql"
  - "docs/features/board-protocol/migrations/20261027120006_create_association_settings.down.sql"
  - "docs/features/board-protocol/migrations/20261027120007_add_association_and_chair_to_meetings.up.sql"
  - "docs/features/board-protocol/migrations/20261027120007_add_association_and_chair_to_meetings.down.sql"
  - "docs/features/board-protocol/migrations/20261027120011_create_idx_meetings_chair.up.sql"
  - "docs/features/board-protocol/migrations/20261027120011_create_idx_meetings_chair.down.sql"
owner: "Andrii Opryshko"
estimate: "S"
status: "todo"
---

# T02 — Promote the roster, settings and meetings-chair migrations

## Why

The roster with terms decides who is a current член правління (AC-16, AC-19); settings hold the meeting form, signers and register start (AC-42, [ADR-0017](../adr/0017-reserve-the-protocol-number-at-first-print-and-never-reuse-it.md)). Tables per [data-model.md](../data-model.md) §roster and settings.

**Satisfies:** [PRD §5](../PRD.md) AC-42, AC-02 · **Blocked by:** [T01](./T01-stage-identity-migrations.md)

## What

- Promote 05 and 06 as they are.
- From 07 promote only `chair_board_member_id` (+ its FK): `association_id` on meetings is already live (00005–00007); 08–10 are live too — skip them.
- Promote 11 as its own `-- +goose NO TRANSACTION` file (CONCURRENTLY index on an existing table).

## Definition of Done

- [ ] Acceptance: board_members, association_settings, protocol_registers and meetings.chair_board_member_id with its index are live and roundtrip cleanly.
- [ ] the new goose files apply and revert cleanly (`/roundtrip`)
- [ ] the chair index file holds exactly one `CREATE INDEX CONCURRENTLY` statement
- [ ] existing meetings tests stay green
- [ ] `mise run lint` and `mise run test` green

## Notes

Migration lane. The staged 07 file is partially live — edit the promoted copy, not the staged pair.
