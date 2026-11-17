---
id: T87
title: "Tell each roster entry what its decision did to it, and refuse to reopen an entry an approval ended"
layer: "ports"
deps: []
acs: ["AC-19", "AC-20"]
files_hint:
  - "services/core/api/openapi.yaml"
  - "docs/features/association-onboarding/contracts/openapi.yaml"
  - "services/core/internal/governance/adapters/postgres/queries/roster.sql"
  - "services/core/internal/governance/adapters/postgres/roster.go"
  - "services/core/internal/governance/app/roster.go"
  - "services/core/internal/governance/ports/"
  - "services/core/internal/governance/adapters/http/board_members.go"
  - "services/core/internal/governance/roster_endpoints_test.go"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T87 — Tell each roster entry what its decision did to it, and refuse to reopen an entry an approval ended

## Why

[Review of 2026-11-17](../_review/review-2026-11-17.md), N01. [PRD](../PRD.md) AC-19: the previous голова's «whole roster entry ends … if they stay on the board the board adds them again». Today the board can clear the term end of the ended entry; it stays голова правління under the transfer's decision, so the association has two current heads. The roster screen guesses the ended entry by full name.

**Satisfies:** [PRD §5](../PRD.md) AC-19, AC-20 · **Blocked by:** —

## What

- `BoardMember` gains `decision_effect`: `seated`, `ended` or null (delta and live contract, `mise run gen`). Governance computes it with the first-revision rule of [T86](./T86-decision-entries-by-first-revision.md), in its own query.
- `ChangeMember` refuses a term end that is cleared or later than the one the approval set, on an entry with `decision_effect` `ended`. Other fields are saved as before.
- A person added again is a new entry with the position член правління and no decision (unchanged `CreateMember`).

## Definition of Done

- [ ] Acceptance: Handler tests show `listBoardMembers` returning `decision_effect` `seated` or `ended` for the entries of a transfer, and `updateBoardMember` refusing with 422 `governance.roster_entry_invalid` on `/term_end` to clear or move later the term end of an entry an approval ended, while a name fix is saved.
- [ ] `roster_endpoints_test.go`: both effects after a transfer, null for an entry without a decision, the refusal, the saved name fix
- [ ] a re-added previous голова is a second entry, position `member`, and the association has one current голова
- [ ] the contract and its generated code land in the same commit
- [ ] `mise run lint` and `mise run test` green

## Notes

The refusal follows the PRD wording. If the product owner prefers that reopening resets the entry to член правління without a decision, only `ChangeMember` and its test change.
