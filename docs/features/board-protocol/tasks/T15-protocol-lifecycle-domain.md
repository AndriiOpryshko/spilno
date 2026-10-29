---
id: T15
title: "Model the protocol lifecycle: states, revision, transitions and refusals"
layer: "domain"
deps: []
acs: ["AC-05d", "AC-14b", "AC-17", "AC-32", "AC-35", "AC-36", "AC-37", "AC-38", "AC-39", "AC-40"]
files_hint:
  - "services/core/internal/governance/domain/protocol.go"
  - "services/core/internal/governance/domain/protocol_test.go"
owner: "Andrii Opryshko"
estimate: "L"
status: "todo"
---

# T15 — Model the protocol lifecycle: states, revision, transitions and refusals

## Why

[ADR-0009](../adr/0009-track-draft-lifecycle-with-states-and-revision-numbers.md): чернетка ⇄ готовий до голосування → готовий до підпису → підписаний; знятий з розгляду from any non-final state with a reason. [sad §6](../sad.md) Flows 3, 6, 9.

**Satisfies:** [PRD §5](../PRD.md) AC-05d, AC-14b, AC-17, AC-32, AC-35, AC-36, AC-37, AC-38, AC-39, AC-40 · **Blocked by:** —

## What

- Protocol aggregate: state, revision, reserved number, transitions with author and time.
- Any edit or «повернути в чернетку» bumps revision, returns to чернетка, closes the open round, voids the current print copy (AC-05d, AC-14b).
- Refusals: locked (AC-17, AC-38), withdrawn (AC-39 + «copy as new draft»), not ready for signing (AC-40), withdrawal needs a reason.

## Definition of Done

- [ ] Acceptance: Table-driven unit tests cover every allowed and refused transition of the Stage 1 state machine, revision bumps and the stale-revision conflict.
- [ ] table test over (state × action) with the expected next state or domain error code
- [ ] test: the number survives return to чернетка and withdrawal (AC-14b, AC-37)
- [ ] domain-pure depguard clean
- [ ] `mise run lint` and `mise run test` green

## Notes

Pure domain — day-one parallel branch.
