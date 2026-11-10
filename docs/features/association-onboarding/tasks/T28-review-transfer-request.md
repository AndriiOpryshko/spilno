---
id: T28
title: "Show competing запити and the storage account in the review of a запит"
layer: "ports"
deps: ["T24", "T27"]
acs: ["AC-19b", "AC-28b"]
files_hint:
  - "services/core/internal/identity/app/"
  - "services/core/internal/identity/ports/ports.go"
  - "services/core/internal/identity/adapters/postgres/"
  - "services/core/internal/identity/adapters/http/"
owner: "Andrii Opryshko"
estimate: "S"
status: "todo"
---

# T28 — Show competing запити and the storage account in the review of a запит

## Why

[PRD](../PRD.md) AC-19b, AC-28b («the оператор sees which account holds the files»); [ADR-0003](../adr/0003-enforce-association-isolation-with-postgres-row-level-security.md): the only association-owned value an operator use case reads is the storage account address.

**Satisfies:** [PRD §5](../PRD.md) AC-19b, AC-28b · **Blocked by:** [T24](./T24-operator-queue-review-check.md), [T27](./T27-file-transfer-request.md)

## What

- `getApplicationForReview` for a transfer: `competing_application_ids` from the pending запити of the association; `storage_account_email` from `google_connections`, read inside `InAssociation` for that association; `null` when no storage is connected.

## Definition of Done

- [ ] Acceptance: The review of a запит на передачу lists the other pending запити of the same association and the address of the account that holds its files, read inside InAssociation and nothing else.
- [ ] two pending запити for one association each list the other
- [ ] a connected storage → its address; none → `null`; for a заявка the field is `null`
- [ ] the same оператор, off that roster, is still refused on every association path (asserted next to it)
- [ ] `mise run lint` and `mise run test` green

## Notes

No general bypass: one column, one query (ADR-0003, ADR-0009).
