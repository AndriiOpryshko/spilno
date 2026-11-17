---
id: T92
title: "Resolve the file storage per association"
layer: "ports"
deps: []
acs: ["AC-26", "AC-28"]
files_hint:
  - "services/core/internal/governance/ports/files.go"
  - "services/core/internal/governance/app/files.go"
  - "services/core/internal/governance/app/cases.go"
  - "services/core/internal/governance/module.go"
  - "services/core/internal/governance/adapters/filesfake/"
  - "services/core/cmd/api/main.go"
  - "services/core/internal/governance/cases_endpoints_test.go"
  - "services/core/internal/governance/file_download_test.go"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T92 — Resolve the file storage per association

## Why

[Review of 2026-11-17](../_review/review-2026-11-17.md), N09. The `Files` port is one per process and takes no association, while [board-protocol ADR-0005](../../board-protocol/adr/0005-store-files-in-association-owned-google-drive-folder.md) gives every association its own folder and the Drive adapter is built per association. The development fake is shared by every association.

**Satisfies:** [PRD §5](../PRD.md) AC-26, AC-28 · **Blocked by:** —

## What

- `WithFiles` takes a resolver `func(ctx, association) (Files, error)`; `FileService` and `CaseService` ask it per call.
- `NewDevFiles` returns a resolver that keeps one in-memory storage per association.
- `cmd/api` passes the development resolver under `SPILNO_DEV_STORAGE`, as today.

## Definition of Done

- [ ] Acceptance: Integration tests on two associations show a file uploaded in one is not readable through the storage of the other, and an association without a storage answers 409 `governance.drive_not_connected` on upload and 503 `governance.storage_unavailable` on download.
- [ ] `testdb.NewTwoAssociations(t)`: upload in A, download of the same storage id through B fails
- [ ] the e2e stack (`SPILNO_DEV_STORAGE=true`) behaves as before
- [ ] `mise run lint` and `mise run test` green

## Notes

The real resolver is [T97](./T97-wire-drive-storage-per-association.md).
