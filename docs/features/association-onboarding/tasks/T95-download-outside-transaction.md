---
id: T95
title: "Read a download from the storage outside the transaction and answer 503 when the read fails"
layer: "app"
deps: ["T91", "T92", "T93"]
acs: ["AC-26"]
files_hint:
  - "services/core/internal/governance/app/files.go"
  - "services/core/internal/governance/adapters/http/files.go"
  - "services/core/internal/governance/file_download_test.go"
  - "services/core/internal/governance/adapters/filesfake/"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T95 — Read a download from the storage outside the transaction and answer 503 when the read fails

## Why

[Review of 2026-11-17](../_review/review-2026-11-17.md), Q01, Q04. A 20 MB download sits in memory twice (the digest check and the guard's buffer) under an open transaction, and a read error from the storage is a 500 though the contract advertises 503.

**Satisfies:** [PRD §5](../PRD.md) AC-26 · **Blocked by:** [T91](./T91-governance-actor-lookup-injected.md), [T92](./T92-files-per-association.md), [T93](./T93-guard-check-then-handler-for-outside-io.md)

## What

- `Download`: find the record and check the pin in a transaction, then open, read and verify outside it.
- A read error that is not the request's own cancellation becomes `CodeStorageUnavailable`.

## Definition of Done

- [ ] Acceptance: Handler tests show the record and the pinned check read in a short transaction, the bytes read with no transaction open and held in memory once, and a storage failure in the middle of the read answering 503 `governance.storage_unavailable`.
- [ ] `file_download_test.go`: a fake that fails after some bytes; the read-only refusal and the byte comparison still pass
- [ ] `mise run lint` and `mise run test` green
