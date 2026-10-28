---
id: T01
title: "Promote the identity migrations: accounts, sessions, google_connections"
layer: "migration"
deps: []
acs: ["AC-19", "AC-16"]
files_hint:
  - "docs/features/board-protocol/migrations/20261027120002_create_accounts.up.sql"
  - "docs/features/board-protocol/migrations/20261027120002_create_accounts.down.sql"
  - "docs/features/board-protocol/migrations/20261027120003_create_sessions.up.sql"
  - "docs/features/board-protocol/migrations/20261027120003_create_sessions.down.sql"
  - "docs/features/board-protocol/migrations/20261027120004_create_google_connections.up.sql"
  - "docs/features/board-protocol/migrations/20261027120004_create_google_connections.down.sql"
owner: "Andrii Opryshko"
estimate: "S"
status: "todo"
---

# T01 — Promote the identity migrations: accounts, sessions, google_connections

## Why

Sign-in needs accounts and server-side sessions ([ADR-0003](../adr/0003-sign-in-with-oidc-and-postgres-sessions.md)); Drive access needs the association's Google connection ([ADR-0005](../adr/0005-store-files-in-association-owned-google-drive-folder.md)). Tables per [data-model.md](../data-model.md) §identity.

**Satisfies:** [PRD §5](../PRD.md) AC-19, AC-16 · **Blocked by:** —

## What

- Promote staged pairs 02–04 with `/promote-migration board-protocol` into the next goose numbers after `00008`.
- Staged 00 (`associations`) and 01 (bootstrap) are already live as `00004_create_associations.sql` — skip them.

## Definition of Done

- [ ] Acceptance: The three identity migrations are live goose files that apply, revert and re-apply cleanly in `mise run test`.
- [ ] staged pairs 02–04 are live goose files under `services/core/migrations/`
- [ ] `/roundtrip` (up → down-to baseline → up) shows an empty `pg_dump --schema-only` diff
- [ ] `mise run gen` regenerates sqlc with no diff outside identity
- [ ] `mise run lint` and `mise run test` green

## Notes

Migration lane: serialized with T02–T05 in this order.
