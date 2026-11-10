---
status: accepted
date: 2026-11-05
feature: association-onboarding
sad_section: "§5"
deciders: ["Andrii Opryshko", "Claude (ao-architecture)"]
---

# 0006 Apply approvals to the roster through a governance facade in one transaction

## Context and problem statement

Approving a заявка creates an association (identity) and its first член правління with the position голова правління (governance, AC-08). Approving a запит на передачу ends the previous голова's term and seats the new one (AC-19). The roster has no position today, so nothing marks the голова. The decision and the roster change must never diverge.

## Decision drivers

- PRD AC-08, AC-09, AC-19, AC-19c, AC-20: roster entries link to the decision.
- PRD AC-19: the previous term ends the day before the new term start, no later than the day before approval.
- Board-protocol [ADR-0012](../../board-protocol/adr/0012-integrate-contexts-through-synchronous-module-facades.md): contexts integrate through synchronous `module.go` facades.

## Blast radius

- Irreversible: yes — the roster is legal history.
- Multi-context: yes — identity and governance.
- Legitimate alternative: yes — an event or River job that governance applies after identity commits.
- Touches the database: yes — `board_member_revisions` gains a position and a decision reference.

## Considered options

1. **identity's use case calls `governance.SeatHead` and `governance.EndTerm` inside one transaction carried in `ctx` by `platform/db`.**
2. **identity commits the decision and enqueues a River job**; governance applies the roster change after.

## Decision outcome

**Chosen:** Option 1. Either all of decision, association and roster change are recorded or none is; no window where the decision exists and the roster does not.

**Amended 2026-11-10** (PRD AC-17, AC-19, AC-19b as clarified): the applicant of a запит may already be on the roster. `SeatHead` then appends a revision to the person's current entry with the position голова правління and the decision reference and creates no second entry — the path AC-19c already used; for a person off the roster it creates the entry from the typed term start. The re-seated entry keeps its own term; the term typed in the запит stays with the filing and is reached through the decision reference. `EndTerm` ends the whole entry of the голова recorded by the previous approval and touches no other member's entry; it is not called when no голова is recorded (AC-19c) or when that голова is the applicant. A competing запит approved later goes through the same two calls as a new transfer.

### Consequences

- Good: AC-19's «access ends at the moment of approval» holds from the same commit.
- Bad: governance's facade must accept a transaction from `ctx`; contexts share one transaction (still no cross-context table access).
- Neutral: the hand-made association gets its голова by the same path (AC-19c, with no term to end); the roster comment «re-election is a new entry» of `migrations/00012` keeps holding for the board's own edits, a re-seat by approval is the one exception (AC-19).

### Migration and rollback

- Migration: `board_member_revisions` + `position` (nullable → backfill `член правління` for existing rows → `SET NOT NULL`, 3 steps) and `decision_id` (nullable, id only, no foreign key across contexts).
- Rollback: down migrations drop the two columns; roster terms are kept.

## Links

- PRD: [PRD.md](../PRD.md) — AC-08, AC-09, AC-17, AC-19, AC-19b, AC-19c, AC-20, §9
- SAD: [sad.md](../sad.md) §5, §6
- Related ADR: board-protocol [ADR-0012](../../board-protocol/adr/0012-integrate-contexts-through-synchronous-module-facades.md); [0003](./0003-enforce-association-isolation-with-postgres-row-level-security.md)
