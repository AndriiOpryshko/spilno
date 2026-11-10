---
id: T01
title: "Move the governance and finance operations under `/<context>/associations/{association_id}/`"
layer: "wiring"
deps: []
acs: ["AC-04d", "AC-07"]
files_hint:
  - "services/core/api/openapi.yaml"
  - "services/core/internal/governance/adapters/http/"
  - "services/core/internal/governance/meetings_test.go"
  - "services/core/internal/governance/stubs_test.go"
  - "services/core/internal/finance/adapters/http/"
  - "services/core/internal/finance/module_test.go"
  - "apps/web/src/features/governance/api.ts"
  - "apps/web/src/features/governance/types.ts"
owner: "Andrii Opryshko"
estimate: "L"
status: "todo"
---

# T01 — Move the governance and finance operations under `/<context>/associations/{association_id}/`

## Why

[ADR-0002](../adr/0002-put-the-association-id-in-every-association-scoped-path.md): the association is explicit in every association-scoped path. [sad](../sad.md) §11 ranks the collision with board-protocol tasks in flight as the one High risk and asks for the path move as the first task of this feature, in one PR; [api-sync-report](../contracts/api-sync-report.md) §E «Path move» lists the mechanics.

**Satisfies:** [PRD §5](../PRD.md) AC-04d, AC-07 · **Blocked by:** —

## What

- In `services/core/api/openapi.yaml`: every `/governance/<rest>` becomes `/governance/associations/{association_id}/<rest>` and every `/finance/<rest>` becomes `/finance/associations/{association_id}/<rest>`; add the `AssociationId` parameter component; operationIds do not change. `/system/health` and the `/identity/*` operations stay where they are.
- `mise run gen`; adapt the governance handler, its 30 stubs and the finance handler to the regenerated request objects. Meetings take the association from the path instead of the fixed id; finance accepts the id (it is stored from T13).
- Update `meetings_test.go`, `stubs_test.go` and `finance/module_test.go` to the new paths; make `features/governance/api.ts` and `types.ts` compile against the regenerated client. Web routes move in T35.

## Definition of Done

- [ ] Acceptance: Every governance and finance operation of the live contract carries {association_id} in its path, generated code is in sync and a meeting created under one association is not listed under another.
- [ ] a Go test over the embedded spec: every operation tagged `governance` or `finance` has the `{association_id}` path parameter (T21 reuses the walk)
- [ ] integration test: a meeting created under association A is not in the list under B's path
- [ ] `mise run gen:check` shows no diff
- [ ] `mise run lint` and `mise run test` green

## Notes

No external client exists, so no old path is kept (ADR-0002). Merge this before any board-protocol task that adds a governance handler or a web route (board-protocol T10, T21–T31, T36–T43): they then build on the new paths. Until T19 nothing checks who calls a path — that is today's state, not a regression. Shares the `openapi.yaml` lane with T02.
