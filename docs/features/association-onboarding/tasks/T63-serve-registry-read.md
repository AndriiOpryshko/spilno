---
id: T63
title: "Serve the реєстр протоколів: search and the signed protocol"
layer: "ports"
deps: ["T62"]
acs: ["AC-09", "AC-26"]
files_hint:
  - "services/core/internal/governance/app/registry.go"
  - "services/core/internal/governance/adapters/postgres/registry.go"
  - "services/core/internal/governance/adapters/postgres/queries/registry.sql"
  - "services/core/internal/governance/adapters/http/registry.go"
  - "services/core/internal/governance/adapters/http/stubs.go"
  - "services/core/internal/governance/governancetest/"
  - "services/core/internal/identity/middleware_readonly_test.go"
owner: "Andrii Opryshko"
estimate: "L"
status: "done"
---

# T63 — Serve the реєстр протоколів: search and the signed protocol

## Why

[Review of 2026-11-10](../_review/review-2026-11-10.md), F15. [PRD](../PRD.md) AC-09 (the new голова opens the registry) and AC-26 (a member of a вимкнене ОСББ reads it as before). `SearchRegistry` and `GetSignedProtocol` answer 501, and the e2e accepts that.

**Satisfies:** [PRD §5](../PRD.md) AC-09, AC-26 · **Blocked by:** [T62](./T62-serve-roster-endpoints.md)

## What

- The read side of [board-protocol T31](../../board-protocol/tasks/T31-registry-search.md) over `registry_entries` and `protocol_snapshots`, which are already migrated.
- The filters of the contract (counterparty, amount or range, subject, date period) and `governance.search_invalid`.
- A test fixture that seeds a signed protocol with its snapshot and file row, since nothing in the app can lock a protocol before board-protocol T29.

## Definition of Done

- [ ] Acceptance: Integration tests over seeded signed protocols show `searchRegistry` returning cursor pages newest first and an empty page for a new association, and `getSignedProtocol` returning the signed protocol with its file reference, for an enabled and for a вимкнене ОСББ alike.
- [ ] one test per filter, cursor paging, `search_invalid`
- [ ] isolation: the second association's entries are never returned
- [ ] both operations answer 200 while the association is disabled (`middleware_readonly_test.go` allow-list, now with a real handler)
- [ ] no Drive file or revision id in any response body
- [ ] `mise run lint` and `mise run test` green

## Notes

Carries board-protocol T31 whole; its `EXPLAIN` checks move here.
