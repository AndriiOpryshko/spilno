---
id: T51
title: "Carry the filed term end into the roster and end the previous голова whatever their term end"
layer: "app"
deps: ["T50"]
acs: ["AC-08", "AC-19", "AC-19c", "AC-30"]
files_hint:
  - "services/core/internal/identity/adapters/postgres/queries/approval.sql"
  - "services/core/internal/identity/adapters/postgres/approval.go"
  - "services/core/internal/identity/ports/"
  - "services/core/internal/identity/approval.go"
  - "services/core/cmd/api/main.go"
  - "services/core/internal/governance/app/roster.go"
  - "services/core/internal/governance/domain/board_member.go"
  - "services/core/internal/identity/approval_test.go"
  - "services/core/internal/identity/approval_transfer_test.go"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T51 — Carry the filed term end into the roster and end the previous голова whatever their term end

## Why

[Review of 2026-11-10](../_review/review-2026-11-10.md), F05, F06. [PRD](../PRD.md) AC-08 copies the term from the заявка; [ADR-0006](../adr/0006-apply-approvals-to-the-roster-through-a-governance-facade-in-one-transaction.md) ends the whole entry of the previous голова. Today the term end is dropped on seating, and `GetEntryOfDecision` skips an entry that has a term end, so the two defects hide each other.

**Satisfies:** [PRD §5](../PRD.md) AC-08, AC-19, AC-19c, AC-30 · **Blocked by:** [T50](./T50-kyiv-calendar-day.md)

## What

- `GetLatestFilingForApproval` selects `term_end`; `SeatHeadFunc` and the wiring in `cmd/api` carry it into `governance.Term.End`.
- `GetEntryOfDecision` returns the latest entry of the decision whatever its `term_end`; `EndTerm` clamps to the earlier of the existing end and the day before the new term start, never before the entry's first day (SAD §8).
- Fix both in one change: storing the term end without the second fix would stop every later transfer from ending the previous голова.

## Definition of Done

- [ ] Acceptance: Integration tests show a голова seated with the term end from the заявка, and a transfer that ends a previous голова whose entry already has a future term end, leaving exactly one current head.
- [ ] a connection approval with term end 2026-12-31 stores that end on the roster entry
- [ ] a transfer over a head with a future term end ends that entry and seats the applicant; `CurrentMember` refuses the previous голова on the next request
- [ ] the hand-made association of AC-19c (a head re-seated from an existing entry with a term end) is covered
- [ ] a previous head whose term already ended keeps their own end date
- [ ] `mise run lint` and `mise run test` green

## Notes

`mise run gen` after the query change.
