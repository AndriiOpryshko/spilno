---
status: accepted
date: 2026-10-25
feature: board-protocol
sad_section: "§4"
deciders: ["Andrii Opryshko"]
---

# 0015 Run background jobs in cmd/worker with River on Postgres

## Context and problem statement

ADR 0001 moves slow and external work out of `cmd/api` into `cmd/worker`. In Stage 1 that work is the assistant draft (≤ 30 s p95, AC-03), rendering the print copy (≤ 5 s p95, AC-25, ADR 0018), applying Drive read-only restrictions after lock or withdrawal (ADR 0004) and the nightly integrity check (ADR 0019). ADR 0008 chose a hand-written `governance_jobs` table claimed with `FOR UPDATE SKIP LOCKED`; retries, timeouts, unique jobs, periodic scheduling and stuck-job recovery would all be our code. This ADR supersedes ADR 0008.

## Decision drivers

- PRD §6: draft from the помічник протоколу ≤ 30 s p95; print copy (PDF) rendering ≤ 5 s p95.
- QG-1: a committed state change and its follow-up job never diverge (no lost job, no job for a rolled-back change).
- SAD §2: one Postgres, pgx/v5, goose is the only migrator (repo ADR 0004); no broker in the stack.
- One developer: no queue code to write and test by hand.

## Blast radius

- Irreversible: yes — every asynchronous flow, its retry semantics and its job schema depend on the queue library; moving off it is ≥ 3 days.
- Multi-context: yes — `cmd/api` enqueues, `cmd/worker` runs; `governance` handlers, `platform/jobs`, the web polling job state.
- Legitimate alternative: yes — the hand-written job table of ADR 0008.
- Touches the database: yes — River's tables (`river_job`, `river_leader`, `river_queue`, …) in the core database.

## Considered options

1. **River** — Go job queue on Postgres with pgx/v5: transactional enqueue, retries with backoff, unique jobs, periodic jobs, job state queries.
2. **Hand-written `governance_jobs` table** (ADR 0008) — goose migration, sqlc queries, a poller with `FOR UPDATE SKIP LOCKED`; retries, timeouts and scheduling written by us.

## Decision outcome

**Chosen:** option 1. Transactional enqueue removes the lost-job and phantom-job cases with no new service, and retries, stuck-job rescue and the nightly schedule already exist and are tested upstream.

- `internal/platform/jobs` builds the River client from the shared pgx pool: insert-only in `cmd/api`, insert + work in `cmd/worker`. Contexts enqueue through their own `module.go`, inside the same transaction as the state change (`InsertTx`).
- Job kinds (Stage 1): `assistant_draft`, `render_pdf` (ADR 0018), `apply_drive_restrictions` (ADR 0004) and the periodic `verify_integrity` (nightly, ADR 0019). `read_signed_scan` comes with AC-28b in Stage 2; `sync_doc_access` is removed with ADR 0006.
- Payloads carry ids only (association, protocol, revision, file) plus the W3C trace context of the request, never document text or personal data.
- Handlers are idempotent: each checks the target state first (e.g. a print copy for this revision already exists → done), so a retry after a crash is safe. `assistant_draft` and `render_pdf` are unique per (protocol, revision) while queued or running.
- The web reads job state through a `governance` endpoint (pending / done / failed with retry offered) and polls it with TanStack Query.
- River's schema SQL is copied into a goose migration; a River upgrade that changes the schema adds a new goose migration.

### Consequences

- Good: no lost or orphaned jobs; failures end in a visible state the UI can show; the nightly job is declared in code.
- Good: no new container in compose; a second worker replica is safe at any time.
- Bad: a third-party dependency whose schema migrations we mirror into goose by hand on every upgrade (SAD §11 accepted debt).
- Neutral: job rows are not legal records; River's own retention deletes finished jobs.

### Migration and rollback

- Migration: one goose migration with River's schema for the pinned River version; no existing rows.
- Rollback: drain the queue, drop River's tables; protocol data is unaffected, because job results are written to `governance` tables.

## Links

- PRD: [PRD.ao-prd.md](../PRD.ao-prd.md) — US-02, US-06, US-07; AC-03, AC-25, AC-28; §6 latency and Integrity rows
- SAD: [sad.ao-architecture.md](../sad.ao-architecture.md) §4, §5, §6, §8
- Related ADR: supersedes [0008 Postgres job table](./0008-queue-ai-drive-and-pdf-work-in-a-postgres-job-table-processed-by-cmd-worker.md) · [0001 Surfaces](./0001-deliver-board-protocol-as-core-endpoints-spa-routes-and-a-worker-process.md) · repo [0004 Migrations and queries](../../../adr/0004-migrations-and-queries.md)
