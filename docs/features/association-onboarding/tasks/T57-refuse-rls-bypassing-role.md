---
id: T57
title: "Refuse to start when the serving role bypasses row-level security"
layer: "wiring"
deps: []
acs: ["AC-04d", "AC-07"]
files_hint:
  - "services/core/internal/platform/config/config.go"
  - "services/core/internal/platform/postgres/"
  - "services/core/cmd/api/main.go"
  - "mise.toml"
  - "README.md"
  - ".env.example"
owner: "Andrii Opryshko"
estimate: "S"
status: "todo"
---

# T57 — Refuse to start when the serving role bypasses row-level security

## Why

[Review of 2026-11-10](../_review/review-2026-11-10.md), F16. [ADR-0003](../adr/0003-enforce-association-isolation-with-postgres-row-level-security.md): the API serves as a role without `BYPASSRLS`. With `DATABASE_URL` unset the default is the owner, and nothing checks the role at startup.

**Satisfies:** [PRD §5](../PRD.md) AC-04d, AC-07 · **Blocked by:** —

## What

- After opening the serving pool, read `rolsuper OR rolbypassrls` for `current_user` and fail `run()` when it is true.
- `DATABASE_URL` has no default; `MIGRATIONS_DATABASE_URL` stays explicit. `mise run dev`, CI and README name both.

## Definition of Done

- [ ] Acceptance: The API exits with a clear error when the role of `DATABASE_URL` is a superuser or has `BYPASSRLS`, and `DATABASE_URL` has no owner default.
- [ ] test: starting with the owner URL as `DATABASE_URL` fails with the message
- [ ] `mise run dev` and `mise run e2e` still start (both already use `spilno_app_login`)
- [ ] `mise run lint` and `mise run test` green
