---
id: T98
title: "Answer 413 for every oversized body and limit the bodies of finance and system"
layer: "infra"
deps: []
acs: ["AC-02"]
files_hint:
  - "services/core/internal/platform/httpapi/bodylimit.go"
  - "services/core/internal/platform/httpapi/problem.go"
  - "services/core/internal/platform/httpapi/problem_test.go"
  - "services/core/internal/finance/module.go"
  - "services/core/internal/system/module.go"
  - "services/core/internal/governance/adapters/http/cases.go"
  - "services/core/internal/governance/cases_endpoints_test.go"
  - "services/core/internal/identity/applications_file_test.go"
  - "services/core/internal/finance/module_test.go"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T98 — Answer 413 for every oversized body and limit the bodies of finance and system

## Why

[Review of 2026-11-17](../_review/review-2026-11-17.md), Q02. A body without `Content-Length` trips `MaxBytesReader` inside the validator and is reported as a 422 field error, or as a 500 from `readUpload`; finance has no limit at all.

**Satisfies:** [PRD §5](../PRD.md) AC-02 · **Blocked by:** —

## What

- `validationErrorHandler`, `StrictRequestError` and `readUpload` map `*http.MaxBytesError` to the 413 problem.
- `BodyLimit` on the finance and system groups.

## Definition of Done

- [ ] Acceptance: Tests show a chunked body over the limit answering 413 problem+json on an identity, a governance and a finance operation, and the 21 MB multipart limit of governance refused before the handler.
- [ ] `problem_test.go`: the mapping
- [ ] one chunked-oversize test per context
- [ ] the governance multipart limit
- [ ] `mise run lint` and `mise run test` green
