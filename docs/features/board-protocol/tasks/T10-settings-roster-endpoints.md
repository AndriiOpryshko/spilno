---
id: T10
title: "Serve settings and roster: use cases, Postgres repositories, handlers and CurrentMember"
layer: "ports"
deps: ["T02", "T06", "T07", "T09"]
acs: ["AC-42", "AC-02"]
files_hint:
  - "services/core/internal/governance/app/settings.go"
  - "services/core/internal/governance/app/roster.go"
  - "services/core/internal/governance/adapters/postgres/settings.go"
  - "services/core/internal/governance/adapters/postgres/board_members.go"
  - "services/core/internal/governance/adapters/postgres/queries/settings.sql"
  - "services/core/internal/governance/adapters/postgres/queries/board_members.sql"
  - "services/core/internal/governance/adapters/http/settings.go"
  - "services/core/internal/governance/adapters/http/board_members.go"
  - "services/core/internal/governance/module.go"
owner: "Andrii Opryshko"
estimate: "L"
status: "todo"
---

# T10 — Serve settings and roster: use cases, Postgres repositories, handlers and CurrentMember

## Why

[sad §6 Flow 12](../sad.md); [ADR-0015](../adr/0015-scope-owned-rows-to-one-association-with-explicit-query-filters.md) (every query filters by `association_id`); `CurrentMember(ctx, email, date)` is what identity asks ([sad §5](../sad.md) context interactions).

**Satisfies:** [PRD §5](../PRD.md) AC-42, AC-02 · **Blocked by:** [T02](./T02-stage-roster-settings-migrations.md), [T06](./T06-problem-codes-registry.md), [T07](./T07-merge-contract-identity-skeleton.md), [T09](./T09-settings-roster-domain.md)

## What

- Use cases, sqlc queries and repositories for settings revisions and roster entries (+ `board_member_revisions`).
- Handlers replacing the T07 stubs; `module.go` exposes `CurrentMember`.

## Definition of Done

- [ ] Acceptance: Handler tests show getAssociationSettings / saveAssociationSettings / list/create/updateBoardMember return the contract shapes, 422 `governance.settings_invalid` / `governance.roster_entry_invalid` name every field (AC-02), and 409 `governance.register_start_locked` after the first number.
- [ ] integration tests on `testdb.New(t)` for each operation incl. AC-02 and AC-42 errors
- [ ] isolation test: rows of a second association are never returned ([ADR-0015](../adr/0015-scope-owned-rows-to-one-association-with-explicit-query-filters.md))
- [ ] `CurrentMember` returns the person only when a term covers the date
- [ ] `mise run lint` and `mise run test` green