---
id: T54
title: "Commit the association transaction before the response leaves, and keep outside calls out of it"
layer: "wiring"
deps: ["T50"]
acs: ["AC-25", "AC-27"]
files_hint:
  - "services/core/internal/identity/guard.go"
  - "services/core/internal/platform/db/db.go"
  - "services/core/internal/platform/db/db_integration_test.go"
  - "services/core/internal/identity/app/storage.go"
  - "services/core/internal/identity/middleware_test.go"
  - "services/core/internal/identity/storage_test.go"
owner: "Andrii Opryshko"
estimate: "L"
status: "done"
---

# T54 — Commit the association transaction before the response leaves, and keep outside calls out of it

## Why

[Review of 2026-11-10](../_review/review-2026-11-10.md), B2. [ADR-0003](../adr/0003-enforce-association-isolation-with-postgres-row-level-security.md): every association-scoped request runs in one `InAssociation` transaction opened by the guard. The guard commits after the handler has written its response, so a failed commit follows a 2xx, and the Google code exchange runs inside the transaction.

**Satisfies:** [PRD §5](../PRD.md) AC-25, AC-27 · **Blocked by:** [T50](./T50-kyiv-calendar-day.md)

## What

- Buffer the handler's response in the guard and send it only after a successful commit; a failed commit answers one 500 problem.
- A handler that answered 4xx or 5xx rolls back instead of committing partial writes.
- Move the Google code exchange of the storage consent out of the transaction; governance's later Drive work follows the same rule (note it on board-protocol T21).

## Definition of Done

- [ ] Acceptance: A commit that fails after the handler ran answers 500 and no success body reaches the client, and the storage consent exchanges the code outside the association transaction.
- [ ] test with a commit that fails: the client gets 500 only, and the write is absent
- [ ] test: a handler that writes and then answers 409 leaves nothing
- [ ] streamed downloads still stream: buffering applies to non-download responses, or the download reads before the commit and streams after
- [ ] `mise run lint` and `mise run test` green

## Notes

The riskiest task of the follow-up: it changes the path of every association-scoped request. The review did not verify B2; start by reproducing it.
