---
id: T29
title: "Approve a запит на передачу: seat the new голова and end the previous term in one transaction"
layer: "ports"
deps: ["T26", "T28"]
acs: ["AC-16c", "AC-18b", "AC-19", "AC-19b", "AC-19c", "AC-20", "AC-24b"]
files_hint:
  - "services/core/internal/identity/app/"
  - "services/core/internal/identity/ports/ports.go"
  - "services/core/internal/identity/adapters/postgres/"
  - "services/core/internal/identity/adapters/http/"
owner: "Andrii Opryshko"
estimate: "L"
status: "todo"
---

# T29 — Approve a запит на передачу: seat the new голова and end the previous term in one transaction

## Why

[sad](../sad.md) §6 Flow 3; [ADR-0006](../adr/0006-apply-approvals-to-the-roster-through-a-governance-facade-in-one-transaction.md) as amended 2026-11-10 (re-seat, the end date, the competing запит); [sad](../sad.md) §10 QG-2.

**Satisfies:** [PRD §5](../PRD.md) AC-16c, AC-18b, AC-19, AC-19b, AC-19c, AC-20, AC-24b · **Blocked by:** [T26](./T26-approve-connection-application.md), [T28](./T28-review-transfer-request.md)

## What

- `ApproveApplication` for a transfer: T09 `Approve` with the extract's ЄДРПОУ against the association's code, or the typed one when the association has none → `InAssociation(association)` → record the ЄДРПОУ if missing → when a голова was confirmed by an earlier approval and is not the applicant, `Membership.EndTerm` → `Membership.SeatHead` → the decision → application `approved` → `confirmed_by_decision_id`.
- The association's state is not touched; another pending запит stays `pending`.

## Definition of Done

- [ ] Acceptance: Approving a запит seats or re-seats the applicant as голова правління, ends the whole entry of the голова confirmed by the previous approval on the day before the new term start, records the decision on both, and the previous голова's next request is refused.
- [ ] AC-19: term start 2026-11-20, approved 2026-11-25 → the new голова is current from 2026-11-20, the previous entry ends 2026-11-19, the other entries are unchanged, both revisions carry the decision id; an applicant already on the roster gets no second entry
- [ ] the previous голова's session, opened before the approval, gets 403 `identity.not_current_member` on its next request
- [ ] AC-19b: approving the second запит later ends the entry of the голова confirmed first, never before that entry's first day
- [ ] AC-19c: the hand-made association gets its ЄДРПОУ, the member's entry gets `head`, no term ends; while its голова is the оператор → 409 `identity.operator_own_application`
- [ ] AC-18b: no check or another code → refused, the roster unchanged; AC-24b: a disabled association stays disabled; a failing roster step undoes everything
- [ ] `mise run lint` and `mise run test` green

## Notes

Known edge (SAD §11): a previous entry that began on the approval day keeps that one day.
