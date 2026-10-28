---
id: T20
title: "Persist the protocol aggregate with a revision guard"
layer: "infra"
deps: ["T03", "T16"]
acs: ["AC-29", "AC-05d", "AC-03"]
files_hint:
  - "services/core/internal/governance/adapters/postgres/protocols.go"
  - "services/core/internal/governance/adapters/postgres/queries/protocols.sql"
  - "services/core/internal/governance/adapters/postgres/queries/agenda_items.sql"
  - "services/core/internal/governance/adapters/postgres/queries/slots.sql"
  - "services/core/internal/governance/ports/protocols.go"
owner: "Andrii Opryshko"
estimate: "L"
status: "todo"
---

# T20 — Persist the protocol aggregate with a revision guard

## Why

[ADR-0009](../adr/0009-track-draft-lifecycle-with-states-and-revision-numbers.md) (revision guard), [ADR-0015](../adr/0015-scope-owned-rows-to-one-association-with-explicit-query-filters.md) (explicit association filter), [data-model.md](../data-model.md) §protocol aggregate and §Indexes.

**Satisfies:** [PRD §5](../PRD.md) AC-29, AC-05d, AC-03 · **Blocked by:** [T03](./T03-stage-case-protocol-migrations.md), [T16](./T16-agenda-slots-domain.md)

## What

- Repository port + Postgres adapter; `UPDATE … WHERE revision = $n` guard.
- Load query feeds getProtocol (meeting date, roster in force on that date, case documents — AC-03).

## Definition of Done

- [ ] Acceptance: Integration tests on `testdb.New(t)` save and load a protocol with agenda items, slots, values and confirmations, reject a write with a stale revision as `governance.draft_revision_conflict`, record each transition and never return another association's rows.
- [ ] integration tests incl. stale revision and second-association isolation
- [ ] each sqlc query is served by an index named in [data-model.md](../data-model.md)
- [ ] `mise run lint` and `mise run test` green