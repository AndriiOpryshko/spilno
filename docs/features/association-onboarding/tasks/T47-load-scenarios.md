---
id: T47
title: "Add the load scenarios for the latency and throughput targets of PRD §6"
layer: "tests"
deps: ["T26", "T30", "T32"]
acs: ["AC-01", "AC-05", "AC-08", "AC-29"]
files_hint:
  - "tests/load/"
  - "mise.toml"
  - "docs/features/association-onboarding/_review/"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T47 — Add the load scenarios for the latency and throughput targets of PRD §6

## Why

[PRD](../PRD.md) §6 latency and throughput rows; [test-plan](../test-plan.md) «NFR validation (load)»: pre-release and on schedule, not on every PR.

**Satisfies:** [PRD §5](../PRD.md) AC-01, AC-05, AC-08, AC-29 · **Blocked by:** [T26](./T26-approve-connection-application.md), [T30](./T30-disable-enable-association.md), [T32](./T32-memberships-and-association-card.md)

## What

- No load tool is in the repo: add one suited to the stack behind `mise run load`, outside `mise run test`.
- Seed 20 associations and 200 applications with extracts; scenarios: read mix at 50 req/s for 5 min (p95 ≤ 200 ms, no errors); approvals, rejections, disablings and enablings at 5 req/s (p95 ≤ 500 ms; refusals of a decided заявка are not errors); 20 filings of the 5 MB fixture (p95 ≤ 5 s).

## Definition of Done

- [ ] Acceptance: Three load scenarios run against the local stack, fail when a PRD §6 threshold is missed, and their first results are recorded.
- [ ] `mise run load` exits non-zero when a threshold is missed
- [ ] the first run's numbers are in `docs/features/association-onboarding/_review/load-<date>.md`
- [ ] `mise run lint` and `mise run test` green

## Notes

Runs on the app-role pool with row-level security on: every scoped read is inside a transaction (ADR-0003 «Bad»), and this is where its cost shows.
