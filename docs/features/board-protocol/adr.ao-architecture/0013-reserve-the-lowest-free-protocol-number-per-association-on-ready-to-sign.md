---
status: accepted
date: 2026-10-25
feature: board-protocol
sad_section: "§5"
deciders: ["Andrii Opryshko"]
---

# 0013 Reserve the lowest free protocol number per association on ready to sign

## Context and problem statement

A draft has no number (AC-08). Marking it готовий до підпису takes "the next free number of the association's protocol register (one sequence, no duplicates)" printed on the copy (AC-25); returning it to чернетка or withdrawing it releases the number (AC-26, AC-34c); a signed protocol keeps it (AC-28). Two members may mark two drafts ready at the same moment.

## Decision drivers

- AC-25: no duplicates; AC-26 / AC-34c: numbers are released.
- The paper register should have no holes the ревізійна комісія would ask about.
- QG-1: a signed number never changes.

## Blast radius

- Irreversible: yes — numbers are printed on signed paper.
- Multi-context: no — inside `governance`.
- Legitimate alternative: yes — always maximum + 1, leaving released numbers as gaps.
- Touches the database: yes — partial unique index and numbering query.

## Considered options

1. **Lowest free number** not held by a готовий до підпису or підписаний protocol.
2. **Maximum + 1**, released numbers stay as gaps.

## Decision outcome

**Chosen:** option 1. In the transition transaction `app/` takes `pg_advisory_xact_lock` keyed by the association (a lock held until commit, so concurrent "ready" calls queue), computes the lowest positive integer not used by a `ready_to_sign` or `signed` protocol of the association, and writes it. A partial unique index `(association_id, number) WHERE state IN ('ready_to_sign','signed')` is the final guard. Returning to чернетка or withdrawing sets `number` to null in the same transaction.

### Consequences

- Good: the register has no holes; duplicates are impossible.
- Bad: a re-used number can be dated later than a higher number (No 14 after No 15); the registry sorts by meeting date and shows the number.

### Migration and rollback

- Migration: `protocols.number int NULL` and the partial unique index in the protocols migration.
- Rollback: drop the index only; signed numbers stay.

## Links

- PRD: [PRD.ao-prd.md](../PRD.ao-prd.md) — AC-08, AC-25, AC-26, AC-28, AC-34c
- SAD: [sad.ao-architecture.md](../sad.ao-architecture.md) §5
- Related ADR: [0009 Guard signed protocols](./0009-guard-signed-protocols-with-a-database-trigger-and-a-snapshot-hash.md)
