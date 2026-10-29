---
id: T06
title: "Add the problem `code` member to httpapi with a governance error registry"
layer: "infra"
deps: []
acs: ["AC-02", "AC-42"]
files_hint:
  - "services/core/internal/platform/httpapi/"
  - "services/core/internal/governance/domain/errors.go"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T06 — Add the problem `code` member to httpapi with a governance error registry

## Why

The contract advertises 20 `<context>.<error_name>` codes but the repo has only Go sentinels with `about:blank` (contracts/api-sync-report.md check 2, deferred). .claude/rules/openapi.md requires one code = one definition.

**Satisfies:** [PRD §5](../PRD.md) AC-02, AC-42 · **Blocked by:** —

## What

- Extend the problem writer with `code` and a typed domain error carrying code, status and `errors[]` (pointer, detail).
- Add the governance codes to a registry in `domain/errors.go`; identity codes arrive with T08.

## Definition of Done

- [ ] Acceptance: httpapi writes the `code` extension member (`<context>.<error_name>`) with `type: about:blank` for registered domain errors, exactly as contracts/openapi.yaml `Problem`; validator 422s keep no `code`; a unit test checks every contract code has one registry entry.
- [ ] unit test: a registered error renders `type`, `status`, `code`, `detail`, `errors[]`
- [ ] unit test: every `governance.*` code in the contract delta exists in the registry
- [ ] existing meetings / contracts tests stay green
- [ ] `mise run lint` and `mise run test` green