---
status: accepted
date: 2026-10-25
feature: board-protocol
sad_section: "§4"
deciders: ["Andrii Opryshko"]
---

# 0008 Queue AI, Drive and PDF work in a Postgres job table processed by cmd/worker

## Context and problem statement

ADR 0001 moves slow and external work into `cmd/worker`: the assistant draft (≤ 30 s p95), reading the signed scan (AC-28b), rendering the printable PDF (≤ 5 s p95), sharing and revoking Doc access (ADR 0006) and the daily integrity check (ADR 0009). `core` must hand work to the worker without losing it when either process restarts, and the web client must learn when it is done.

## Decision drivers

- PRD §6: draft ≤ 30 s p95, printable protocol ≤ 5 s p95, upload ≤ 3 s p95.
- SAD §2: no broker in the stack; one Postgres.
- QG-1: a lock and its follow-up jobs must not diverge.

## Blast radius

- Irreversible: yes — the job contract is used by every asynchronous flow.
- Multi-context: yes — `core` and `worker`, and the web polling the job status.
- Legitimate alternative: yes — synchronous draft in `core`, worker only for the rest.
- Touches the database: yes — new job table.

## Considered options

1. **Postgres job table, enqueued in the same transaction, claimed with `FOR UPDATE SKIP LOCKED`.**
2. **Synchronous draft in `core`, job table for the rest.**

## Decision outcome

**Chosen:** option 1. `governance_jobs` (id, association id, kind, payload jsonb, status queued / running / done / failed, attempts, run after, last error, result reference, timestamps). `core` inserts a job in the same transaction as the change that needs it, so a committed change always has its job. `cmd/worker` claims due jobs with `SELECT … FOR UPDATE SKIP LOCKED` (two workers never take the same job), runs the handler registered by `governance/module.go`, and retries with exponential backoff up to a per-kind limit; a failed job is visible on the draft. Kinds: `assistant_draft`, `read_signed_scan`, `render_pdf`, `sync_doc_access`, `apply_drive_restrictions`, `verify_integrity` (scheduled daily). The web shows "помічник пише…" and polls the job every 2 s with TanStack Query.

### Consequences

- Good: no new infrastructure; jobs survive restarts; one trace links the request and its job.
- Bad: polling adds up to 2 s to the visible draft time; the 30 s p95 is measured from request to filled draft (SAD §10).
- Neutral: handlers must be idempotent (safe to run twice), because a crash after the external call retries the job.

### Migration and rollback

- Migration: new table `governance_jobs` with an index on (status, run_after).
- Rollback: drop when the worker is removed; done jobs older than 90 days are deleted by the worker (jobs are not legal records).

## Links

- PRD: [PRD.ao-prd.md](../PRD.ao-prd.md) — US-02, US-07, AC-03, AC-28b; §6 latency and Integrity rows
- SAD: [sad.ao-architecture.md](../sad.ao-architecture.md) §4, §6
- Related ADR: [0001 Surfaces](./0001-deliver-board-protocol-as-core-endpoints-spa-routes-and-a-worker-process.md) · repo [0004 Migrations and queries](../../../adr/0004-migrations-and-queries.md)
