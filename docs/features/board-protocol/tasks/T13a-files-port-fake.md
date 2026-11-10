---
id: T13a
title: "Declare the Files port and an in-memory fake that passes a shared contract test"
layer: "infra"
parent: T13
acs: ["AC-26", "AC-14", "AC-17"]
scope: governance
package: services/core/internal/governance/adapters/filesfake
test_cmd: "cd services/core && go test -count=1 ./internal/governance/..."
test_files:
  - services/core/internal/governance/adapters/filestest/contract.go
  - services/core/internal/governance/adapters/filesfake/fake_test.go
estimate: "half a day"
status: "done"
---

# T13a — Declare the Files port and an in-memory fake that passes a shared contract test

Slice of [T13](./T13-drive-files-adapter.md) that needs no Google account: the port, the fake every later
use case and worker test runs on, and the contract test the Drive adapter (T13b, after
[T12](./T12-google-connection-token-source.md)) must pass as well. [ADR-0005](../adr/0005-store-files-in-association-owned-google-drive-folder.md),
[sad §6 Flow 8](../sad.md) (storage unavailable → `governance.storage_unavailable`, state unchanged).

## What

- `governance/ports/files.go`: the `Files` port — upload (sha256 computed while streaming), stream back,
  stat (sha256, revision, trashed, read-only as the storage reports them), set read-only, untrash —
  plus `ErrFileNotFound` and `ErrChecksumMismatch`.
- `domain.CodeStorageUnavailable` for `governance.storage_unavailable`, used in `ErrorRegistry`.
- `governance/adapters/filesfake`: in-memory fake with two faults a test can switch on: the storage is
  unavailable (every call → `storage_unavailable`), the storage reports a wrong sha256 (upload →
  `ErrChecksumMismatch`, nothing stored).
- `governance/adapters/filestest`: the shared contract suite (written first, part of the acceptance test).

## Definition of Done

- [ ] Acceptance: `filestest.Run` passes for the fake; an unavailable storage answers `governance.storage_unavailable`
  from every method; a sha256 the storage reports differently from the one computed while streaming fails the upload
  with `ErrChecksumMismatch` and stores nothing.
- [ ] the acceptance tests (`filestest/contract.go`, `filesfake/fake_test.go`) are unchanged
- [ ] no Google SDK import (the fake is stdlib only); depguard clean
- [ ] `go test ./internal/governance/...` and `golangci-lint run ./internal/governance/...` green
