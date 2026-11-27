---
id: T58
title: "Pass the storage check into governance from cmd/api and count the pending age from the latest filing"
layer: "wiring"
deps: ["T57"]
acs: ["AC-05", "AC-28"]
files_hint:
  - "services/core/internal/governance/module.go"
  - "services/core/cmd/api/main.go"
  - "services/core/internal/identity/metrics.go"
  - "services/core/internal/identity/pending_age_gauge_test.go"
  - "services/core/internal/identity/adapters/postgres/queries/applications.sql"
owner: "Andrii Opryshko"
estimate: "S"
status: "done"
---

# T58 — Pass the storage check into governance from cmd/api and count the pending age from the latest filing

## Why

[Review of 2026-11-10](../_review/review-2026-11-10.md), B7, B8. Contexts meet only through `module.go` and are wired in `cmd/api` (repo ADR 0002); `governance.New` builds a second identity module instead. The pending-age gauge counts from the first filing, so a refile raises a false alert, and it drops its errors.

**Satisfies:** [PRD §5](../PRD.md) AC-05, AC-28 · **Blocked by:** [T57](./T57-refuse-rls-bypassing-role.md)

## What

- `cmd/api` passes `idm.HasStorageConnection` into `governance.New`.
- The gauge query uses the latest filing's `created_at`; the callback logs its error and `RegisterMetrics` returns its error from `run()`.

## Definition of Done

- [ ] Acceptance: `governance.New` takes the storage check as an argument and imports no identity constructor, and the gauge reports the age of the latest filing of the oldest pending application and logs a failed query.
- [ ] the depguard rule still passes with the identity import gone from `governance/module.go`
- [ ] gauge test: an application refiled today reports an age of zero days
- [ ] `mise run lint` and `mise run test` green
