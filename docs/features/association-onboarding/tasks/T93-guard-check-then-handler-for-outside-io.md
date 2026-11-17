---
id: T93
title: "Run operations with outside I/O after the guard's check transaction, not inside it"
layer: "wiring"
deps: []
acs: ["AC-25", "AC-28"]
files_hint:
  - "services/core/internal/identity/guard.go"
  - "services/core/internal/platform/db/db.go"
  - "services/core/internal/platform/db/db_integration_test.go"
  - "services/core/internal/identity/middleware_test.go"
  - "services/core/internal/identity/middleware_readonly_test.go"
  - "services/core/internal/identity/commit_before_response_test.go"
  - "services/core/internal/governance/adapters/http/cases.go"
  - "services/core/internal/governance/adapters/http/files.go"
  - "services/core/internal/identity/adapters/http/storage.go"
  - "docs/features/association-onboarding/adr/0003-enforce-association-isolation-with-postgres-row-level-security.md"
  - "docs/features/association-onboarding/sad.md"
owner: "Andrii Opryshko"
estimate: "L"
status: "todo"
---

# T93 — Run operations with outside I/O after the guard's check transaction, not inside it

## Why

[Review of 2026-11-17](../_review/review-2026-11-17.md), N06, Q01. The guard runs the whole handler inside `db.InAssociation`: a 21 MB body is buffered and a file is sent to the storage while a pooled connection and a transaction are held, against the contract («streamed to storage before the transaction») and board-protocol SAD Flow 8. `db.Detached` hides the transaction, it does not end it.

**Satisfies:** [PRD §5](../PRD.md) AC-25, AC-28 · **Blocked by:** —

## What

- The guard knows the operations with outside I/O by operation id. For them it checks the session, the roster and the read-only rule in a short transaction, commits, and calls the handler with the actor and a scope the handler opens its own transactions from (`db.InScope(ctx, fn)` or the like).
- Their response is written straight to the client: there is no later commit to wait for.
- ADR-0003 and SAD §8 get a dated amendment: which operations, why, and that each of their transactions is still association-scoped.

## Definition of Done

- [ ] Acceptance: For `addCaseDocument`, `downloadFile` and `completeStorageConnection` a test shows that no pool connection is held while the request body is read and while the handler runs its outside call, the roster and read-only refusals are unchanged, and every other operation still runs in one transaction.
- [ ] `middleware_test.go`: a slow body on `addCaseDocument` holds no connection (pool statistics or a pool of one)
- [ ] `middleware_readonly_test.go` still walks every operation, the three included
- [ ] a handler of the three that reads `db.From(ctx)` outside its own transaction fails loudly, not silently unscoped
- [ ] `mise run lint` and `mise run test` green

## Notes

The one task with a wide blast radius: it touches the path of every association-scoped request. The three handlers move to the new shape in [T94](./T94-upload-outside-transaction.md), [T95](./T95-download-outside-transaction.md) and [T96](./T96-code-exchange-outside-transaction.md); until then they open one transaction around their whole body, so this task alone changes no behaviour of theirs but the body read.
