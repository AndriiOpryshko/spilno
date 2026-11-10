---
id: T34
title: "Refuse case documents and assistant turns while the storage is not connected"
layer: "ports"
deps: ["T02", "T03", "T19"]
acs: ["AC-28"]
files_hint:
  - "services/core/internal/governance/adapters/http/"
  - "services/core/internal/governance/module.go"
  - "services/core/internal/identity/module.go"
owner: "Andrii Opryshko"
estimate: "S"
status: "todo"
---

# T34 — Refuse case documents and assistant turns while the storage is not connected

## Why

[sad](../sad.md) §6 Flow 11; [PRD](../PRD.md) AC-28; [api-sync-report](../contracts/api-sync-report.md) §E «Storage not connected».

**Satisfies:** [PRD §5](../PRD.md) AC-28 · **Blocked by:** [T02](./T02-merge-contract-delta.md), [T03](./T03-identity-problem-codes.md), [T19](./T19-association-and-operator-guards.md)

## What

- `identity/module.go` exposes whether the actor's association has a storage connection; the governance handlers of `addCaseDocument` and `createAssistantTurn` ask it first.

## Definition of Done

- [ ] Acceptance: addCaseDocument and createAssistantTurn answer 409 governance.drive_not_connected for an association without a storage connection and reach their existing behaviour for one that has it.
- [ ] integration + contract: no connection → 409 with the code and «Спершу підключіть сховище документів ОСББ», saying any член правління can connect it; nothing is stored
- [ ] with a connection the call goes on to the handler's current behaviour (a 501 stub until board-protocol T21 and T25)
- [ ] `mise run lint` and `mise run test` green

## Notes

The guard is written before the stubs are filled; board-protocol T21 and T25 keep it as their first step.
