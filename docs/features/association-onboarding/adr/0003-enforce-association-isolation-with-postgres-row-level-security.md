---
status: accepted
date: 2026-11-05
feature: association-onboarding
sad_section: "§4"
deciders: ["Andrii Opryshko", "Claude (ao-architecture)"]
---

# 0003 Enforce association isolation with Postgres row-level security

## Context and problem statement

Board-protocol [ADR-0015](../../board-protocol/adr/0015-scope-owned-rows-to-one-association-with-explicit-query-filters.md) scoped rows by `association_id` with explicit sqlc filters and deferred row-level security «until a second association joins». This feature brings the second association. A missed filter means a stranger reads protocols, the top risk of brief §10. `finance.contracts` has no association column at all.

## Decision drivers

- PRD §6 Isolation: 0 records of association B readable by a person not on B's roster, in a release-blocking check.
- PRD §6.1 abuse case: cross-association read.
- Product owner 2026-11-05: row-level security now.
- `.claude/rules/migrations.md`: no triggers, no stored procedures; 3-step for a new NOT NULL column.

## Blast radius

- Irreversible: yes — roles, policies and the request transaction model shape every query and test.
- Multi-context: yes — identity, governance, finance, platform.
- Legitimate alternative: yes — explicit filters plus a guard test only.
- Touches the database: yes — policies on every association-owned table, two roles, a new column on `finance.contracts`.

## Considered options

1. **Row-level security now:** `ENABLE` and `FORCE ROW LEVEL SECURITY` on every table with `association_id`; policy `association_id = current_setting('app.association_id', true)::uuid`; the app connects as `spilno_app` (not owner, no `BYPASSRLS`); explicit filters stay.
2. **Explicit filters plus a release-blocking test and a query lint**, no row-level security.

## Decision outcome

**Chosen:** Option 1. The database refuses another association's rows even when a query forgets its filter; the explicit filters of ADR-0015 remain as a second line and the isolation test (SAD §10 QG-1) as the third.

Shape:
- `platform/db.InAssociation(ctx, associationID, actorEmail, fn)` begins a transaction, runs `set_config('app.association_id', $1, true)` and `set_config('app.actor_email', $2, true)` and passes the transaction to `fn` through `ctx`; the identity middleware opens it for every association-scoped request (the membership check runs inside it). The оператор's approval opens it with the id of the association it is creating (or transferring), so the roster insert passes the policy. River jobs do the same from their payload.
- An unset association yields no rows (fail closed); `platform.db.association_unset` counts it.
- The roster is readable across associations only for the person's own entries: the policy on `board_member_revisions` adds `OR sign_in_email = current_setting('app.actor_email', true)`, so `governance.MembershipsOf` works without bypassing the policy.
- Tables without an association (`accounts`, `sessions`, applications and extracts before approval) carry no policy and are reached only by account or by the оператор in `identity/app`. When the оператор reviews a запит на передачу, `identity/app` reads only the storage account address from `google_connections` inside `InAssociation` for that association (AC-28b); the оператор never gets a general bypass.

### Consequences

- Good: a forgotten filter is no longer a leak; the оператор role needs no bypass because the operator use cases never read association-owned tables (AC-16b); an оператор who is also a член правління reads them as a member, through `InAssociation` like anyone ([ADR-0009](./0009-treat-the-operator-role-and-a-roster-membership-as-independent-grants-on-one-account.md), clarified 2026-11-10).
- Bad: every scoped read runs inside a transaction; two roles to operate locally, in tests and in production; about one week more effort.
- Neutral: supersedes the «RLS later» trigger of board-protocol ADR-0015; the column and filters of ADR-0015 stay.

### Migration and rollback

- Migration: (1) a goose migration creates `spilno_app` and grants it data privileges (`SELECT, INSERT` only on insert-only tables); (2) `finance.contracts` + `association_id` nullable → backfill to the one existing association → `SET NOT NULL`, three migrations, then `CREATE INDEX CONCURRENTLY` in its own `-- +goose NO TRANSACTION` migration; (3) one migration per context enabling and forcing RLS with its policies; `DATABASE_URL` switches to `spilno_app` in the same release.
- Rollback: down migrations drop the policies and disable RLS; the column on `finance.contracts` stays (data kept); the app can return to the owner role.

## Links

- PRD: [PRD.md](../PRD.md) — §6 Isolation, AC-07, AC-04d
- SAD: [sad.md](../sad.md) §4, §7, §10
- Related ADR: [0002](./0002-put-the-association-id-in-every-association-scoped-path.md); board-protocol [ADR-0015](../../board-protocol/adr/0015-scope-owned-rows-to-one-association-with-explicit-query-filters.md); repo [0004 Migrations and queries](../../../adr/0004-migrations-and-queries.md)
