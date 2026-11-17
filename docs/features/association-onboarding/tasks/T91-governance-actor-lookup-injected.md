---
id: T91
title: "Pass the read-only lookup into governance from cmd/api and refuse a request without an actor"
layer: "wiring"
deps: ["T90"]
acs: ["AC-26"]
files_hint:
  - "services/core/internal/governance/adapters/http/files.go"
  - "services/core/internal/governance/adapters/http/handler.go"
  - "services/core/internal/governance/module.go"
  - "services/core/cmd/api/main.go"
  - "services/core/internal/governance/file_download_test.go"
owner: "Andrii Opryshko"
estimate: "S"
status: "done"
---

# T91 — Pass the read-only lookup into governance from cmd/api and refuse a request without an actor

## Why

[Review of 2026-11-17](../_review/review-2026-11-17.md), Q09. B8 removed the hidden identity dependency from the governance module; the HTTP adapter brought one back for `ActorFrom`, and a missing actor reads as «not read-only», so the pinned-file check of [ADR-0007](../adr/0007-enforce-a-disabled-association-as-read-only-in-the-identity-middleware.md) fails open.

**Satisfies:** [PRD §5](../PRD.md) AC-26 · **Blocked by:** [T90](./T90-one-definition-per-problem-code.md)

## What

- `governance.New` takes a function that reports whether the request's association is read-only and why, wired in `cmd/api` next to `hasStorage`.
- No actor: the handler answers the same 403 as a person off the roster.

## Definition of Done

- [ ] Acceptance: The governance HTTP adapter imports nothing from identity, and a download that reaches the handler without an actor in the context is refused, shown by a handler test.
- [ ] handler test: a context without an actor is refused
- [ ] `go list -deps` of the governance adapters shows no identity package
- [ ] `mise run lint` and `mise run test` green
