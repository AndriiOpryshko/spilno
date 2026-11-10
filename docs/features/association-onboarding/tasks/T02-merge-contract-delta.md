---
id: T02
title: "Merge the onboarding contract delta and stub the new identity operations"
layer: "wiring"
deps: ["T01"]
acs: ["AC-01", "AC-10", "AC-17", "AC-21", "AC-27", "AC-29"]
files_hint:
  - "services/core/api/openapi.yaml"
  - "services/core/internal/identity/adapters/http/handler.go"
  - "services/core/internal/identity/module_test.go"
  - "docs/features/association-onboarding/contracts/api-sync-report.md"
owner: "Andrii Opryshko"
estimate: "L"
status: "todo"
---

# T02 — Merge the onboarding contract delta and stub the new identity operations

## Why

[contracts/openapi.yaml](../contracts/openapi.yaml) is a delta that `implement-tasks` merges into the one live contract (`.claude/rules/openapi.md`); [api-sync-report](../contracts/api-sync-report.md) §E is the merge checklist. Merged once with stubs so the ui lane (T35–T42) can start on generated hooks.

**Satisfies:** [PRD §5](../PRD.md) AC-01, AC-10, AC-17, AC-21, AC-27, AC-29 · **Blocked by:** [T01](./T01-move-paths-under-association.md)

## What

- Merge per [api-sync-report](../contracts/api-sync-report.md) §E: the 20 new `identity` operations with their parameters, responses, examples and schemas; `Principal` becomes `{account_id, email, is_operator}`; `getSession` and `completeSignIn` lose their 403; `BoardMember` gains `position` and `decision_id`.
- Every moved governance and finance operation answers 403 through `MemberOrDisabled`; `addCaseDocument` and `createAssistantTurn` gain 409 `governance.drive_not_connected` with its example.
- `mise run gen`; `identity` `Handler` gets one stub per new operation returning `httpapi.ErrNotImplemented`, as `governance/adapters/http/stubs.go` does.
- Lint the merged contract with `npx @stoplight/spectral-cli lint`; note the merge in `api-sync-report.md`.

## Definition of Done

- [ ] Acceptance: The live contract holds the 20 new identity operations and the changed components, generated server and client are in sync, and every new operation answers 501 problem+json from a stub.
- [ ] a stubs test (as `governance/stubs_test.go`): each of the 20 new operations answers 501 `application/problem+json`
- [ ] spectral reports 0 errors on `services/core/api/openapi.yaml`
- [ ] `mise run gen:check` shows no diff; `pnpm --filter @spilno/web typecheck` passes on the regenerated hooks
- [ ] `mise run lint` and `mise run test` green

## Notes

Accepted deviation from «contract and handler in the same PR», as in board-protocol T07: each ports task removes its own stubs. `info.version` moves only with a line that says why (`.claude/rules/openapi.md`). Two deferrals of [api-sync-report](../contracts/api-sync-report.md) §D stay open and are not invented here: the response for a bad storage `state` (see T33) and editing the association's name and address (no AC).
