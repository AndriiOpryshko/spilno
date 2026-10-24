---
status: Accepted
owner: "Andrii Opryshko"
reviewers: []
updated_at: "2026-10-25"
feature_size: L
stage: "04-05"
ticket: ""
---

# 0006 — Run background work in the worker through a River job queue in Postgres

- **Status:** Accepted
- **Date:** 2026-10-25
- **Deciders:** Andrii Opryshko (Architect), with Claude during the architecture-design pass

## Context

ADR-0001 puts slow and scheduled work in a separate worker process. The API must hand it work: read an uploaded signed scan, read a newly attached document, answer a conversation turn with a proposal, copy a file in Drive, and run the nightly integrity check. If the domain write ("scan uploaded", "document attached") commits but the job is lost, a draft waits forever for a proposal; if the job runs but the write rolled back, it works on data that does not exist.

## Decision drivers

- PRD §6: proposal p95 ≤ 60 s; signed-copy upload p95 ≤ 20 s including the read — retries must be bounded and visible.
- PRD §6: periodic fingerprint comparison of every locked protocol.
- §2: Postgres is the only datastore; goose is the only migrator (repo ADR 0004); pgx/v5.

## Considered options

1. **River** — a Go job queue on a Postgres table with pgx/v5: transactional enqueue, retries with backoff, unique jobs, periodic jobs.
2. **Hand-rolled `jobs` table** — goose migration, sqlc queries, a poller using `SELECT … FOR UPDATE SKIP LOCKED`; retries, timeouts and scheduling written by us.

## Decision outcome

**Chosen:** Option 1. Enqueuing in the same transaction as the domain write removes the lost-job and phantom-job cases with no new infrastructure, and the retry, stuck-job and periodic-job handling already exists and is tested. River's own schema SQL is copied into a goose migration (and refreshed by a new migration when River is upgraded), so goose remains the single migrator and sqlc keeps reading one migrations directory.

Job kinds (contracts detailed by api-forge in `contracts/events.md`): `cases.read_document`, `governance.read_signed_scan`, `assistant.propose`, `gdrive.copy_file`, and periodic `governance.verify_locked_protocols` (nightly). No backup job in the first release (SAD §1 decision override). Each job carries ids only, never document content. The API enqueues through each context's `module.go`; job handlers live in the owning context and are registered by `cmd/worker`. Results are written back through the owning context, and the SPA polls them (§4 UI-architecture).

## Consequences

**Positive**
- No lost or orphaned jobs; failures retry with backoff and end in a visible "failed" state the UI can show.
- Nightly jobs are declared in code next to the work they do.
- No new service in `infra/compose.yaml`.

**Negative**
- A new dependency whose schema migrations we must mirror into goose on every River upgrade.
- Job tables live in the same database as the registry; a runaway retry loop adds write load there.

**Neutral**
- The hand-rolled table stays possible later behind the same enqueue functions in `module.go`.

## Links

- PRD: [[../PRD.md]] §6
- SAD: [[../sad.md]] §4, §5, §6, §7
- Related ADR: [[0001-build-backend-web-and-worker-surfaces]], [[0004-use-gemini-api-for-document-reading-and-drafting]]
