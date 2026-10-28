---
id: T07
title: "Merge the contract delta, add the identity context skeleton and regenerate server and client"
layer: "wiring"
deps: []
acs: ["AC-19"]
files_hint:
  - "services/core/api/openapi.yaml"
  - "services/core/internal/identity/module.go"
  - "services/core/internal/identity/adapters/http/"
  - "services/core/internal/governance/adapters/http/"
  - "services/core/sqlc.yaml"
  - "services/core/.golangci.yml"
  - "services/core/cmd/api/main.go"
  - "apps/web/src/api/gen/"
owner: "Andrii Opryshko"
estimate: "L"
status: "todo"
---

# T07 — Merge the contract delta, add the identity context skeleton and regenerate server and client

## Why

One contract per repo (repo ADR 0003, .claude/rules/openapi.md): the delta in `contracts/openapi.yaml` merges into the live spec. Merging it once lets the web tasks start on generated hooks while the backend fills the stubs. New-context wiring per [sad §5](../sad.md).

**Satisfies:** [PRD §5](../PRD.md) — (enabling / NFR task) · **Blocked by:** —

## What

- Merge `docs/features/board-protocol/contracts/openapi.yaml` into `services/core/api/openapi.yaml` (SessionCookie scheme, identity and governance operations, schemas).
- Create `internal/identity/` (module.go, `adapters/http/gen/oapi-codegen.yaml`), an `identity` entry in `sqlc.yaml`, the `identity-private` depguard rule and widen the `_test.go` exclusion to `(system|finance|governance|identity)-private`.
- Register `identity.New(...)` in `cmd/api/main.go` with `r.Group` (never `r.Mount`).
- Stub every new StrictServerInterface method with a 501 problem; later tasks replace them.
- Do not enforce the session yet (T11) — existing e2e must stay green.

## Definition of Done

- [ ] Acceptance: `mise run gen` and `mise run gen:check` pass with the merged delta; every new operation answers a 501 problem from a stub; `mise run lint` and `mise run test` stay green.
- [ ] `mise run gen` + `mise run gen:check` clean
- [ ] a test calls one stub per context and gets 501 `application/problem+json`
- [ ] `mise run lint` (depguard) and `mise run test` green
- [ ] `mise run lint` and `mise run test` green

## Notes

Deviation, accepted: the rule «contract and handler land in the same PR» is met by stubs here and real handlers in T10, T08, T21–T31; each of those removes its stubs. Overlaps `cmd/api/main.go` with T11 and `.golangci.yml` with T24.
