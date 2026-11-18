---
id: T96
title: "Exchange the storage consent code with Google outside the transaction"
layer: "app"
deps: ["T93"]
acs: ["AC-27", "AC-28b"]
files_hint:
  - "services/core/internal/identity/storage.go"
  - "services/core/internal/identity/app/storage.go"
  - "services/core/internal/identity/adapters/http/storage.go"
  - "services/core/internal/identity/storage_test.go"
owner: "Andrii Opryshko"
estimate: "M"
status: "done"
---

# T96 — Exchange the storage consent code with Google outside the transaction

## Why

[Review of 2026-11-17](../_review/review-2026-11-17.md), B2. The follow-up passed `db.Detached(ctx)` to the exchange, which only hides the transaction; it stays open for the whole call to Google.

**Satisfies:** [PRD §5](../PRD.md) AC-27, AC-28b · **Blocked by:** [T93](./T93-guard-check-then-handler-for-outside-io.md)

## What

- `Complete`: state and membership checks, then the exchange, then one scoped transaction for the write.
- `db.Detached` is removed if nothing else uses it.

## Definition of Done

- [ ] Acceptance: A test shows `completeStorageConnection` exchanging the code with no transaction open and writing the connection in one transaction afterwards, with the refusals of a bad state, a refused code and a member's sign-in address unchanged.
- [ ] `storage_test.go`: the exchange runs on a pool of one connection while another request is served
- [ ] `mise run lint` and `mise run test` green
