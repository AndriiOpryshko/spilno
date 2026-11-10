---
id: T25
title: "Reject an application with a reason"
layer: "ports"
deps: ["T09", "T24"]
acs: ["AC-13", "AC-14", "AC-15", "AC-16c"]
files_hint:
  - "services/core/internal/identity/app/"
  - "services/core/internal/identity/ports/ports.go"
  - "services/core/internal/identity/adapters/postgres/"
  - "services/core/internal/identity/adapters/http/"
owner: "Andrii Opryshko"
estimate: "S"
status: "todo"
---

# T25 — Reject an application with a reason

## Why

[sad](../sad.md) §6 Flow 8; [PRD](../PRD.md) §10 «Очікує → Відхилена».

**Satisfies:** [PRD §5](../PRD.md) AC-13, AC-14, AC-15, AC-16c · **Blocked by:** [T09](./T09-application-lifecycle-domain.md), [T24](./T24-operator-queue-review-check.md)

## What

- Use case: T09 `Reject` → insert the decision (`rejected`, filing, reason, decided by) and set the state, in one transaction; `uq_application_decisions_filing_id` backs «one decision per filing».
- Handler `RejectApplication`; queries `CreateDecision`, `GetDecisionOfFiling`.

## Definition of Done

- [ ] Acceptance: rejectApplication records one decision with the reason, date and оператор and sets the application rejected, refuses an empty reason, a decided application and the оператор's own filing, and creates no association.
- [ ] AC-13: an empty reason → 422 `identity.reason_required`; the application stays `pending`
- [ ] AC-14: `rejected` with reason, date and оператор; no association row; the applicant's `getMyApplication` shows the reason
- [ ] AC-15: a second decision → 409 `identity.application_not_pending`; the stored decision is unchanged
- [ ] AC-16c: the оператор's own filing → 409 `identity.operator_own_application`; it stays `pending`; both kinds covered
- [ ] `mise run lint` and `mise run test` green

## Notes

A rejection may come before any check (`check_id` is nullable).
