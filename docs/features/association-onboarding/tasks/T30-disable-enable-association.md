---
id: T30
title: "Disable and enable an association as the оператор"
layer: "ports"
deps: ["T11", "T19"]
acs: ["AC-16", "AC-21", "AC-22", "AC-23", "AC-24"]
files_hint:
  - "services/core/internal/identity/app/"
  - "services/core/internal/identity/ports/ports.go"
  - "services/core/internal/identity/adapters/postgres/"
  - "services/core/internal/identity/adapters/http/"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T30 — Disable and enable an association as the оператор

## Why

[sad](../sad.md) §6 Flow 9; [ADR-0007](../adr/0007-enforce-a-disabled-association-as-read-only-in-the-identity-middleware.md): the state and an insert-only history; [PRD](../PRD.md) US-06.

**Satisfies:** [PRD §5](../PRD.md) AC-16, AC-21, AC-22, AC-23, AC-24 · **Blocked by:** [T11](./T11-association-state-access-matrix.md), [T19](./T19-association-and-operator-guards.md)

## What

- `listAssociationsForOperator`, `disableAssociation`, `enableAssociation` with the rules of T11.
- Queries `ListAssociations`, `GetAssociation`, `UpdateAssociationState`, `CreateStateChange`, `ListStateChanges`; state and its change row in one transaction.

## Definition of Done

- [ ] Acceptance: The оператор lists associations, disables one with a reason and enables it again, each change an insert-only row with reason, оператор and time, and the refusals of AC-22 and AC-24 record nothing.
- [ ] AC-21: `disabled`, with a state change holding the reason, the оператор and the time
- [ ] AC-22: an empty reason → 422 `identity.reason_required`; still `connected`, no row
- [ ] AC-23: `connected` again; the enabling is a row; the association's other tables are row-for-row as before
- [ ] AC-24: the same state → 409 `identity.association_state_unchanged`, no row; AC-16: anyone else → 403 on all three
- [ ] `mise run lint` and `mise run test` green

## Notes

The list reads `associations` only (no policy on it); it never touches an association-owned table.
