---
id: T13
title: "Scope finance contracts to the association in the path"
layer: "app"
deps: ["T01", "T04", "T05"]
acs: ["AC-04d", "AC-07"]
files_hint:
  - "services/core/internal/finance/"
owner: "Andrii Opryshko"
estimate: "S"
status: "todo"
---

# T13 — Scope finance contracts to the association in the path

## Why

[ADR-0003](../adr/0003-enforce-association-isolation-with-postgres-row-level-security.md) migration step 2 and [ADR-0002](../adr/0002-put-the-association-id-in-every-association-scoped-path.md): `finance.contracts` gets `association_id`; the app must write it before the backfill (T16) and the `NOT NULL` (T17).

**Satisfies:** [PRD §5](../PRD.md) AC-04d, AC-07 · **Blocked by:** [T01](./T01-move-paths-under-association.md), [T04](./T04-in-association-transaction.md), [T05](./T05-promote-expand-migrations.md)

## What

- `finance.sql`: `CreateContract` writes `association_id`; `GetContract` and `ListContracts` filter on it (`idx_contracts_association_created` arrives with T17).
- The repository takes its runner from `db.From(ctx)` and falls back to the pool until T19 opens the transaction; the handler passes the path's association.

## Definition of Done

- [ ] Acceptance: A contract created under association A is written with its association_id and is neither listed nor opened under association B.
- [ ] integration test: create under A; list under B is empty; get under B answers 404 with the same body as a missing id
- [ ] `mise run gen:check` clean
- [ ] `mise run lint` and `mise run test` green

## Notes

Explicit filter kept as the second line next to row-level security (board-protocol ADR-0015).
