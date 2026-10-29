---
id: T13
title: "Implement the Drive Files adapter with sha256 checks and an in-memory fake"
layer: "infra"
deps: ["T12"]
acs: ["AC-26", "AC-14", "AC-17"]
files_hint:
  - "services/core/internal/governance/ports/files.go"
  - "services/core/internal/governance/adapters/gdrive/"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T13 — Implement the Drive Files adapter with sha256 checks and an in-memory fake

## Why

[ADR-0005](../adr/0005-store-files-in-association-owned-google-drive-folder.md); [sad §6 Flow 8](../sad.md) (storage unavailable → `governance.storage_unavailable`).

**Satisfies:** [PRD §5](../PRD.md) AC-26, AC-14, AC-17 · **Blocked by:** [T12](./T12-google-connection-token-source.md)

## What

- Port in `governance/ports/files.go`; Drive adapter in `adapters/gdrive`; fake for unit/integration tests.
- Map Drive auth/5xx errors to a storage-unavailable domain error.

## Definition of Done

- [ ] Acceptance: The `Files` port (upload, stream, sha256, set read-only, untrash) has a Drive adapter and an in-memory fake that pass the same contract test; the Drive adapter compares its computed sha256 with Drive's.
- [ ] shared contract test passes for the fake
- [ ] Drive adapter contract test runs behind a build tag / env with a test Google account (skipped in CI by default)
- [ ] sha256 mismatch is reported as an error
- [ ] `mise run lint` and `mise run test` green