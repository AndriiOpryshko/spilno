---
id: T97
title: "Build each association's Drive storage from its Google connection in cmd/api"
layer: "wiring"
deps: ["T94", "T95"]
acs: ["AC-26", "AC-28"]
files_hint:
  - "services/core/cmd/api/main.go"
  - "services/core/internal/governance/module.go"
  - "services/core/internal/governance/adapters/gdrive/"
  - "services/core/internal/identity/google.go"
  - "services/core/internal/identity/app/google_connections.go"
  - "services/core/internal/platform/config/google.go"
  - "services/core/internal/platform/config/google_test.go"
  - "services/core/internal/governance/cases_endpoints_test.go"
  - ".env.example"
owner: "Andrii Opryshko"
estimate: "L"
status: "done"
---

# T97 — Build each association's Drive storage from its Google connection in cmd/api

## Why

[Review of 2026-11-17](../_review/review-2026-11-17.md), N09. Only the development fake is wired, so AC-26 and AC-28 hold on one machine. board-protocol T13b left «building the HTTP client of an association from `identity.GoogleTokenSource`, finding its folder id and handing the adapter to the governance use cases» to the first use case that stores a file; that use case was built here (T64, T65).

**Satisfies:** [PRD §5](../PRD.md) AC-26, AC-28 · **Blocked by:** [T94](./T94-upload-outside-transaction.md), [T95](./T95-download-outside-transaction.md)

## What

- Identity exposes the folder id of an association's connection next to `GoogleTokenSource`.
- Governance exposes `NewDriveFiles(client, folderID, apiURL, uploadURL)`; `cmd/api` builds the resolver of [T92](./T92-files-per-association.md) from the two, with the Drive addresses from the configuration (Google's by default).
- The fake Drive of the adapter's tests moves to a package the module test can import.

## Definition of Done

- [ ] Acceptance: With `SPILNO_DEV_STORAGE` unset, a test API whose Drive addresses point at a fake Drive uploads a case document into the folder of the association's connection with its token and downloads it back byte for byte; an association without a connection or a folder answers `governance.drive_not_connected` / `governance.storage_unavailable`.
- [ ] module-level test through HTTP: upload, download, bytes equal, the request to the fake Drive carries the association's token and folder
- [ ] two associations use two folders
- [ ] `config` test for the two Drive addresses
- [ ] the tracker's known gap about `SPILNO_DEV_STORAGE` is reworded: what remains is a run against real Google
- [ ] `mise run lint` and `mise run test` green

## Notes

Cannot be proven against real Google here (no test account, manual consent); the task says so in the tracker. The token lookup reads `google_connections`, which is under row-level security: resolve it inside a scoped transaction, before the outside call.
