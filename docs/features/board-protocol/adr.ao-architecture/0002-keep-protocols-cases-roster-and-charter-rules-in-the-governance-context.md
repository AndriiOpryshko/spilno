---
status: accepted
date: 2026-10-25
feature: board-protocol
sad_section: "§4"
deciders: ["Andrii Opryshko"]
---

# 0002 Keep protocols, cases, the board roster and charter rules in the governance context

## Context and problem statement

The feature adds protocols with versions, slots, votes and overrides; справи with documents; board terms; правила статуту; частки будинку. `governance` already owns board meetings ("board meetings and their decisions", `module.go`), and PRD §5.1 names every error `governance.*`. Repo ADR 0002 forbids one context reading another's tables, so the owner of each table must be decided now (PRD §9, AC-02, AC-19, AC-33).

## Decision drivers

- PRD §9: a protocol links to exactly one meeting, a meeting has at most one protocol.
- PRD §10: each блокуюча перевірка reads the meeting, the board on the meeting date and the rules on the meeting date.
- PRD §5.1: error codes are `governance.<error_name>`.
- Repo ADR 0002: contexts talk only through `module.go`.

## Blast radius

- Irreversible: yes — table ownership, OpenAPI tag and `/governance/` paths are hard to move once protocols are signed.
- Multi-context: yes — `governance` and `finance` (contracts as джерела коштів), and the web client's generated hooks.
- Legitimate alternative: yes — a new `protocols` context, or `charter` + `protocols`.
- Touches the database: yes — owner of the new tables.

## Considered options

1. **Extend `governance`** — one context owns meetings, protocols, справи, board terms, правила статуту, частки будинку.
2. **New `protocols` context** — `governance` keeps meetings, roster and rules and lends them through `module.go`.
3. **Split into `charter` and `protocols`** — three contexts.

## Decision outcome

**Chosen:** option 1. The checks and the "one protocol per meeting" invariant stay inside one transaction, the PRD error codes stay as approved, and no new context wiring is needed for these aggregates.

### Consequences

- Good: no cross-context call in the hot path of a state transition; a unique constraint enforces one protocol per meeting.
- Bad: `governance` becomes the largest context (about six aggregates); files in `domain/`, `app/` and `adapters/postgres/queries/` are split per aggregate to stay navigable.
- Neutral: `finance` exposes `ContractsInForce(ctx, association, date)` through `module.go`; `governance` never reads `contracts` directly.

### Migration and rollback

- Migration: new goose migrations add `governance`-owned tables next to `meetings`; `meetings` itself gains only `association_id` (ADR 0010).
- Rollback: down migrations drop only empty new tables; once a protocol is signed, tables are never dropped (PRD §9 Reversibility).

## Links

- PRD: [PRD.ao-prd.md](../PRD.ao-prd.md) — §5.1, §9, AC-02, AC-10, AC-19
- SAD: [sad.ao-architecture.md](../sad.ao-architecture.md) §4, §5
- Related ADR: repo [0002 Bounded contexts](../../../adr/0002-bounded-contexts.md)
