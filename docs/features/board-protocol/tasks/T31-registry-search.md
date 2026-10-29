---
id: T31
title: "Search the реєстр протоколів and open a signed protocol"
layer: "ports"
deps: ["T29"]
acs: ["AC-18", "AC-41", "AC-15", "AC-19"]
files_hint:
  - "services/core/internal/governance/app/registry.go"
  - "services/core/internal/governance/adapters/postgres/registry.go"
  - "services/core/internal/governance/adapters/postgres/queries/registry.sql"
  - "services/core/internal/governance/adapters/http/registry.go"
owner: "Andrii Opryshko"
estimate: "L"
status: "todo"
---

# T31 — Search the реєстр протоколів and open a signed protocol

## Why

[sad §6 Flow 13](../sad.md); registry reads only `signed` snapshots ([ADR-0007](../adr/0007-lock-protocols-as-fingerprinted-immutable-snapshots.md)); `governance.search_invalid`; p95 ≤ 200 ms (PRD §6).

**Satisfies:** [PRD §5](../PRD.md) AC-18, AC-41, AC-15, AC-19 · **Blocked by:** [T29](./T29-lock-with-signed-copy.md)

## What

- Queries over `registry_entries` with the indexes in [data-model.md](../data-model.md) §Indexes.

## Definition of Done

- [ ] Acceptance: Integration tests over 30 seeded signed protocols show searchRegistry filters by counterparty, exact amount or range, subject and date period with cursor pages, and getSignedProtocol returns the signed copy link, meeting facts, cited documents, correction note and poll history labelled informal.
- [ ] integration tests per filter + cursor paging + `search_invalid`
- [ ] EXPLAIN on each query uses its index (asserted in a test or noted in the PR)
- [ ] `mise run lint` and `mise run test` green