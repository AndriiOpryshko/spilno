---
id: T62
title: "Serve the roster: list, add and change члени правління with their position and decision"
layer: "ports"
deps: ["T51", "T58", "T59"]
acs: ["AC-09", "AC-16b", "AC-20", "AC-25"]
files_hint:
  - "services/core/internal/governance/app/roster.go"
  - "services/core/internal/governance/adapters/postgres/roster.go"
  - "services/core/internal/governance/adapters/postgres/roster_write.go"
  - "services/core/internal/governance/adapters/postgres/queries/roster.sql"
  - "services/core/internal/governance/adapters/http/board_members.go"
  - "services/core/internal/governance/adapters/http/stubs.go"
  - "services/core/internal/governance/module.go"
  - "services/core/internal/identity/stubs_test.go"
owner: "Andrii Opryshko"
estimate: "L"
status: "done"
---

# T62 — Serve the roster: list, add and change члени правління with their position and decision

## Why

[Review of 2026-11-10](../_review/review-2026-11-10.md), F15. [PRD](../PRD.md) AC-09 (the approved голова adds the other члени правління), AC-20 (the roster shows the decision behind each голова) and AC-16b; [sad](../sad.md) §6 Flows 2 and 12. `ListBoardMembers`, `CreateBoardMember` and `UpdateBoardMember` answer 501.

**Satisfies:** [PRD §5](../PRD.md) AC-09, AC-16b, AC-20, AC-25 · **Blocked by:** [T51](./T51-carry-term-end-end-previous-head.md), [T58](./T58-storage-check-injection-and-pending-gauge.md), [T59](./T59-index-lower-sign-in-email.md)

## What

- The roster half of [board-protocol T10](../../board-protocol/tasks/T10-settings-roster-endpoints.md): use cases, queries and handlers for the three board-member operations on the existing roster repository and `roster_write.go`.
- Entries carry `position`, `decision_id`, the term and the sign-in address; a seat created by an approval cannot have its `position` or `decision_id` rewritten here.
- The settings operations stay stubs for board-protocol T10.

## Definition of Done

- [ ] Acceptance: Handler tests show the approved голова listing the roster with `position` and `decision_id` on each entry, adding and changing a член правління, a 422 that names every broken field, and the three operations refused for a person off the roster and for a вимкнене ОСББ.
- [ ] integration tests with `testdb.NewTwoAssociations(t)`: rows of the second association are never returned
- [ ] 422 `governance.roster_entry_invalid` names every field at once
- [ ] create and update answer 403 `identity.association_disabled` while disabled; list is refused too (only the registry reads are open, ADR-0007)
- [ ] an оператор on the roster lists it as a member; off the roster gets `identity.not_current_member`
- [ ] `mise run lint` and `mise run test` green

## Notes

Carries the roster half of board-protocol T10; that task keeps the settings half and `governance.register_start_locked`.
