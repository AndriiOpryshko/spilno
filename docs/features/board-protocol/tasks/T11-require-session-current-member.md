---
id: T11
title: "Require a session and a current член правління on the governance and finance routes"
layer: "wiring"
deps: ["T08", "T10"]
acs: ["AC-19", "AC-16"]
files_hint:
  - "services/core/cmd/api/main.go"
  - "services/core/internal/identity/adapters/http/middleware.go"
  - "services/core/internal/platform/httpapi/principal.go"
  - "services/core/internal/governance/meetings_test.go"
  - "services/core/internal/finance/"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T11 — Require a session and a current член правління on the governance and finance routes

## Why

AC-19, AC-16; [ADR-0003](../adr/0003-sign-in-with-oidc-and-postgres-sessions.md); [sad §6 Flow 7](../sad.md). Membership is checked on every request, before any resource is read.

**Satisfies:** [PRD §5](../PRD.md) AC-19, AC-16 · **Blocked by:** [T08](./T08-identity-oidc-sessions.md), [T10](./T10-settings-roster-endpoints.md)

## What

- `RequireSession` + current-member check wrapped around the governance and finance groups in `cmd/api/main.go`.
- `ActorFrom(ctx)` → account, person, association; existing governance and finance tests get a signed-in test helper.

## Definition of Done

- [ ] Acceptance: Integration tests show no session → 401 `identity.session_required`, an ended or never-held term → 403 `identity.not_current_member` with the same body for an existing and a missing id, `/system/health` stays public and the existing e2e smoke passes.
- [ ] integration tests for 401, 403 (existing vs missing id give an identical body)
- [ ] `mise run e2e` health smoke stays green
- [ ] existing finance and meetings tests pass with the new helper
- [ ] `mise run lint` and `mise run test` green

## Notes

Shares `cmd/api/main.go` with T07 (serialized by dep).
