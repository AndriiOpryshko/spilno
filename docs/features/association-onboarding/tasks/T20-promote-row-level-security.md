---
id: T20
title: "Promote row-level security for the governance, finance and identity tables"
layer: "migration"
deps: ["T12", "T13", "T17", "T19"]
acs: ["AC-04d", "AC-07", "AC-16b"]
files_hint:
  - "docs/features/association-onboarding/migrations/NNNNN_enable_rls_governance.sql"
  - "docs/features/association-onboarding/migrations/NNNNN_enable_rls_finance.sql"
  - "docs/features/association-onboarding/migrations/NNNNN_enable_rls_identity.sql"
  - "docs/features/association-onboarding/migrations/MANIFEST.md"
  - "services/core/internal/platform/db/rls_test.go"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T20 — Promote row-level security for the governance, finance and identity tables

## Why

[ADR-0003](../adr/0003-enforce-association-isolation-with-postgres-row-level-security.md): the database refuses another association's rows even when a query forgets its filter. [MANIFEST](../migrations/MANIFEST.md): these files ship together with the app that connects as `spilno_app` (T12) and runs every scoped request through `InAssociation` (T19) — enabled earlier, they would hide every row.

**Satisfies:** [PRD §5](../PRD.md) AC-04d, AC-07, AC-16b · **Blocked by:** [T12](./T12-two-database-roles.md), [T13](./T13-finance-contracts-association.md), [T17](./T17-promote-not-null-contracts.md), [T19](./T19-association-and-operator-guards.md)

## What

- Promote staged files 17–19 in manifest order; `/roundtrip`.
- `rls_test.go` on `testdb.NewTwoAssociations(t)`: walks the catalogue for every table with an `association_id` column, so a table added later is checked without editing the test.
- Move the integration tests of governance, finance and the association-scoped part of identity to the app-role pool.

## Definition of Done

- [ ] Acceptance: As spilno_app, every table with an association_id returns no rows without an association set and only that association's rows with one set, and the integration tests of all three contexts pass on the app-role pool.
- [ ] for each such table: no association set → 0 rows; A set → only A's rows; an insert for B while A is set is refused
- [ ] `board_member_revisions`: with A set and the actor's address, the actor's own entries in B are visible and nobody else's (`MembershipsOf`)
- [ ] the owner role still reads every row (goose, backfills)
- [ ] `/roundtrip` clean; `mise run test` green on the app-role pool
- [ ] `mise run lint` and `mise run test` green

## Notes

`FORCE ROW LEVEL SECURITY` applies to the owner too, so the migration role needs `BYPASSRLS` (a superuser in compose and testdb; header of the staged files). River jobs must open `InAssociation` from their payload when board-protocol T14 builds the worker.
