---
status: Accepted
owner: "Andrii Opryshko"
reviewers: []
updated_at: "2026-10-25"
feature_size: L
stage: "04-05"
ticket: ""
---

# 0015 — Scope owned rows to one association with an association_id column and explicit query filters, without RLS

- **Status:** Accepted
- **Date:** 2026-10-25
- **Deciders:** Andrii Opryshko (Architect), with Claude during the Stage 1 rescope

## Context

PRD §3: the first release serves one association — "records carry the association, but there is no multi-association scope". The rollout plan is that neighbouring associations may follow (PRD §1). Adding an association column to tables full of locked, immutable snapshots later is a painful migration; enforcing isolation between associations that do not exist yet costs effort now. The existing `meetings` and `contracts` tables have no association column.

## Decision drivers

- PRD §3: records carry the association from the first row.
- QG-3: no cross-association leak once a second association exists.
- Locked snapshots are immutable (ADR-0007): the association id must be in them from the start.
- One developer, one association: no isolation machinery that cannot be tested against a real second tenant.

## Considered options

1. **`association_id` column + explicit filter in every sqlc query, no RLS.**
2. **`association_id` + Postgres Row-Level Security** with a session variable set per transaction and separate owner and app roles.
3. **No association column until a second association joins.**

## Decision outcome

**Chosen:** Option 1. Every table this feature owns (in `governance` and `identity`) carries `association_id uuid NOT NULL` holding the id of the one row in `identity.associations` (a foreign key inside `identity`, an id only in `governance` — no foreign keys across contexts); `meetings` gains it and is backfilled to that row in the same migration; `finance.contracts` is left as is in Stage 1. Every sqlc query on these tables takes `association_id` as a parameter and filters on it; the value comes from `ActorFrom(ctx)` ([ADR-0014](./0014-keep-board-protocol-in-governance-and-add-only-an-identity-context.md)), never from the request body. Unique keys include it (e.g. `(association_id, number)` on protocols). RLS is introduced by its own ADR when a second association joins.

## Consequences

**Positive**
- Data, snapshots and unique keys are multi-association-ready without a later backfill of immutable rows.
- No roles, policies or session variables to operate now; queries stay plain sqlc.

**Negative**
- Isolation rests on every query remembering the filter; a missed filter is invisible with one association. Mitigation: a test helper seeds a second association in `testdb.New(t)` and the integration tests assert none of its rows come back; a review checklist item for new queries.
- Moving to RLS later still needs roles, policies and a migration window.

**Neutral**
- The association row also anchors the settings (AC-42) and the Google connection.

## Links

- PRD: [PRD.md](../PRD.md) §1, §3, §6.1, AC-19, AC-42
- SAD: [sad.md](../sad.md) §4, §5, §8, §11
- Related ADR: [ADR-0007](./0007-lock-protocols-as-fingerprinted-immutable-snapshots.md), [ADR-0014](./0014-keep-board-protocol-in-governance-and-add-only-an-identity-context.md), [ADR-0017](./0017-reserve-the-protocol-number-at-first-print-and-never-reuse-it.md)
