---
id: T12
title: "Store the association's Google connection and expose GoogleTokenSource"
layer: "infra"
deps: ["T01", "T07"]
acs: ["AC-26", "AC-14"]
files_hint:
  - "services/core/internal/identity/adapters/postgres/google_connections.go"
  - "services/core/internal/identity/adapters/postgres/queries/google_connections.sql"
  - "services/core/internal/identity/app/google.go"
  - "services/core/internal/platform/google/"
  - "services/core/cmd/google-connect/"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T12 — Store the association's Google connection and expose GoogleTokenSource

## Why

Files live in the association-owned Drive folder under `drive.file` ([ADR-0005](../adr/0005-store-files-in-association-owned-google-drive-folder.md)); the refresh token is stored encrypted ([sad §5](../sad.md) identity row). There is no member-facing story: consent is an operator step (contracts/api-sync-report.md deferral).

**Satisfies:** [PRD §5](../PRD.md) AC-26, AC-14 · **Blocked by:** [T01](./T01-stage-identity-migrations.md), [T07](./T07-merge-contract-identity-skeleton.md)

## What

- `google_connections` repository with AES-GCM encryption, key from config.
- One-time operator consent command `cmd/google-connect` (no API surface); creates the `spilno/` folder.
- `platform/google`: Drive and Docs clients built from a token source.

## Definition of Done

- [ ] Acceptance: A unit test round-trips the refresh token through AES-GCM, an integration test stores and reads a connection, and `GoogleTokenSource(ctx, association)` returns a token source built from it.
- [ ] unit test: encrypt/decrypt round trip; wrong key fails
- [ ] integration test: store + read connection; `GoogleTokenSource` returns a source
- [ ] no token or file content in logs (sloglint + review)
- [ ] `mise run lint` and `mise run test` green