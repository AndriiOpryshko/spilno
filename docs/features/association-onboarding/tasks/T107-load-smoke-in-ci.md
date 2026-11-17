---
id: T107
title: "Run the load smoke of PRD §6 in CI"
layer: "tests"
deps: ["T99"]
acs: ["AC-01", "AC-05", "AC-08", "AC-29"]
files_hint:
  - ".github/workflows/ci.yml"
  - "tests/load/"
  - "mise.toml"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T107 — Run the load smoke of PRD §6 in CI

## Why

[Review of 2026-11-17](../_review/review-2026-11-17.md), Q10. [PRD](../PRD.md) §6 measures the throughput requirement by a «load smoke test in CI»; the scenarios of T47 run only by hand.

**Satisfies:** [PRD §5](../PRD.md) AC-01, AC-05, AC-08, AC-29 · **Blocked by:** [T99](./T99-sign-in-state-in-sealed-cookie.md)

## What

- A `load` job after `core`, on the Postgres service of the e2e job, with a short duration fit for a shared runner.
- The thresholds are the ones of PRD §6; a note in the job says what a shared runner can and cannot prove.

## Definition of Done

- [ ] Acceptance: A CI job starts the API as the e2e job does, runs `mise run load` with the PRD §6 thresholds and fails the workflow when one is missed.
- [ ] the job is green on the PR and red with a threshold set to 1 ms (checked once, not committed)
- [ ] `mise run lint` and `mise run test` green

## Notes

If the shared runner is too noisy for the p95 threshold, the job keeps the throughput and error-rate thresholds and the record says so.
