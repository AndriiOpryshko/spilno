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
status: "in_progress"
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

## Status (2026-11-17)

The port and the fake are done by [T13a](./T13a-files-port-fake.md), the adapter on a fake Drive by
[T13b](./T13b-drive-files-adapter.md); both pass the shared contract of `adapters/filestest`, the sha256 check
included. association-onboarding built on them after its review of 2026-11-17:

- [T92](../../association-onboarding/tasks/T92-files-per-association.md) (commit `159505b`): the storage is resolved per association
  (`ports.FilesResolver`), as ADR-0005 gives every association its own folder.
- [T94](../../association-onboarding/tasks/T94-upload-outside-transaction.md) (commit `d38dbc5`): `Files.Remove` on the port, the fake,
  the contract and the adapter.
- [T97](../../association-onboarding/tasks/T97-wire-drive-storage-per-association.md) (commit `f6fde66`): `cmd/api` builds the adapter
  of each association from its Google connection.

Left here: the contract run against real Google, behind a build tag or an environment variable with a test
Google account (the third item of the Definition of Done). No such test exists yet; the adapter and its wiring
have run against fakes only. A real connection cannot be created before that either: `cmd/api` builds the
storage consent without its Drive address
([security review](../../association-onboarding/_review/security-2026-11-10.md), SEC-13).
