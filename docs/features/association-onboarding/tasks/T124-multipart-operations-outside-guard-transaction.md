---
id: T124
title: "Run every multipart operation outside the guard's transaction and fail the build when one is missed"
layer: "wiring"
deps: []
acs: ["AC-25", "AC-28"]
files_hint:
  - "services/core/internal/identity/guard.go"
  - "services/core/internal/identity/middleware_outside_io_test.go"
  - "services/core/internal/governance/module.go"
  - "services/core/internal/governance/adapters/http/stubs.go"
  - "docs/features/association-onboarding/adr/0003-enforce-association-isolation-with-postgres-row-level-security.md"
  - "docs/features/association-onboarding/sad.md"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T124 — Run every multipart operation outside the guard's transaction and fail the build when one is missed

## Why

[Second re-review of 2026-11-17](../_review/review-2026-11-17-2.md), S04. The outside-I/O list of T93 is opt-in. `lockProtocolWithSignedCopy` (and `createAssistantTurn`, by the reviewer's read) takes a multipart body behind the guard and is not on it, so its body is buffered while a pool connection is held. A new multipart operation gets the same by default.

**Satisfies:** [PRD §5](../PRD.md) AC-25, AC-28 · **Blocked by:** —

## What

- Every operation under `{association_id}` whose request body is `multipart/form-data` in the contract is on the list.
- A test walks the contract and fails for one that is not.
- The amendment of ADR-0003 and the SAD §8 row name the rule, not three operations.

## Definition of Done

- [ ] Acceptance: A test that reads `services/core/api/openapi.yaml` fails when an association-scoped operation with a multipart body is missing from the outside-I/O list, and the pool test of T93 shows no connection held while the body of `lockProtocolWithSignedCopy` is read.
- [ ] spec-walking test in identity
- [ ] `middleware_outside_io_test.go`: the lock operation holds 0 connections while its body is read
- [ ] the refusals of a вимкнене ОСББ and of a person off the roster still precede the body (existing tests)
- [ ] ADR-0003 amendment and SAD §8 reworded
- [ ] `mise run lint` and `mise run test` green

## Notes

Not decided by the request's content type: a client chooses that header and could move any operation out of its transaction. The stubs answer 501 either way; board-protocol's handlers get their own transactions as `addCaseDocument` did.
