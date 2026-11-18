---
id: T130
title: "Bring the contract in line with the sign-in cookie, the 413 answers, the term-end refusal and the decision effect"
layer: "wiring"
deps: ["T114"]
acs: ["AC-02", "AC-05", "AC-19", "AC-20", "AC-26"]
files_hint:
  - "services/core/api/openapi.yaml"
  - "docs/features/association-onboarding/contracts/openapi.yaml"
  - "services/core/internal/identity/adapters/http/gen/"
  - "services/core/internal/governance/adapters/http/gen/"
  - "services/core/internal/finance/adapters/http/gen/"
  - "services/core/internal/system/adapters/http/gen/"
  - "services/core/internal/governance/adapters/http/board_members.go"
  - "apps/web/src/api/gen/"
  - "apps/web/src/features/governance/roster/"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T130 — Bring the contract in line with the sign-in cookie, the 413 answers, the term-end refusal and the decision effect

## Why

[Second re-review of 2026-11-17](../_review/review-2026-11-17-2.md), R08, R10, R11. `startSignIn` still describes server-side state and lists no `Set-Cookie`; `completeSignIn` does not name the cleared cookie or the 401 of a missing or foreign cookie; no operation advertises 413; the 422 of `updateBoardMember` does not name the refusal on `/term_end`; `enum: [seated, ended, null]` generates a Go constant `"<nil>"`; the delta carries neither `decision_effect` nor the 403 of `downloadFile`.

**Satisfies:** [PRD §5](../PRD.md) AC-02, AC-05, AC-19, AC-20, AC-26 · **Blocked by:** [T114](./T114-refuse-another-storage-account-while-files-exist.md)

## What

- Sign-in: both descriptions, `Set-Cookie` on the 302 and on the callback, the 401 text.
- A shared `PayloadTooLarge` response on the operations that take a file (`fileApplication`, `refileApplication`, `addCaseDocument`).
- `updateBoardMember` 422 names the ended-entry refusal.
- `DecisionEffect` is a string enum by `$ref`, nullable through `oneOf` with `type: "null"`.
- The delta carries `BoardMember.decision_effect` and the 403 of `downloadFile`; `mise run gen`.

## Definition of Done

- [ ] Acceptance: `mise run gen` leaves no diff, the generated Go has no `"<nil>"` constant, spectral reports 0 errors on the live contract and on the delta, and the delta equals the live contract for every operation and schema it carries.
- [ ] `npx @stoplight/spectral-cli lint` on both files: 0 errors
- [ ] a test or grep in CI's gen check: no `<nil>` in `governance.gen.go`
- [ ] the roster handler and the web roster compile against the regenerated types; their tests pass
- [ ] a contract test asserts the 413 body of an oversized filing against the advertised response
- [ ] `mise run lint` and `mise run test` green

## Notes

Contract lane: lands after [T114](./T114-refuse-another-storage-account-while-files-exist.md), which adds its code to the same file. No `info.version` bump without a CHANGELOG line (openapi.md).
