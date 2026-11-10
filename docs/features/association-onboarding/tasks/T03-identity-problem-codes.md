---
id: T03
title: "Register the identity problem codes of PRD §5.1 and governance.drive_not_connected"
layer: "infra"
deps: []
acs: ["AC-02", "AC-03", "AC-03b", "AC-04", "AC-04b", "AC-12", "AC-13", "AC-15", "AC-16", "AC-16c", "AC-18", "AC-22", "AC-24", "AC-25", "AC-28", "AC-28b", "AC-30"]
files_hint:
  - "services/core/internal/identity/domain/errors.go"
  - "services/core/internal/identity/domain/errors_test.go"
  - "services/core/internal/governance/domain/errors.go"
  - "services/core/internal/governance/domain/errors_test.go"
owner: "Andrii Opryshko"
estimate: "S"
status: "todo"
---

# T03 — Register the identity problem codes of PRD §5.1 and governance.drive_not_connected

## Why

`.claude/rules/openapi.md`: one error `code` in the contract = one definition in the repo's registry, returning the same string. [PRD](../PRD.md) §5.1 plans 15 new `identity.*` codes and `governance.drive_not_connected`.

**Satisfies:** [PRD §5](../PRD.md) AC-02, AC-03, AC-03b, AC-04, AC-04b, AC-12, AC-13, AC-15, AC-16, AC-16c, AC-18, AC-22, AC-24, AC-25, AC-28, AC-28b, AC-30 · **Blocked by:** —

## What

- `identity/domain/errors.go`: constants and coded errors for the `identity.*` codes of [PRD](../PRD.md) §5.1 (`identity.session_required` and `identity.not_current_member` included), in the shape of `governance/domain/errors.go` (`Code…` constants, `CodedFieldsError` for the codes that name fields).
- The status of each code as the contract advertises it: 422 for `application_invalid` and `reason_required`, 403 for `not_operator`, `not_current_member`, `association_disabled`, 401 for `session_required`, 409 for the rest.
- `governance/domain/errors.go` gains `CodeDriveNotConnected`.

## Definition of Done

- [ ] Acceptance: Every code of PRD §5.1 has exactly one constant whose string equals the catalogue, with its HTTP status, and a test fails when a code is missing or misspelt.
- [ ] unit test: the set of constants equals the `Code` column of PRD §5.1, string for string
- [ ] unit test: the status table matches `x-problem-types` of the feature contract
- [ ] `mise run lint` and `mise run test` green

## Notes

Lands first so T08, T09 and T11 work in separate files of `identity/domain` without touching `errors.go`.
