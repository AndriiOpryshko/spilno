---
status: superseded
superseded_by: "0020"
date: 2026-10-25
feature: board-protocol
sad_section: "§5"
deciders: ["Andrii Opryshko"]
---

# 0011 Keep charter rules, board terms and building shares as append-only effective-dated rows

> **Status:** superseded by [0020](./0020-keep-board-terms-and-building-shares-as-dated-rows-and-charter-rules-as-current-settings.md) (2026-10-25): narrowed — board terms and building shares stay dated; charter rules with effective dates (AC-02) move to Stage 2.

## Context and problem statement

A protocol takes the board composition, частки будинку, кворум, majority, competence limit and allowed meeting forms as of its meeting date, not as of today (US-01, AC-01, AC-02, AC-07). A signed protocol names the rules it applied (AC-19, AC-36), so a rule row once applied must never change. Records are kept for years (PRD §6 Retention).

## Decision drivers

- QG-2: the rule in force on the meeting date, reproducibly.
- QG-1: rows referenced by a signed protocol are immutable (ADR 0009).
- AC-21b: a missing rule warns instead of blocking, so "no rule in force" is a valid answer.

## Blast radius

- Irreversible: yes — years of rule history and references from signed protocols.
- Multi-context: no — inside `governance`.
- Legitimate alternative: yes — validity ranges (`daterange`) with an exclusion constraint.
- Touches the database: yes — three new tables and their history semantics.

## Considered options

1. **Append-only rows with `effective_from`**; a correction is a new row with `supersedes_id`.
2. **Validity ranges `valid_from..valid_to`** with an exclusion constraint; adding a rule closes the previous one with an UPDATE.

## Decision outcome

**Chosen:** option 1. `charter_rules` (association, kind: `quorum`, `majority`, `competence_limit`, `meeting_forms`, `non_voting_counts_toward_item_quorum`; value as jsonb; `effective_from`; source document; `supersedes_id`; recorded by and at). `building_shares` (inter-building decision document, `effective_from`, share per building in basis points). `board_terms` (person name, `account_id`, role chair or member, `term_from`, `term_to`, election protocol document). The rule in force for a kind on date D is the row with the latest `effective_from ≤ D` that no other row supersedes. Rows are never updated or deleted: a trigger of the same kind as ADR 0009 refuses UPDATE and DELETE on all three tables, so a row a signed protocol applied cannot change.

### Consequences

- Good: "rule on a date" is one indexed query; full history stays for the ревізійна комісія.
- Bad: two rows with the same kind and date are resolved by `supersedes_id`; the UI must offer "correct" rather than "edit".
- Neutral: board terms naturally carry a range (`term_from`, `term_to`) and are append-only too; an early end of a term is a superseding row.

### Migration and rollback

- Migration: new tables, no existing rows; before the first protocol a член правління records the board, shares and rules (PRD §9 Rollout).
- Rollback: drop while no protocol references them; afterwards they stay.

## Links

- PRD: [PRD.ao-prd.md](../PRD.ao-prd.md) — US-01, AC-01, AC-02, AC-07, AC-21b
- SAD: [sad.ao-architecture.md](../sad.ao-architecture.md) §5
- Related ADR: [0007 Pure checks](./0007-evaluate-blocking-checks-as-a-pure-domain-function-and-snapshot-them-on-transition.md) · [0009 Guard signed protocols](./0009-guard-signed-protocols-with-a-database-trigger-and-a-snapshot-hash.md)
