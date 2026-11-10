---
id: T26
title: "Approve a заявка на підключення: create the association and seat its first голова in one transaction"
layer: "ports"
deps: ["T09", "T15", "T20", "T24"]
acs: ["AC-03", "AC-03b", "AC-04c", "AC-08", "AC-09", "AC-12", "AC-15", "AC-16c"]
files_hint:
  - "services/core/internal/identity/app/"
  - "services/core/internal/identity/ports/ports.go"
  - "services/core/internal/identity/adapters/postgres/"
  - "services/core/internal/identity/adapters/http/"
owner: "Andrii Opryshko"
estimate: "L"
status: "todo"
---

# T26 — Approve a заявка на підключення: create the association and seat its first голова in one transaction

## Why

[sad](../sad.md) §6 Flow 2; [ADR-0006](../adr/0006-apply-approvals-to-the-roster-through-a-governance-facade-in-one-transaction.md): one transaction across identity and the governance facade; [ADR-0003](../adr/0003-enforce-association-isolation-with-postgres-row-level-security.md): the approval opens `InAssociation` with the id of the association it creates.

**Satisfies:** [PRD §5](../PRD.md) AC-03, AC-03b, AC-04c, AC-08, AC-09, AC-12, AC-15, AC-16c · **Blocked by:** [T09](./T09-application-lifecycle-domain.md), [T15](./T15-seat-head-end-term.md), [T20](./T20-promote-row-level-security.md), [T24](./T24-operator-queue-review-check.md)

## What

- Use case: T09 `Approve` → begin → insert the association (ЄДРПОУ, name, address of the latest filing, `connected`); a hit on `uq_associations_edrpou` is `identity.association_already_connected` → `InAssociation(new id, applicant's address)` → `Membership.SeatHead` with the term of the filing → insert the decision (`approved`, check, association, document relied on) → application `approved` with its association → `confirmed_by_decision_id`.
- Handler `ApproveApplication`; queries `CreateAssociation`, `SetAssociationConfirmation`.

## Definition of Done

- [ ] Acceptance: approveApplication creates the association, the decision and the applicant's roster entry as голова правління together or not at all, refuses every guard of AC-03, AC-03b, AC-12, AC-15 and AC-16c, and two racing approvals for one code create exactly one association.
- [ ] AC-08: the association has the code, name and address of the заявка; the decision has the date and the оператор; the applicant is on the roster as `head` with the term of the заявка
- [ ] AC-09: the applicant is the current member of the new association from the term start
- [ ] AC-03: no check, or a check with the signature found invalid → 409, stays `pending`, no association; AC-03b → 409; AC-12: an open розбіжність without a document → 409, with one of ≤ 500 characters → approved and stored, a longer one → 422
- [ ] AC-04c: two approvals at once → one association, the other answers 409 `identity.association_already_connected` and can still be rejected
- [ ] the roster step failing → no association, no decision, the заявка `pending`; AC-15 and AC-16c answer their codes
- [ ] `mise run lint` and `mise run test` green

## Notes

Runs on the app-role pool with row-level security on (T20). Write latency target ≤ 500 ms p95 (PRD §6): one transaction, no file.
