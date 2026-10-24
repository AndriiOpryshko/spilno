---
status: accepted
date: 2026-10-25
feature: board-protocol
sad_section: "§4"
deciders: ["Andrii Opryshko"]
---

# 0007 Evaluate blocking checks as a pure domain function and snapshot them on transition

## Context and problem statement

Кворум, кворум пункту, majority, the competence limit, the allowed meeting form, unconfirmed or unsourced numbers and text-slot mismatch block a draft; other findings are попередження (AC-06, AC-09, AC-12, AC-19-AC-21b, AC-25-AC-27). PRD §10 lists which checks an обхід блокування lifts. The checks must agree with hand calculation on past protocols (QG-2), run again after every change (AC-26), and stay visible with the signed protocol (AC-36).

## Decision drivers

- QG-2: results reproducible from the roster and правила статуту in force on the meeting date.
- PRD §10: override table; AC-06 and AC-09 are never lifted by an override.
- CONTEXT "помічник протоколу": the AI judges nothing.

## Blast radius

- Irreversible: no — the function can move or be split without data changes.
- Multi-context: no — inside `governance`.
- Legitimate alternative: yes — store every check result as rows and update them on each change.
- Touches the database: yes — the result snapshot stored on the transition to готовий до підпису.

## Considered options

1. **Pure function, computed on read, snapshot on transition.**
2. **Persisted check rows** updated on every change of a draft.

## Decision outcome

**Chosen:** option 1. `governance/domain.Evaluate(Snapshot) []Check` has no I/O. A `Snapshot` holds the text (current or frozen), slots with sources and confirmations, meeting facts (attendance, неучасть у голосуванні, vote counts, meeting form, головуючий), the board on the meeting date and the правила статуту in force on that date (ADR 0011). A `Check` has a PRD §5.1 code, kind blocking or warning, whether an override lifts it, the agenda item and the rule or slot it names. `app/` builds the snapshot in one read transaction, runs it on every draft read and before every transition, and on готовий до підпису stores the check list and the ids of the applied rules in `protocol_check_snapshots` (jsonb, a JSON column). Overrides are rows that reference a check code and carry author, time and reason (AC-23, AC-24).

### Consequences

- Good: table-driven unit tests per AC, plus a regression set of the association's past protocols (SAD §10 QG-2).
- Good: the signed protocol shows exactly the checks and rules that applied when it was marked ready.
- Bad: every draft read rebuilds the snapshot; at ~20 protocols per association per year this is within the 200 ms read target.

### Migration and rollback

- Migration: new tables `protocol_check_snapshots`, `protocol_overrides`.
- Rollback: drop while empty; rows of signed protocols are kept (ADR 0009).

## Links

- PRD: [PRD.ao-prd.md](../PRD.ao-prd.md) — US-06, AC-06, AC-09, AC-19, AC-19b, AC-19c, AC-20, AC-21, AC-21b, AC-23, AC-24; §10 override table
- SAD: [sad.ao-architecture.md](../sad.ao-architecture.md) §4, §10
- Related ADR: [0009 Guard signed protocols](./0009-guard-signed-protocols-with-a-database-trigger-and-a-snapshot-hash.md)
