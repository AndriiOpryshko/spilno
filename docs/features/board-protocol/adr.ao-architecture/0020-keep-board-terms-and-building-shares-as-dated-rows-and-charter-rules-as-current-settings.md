---
status: accepted
date: 2026-10-25
feature: board-protocol
sad_section: "§5"
deciders: ["Andrii Opryshko"]
---

# 0020 Keep board terms as append-only dated rows, and charter rules as current association settings

> Amended 2026-10-25: narrowed to board terms and the current settings — частки будинку and the allocation of an amount between buildings are not in Stage 1 (PRD.ao-prd.md AC-07 moved to «Deferred to Stage 2»); `building_shares` returns with that AC in Stage 2 as an append-only dated table of the same shape. Board terms carry the member's `sign_in_email` instead of `account_id`, because sign-in matches the roster (ADR 0022). The file name keeps its original slug so links stay stable.

## Context and problem statement

Stage 1 still needs the board as of a date: rights come from a term in force on the date of the action (AC-32, PRD §10, checked before every request by ADR 0022), and members shown in a draft come from the board on the meeting date (AC-01, AC-08). Charter rules with effective dates (AC-02) moved to Stage 2; Stage 1 stores only the current meeting form, signers, register start number and current правила статуту in the association settings, refuses the first draft while they are missing (AC-45) and checks nothing against them. Per-building amounts from частки будинку (AC-07) are Stage 2 as well. ADR 0011 made charter rules, board terms and shares effective-dated; this ADR narrows it and supersedes ADR 0011.

## Decision drivers

- AC-01, AC-08, AC-32: board on a date, for display and for authorization.
- AC-45: settings present before the first draft; printed on every copy; no state change blocked by them.
- QG-1: rows a signed protocol used must not change (they are also copied into its snapshot, ADR 0019).

## Blast radius

- Irreversible: yes — years of board history and the dates authorization is computed from.
- Multi-context: no — inside `governance`; `identity` only asks it whether a sign-in e-mail is a current член правління (ADR 0022).
- Legitimate alternative: yes — validity ranges with UPDATE (ADR 0011 option 2), or effective-dated charter rules now (ADR 0011).
- Touches the database: yes — `board_terms`, `association_settings`.

## Considered options

1. **Board terms append-only with dates; charter rules as one current settings row.**
2. **Everything effective-dated now** (ADR 0011) — charter rule history built before any rule is checked.

## Decision outcome

**Chosen:** option 1. `board_terms` (association, person name, `sign_in_email`, role chair or member, `term_from`, `term_to`, election protocol document, `supersedes_id`, recorded by and at) is append-only: a correction or an early end is a new row that supersedes the old one, and a trigger refuses UPDATE and DELETE. "Current член правління" is a term with `term_from ≤ today ≤ term_to` (Europe/Kyiv) that no row supersedes; `governance/module.go` answers it for sign-in (ADR 0022). `association_settings` (association PK, allowed meeting form, signers, `register_start_number`, current правила статуту as jsonb, print template revision, updated by and at) is one mutable row; each protocol copies the values it printed into its snapshot (ADR 0019), so later edits do not reach signed protocols.

**Stage 2 note:** `building_shares` (association, inter-building decision document, `effective_from`, share per building in basis points, `supersedes_id`, append-only under the same trigger) and the largest-remainder split come with AC-07.

### Consequences

- Good: authorization and the members slot are one indexed query; history stays for the ревізійна комісія.
- Good: Stage 2 adds `charter_rules` with `effective_from` and `building_shares` next to the settings without migrating board terms.
- Bad: the UI offers "correct" rather than "edit" for terms.
- Neutral: the current правила статуту in settings are informational in Stage 1.

### Migration and rollback

- Migration: new tables, no existing rows; before the first protocol a член правління records the board and the settings (PRD §9 Rollout); the first roster entry comes from the setup seed (ADR 0022).
- Rollback: drop while no protocol references them; afterwards they stay.

## Links

- PRD: [PRD.ao-prd.md](../PRD.ao-prd.md) — US-01; AC-01, AC-08, AC-32, AC-45; §3 Deferred "Charter rules with effective dates (AC-02)", "Allocation between buildings (AC-07)"
- SAD: [sad.ao-architecture.md](../sad.ao-architecture.md) §5, §8
- Related ADR: supersedes [0011 Effective-dated rows](./0011-keep-charter-rules-board-terms-and-building-shares-as-append-only-effective-dated-rows.md) · [0019 Lock snapshot](./0019-lock-a-protocol-as-one-fingerprinted-snapshot-row-guarded-by-a-trigger.md) · [0022 Roster sign-in](./0022-sign-in-with-google-and-admit-only-current-board-members-from-the-governance-roster.md)
