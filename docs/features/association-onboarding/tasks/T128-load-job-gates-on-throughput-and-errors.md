---
id: T128
title: "Fail the CI load job on throughput and errors and report its latency"
layer: "tests"
deps: []
acs: ["AC-01", "AC-05", "AC-08", "AC-29"]
files_hint:
  - ".github/workflows/ci.yml"
  - "tests/load/main.go"
  - "tests/load/stats.go"
  - "mise.toml"
owner: "Andrii Opryshko"
estimate: "S"
status: "done"
---

# T128 — Fail the CI load job on throughput and errors and report its latency

## Why

[Second re-review of 2026-11-17](../_review/review-2026-11-17-2.md), S06. The job fails on the PRD p95 limits on a shared runner, and its own comment calls a single red p95 a reason to rerun: a check its author expects to be noise.

**Satisfies:** [PRD §5](../PRD.md) AC-01, AC-05, AC-08, AC-29 · **Blocked by:** —

## What

- In CI the p95 limits are reported, not enforced (a flag of the load client); throughput and the error rate still fail the job.
- `mise run load` keeps the PRD §6 limits for the run before a release.

## Definition of Done

- [ ] Acceptance: The CI load job fails when the read path does not keep 50 req/s or any request fails, prints the three p95 values next to their PRD limits without failing on them, and `mise run load` without the flag still exits 1 on a missed p95.
- [ ] a unit test of the load client's gate: report-only mode, enforcing mode
- [ ] the workflow comment says what the job proves
- [ ] the `load` job passes on the PR
- [ ] `mise run lint` and `mise run test` green

## Notes

PRD §6 names a «load smoke test in CI» for the throughput requirement; the latency numbers stay a release check, recorded in PRD §8 by [T132](./T132-record-open-gaps-in-prd.md).
