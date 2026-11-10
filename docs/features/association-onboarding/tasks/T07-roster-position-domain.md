---
id: T07
title: "Give roster entries a position and a decision reference, with the date rules of a transfer"
layer: "domain"
deps: []
acs: ["AC-08", "AC-19", "AC-19b", "AC-19c"]
files_hint:
  - "services/core/internal/governance/domain/board_member.go"
  - "services/core/internal/governance/domain/board_member_test.go"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T07 — Give roster entries a position and a decision reference, with the date rules of a transfer

## Why

[ADR-0006](../adr/0006-apply-approvals-to-the-roster-through-a-governance-facade-in-one-transaction.md) as amended 2026-11-10; [sad](../sad.md) §8 «Dates». `BoardMemberEntry` has no position and no decision reference today ([data-model](../data-model.md) «Drift»).

**Satisfies:** [PRD §5](../PRD.md) AC-08, AC-19, AC-19b, AC-19c · **Blocked by:** —

## What

- `Position` with the closed set `head` / `member` and `DecisionID` on `BoardMemberEntry`; `Validate` names an unknown position as a broken field.
- `PreviousHeadTermEnd(newTermStart, previousEntryStart)`: the day before the new term start, never earlier than the first day of the previous entry.
- `Reseat(entry, decisionID)`: the next revision of the same entry with the position `head` and the decision reference; the entry keeps its own term.

## Definition of Done

- [ ] Acceptance: Unit tests prove the position on an entry, the end date of the previous голова (the day before the new term start, never before the entry's first day) and the re-seat of an entry without a second one.
- [ ] table test: new term start 2026-11-20 gives the end 2026-11-19; an entry that began on 2026-11-25 with a new term start 2026-11-20 ends on 2026-11-25
- [ ] a re-seat keeps `Term` and sets `head` and the decision; no second entry is produced
- [ ] the existing tests of `board_member_test.go` still pass
- [ ] `mise run lint` and `mise run test` green

## Notes

Pure domain, starts on day one. PRD AC-19 does not yet state the «never earlier than the first day» rule; SAD §8 and ADR-0006 do (api-sync-report §D).
