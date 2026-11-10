---
id: T46
title: "Prove disabling, the read-only association, the storage connection and every access denial end to end (US-06, US-07, US-08)"
layer: "tests"
deps: ["T21", "T31", "T33", "T34", "T40", "T42", "T44"]
acs: ["AC-04d", "AC-07", "AC-09", "AC-16", "AC-16b", "AC-21", "AC-22", "AC-23", "AC-24", "AC-25", "AC-26", "AC-27", "AC-28", "AC-28b"]
files_hint:
  - "apps/web/e2e/"
owner: "Andrii Opryshko"
estimate: "L"
status: "todo"
---

# T46 — Prove disabling, the read-only association, the storage connection and every access denial end to end (US-06, US-07, US-08)

## Why

[test-plan](../test-plan.md) rows of US-06, US-07, US-08 and of the access criteria; [PRD](../PRD.md) §9 rollout step 4.

**Satisfies:** [PRD §5](../PRD.md) AC-04d, AC-07, AC-09, AC-16, AC-16b, AC-21, AC-22, AC-23, AC-24, AC-25, AC-26, AC-27, AC-28, AC-28b · **Blocked by:** [T21](./T21-isolation-release-check.md), [T31](./T31-disabled-association-read-only.md), [T33](./T33-connect-storage-in-app.md), [T34](./T34-refuse-documents-without-storage.md), [T40](./T40-ui-operator-associations.md), [T42](./T42-ui-storage-and-suspended-notice.md), [T44](./T44-e2e-connection-flows.md)

## What

- Flow US-06: disable with a reason → a member sees «Доступ призупинено: …» → enable → the member changes a record again. Flow US-07: in a disabled association a change is refused and the registry stays readable. Flow US-08: connect the storage with the fake provider and read the account address.
- Refusals: AC-22, AC-24, AC-28, AC-28b. Access: two windows hold two associations (AC-04d); an applicant opens an association address (AC-07); no operator section and no access by its address (AC-16); the оператор sees the queue and their own association only (AC-16b).

## Definition of Done

- [ ] Acceptance: mise run e2e passes the flows of US-06, US-07 and US-08 and the access denials of AC-04d, AC-07, AC-16 and AC-16b, with the storage provider faked.
- [ ] three flow specs and one access spec green on the harness of T44
- [ ] two browser contexts of one person stay in their own association after reloads
- [ ] steps that need board-protocol screens are `test.fixme` with the task named: downloading a signed protocol (T29, T31, T43 there) and attaching a case document (T21, T36 there); the guard behind both is already proven in T31 and T34
- [ ] `mise run lint` and `mise run test` green

## Notes

The last task before the security review; if board-protocol's registry is merged by then, the two `fixme` steps are filled in here.
