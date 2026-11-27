---
id: T86
title: "Find the голова a decision seated and the entry it ended by the first revision that carries the decision"
layer: "infra"
deps: []
acs: ["AC-19", "AC-19b", "AC-20", "AC-29"]
files_hint:
  - "services/core/internal/identity/adapters/postgres/queries/approval.sql"
  - "services/core/internal/identity/adapters/postgres/queries/card.sql"
  - "services/core/internal/identity/adapters/postgres/approval.go"
  - "services/core/internal/identity/adapters/postgres/card.go"
  - "services/core/internal/identity/approval_transfer_test.go"
  - "services/core/internal/identity/card_test.go"
owner: "Andrii Opryshko"
estimate: "M"
status: "done"
---

# T86 — Find the голова a decision seated and the entry it ended by the first revision that carries the decision

## Why

[Review of 2026-11-17](../_review/review-2026-11-17.md), N01. [PRD](../PRD.md) AC-19b: a later transfer «ends the term of the голова confirmed by the first approval». `GetEntryOfDecision` takes the newest revision that carries the decision and `ListCardHistory` takes the ended entry by revision order; `ChangeMember` copies the decision into every later revision, so one edit of the previous голова's entry turns both answers around.

**Satisfies:** [PRD §5](../PRD.md) AC-19, AC-19b, AC-20, AC-29 · **Blocked by:** —

## What

- `GetEntryOfDecision` and the «ended» subquery of `ListCardHistory` look only at the first revision of each board member that carries the decision: the member whose first such revision is the latest is the one the decision seated, an earlier one is the entry it ended.
- A later edit of either entry changes neither answer.
- `mise run gen` after the queries change.

## Definition of Done

- [ ] Acceptance: Integration tests show that after a transfer, an edit of the previous голова's entry and a second transfer, the second transfer ends the голова the first one seated, and the card names the previous голова as ended.
- [ ] `approval_transfer_test.go`: transfer, `updateBoardMember` on the previous голова (name fix), second transfer; the голова of the first transfer has no current entry and is refused on the next request
- [ ] `card_test.go`: transfer, edit of the previous голова, card read; `ended_head_full_name` is the previous голова
- [ ] the existing cases (re-seated applicant, AC-19c with nothing ended) still pass
- [ ] `mise run lint` and `mise run test` green

## Notes

Starts with the failing test. No migration: the rule rests on the order «ended entry first, seated entry second» that `approval.go` already keeps. The alternative, a column that records what a decision did to a revision, needs an expand and a backfill of `board_member_revisions`; take it only if the test shows the order cannot be trusted.
