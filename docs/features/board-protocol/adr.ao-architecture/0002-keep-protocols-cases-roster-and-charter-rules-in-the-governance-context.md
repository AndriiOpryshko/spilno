---
status: accepted
date: 2026-10-25
feature: board-protocol
sad_section: "§4"
deciders: ["Andrii Opryshko"]
---

# 0002 Keep protocols, cases, the board roster and charter rules in the governance context

## Context and problem statement

The feature adds protocols with revisions, slots, poll rounds and votes, print copies and meeting facts; справи with documents; board terms; association settings (current правила статуту); частки будинку. `governance` already owns board meetings ("board meetings and their decisions", `module.go`), and PRD §5.1 names every error `governance.*`. Repo ADR 0002 forbids one context reading another's tables, so the owner of each table must be decided now (PRD §9, AC-01, AC-03, AC-33). Note 2026-10-25: Stage 1 rescope — blocking checks and finance contracts moved to Stage 2 (ADR 0007, 0012 deferred); the ownership decision is unchanged. Amended 2026-10-25: частки будинку and allocation between buildings (AC-07) are Stage 2 too, so Stage 1 adds no `building_shares` table; the roster rows carry each member's sign-in e-mail (ADR 0020, 0022).

## Decision drivers

- PRD §9: a protocol links to exactly one meeting, a meeting has at most one protocol.
- PRD §10: rights and the members slot read the board on a date; the protocol is 1:1 with a meeting (blocking checks that read meeting, board and rules together follow in Stage 2).
- PRD §5.1: error codes are `governance.<error_name>`.
- Repo ADR 0002: contexts talk only through `module.go`.

## Blast radius

- Irreversible: yes — table ownership, OpenAPI tag and `/governance/` paths are hard to move once protocols are signed.
- Multi-context: yes — `governance` and `identity` (actor), and the web client's generated hooks; `finance` joins in Stage 2 (ADR 0012).
- Legitimate alternative: yes — a new `protocols` context, or `charter` + `protocols`.
- Touches the database: yes — owner of the new tables.

## Considered options

1. **Extend `governance`** — one context owns meetings, protocols, справи, board terms, association settings, частки будинку.
2. **New `protocols` context** — `governance` keeps meetings, roster and rules and lends them through `module.go`.
3. **Split into `charter` and `protocols`** — three contexts.

## Decision outcome

**Chosen:** option 1. State transitions and the "one protocol per meeting" invariant stay inside one transaction, the PRD error codes stay as approved, and no new context wiring is needed for these aggregates.

### Consequences

- Good: no cross-context call in the hot path of a state transition; a unique constraint enforces one protocol per meeting.
- Bad: `governance` becomes the largest context (about six aggregates); files in `domain/`, `app/` and `adapters/postgres/queries/` are split per aggregate to stay navigable.
- Neutral: in Stage 1 a contract is only a document of a справа; `ContractsInForce` through `finance/module.go` comes with ADR 0012 in Stage 2, and `governance` never reads `contracts` directly.

### Migration and rollback

- Migration: new goose migrations add `governance`-owned tables next to `meetings`: `protocols` (meeting id unique — one protocol per meeting, state, revision, number, `corrects_protocol_id`), `protocol_items` (agenda item, order), `protocol_slots` (kind, value, origin system / ai / human, status, source document version and fragment), `slot_confirmations` (value, confirmed or corrected by, at; proposed value kept), `poll_rounds` (round number, revision, opened and closed at, close reason) and `protocol_votes` (round, member, item, choice, at), `cases` (one per agenda item) with `case_documents` (version of an earlier document) and file metadata in `governance_files`; plus the tables of ADR 0017, 0018, 0019, 0020, 0021. River's tables come with ADR 0015. `meetings` itself gains only `association_id` (ADR 0016).
- Rollback: down migrations drop only empty new tables; once a protocol is signed, tables are never dropped (PRD §9 Reversibility).

## Links

- PRD: [PRD.ao-prd.md](../PRD.ao-prd.md) — §5.1, §9, AC-01, AC-03, AC-33; Stage 2: AC-02, AC-07, AC-10, AC-19
- SAD: [sad.ao-architecture.md](../sad.ao-architecture.md) §4, §5
- Related ADR: repo [0002 Bounded contexts](../../../adr/0002-bounded-contexts.md)
