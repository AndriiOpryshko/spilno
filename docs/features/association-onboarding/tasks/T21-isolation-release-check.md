---
id: T21
title: "Add the release-blocking isolation test over every association-scoped operation of the contract"
layer: "tests"
deps: ["T02", "T20"]
acs: ["AC-04d", "AC-07", "AC-16b"]
files_hint:
  - "services/core/internal/platform/isolation_test.go"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T21 — Add the release-blocking isolation test over every association-scoped operation of the contract

## Why

[PRD](../PRD.md) §6 «Isolation» (release-blocking) and §9 rollout step 1; [sad](../sad.md) §10 QG-1.

**Satisfies:** [PRD §5](../PRD.md) AC-04d, AC-07, AC-16b · **Blocked by:** [T02](./T02-merge-contract-delta.md), [T20](./T20-promote-row-level-security.md)

## What

- `isolation_test.go` builds the router as `cmd/api` does on `testdb.NewTwoAssociations(t)`, reads the embedded spec and takes every operation with an `{association_id}` parameter — a new one is picked up without editing the test.
- Persons: on A's roster only; on both rosters; the оператор's address on A's roster only. Each calls every operation with B's id and with A's id.

## Definition of Done

- [ ] Acceptance: A test in mise run test calls every operation of the contract that has {association_id} with another association's id as three seeded persons and fails on any record of B returned to a person off B's roster.
- [ ] a person off B's roster gets 403 `identity.not_current_member` on every operation under B and no field of B in any body
- [ ] the person on both rosters is not refused by the guard under either id
- [ ] the оператор on A only is refused under B like anyone
- [ ] the test is part of `mise run test`, which gates the release
- [ ] `mise run lint` and `mise run test` green

## Notes

Stubbed governance operations answer 501 after the guard; the test asserts on the guard, so it holds before and after board-protocol fills them in.
