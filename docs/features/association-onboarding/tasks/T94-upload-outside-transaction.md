---
id: T94
title: "Upload a case document before the write transaction and remove the file when the write fails"
layer: "app"
deps: ["T92", "T93"]
acs: ["AC-28"]
files_hint:
  - "services/core/internal/governance/app/cases.go"
  - "services/core/internal/governance/adapters/http/cases.go"
  - "services/core/internal/governance/ports/files.go"
  - "services/core/internal/governance/adapters/filesfake/"
  - "services/core/internal/governance/adapters/filestest/"
  - "services/core/internal/governance/adapters/gdrive/files.go"
  - "services/core/internal/governance/adapters/gdrive/files_test.go"
  - "services/core/internal/governance/cases_endpoints_test.go"
owner: "Andrii Opryshko"
estimate: "M"
status: "done"
---

# T94 — Upload a case document before the write transaction and remove the file when the write fails

## Why

[Review of 2026-11-17](../_review/review-2026-11-17.md), N06. A failed insert or commit after `files.Upload` leaves a stored file with no record; board-protocol SAD Flow 8 expects the upload before the transaction and the orphan removed. `ErrChecksumMismatch` falls through to 500.

**Satisfies:** [PRD §5](../PRD.md) AC-28 · **Blocked by:** [T92](./T92-files-per-association.md), [T93](./T93-guard-check-then-handler-for-outside-io.md)

## What

- `AddDocument`: checks (case, previous version, file) in a short transaction; upload; then one transaction for the file and the document.
- `Files` gains `Remove(ctx, id)`; the fake, the contract test and the Drive adapter implement it. A failed write removes the uploaded file, best effort, and logs a failed removal with the storage id.
- `ErrChecksumMismatch` maps to 503.

## Definition of Done

- [ ] Acceptance: An integration test with a failing insert after a successful upload shows no file left in the storage and a 5xx problem, a checksum mismatch answers 503 `governance.storage_unavailable`, and a successful upload records the file and the document in one transaction opened after the upload.
- [ ] `cases_endpoints_test.go`: failing insert, checksum mismatch, the happy path
- [ ] `filestest`: `Remove` is idempotent and a removed file is not found
- [ ] the log line of a failed removal carries the storage id and no file name
- [ ] `mise run lint` and `mise run test` green

## Notes

The scheduled `remove_orphan_file` job stays with board-protocol T14 (the River worker).
