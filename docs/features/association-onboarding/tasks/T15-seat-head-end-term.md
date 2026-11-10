---
id: T15
title: "Seat a голова and end a term through the governance facade inside the caller's transaction"
layer: "app"
deps: ["T14"]
acs: ["AC-08", "AC-19", "AC-19c", "AC-20"]
files_hint:
  - "services/core/internal/governance/app/roster.go"
  - "services/core/internal/governance/app/roster_test.go"
  - "services/core/internal/governance/adapters/postgres/"
  - "services/core/internal/governance/module.go"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T15 — Seat a голова and end a term through the governance facade inside the caller's transaction

## Why

[ADR-0006](../adr/0006-apply-approvals-to-the-roster-through-a-governance-facade-in-one-transaction.md): the approval and the roster change commit together or not at all; [sad](../sad.md) §5 «Roster change on approval», §6 Flows 2 and 3.

**Satisfies:** [PRD §5](../PRD.md) AC-08, AC-19, AC-19c, AC-20 · **Blocked by:** [T14](./T14-roster-read-facade.md)

## What

- `SeatHead(ctx, association, person, term, decisionID)`: a person off the roster gets an entry from the typed term start with the position `head`; a person with a current entry gets a new revision of that entry (T07 `Reseat`), never a second entry.
- `EndTerm(ctx, association, memberID, date, decisionID)`: a revision that ends the whole entry on the date clamped by T07's rule, with the decision reference; no other member's entry changes.
- Both refuse to run without a transaction in `ctx`.

## Definition of Done

- [ ] Acceptance: SeatHead creates or re-seats one entry with the decision reference and EndTerm ends one whole entry, both only inside the transaction in ctx, and a failure in the caller leaves the roster unchanged.
- [ ] integration test: seat on an empty roster — one entry, `head`, decision reference set
- [ ] re-seat of a current member — the same entry id, one more revision, the term unchanged
- [ ] after `EndTerm` on yesterday's date, `CurrentMember` is false today; the other entries are byte-equal to before
- [ ] the caller's `fn` returns an error after both calls — the roster is as before
- [ ] `mise run lint` and `mise run test` green

## Notes

Entries seated here have no `created_by_member_id` (written by the system on an оператор's decision). The roster comment «re-election is a new entry» of `00012` keeps holding for the board's own edits; a re-seat by approval is the one exception (ADR-0006).
