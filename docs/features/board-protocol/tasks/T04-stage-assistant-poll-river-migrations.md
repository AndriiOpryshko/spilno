---
id: T04
title: "Promote the assistant-turn and poll migrations and add the River schema"
layer: "migration"
deps: ["T03"]
acs: ["AC-26", "AC-30", "AC-31"]
files_hint:
  - "docs/features/board-protocol/migrations/20261027120015_create_assistant_turns.up.sql"
  - "docs/features/board-protocol/migrations/20261027120015_create_assistant_turns.down.sql"
  - "docs/features/board-protocol/migrations/20261027120017_create_poll_rounds.up.sql"
  - "docs/features/board-protocol/migrations/20261027120017_create_poll_rounds.down.sql"
  - "services/core/go.mod"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T04 — Promote the assistant-turn and poll migrations and add the River schema

## Why

Conversations and proposals (AC-26) and informal poll rounds (AC-30, AC-31). River's tables come into goose so goose stays the only migrator ([ADR-0006](../adr/0006-run-background-work-in-worker-via-river-queue.md), [data-model.md](../data-model.md) §Migrations).

**Satisfies:** [PRD §5](../PRD.md) AC-26, AC-30, AC-31 · **Blocked by:** [T03](./T03-stage-case-protocol-migrations.md)

## What

- Promote staged pair 17 (15 went with T03: 16 depends on it).
- Pin the River version in `go.mod`; generate its schema with `river migrate-get --all --up` / `--down` into one goose file.

## Definition of Done

- [ ] Acceptance: assistant_turns, poll_rounds, poll_votes and River's schema (as one goose file from the pinned River version) are live and roundtrip cleanly.
- [ ] the goose files apply and revert cleanly (`/roundtrip`)
- [ ] the River file reproduces `river migrate-get` output for the pinned version (comment names the version)
- [ ] `mise run lint` and `mise run test` green

## Notes

Migration lane. Every River upgrade needs a new mirrored goose file ([sad §11](../sad.md) accepted debt).
