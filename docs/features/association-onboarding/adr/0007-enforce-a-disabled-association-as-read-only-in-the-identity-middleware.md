---
status: accepted
date: 2026-11-05
feature: association-onboarding
sad_section: "§8"
deciders: ["Andrii Opryshko", "Claude (ao-architecture)"]
---

# 0007 Enforce a disabled association as read-only in the identity middleware

## Context and problem statement

A вимкнене ОСББ lets its members only read and download the signed protocols of the registry; everything else is refused with the reason; enabling restores everything (PRD AC-21, AC-23, AC-25, AC-26). The rule spans every context with association-scoped operations.

## Decision drivers

- PRD AC-25, AC-26; PRD §6: in force from the next request.
- PRD AC-23: all data kept.

## Blast radius

- Irreversible: no — the allow-list can move later.
- Multi-context: yes — identity enforces, governance and finance obey.
- Legitimate alternative: yes — each context checks the state in its own use cases.
- Touches the database: yes — association state and an insert-only state-change history.

## Considered options

1. **The association-scoped middleware, after the membership check (a non-member gets `identity.not_current_member` whatever the state), reads the association state; a disabled association passes only an allow-list of operation ids (registry reads, signed-copy download); everything else gets `identity.association_disabled` with the reason; the actor carries `read_only`.**
2. **Every context checks the state in `app/`.**

## Decision outcome

**Chosen:** Option 1. One place, one test; a new operation is refused by default when the association is disabled.

### Consequences

- Good: fail-closed for new operations.
- Bad: the allow-list names governance operations inside identity.
- Neutral: governance still checks that a downloaded file is pinned by a signed snapshot when `read_only` is set.

### Migration and rollback

- Migration: `associations` + `state` (nullable → backfill `connected` → `SET NOT NULL`), `edrpou` (nullable, unique), `address`; a new insert-only state-change table with reason, оператор and time.
- Rollback: drop the columns and the history table; associations stay.

## Links

- PRD: [PRD.md](../PRD.md) — US-06, US-07, AC-21, AC-23, AC-25, AC-26
- SAD: [sad.md](../sad.md) §8
- Related ADR: [0002](./0002-put-the-association-id-in-every-association-scoped-path.md)
