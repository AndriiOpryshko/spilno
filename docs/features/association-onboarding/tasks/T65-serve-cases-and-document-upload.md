---
id: T65
title: "Serve справи and document upload behind the storage guard"
layer: "ports"
deps: ["T64"]
acs: ["AC-25", "AC-28"]
files_hint:
  - "services/core/internal/governance/domain/case.go"
  - "services/core/internal/governance/app/cases.go"
  - "services/core/internal/governance/adapters/postgres/cases.go"
  - "services/core/internal/governance/adapters/postgres/queries/cases.sql"
  - "services/core/internal/governance/adapters/http/cases.go"
  - "services/core/internal/governance/adapters/http/stubs.go"
  - "services/core/internal/governance/module.go"
owner: "Andrii Opryshko"
estimate: "L"
status: "todo"
---

# T65 — Serve справи and document upload behind the storage guard

## Why

[Review of 2026-11-10](../_review/review-2026-11-10.md), F15. [PRD](../PRD.md) AC-28: a член правління who attaches a document to a справа without a connected storage is told to connect it first. The guard answers that 409 already, but there is no справа to attach to: `ListCases`, `CreateCase` and `GetCase` answer 501.

**Satisfies:** [PRD §5](../PRD.md) AC-25, AC-28 · **Blocked by:** [T64](./T64-serve-file-download-read-only-check.md)

## What

- The rest of [board-protocol T21](../../board-protocol/tasks/T21-cases-endpoints.md): `listCases`, `createCase`, `getCase`, `addCaseDocument` on the migrated `cases`, `case_documents` and `files` tables.
- A newer version of a document is a new row; uploads over 20 MB or outside PDF/JPEG/PNG answer 422 `governance.file_invalid`.
- The storage guard stays the first step of `addCaseDocument`.

## Definition of Done

- [ ] Acceptance: Handler tests show a справа created, listed and opened, a document attached through the fake `Files` once the storage is connected, 409 `governance.drive_not_connected` before that, and every change refused in a вимкнене ОСББ.
- [ ] handler tests for each operation and error with the fake `Files`
- [ ] isolation test with two associations
- [ ] create and upload answer 403 `identity.association_disabled` while disabled
- [ ] no Drive id in any response body
- [ ] `mise run lint` and `mise run test` green

## Notes

Carries board-protocol T21 apart from its domain event for the needs-review hook (board-protocol T23) and the upload-duration metric, which stay there. This is the slice to drop first if the follow-up must shrink: without it the UI half of AC-28 needs a de-scope in PRD §8.
