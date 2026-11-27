---
id: T53
title: "Limit request bodies and read time before the request validator runs"
layer: "infra"
deps: []
acs: ["AC-02"]
files_hint:
  - "services/core/internal/platform/httpapi/"
  - "services/core/cmd/api/main.go"
  - "services/core/internal/identity/module.go"
  - "services/core/internal/governance/module.go"
  - "services/core/internal/identity/applications_file_test.go"
owner: "Andrii Opryshko"
estimate: "M"
status: "done"
---

# T53 — Limit request bodies and read time before the request validator runs

## Why

[Review of 2026-11-10](../_review/review-2026-11-10.md), B1. [PRD](../PRD.md) §6: the extract is at most 5 MB. The request validator buffers a multipart body before the handler's size check and before any session check on `POST /identity/applications`, and the server has no `ReadTimeout`.

**Satisfies:** [PRD §5](../PRD.md) AC-02 · **Blocked by:** —

## What

- A body-limit middleware ahead of `RequestValidator`: about 6 MiB on the multipart operations, 1 MiB elsewhere.
- `ReadTimeout` on the HTTP server next to the existing `ReadHeaderTimeout`.
- The overrun answers 413 as RFC 9457 problem+json; add the response to the contract if the validator would otherwise reject it.

## Definition of Done

- [ ] Acceptance: A request whose body exceeds the limit of its operation is refused with problem+json before the validator reads it, with or without a session, and the server has a read timeout.
- [ ] a 20 MB multipart body without a session is refused and the handler is never entered
- [ ] the existing «5 MB + 1 byte» test still names the file field with 422
- [ ] governance uploads keep their own 20 MB rule (board-protocol T21): the limit is per operation
- [ ] `mise run lint` and `mise run test` green
