---
id: T09
title: "Model association settings and roster entries with all-fields validation"
layer: "domain"
deps: []
acs: ["AC-42", "AC-02"]
files_hint:
  - "services/core/internal/governance/domain/settings.go"
  - "services/core/internal/governance/domain/board_member.go"
  - "services/core/internal/governance/domain/settings_test.go"
  - "services/core/internal/governance/domain/board_member_test.go"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T09 — Model association settings and roster entries with all-fields validation

## Why

PRD AC-02 and AC-42; the register start is locked after the first number ([ADR-0017](../adr/0017-reserve-the-protocol-number-at-first-print-and-never-reuse-it.md)). [sad §6 Flow 12](../sad.md).

**Satisfies:** [PRD §5](../PRD.md) AC-42, AC-02 · **Blocked by:** —

## What

- Value types for settings and a roster entry with a term; `MissingSettings()` and `Validate()` returning all field errors.
- `CoversDate(date)` on a term — the basis of «current член правління».

## Definition of Done

- [ ] Acceptance: Unit tests prove a roster entry without a term start or with an end before the start is rejected naming every broken field, and settings report every missing value (meeting form, signers, register start) at once.
- [ ] unit tests for AC-02 (both cases) and AC-42 (every missing setting named)
- [ ] domain package imports stdlib + uuid only (depguard `domain-pure`)
- [ ] `mise run lint` and `mise run test` green

## Notes

Pure domain — can start on day one in parallel with T01 and T06.
