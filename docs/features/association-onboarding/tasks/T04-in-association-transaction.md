---
id: T04
title: "Add platform/db.InAssociation: a transaction that carries the association and the actor"
layer: "infra"
deps: []
acs: ["AC-04d", "AC-07"]
files_hint:
  - "services/core/internal/platform/db/"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T04 — Add platform/db.InAssociation: a transaction that carries the association and the actor

## Why

[ADR-0003](../adr/0003-enforce-association-isolation-with-postgres-row-level-security.md) «Shape»: row-level security reads `app.association_id` and `app.actor_email`, set per transaction. [ADR-0006](../adr/0006-apply-approvals-to-the-roster-through-a-governance-facade-in-one-transaction.md): the governance facade joins the caller's transaction through `ctx`. [sad](../sad.md) §11: this helper is the only way a handler gets a query runner on scoped tables.

**Satisfies:** [PRD §5](../PRD.md) AC-04d, AC-07 · **Blocked by:** —

## What

- `InAssociation(ctx, pool, associationID, actorEmail, fn)`: begin, `set_config('app.association_id', $1, true)`, `set_config('app.actor_email', $2, true)`, pass the transaction to `fn` through `ctx`, commit, or roll back on an error or a panic.
- `From(ctx)` returns the transaction as the sqlc `DBTX`; a nested `InAssociation` for the same association reuses it, for another association it is an error.
- Counter `platform.db.association_unset`: incremented when a scoped repository asks for a runner and no transaction is in `ctx` ([sad](../sad.md) §7).

## Definition of Done

- [ ] Acceptance: InAssociation runs fn in a transaction where app.association_id and app.actor_email are set, both are gone after it ends, and a scoped runner asked for outside it is counted.
- [ ] integration test on `testdb.New(t)`: inside `fn`, `current_setting('app.association_id', true)` and `app.actor_email` hold the given values; on the same connection after commit they are empty
- [ ] an error from `fn` rolls back what it wrote; a nested call with the same id reuses the transaction, with another id it fails
- [ ] the counter moves by one when `From` is called without a transaction
- [ ] `mise run lint` and `mise run test` green

## Notes

New package `internal/platform/db` (named in SAD §5); `internal/platform/postgres` keeps opening the pool and running goose. No policy exists yet, so nothing depends on the settings until T20.
