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

> **Amended 2026-10-25 (Stage 1 rescope):** job kinds restated for Stage 1 — all owned by `governance` ([ADR-0014](./0014-keep-board-protocol-in-governance-and-add-only-an-identity-context.md)); scan reading and Drive copying are gone. Payloads carry ids plus the W3C trace context, and handlers are idempotent. The nightly database backup is a separate compose job, not a River job ([ADR-0016](./0016-back-up-postgres-nightly-off-the-vm-with-quarterly-restore-drills.md)). The choice of River is unchanged.

> **Amended 2026-10-25 (cross-check with the ao-architecture SAD):** a job kind `governance.apply_drive_restrictions` is added, enqueued in the lock and the withdrawal transaction to set the Drive read-only restriction on the protocol's files ([ADR-0005](./0005-store-files-in-association-owned-google-drive-folder.md)); the nightly job is renamed `governance.verify_protocol_snapshots` because it now covers signed and withdrawn snapshots ([ADR-0007](./0007-lock-protocols-as-fingerprinted-immutable-snapshots.md)) and restores trashed files. The choice of River is unchanged.

## Context

ADR-0001 puts slow and scheduled work in a separate worker process. The API must hand it work: answer a conversation turn with a proposal, render the print copy, clean up a file whose transaction failed, and run the nightly integrity check. If the domain write ("turn accepted", "marked готовий до підпису") commits but the job is lost, a draft waits forever for a proposal or a print copy; if the job runs but the write rolled back, it works on data that does not exist.

## Decision drivers

- PRD §6: proposal p95 ≤ 30 s from request to filled draft; print copy p95 ≤ 5 s — retries must be bounded and visible.
- PRD §6: periodic fingerprint comparison of every locked protocol.
- §2: Postgres is the only datastore; goose is the only migrator (repo ADR 0004); pgx/v5.

## Considered options

1. **River** — a Go job queue on a Postgres table with pgx/v5: transactional enqueue, retries with backoff, unique jobs, periodic jobs.
2. **Hand-rolled `jobs` table** — goose migration, sqlc queries, a poller using `SELECT … FOR UPDATE SKIP LOCKED`; retries, timeouts and scheduling written by us.

## Decision outcome

**Chosen:** Option 1. Enqueuing in the same transaction as the domain write removes the lost-job and phantom-job cases with no new infrastructure, and the retry, stuck-job and periodic-job handling already exists and is tested. River's own schema SQL is copied into a goose migration (and refreshed by a new migration when River is upgraded), so goose remains the single migrator and sqlc keeps reading one migrations directory.

Job kinds (contracts detailed by api-forge in `contracts/events.md`): `governance.propose_draft`, `governance.render_print_copy`, `governance.remove_orphan_file`, `governance.apply_drive_restrictions` (on lock and withdrawal), and periodic `governance.verify_protocol_snapshots` (nightly). Rules:
- A payload carries ids only (protocol id, revision, turn id, file id) plus the W3C `traceparent` of the request that enqueued it, never document content or personal data; the worker continues that trace.
- Handlers are idempotent: each is keyed by protocol id and revision (or file id), checks whether its result is already recorded, and writes the result in one transaction; River unique-job options stop duplicate enqueues of the same key.
- The API enqueues through `governance`'s use cases; handlers live in `governance/jobs` and are registered by `cmd/worker` through `governance.Worker()`. Results are written back through `governance`, and the SPA polls them (SAD §4 UI-architecture).

## Consequences

**Positive**
- No lost or orphaned jobs; failures retry with backoff and end in a visible "failed" state the UI can show.
- A request and its job share one trace in Grafana.
- Nightly jobs are declared in code next to the work they do.
- No new service in `infra/compose.yaml`.

**Negative**
- A new dependency whose schema migrations we must mirror into goose on every River upgrade.
- Job tables live in the same database as the registry; a runaway retry loop adds write load there.

**Neutral**
- The hand-rolled table stays possible later behind the same enqueue functions.

## Links

- PRD: [PRD.md](../PRD.md) §6, AC-26, AC-14c
- SAD: [sad.md](../sad.md) §4, §5, §6, §7, §8
- Related ADR: [ADR-0001](./0001-build-backend-web-and-worker-surfaces.md), [ADR-0004](./0004-use-gemini-api-for-document-reading-and-drafting.md), [ADR-0014](./0014-keep-board-protocol-in-governance-and-add-only-an-identity-context.md)
