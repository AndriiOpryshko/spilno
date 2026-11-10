---
id: T12
title: "Run goose as the owner and serve as spilno_app; give tests an app-role pool with two associations"
layer: "infra"
deps: ["T05"]
acs: ["AC-06", "AC-15"]
files_hint:
  - "services/core/internal/platform/config/config.go"
  - "services/core/cmd/api/main.go"
  - "services/core/internal/platform/postgres/"
  - "services/core/internal/platform/testdb/"
  - "infra/compose.yaml"
  - "mprocs.yaml"
  - ".github/workflows/ci.yml"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T12 — Run goose as the owner and serve as spilno_app; give tests an app-role pool with two associations

## Why

[sad](../sad.md) §7 «Two database roles»; [ADR-0003](../adr/0003-enforce-association-isolation-with-postgres-row-level-security.md) (the app connects without `BYPASSRLS`); [sad](../sad.md) §10 QG-4 and [PRD](../PRD.md) §6 «Integrity», «Retention»; [data-model](../data-model.md) «Test fixtures».

**Satisfies:** [PRD §5](../PRD.md) AC-06, AC-15 · **Blocked by:** [T05](./T05-promote-expand-migrations.md)

## What

- `MIGRATIONS_DATABASE_URL` (owner, runs goose on start; falls back to `DATABASE_URL`) and `DATABASE_URL` (a login role that is a member of `spilno_app`) in `config`, `cmd/api/main.go`, `infra/compose.yaml`, `mprocs.yaml` and CI.
- `testdb.NewApp(t)`: a fresh migrated database and a pool connected as a login role in `spilno_app`; `testdb.NewTwoAssociations(t)`: the same with associations A and B seeded.
- Privilege test for `application_extracts`, `application_filings`, `application_checks`, `application_decisions`, `association_state_changes`.

## Definition of Done

- [ ] Acceptance: The API migrates over MIGRATIONS_DATABASE_URL and serves over a login role in spilno_app, tests get an app-role pool with two seeded associations, and that role cannot update or delete the insert-only tables.
- [ ] integration test: as the app role, `UPDATE` and `DELETE` on each of the five tables are refused by the database; `INSERT` and `SELECT` work
- [ ] `mise run dev` starts with the two URLs; goose runs as the owner
- [ ] the existing `testdb.New(t)` callers are unchanged
- [ ] `mise run lint` and `mise run test` green

## Notes

No password in a migration: the login role is created per environment (compose init, testdb, production) and granted `spilno_app` (header of `NNNNN_create_app_role.sql`). board-protocol T12d edits `config.go` and `main.go` on `feat/core-google-config` — land after it.
