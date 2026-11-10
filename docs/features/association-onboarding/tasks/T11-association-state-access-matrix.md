---
id: T11
title: "Model the association state and the access matrix of PRD §10"
layer: "domain"
deps: ["T03"]
acs: ["AC-07", "AC-16", "AC-16b", "AC-21", "AC-22", "AC-23", "AC-24", "AC-25", "AC-26"]
files_hint:
  - "services/core/internal/identity/domain/association.go"
  - "services/core/internal/identity/domain/association_test.go"
  - "services/core/internal/identity/domain/access.go"
  - "services/core/internal/identity/domain/access_test.go"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T11 — Model the association state and the access matrix of PRD §10

## Why

[PRD](../PRD.md) §10 «Roles and actions» and the ОСББ transitions; [ADR-0007](../adr/0007-enforce-a-disabled-association-as-read-only-in-the-identity-middleware.md) (read-only by an allow-list, fail-closed); [ADR-0009](../adr/0009-treat-the-operator-role-and-a-roster-membership-as-independent-grants-on-one-account.md) (the оператор role is asked on operator paths only).

**Satisfies:** [PRD §5](../PRD.md) AC-07, AC-16, AC-16b, AC-21, AC-22, AC-23, AC-24, AC-25, AC-26 · **Blocked by:** [T03](./T03-identity-problem-codes.md)

## What

- Association states `connected`, `disabled`; `Disable(reason)` refuses an empty reason (`identity.reason_required`) and a disabled association (`identity.association_state_unchanged`); `Enable()` refuses a connected one.
- `Actor` (account, email, association, member, read-only flag, оператор flag) and a pure decision per request: an association path asks the roster only — no member is `identity.not_current_member` whatever the state and whatever the оператор flag; a member of a disabled association passes only the allow-list `searchRegistry`, `getSignedProtocol`, `downloadFile`, anything else is `identity.association_disabled` with the reason; an operator path asks the оператор flag only (`identity.not_operator`).

## Definition of Done

- [ ] Acceptance: A table test equal to the PRD §10 role and action matrix passes, and the association state machine refuses a disabling without a reason and a change to the same state.
- [ ] table test: every cell of the PRD §10 «Roles and actions» table
- [ ] an operation id that is not on the allow-list is refused in a disabled association (fail-closed for a new operation)
- [ ] the ОСББ rows of «Allowed» and «Refused transitions» each have a case
- [ ] `mise run lint` and `mise run test` green

## Notes

The allow-list names governance operations inside identity — accepted in ADR-0007. T19 and T31 call this decision; they add no rule of their own.
