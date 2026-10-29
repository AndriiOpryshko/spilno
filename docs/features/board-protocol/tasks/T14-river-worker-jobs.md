---
id: T14
title: "Wire the River client, the cmd/worker binary and the job status endpoints"
layer: "wiring"
deps: ["T04", "T07", "T11"]
acs: ["AC-26", "AC-14c"]
files_hint:
  - "services/core/internal/platform/jobs/"
  - "services/core/cmd/worker/main.go"
  - "mprocs.yaml"
  - "services/core/internal/governance/ports/jobs.go"
  - "services/core/internal/governance/module.go"
  - "services/core/internal/governance/adapters/http/jobs.go"
  - "services/core/internal/governance/jobs/"
owner: "Andrii Opryshko"
estimate: "L"
status: "todo"
---

# T14 — Wire the River client, the cmd/worker binary and the job status endpoints

## Why

[ADR-0006](../adr/0006-run-background-work-in-worker-via-river-queue.md) (enqueue in the same transaction, ids-only payloads, idempotent handlers); [events.md](../contracts/events.md) (envelope, uniqueness, retry). Slow results reach the SPA by polling getJob ([sad §4](../sad.md)).

**Satisfies:** [PRD §5](../PRD.md) AC-26, AC-14c · **Blocked by:** [T04](./T04-stage-assistant-poll-river-migrations.md), [T07](./T07-merge-contract-identity-skeleton.md), [T11](./T11-require-session-current-member.md)

## What

- `platform/jobs`: River client shared by api (insert-only) and worker (run).
- `cmd/worker/main.go` wires `governance.New(...).Worker()`; worker pane in `mprocs.yaml`.
- `Jobs` port with `EnqueueTx`; getJob / retryJob handlers (failed → retry, Flow 8).

## Definition of Done

- [ ] Acceptance: A job enqueued in the caller's transaction is run by `cmd/worker` in an integration test, a rolled-back transaction leaves no job, and getJob / retryJob return the contract shapes.
- [ ] integration test: commit → job runs; rollback → no job
- [ ] handler tests for getJob (pending/succeeded/failed) and retryJob
- [ ] `mise run dev` starts the worker pane
- [ ] `mise run lint` and `mise run test` green

## Notes

Owns `governance/module.go` after T10 (serialized by dep).
