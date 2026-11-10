---
id: T14
title: "Serve the roster facade: the current member, a person's memberships and the members' sign-in addresses"
layer: "infra"
deps: ["T04", "T05", "T07"]
acs: ["AC-04d", "AC-07", "AC-09", "AC-28b"]
files_hint:
  - "services/core/internal/governance/adapters/postgres/"
  - "services/core/internal/governance/app/roster.go"
  - "services/core/internal/governance/ports/ports.go"
  - "services/core/internal/governance/module.go"
  - "services/core/internal/governance/governancetest/"
owner: "Andrii Opryshko"
estimate: "L"
status: "todo"
---

# T14 — Serve the roster facade: the current member, a person's memberships and the members' sign-in addresses

## Why

[sad](../sad.md) §5 «Context ownership»: governance provides these to identity through `module.go`; [ADR-0004](../adr/0004-let-any-verified-person-sign-in-and-derive-memberships-from-rosters.md): memberships come from the rosters in force today; [data-model](../data-model.md) «Query -> index map» (`CurrentMember`, `MembershipsOf`, `MemberSignInAddresses`).

**Satisfies:** [PRD §5](../PRD.md) AC-04d, AC-07, AC-09, AC-28b · **Blocked by:** [T04](./T04-in-association-transaction.md), [T05](./T05-promote-expand-migrations.md), [T07](./T07-roster-position-domain.md)

## What

- sqlc queries over `board_members` and `board_member_revisions` (the latest revision per member, the term covering the date): `CurrentMember(ctx, association, email, date)`, `MembershipsOf(ctx, email, date)`, `MemberSignInAddresses(ctx, association)`; revisions read and write `position` and `decision_id`.
- Repositories take the runner from `db.From(ctx)`, the meetings repository included; the three functions are exported from `governance/module.go`.
- Fixture `governancetest.Member(t, tx, association, email, position)` with a term covering today.

## Definition of Done

- [ ] Acceptance: governance exposes CurrentMember, MembershipsOf and MemberSignInAddresses over Postgres on the transaction in ctx, and integration tests prove term boundaries, a person on two rosters and the stored position.
- [ ] a term that ended yesterday is not current; a term that ends today is
- [ ] the same address on the rosters of A and B gives two memberships, each with its position
- [ ] `MemberSignInAddresses` returns the addresses of the current members of one association only
- [ ] a revision written without a position fails in Go before it reaches the database
- [ ] `mise run lint` and `mise run test` green

## Notes

[board-protocol T10](../../board-protocol/tasks/T10-settings-roster-endpoints.md) also declares the roster repository and `CurrentMember`. Whichever task lands first creates `adapters/postgres/board_members.go`; the other extends it. T10's handlers must map `position` and `decision_id` into `BoardMember` and write `member` for the entries the board adds. Email comparison is case-insensitive.
