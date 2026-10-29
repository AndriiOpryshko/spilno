---
id: T05
title: "Promote the print-copy, meeting-facts, snapshot, trigger and integrity-run migrations"
layer: "migration"
deps: ["T04"]
acs: ["AC-14c", "AC-14", "AC-15", "AC-17"]
files_hint:
  - "docs/features/board-protocol/migrations/20261027120018_create_print_copies.up.sql"
  - "docs/features/board-protocol/migrations/20261027120018_create_print_copies.down.sql"
  - "docs/features/board-protocol/migrations/20261027120019_create_meeting_facts.up.sql"
  - "docs/features/board-protocol/migrations/20261027120019_create_meeting_facts.down.sql"
  - "docs/features/board-protocol/migrations/20261027120020_create_protocol_snapshots.up.sql"
  - "docs/features/board-protocol/migrations/20261027120020_create_protocol_snapshots.down.sql"
  - "docs/features/board-protocol/migrations/20261027120021_create_protocol_immutability_triggers.up.sql"
  - "docs/features/board-protocol/migrations/20261027120021_create_protocol_immutability_triggers.down.sql"
  - "docs/features/board-protocol/migrations/20261027120022_create_integrity_runs.up.sql"
  - "docs/features/board-protocol/migrations/20261027120022_create_integrity_runs.down.sql"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T05 — Promote the print-copy, meeting-facts, snapshot, trigger and integrity-run migrations

## Why

Lock and withdrawal write immutable fingerprinted snapshots and no protocol is ever deleted ([ADR-0007](../adr/0007-lock-protocols-as-fingerprinted-immutable-snapshots.md), PRD §6 Integrity, Retention).

**Satisfies:** [PRD §5](../PRD.md) AC-14c, AC-14, AC-15, AC-17 · **Blocked by:** [T04](./T04-stage-assistant-poll-river-migrations.md)

## What

- Promote staged pairs 18–22.
- Wrap trigger functions in `-- +goose StatementBegin` / `-- +goose StatementEnd`.

## Definition of Done

- [ ] Acceptance: print_copies, meeting_facts, protocol_snapshots, registry_entries, the no-update/no-delete triggers and integrity_runs are live; an integration test proves UPDATE/DELETE on a snapshot and DELETE on a protocol are rejected.
- [ ] the goose files apply and revert cleanly (`/roundtrip`)
- [ ] an integration test on `testdb.New(t)` shows UPDATE and DELETE on `protocol_snapshots` and DELETE on `protocols` fail
- [ ] `mise run lint` and `mise run test` green

## Notes

Migration lane. The triggers are the one documented exception to «No CREATE TRIGGER» (.claude/rules/migrations.md).
