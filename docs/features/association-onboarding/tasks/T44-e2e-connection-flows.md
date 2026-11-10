---
id: T44
title: "Prove filing, following and deciding a заявка end to end through the UI (US-01, US-02, US-03)"
layer: "tests"
deps: ["T23", "T25", "T26", "T32", "T37", "T39"]
acs: ["AC-01", "AC-02", "AC-03", "AC-03b", "AC-04", "AC-04b", "AC-04c", "AC-05", "AC-06", "AC-08", "AC-09", "AC-10", "AC-11", "AC-11b", "AC-12", "AC-13", "AC-14", "AC-15", "AC-16c", "AC-30"]
files_hint:
  - "apps/web/e2e/"
  - "apps/web/playwright.config.ts"
  - "mprocs.yaml"
  - "mise.toml"
owner: "Andrii Opryshko"
estimate: "L"
status: "todo"
---

# T44 — Prove filing, following and deciding a заявка end to end through the UI (US-01, US-02, US-03)

## Why

[test-plan](../test-plan.md) «E2E-through-UI»: one flow per user story and every refusal, driven through the rendered app; sign-in and storage providers faked.

**Satisfies:** [PRD §5](../PRD.md) AC-01, AC-02, AC-03, AC-03b, AC-04, AC-04b, AC-04c, AC-05, AC-06, AC-08, AC-09, AC-10, AC-11, AC-11b, AC-12, AC-13, AC-14, AC-15, AC-16c, AC-30 · **Blocked by:** [T23](./T23-follow-and-refile-application.md), [T25](./T25-reject-application.md), [T26](./T26-approve-connection-application.md), [T32](./T32-memberships-and-association-card.md), [T37](./T37-ui-my-applications.md), [T39](./T39-ui-operator-decisions.md)

## What

- Harness: the core service in its e2e configuration (the sign-in shim of T18), a `signInAs(address)` helper, a database per suite, each flow under its own ЄДРПОУ.
- Flow US-01: file a заявка. Flow US-02: read the status → rejected → read the reason → refile. Flow US-03: record the check → approve → the applicant opens the new association.
- Refusals marked e2e-through-UI in the test plan for AC-02, 03, 03b, 04, 04b, 04c, 11, 11b, 12, 13, 15, 16c, 30.

## Definition of Done

- [ ] Acceptance: mise run e2e passes the flows of US-01, US-02 and US-03 and their refusals against the real web app, core service and database, with the sign-in provider faked.
- [ ] three flow specs and one refusals spec are green in `mise run e2e`, locally and in CI
- [ ] the specs run in parallel without sharing a code or an account
- [ ] AC-09's step «adds a член правління» is a `test.fixme` naming board-protocol T10 and its roster screen until they are merged; the rest of the flow asserts the association opens
- [ ] `mise run lint` and `mise run test` green

## Notes

`mise run e2e` needs the core API running (CLAUDE.md). No real name, address or code in a fixture.
