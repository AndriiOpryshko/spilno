---
id: T21
title: "Serve cases: create, list, open, add document versions and download files"
layer: "ports"
deps: ["T03", "T11", "T13"]
acs: ["AC-19", "AC-26", "AC-09"]
files_hint:
  - "services/core/internal/governance/domain/case.go"
  - "services/core/internal/governance/app/cases.go"
  - "services/core/internal/governance/adapters/postgres/cases.go"
  - "services/core/internal/governance/adapters/postgres/queries/cases.sql"
  - "services/core/internal/governance/adapters/http/cases.go"
  - "services/core/internal/governance/adapters/http/files.go"
owner: "Andrii Opryshko"
estimate: "L"
status: "todo"
---

# T21 — Serve cases: create, list, open, add document versions and download files

## Why

[ADR-0005](../adr/0005-store-files-in-association-owned-google-drive-folder.md) (Drive storage, metadata + sha256 in Postgres), PRD §6 upload NFR, §6.1 malicious upload; AC-19 (cases hidden outside the term — via T11).

**Satisfies:** [PRD §5](../PRD.md) AC-19, AC-26, AC-09 · **Blocked by:** [T03](./T03-stage-case-protocol-migrations.md), [T11](./T11-require-session-current-member.md), [T13](./T13-drive-files-adapter.md)

## What

- справа + document versions (a newer version of a document is a new row; emits a domain event for T23's needs-review hook).
- multipart upload streamed to `Files`; download streamed through the API.

## Definition of Done

- [ ] Acceptance: Handler tests show listCases / createCase / getCase / addCaseDocument / downloadFile return the contract shapes, uploads over 20 MB or outside PDF/JPEG/PNG get 422 `governance.file_invalid`, Drive failure gives 503 `governance.storage_unavailable`, and downloads are `Content-Disposition: attachment` with no Drive id on the wire.
- [ ] handler tests with the fake `Files` for each operation and error
- [ ] test: response bodies never contain Drive file or revision ids
- [ ] upload-duration metric recorded
- [ ] `mise run lint` and `mise run test` green