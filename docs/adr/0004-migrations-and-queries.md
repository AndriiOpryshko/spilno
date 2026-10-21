---
status: accepted
date: 2026-10-21
---

# 0004 goose migrations and sqlc queries

## Decision outcome

- Migrations: goose, one file per migration with `-- +goose Up` / `-- +goose Down`, in
  `services/core/migrations/`, embedded into the binary. The service applies pending migrations
  on start; tests apply them once into a template database.
- Queries: sqlc with `sql_package: pgx/v5`, one `sqlc.yaml` entry per context; schema is the
  migrations directory (sqlc reads only the Up sections), so generation needs no live database.
- Dynamic filters use `sqlc.narg` with `IS NULL OR` and a whitelisted `ORDER BY CASE`; a query
  that outgrows that is written by hand on pgx and reuses the sqlc row type.
- Squirrel was rejected: last release 2023, marked complete.

### Consequences

- Good: schema, queries and Go types cannot drift; `mise run gen:check` fails CI on stale code.
- Bad: functions and `DO` blocks need `StatementBegin/End` because goose splits on `;`.
