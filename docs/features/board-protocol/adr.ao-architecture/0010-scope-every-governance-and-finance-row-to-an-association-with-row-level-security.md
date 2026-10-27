---
status: superseded
superseded_by: "0016"
date: 2026-10-25
feature: board-protocol
sad_section: "§5"
deciders: ["Andrii Opryshko"]
---

# 0010 Scope every governance and finance row to an association with Row-Level Security and backfill existing rows to the first one

> **Status:** superseded by [0016](./0016-scope-governance-rows-to-the-single-association-with-an-explicit-filter-in-every-query.md) (2026-10-25): Stage 1 serves one association; an explicit `association_id` filter replaces RLS, which moves to SAD §11 accepted debt.

## Context and problem statement

The app stores no association today (SAD §2). PRD §6.1 classes the records as Confidential, AC-33 answers "not found" to a member of another association, and PRD §9 assigns existing meetings and contracts to the first association, confirmed by a член правління. This changes what every existing and future row of `governance` and `finance` means.

## Decision drivers

- QG-3: one missed filter must not leak another association's protocol.
- PRD §9: existing meetings and contracts go to the first association; no other backfill.
- Repo ADR 0002: no cross-context table access, so no foreign key from `governance`/`finance` to `identity.associations`.

## Blast radius

- Irreversible: yes — every table and query carries the scope from now on.
- Multi-context: yes — `governance`, `finance`, `identity` (source of the association) and `platform/postgres`.
- Legitimate alternative: yes — the column with a filter in every query and no RLS.
- Touches the database: yes — new column, backfill of existing rows, policies.

## Considered options

1. **`association_id` column + filter in every sqlc query** — no database policy.
2. **`association_id` column + Postgres Row-Level Security** — policies hide rows of other associations even when a query forgets the filter.

## Decision outcome

**Chosen:** option 2. Every `governance` and `finance` table has `association_id uuid NOT NULL` and `ENABLE` + `FORCE ROW LEVEL SECURITY` with the policy `association_id = current_setting('app.association_id')::uuid` for reads and writes. Requests run inside `platform/postgres.InAssociation(ctx, pool, association, fn)`, which opens a transaction and runs `SET LOCAL app.association_id` from `identity.ActorFrom(ctx)`; a query outside it sees no rows. Queries still filter by `association_id` explicitly, so plans use the indexes. `identity` tables are not under RLS (it resolves the association), nor is `governance_jobs` (the worker claims across associations and then runs each job inside `InAssociation`). The app connects as role `spilno_app` (no `BYPASSRLS`). As repo ADR 0004 requires, `cmd/api` still applies pending migrations on start: it opens a separate short-lived connection as `spilno_owner` (`DATABASE_OWNER_URL`), runs goose, closes it, and then serves through the `spilno_app` pool; repo ADR 0004 is not superseded. `testdb.New(t)` creates both roles.

### Consequences

- Good: a forgotten filter returns nothing instead of another association's records; AC-33 holds by construction.
- Bad: every request is a transaction; connection setup and test helpers get more complex; two database roles in config.
- Bad: debugging with `psql` needs `SET app.association_id` first.
- Neutral: an integration test "two associations" per route stays in the test plan as a second guard.

### Migration and rollback

- Migration: (1) add nullable `association_id` to `meetings` and `contracts`; (2) the identity migration seeds the first association; a member confirms it at setup; (3) backfill existing `meetings` and `contracts` to it; (4) set `NOT NULL`, add indexes, enable and force RLS, create the policies and roles. New tables are created with the column and policies from the start.
- Rollback: disable RLS and drop the policies; the column stays because signed protocols depend on it.

## Links

- PRD: [PRD.ao-prd.md](../PRD.ao-prd.md) — AC-33, §6.1 abuse case "cross-association access", §8 association OQ, §9 Association scope
- SAD: [sad.ao-architecture.md](../sad.ao-architecture.md) §5, §8
- Related ADR: [0003 Sign in with Google](./0003-sign-in-with-google-in-a-new-identity-context.md) · repo [0002 Bounded contexts](../../../adr/0002-bounded-contexts.md)
