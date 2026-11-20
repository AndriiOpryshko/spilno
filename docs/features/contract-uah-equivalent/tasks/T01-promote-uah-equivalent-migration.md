---
id: T01
title: "Promote the migration that adds the hryvnia equivalent columns to contracts"
layer: "migration"
deps: []
acs: ["AC-01"]
files_hint:
  - "docs/features/contract-uah-equivalent/migrations/NNNNN_add_contracts_uah_equivalent.sql"
  - "services/core/migrations/"
  - "services/core/internal/finance/adapters/postgres/sqlcgen/"
owner: "Andrii Opryshko"
estimate: "S"
status: "todo"
---

# T01 — Promote the migration that adds the hryvnia equivalent columns to contracts

## Why

[data-model](../data-model.md) adds `uah_amount_kopiyky` and `uah_rate_micros` to `contracts` ([ADR-0002](../adr/0002-store-the-hryvnia-equivalent-on-the-contract-and-never-recompute-it.md)).

**Satisfies:** [PRD §5](../PRD.md) AC-01 · **Blocked by:** —

## What

- Promote [`NNNNN_add_contracts_uah_equivalent.sql`](../migrations/NNNNN_add_contracts_uah_equivalent.sql) with `mise run migrate:new add_contracts_uah_equivalent` (never hand-type the number) and copy the body.
- Run `mise run gen`; commit the regenerated `sqlcgen` with the migration.
- No query changes yet (T06).

## Definition of Done

- [ ] Acceptance: The staged migration is live under a goose number from `mise run migrate:new`, applies and reverts cleanly, `mise run gen` leaves sqlcgen with the two nullable columns and every existing test stays green.
- [ ] integration: the repo's migration test (`platform/postgres/migrate_test.go`) applies every migration up and down
- [ ] integration: a test reads the two columns of a fresh `contracts` row as NULL
- [ ] `mise run lint` and `mise run test` green
