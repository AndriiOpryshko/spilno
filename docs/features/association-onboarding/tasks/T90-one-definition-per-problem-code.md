---
id: T90
title: "Advertise the disabled refusal of a file download and define every problem code once"
layer: "ports"
deps: []
acs: ["AC-25", "AC-26"]
files_hint:
  - "services/core/api/openapi.yaml"
  - "docs/features/association-onboarding/contracts/openapi.yaml"
  - "services/core/internal/governance/adapters/http/files.go"
  - "services/core/internal/governance/domain/errors.go"
  - "services/core/internal/governance/domain/case.go"
  - "services/core/internal/governance/app/registry.go"
  - "services/core/internal/governance/app/cases.go"
  - "services/core/internal/identity/module.go"
  - "services/core/internal/governance/file_download_test.go"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T90 — Advertise the disabled refusal of a file download and define every problem code once

## Why

[Review of 2026-11-17](../_review/review-2026-11-17.md), N05, Q08. [openapi rules](../../../../.claude/rules/openapi.md): one code in the contract is one definition in the registry. `downloadFile` answers a code its 403 does not list, typed as a string in governance; two governance codes exist as a registry literal and as a constant.

**Satisfies:** [PRD §5](../PRD.md) AC-25, AC-26 · **Blocked by:** —

## What

- `downloadFile` gets an inline 403 with both codes (delta and live contract, `mise run gen`).
- The identity module exports `CodeAssociationDisabled`; the governance handler uses it.
- `domain.CodeSearchInvalid` is added; the registry, `app/registry.go` and `app/cases.go` use the constants.

## Definition of Done

- [ ] Acceptance: The 403 of `downloadFile` lists `identity.not_current_member` and `identity.association_disabled` with an example of each, the handler takes the code from the constant the identity module exports, and `governance.file_invalid` and `governance.search_invalid` each have one definition that a test compares with the contract.
- [ ] `file_download_test.go` asserts the code against the constant and the response against the contract example
- [ ] no string literal of a problem code is left outside `domain/errors.go` in governance (grep in the task's test or lint note)
- [ ] `mise run lint` and `mise run test` green

## Notes

The description texts of `approveApplication` and the sync report follow in [T110](./T110-contract-texts-and-reconcile.md).
