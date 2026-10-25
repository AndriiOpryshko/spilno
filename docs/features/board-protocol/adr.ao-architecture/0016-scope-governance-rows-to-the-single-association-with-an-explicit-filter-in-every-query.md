---
status: accepted
date: 2026-10-25
feature: board-protocol
sad_section: "§5"
deciders: ["Andrii Opryshko"]
---

# 0016 Scope governance rows to the single association with an explicit filter in every query

## Context and problem statement

Stage 1 serves one association; records carry the association, but there is no multi-association scope (PRD §3, §9 Association scope). ADR 0010 chose Postgres Row-Level Security with `InAssociation` transactions and two database roles, which protects against a second association that Stage 1 does not have, at the cost of a transaction per request and harder tests. AC-33 still answers "not found" for an identifier not stored for the association. This ADR supersedes ADR 0010.

## Decision drivers

- PRD §3 / §9: one association in Stage 1; records carry the association; onboarding a second one is a later feature.
- AC-33: an identifier not stored for the association answers not found and changes nothing.
- PRD §9: existing meetings are assigned to the association, confirmed by a член правління.
- Repo ADR 0002: no foreign key from `governance` to `identity` tables.

## Blast radius

- Irreversible: yes — every governance table and query carries the column from now on; adding RLS later is additive, removing the column is not.
- Multi-context: yes — `governance` (all tables) and `identity` (source of the association id through `ActorFrom`).
- Legitimate alternative: yes — the RLS of ADR 0010.
- Touches the database: yes — new column on `meetings` with a backfill; the column on every new table.

## Considered options

1. **`association_id` column + explicit filter in every sqlc query**, one database role.
2. **`association_id` column + Row-Level Security** (ADR 0010) — policies, `SET LOCAL app.association_id`, roles `spilno_owner` / `spilno_app`.

## Decision outcome

**Chosen:** option 1. Every `governance` table has `association_id uuid NOT NULL`; every sqlc query takes `association_id` as a parameter and filters on it, and every lookup by id is `WHERE id = $1 AND association_id = $2`, so an id of another association is "not found" (AC-33). Repositories take the association from `identity.ActorFrom(ctx)` in `app/`; worker handlers take it from the job payload, and the nightly `verify_integrity` job iterates the association records (one in Stage 1) and passes each id to its queries, so no query runs without it. Indexes lead with `association_id`. The app keeps one database role; migrations run on API start as today (repo ADR 0004). `finance.contracts` stays untouched in Stage 1 (finance is not read by governance until ADR 0012 in Stage 2) and gets the column then.

### Consequences

- Good: no per-request transaction, no second role, plain `psql` debugging; tests stay on `testdb.New(t)` as is.
- Bad: one forgotten filter would leak rows the day a second association exists — mitigated by a review rule and a test that every `queries/*.sql` statement on a governance table mentions `association_id` (SAD §8, §11).
- Neutral: RLS stays possible later as an extra guard over the same column (SAD §11 accepted debt).

### Migration and rollback

- Migration: (1) the identity migration creates the association record (ADR 0003); (2) add nullable `association_id` to `meetings`; (3) backfill existing meetings to that association, confirmed by a член правління at setup (PRD §9); (4) `SET NOT NULL` and an index. New governance tables are created with the column.
- Rollback: drop the index and the `NOT NULL`; the column stays because signed protocols and their meetings carry it.

## Links

- PRD: [PRD.ao-prd.md](../PRD.ao-prd.md) — AC-33; §3 "More than one association"; §8 association OQ; §9 Association scope
- SAD: [sad.ao-architecture.md](../sad.ao-architecture.md) §5, §8, §11
- Related ADR: supersedes [0010 Row-Level Security](./0010-scope-every-governance-and-finance-row-to-an-association-with-row-level-security.md) · [0003 Sign in with Google](./0003-sign-in-with-google-in-a-new-identity-context.md) · repo [0002 Bounded contexts](../../../adr/0002-bounded-contexts.md)
