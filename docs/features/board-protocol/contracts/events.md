---
status: Draft
owner: "Andrii Opryshko"
reviewers: []
updated_at: "2026-10-27"
feature_size: L
---

# Events — board-protocol

Async contract for the flows drawn in `sad.md` §6. In Stage 1 the async surface is a **job queue, not
a pub/sub bus**: the core API enqueues River jobs in the same transaction as the domain write that needs
them, and the worker (`cmd/worker`, `governance/jobs`) runs them ([ADR-0006](../adr/0006-run-background-work-in-worker-via-river-queue.md)).
Each job kind below is derived from an enqueue/take message in a §6 diagram. Nothing is published to other
contexts — `governance` is both producer and consumer.

## Channel: River queue in Postgres (`river_job`)

- **Producer:** `governance` use cases in `cmd/api` (enqueue inside the domain transaction); River's periodic
  scheduler in `cmd/worker` for the nightly job.
- **Consumers:** `governance/jobs` handlers registered by `cmd/worker` through `governance.Worker()`.
- **Delivery:** at-least-once (a handler may run again after a crash or a retry).
- **Ordering:** none across jobs; uniqueness per key (River unique-job options) stops duplicate enqueues.
- **Envelope:** River's job row (`id`, `kind`, `args` JSON, `attempt`, `max_attempts`, `state`, `errors`).
  `args` carry ids plus the W3C `traceparent` of the request that enqueued the job — never document content,
  conversation text or personal data (ADR-0006, SAD §8 Background jobs).
- **Visible to the SPA:** `governance.propose_draft` and `governance.render_print_copy` through
  `GET /governance/jobs/{job_id}` (state folded to `pending | running | completed | failed`) and
  `POST /governance/jobs/{job_id}/retry` ([openapi.yaml](./openapi.yaml) `Job`).

## Event: `governance.propose_draft.v1`

```json
{
  "kind": "governance.propose_draft",
  "args": {
    "version": 1,
    "protocol_id": "<uuid — protocols.id>",
    "assistant_turn_id": "<uuid — assistant_turns.id, the member turn to answer>",
    "protocol_revision": "<int — assistant_turns.protocol_revision the proposal is built on>",
    "traceparent": "<W3C traceparent>"
  }
}
```

- **Required fields:** `version`, `protocol_id`, `assistant_turn_id`, `protocol_revision`, `traceparent`.
- **Origin:** sad.md §6 Flow 1 → `enqueue propose_draft` / `W->>DB: Take the propose_draft job`; Flow 8.
- **Unique key:** `assistant_turn_id` — one proposal per member turn.
- **Effect:** downloads the справа's document versions, calls the text-generation port, validates the structured
  answer, stores `slot_values` (origin `ai`, slots `unconfirmed`), new `agenda_item_versions` and an assistant
  turn; never overwrites a confirmed slot — a new value goes beside it and the slot becomes `needs_review`
  (AC-28b); a stale `protocol_revision` re-bases onto the current draft (ADR-0009).
- **Idempotency:** the handler checks for an assistant turn answering `assistant_turn_id` and skips if present.
- **Metric:** `governance.proposal.duration` (p95 ≤ 30 s, PRD §6).

## Event: `governance.render_print_copy.v1`

```json
{
  "kind": "governance.render_print_copy",
  "args": {
    "version": 1,
    "protocol_id": "<uuid — protocols.id>",
    "protocol_revision": "<int — print_copies.protocol_revision>",
    "traceparent": "<W3C traceparent>"
  }
}
```

- **Required fields:** `version`, `protocol_id`, `protocol_revision`, `traceparent`.
- **Origin:** sad.md §6 Flow 4 → `enqueue render_print_copy` / `W->>DB: Take render_print_copy by protocol id and revision`; Flow 8.
- **Unique key:** `(protocol_id, protocol_revision)`; backed by `uq_print_copies_protocol_revision`.
- **Effect:** copies the pinned Docs layout template, fills the frozen text, the settings (meeting form, signers)
  and the reserved number, leaves blank fields for members present and votes per item, exports the PDF to
  Drive and inserts `files` + `print_copies` (ADR-0013). No poll result on the copy (AC-41).
- **Idempotency:** "skip if already recorded" — an existing `print_copies` row for the key ends the job.
- **Stale run:** if the protocol is no longer `ready_for_signing` at `protocol_revision`, the job completes
  without rendering.
- **Metric:** `governance.print_copy.duration` (p95 ≤ 5 s, PRD §6).

## Event: `governance.apply_drive_restrictions.v1`

```json
{
  "kind": "governance.apply_drive_restrictions",
  "args": {
    "version": 1,
    "protocol_id": "<uuid — protocols.id with a protocol_snapshots row>",
    "traceparent": "<W3C traceparent>"
  }
}
```

- **Required fields:** `version`, `protocol_id`, `traceparent`.
- **Origin:** sad.md §6 Flow 5 (lock) and Flow 9 (withdrawal) → `enqueue apply_drive_restrictions`.
- **Unique key:** `protocol_id`.
- **Effect:** sets Drive `contentRestrictions.readOnly` on every file the snapshot pins — print copies, the signed
  scan, the cited case document versions (ADR-0005, ADR-0007).
- **Idempotency:** setting the restriction is idempotent per file; reruns are safe.

## Event: `governance.remove_orphan_file.v1`

```json
{
  "kind": "governance.remove_orphan_file",
  "args": {
    "version": 1,
    "drive_file_id": "<string — the uploaded Drive file whose transaction rolled back>",
    "traceparent": "<W3C traceparent>"
  }
}
```

- **Required fields:** `version`, `drive_file_id`, `traceparent`.
- **Origin:** sad.md §6 Flow 8 rule — "A file uploaded whose transaction then fails is removed by a
  `remove_orphan_file` job". Carries the Drive id because no `files` row was committed.
- **Unique key:** `drive_file_id`.
- **Effect:** deletes the Drive file if no `files` row references it (`uq_files_drive_file_id`).
- **Enqueue note:** the enqueue cannot share the failed transaction; it is inserted in its own transaction after
  the rollback. <!-- a crash between the rollback and this insert leaves an orphan in Drive; accepted, no sweep in Stage 1 -->

## Event: `governance.verify_protocol_snapshots.v1` (periodic)

```json
{
  "kind": "governance.verify_protocol_snapshots",
  "args": { "version": 1 }
}
```

- **Required fields:** `version`.
- **Origin:** sad.md §6 Flow 10 (nightly integrity check), Flow 3 note (nightly sha256 check of documents of open drafts).
- **Schedule:** River periodic job, nightly; one run at a time (unique by period).
- **Effect:** recomputes every `signed` and `withdrawn` snapshot's fingerprint, compares each pinned file's
  sha256 / revision / trashed flag with Drive, `untrash`es a trashed pinned file, and writes `integrity_runs` +
  `integrity_findings`; for documents of open drafts a changed sha256 marks citing slots `needs_review`.
- **Metrics / alerts:** `governance.integrity.mismatches`, `.restored`, `.verified`; no successful run in 26 h (SAD §7).

## Idempotency & retry

- **Idempotency:** every handler is keyed as listed above and checks whether its result is already recorded
  before writing, in one transaction (ADR-0006). Redelivery of the same River job id is therefore safe.
- **Retry:** River's exponential backoff. Per-kind `max_attempts` is not fixed by any upstream artifact —
  set it in `implement-tasks` against the NFRs (proposal p95 ≤ 30 s, print copy p95 ≤ 5 s) and record it here.
- **Dead-letter:** no separate DLQ — a job that exhausts its attempts is River-`discarded` and shown to the SPA as
  `failed` with a retry offered (Flow 8); slots can still be filled by hand. Alert: > 3 failed jobs of one kind
  in 1 h, oldest available job > 10 min (SAD §7).

## Schema registry

- Registry: this file. The Go args structs live in `services/core/internal/governance/jobs/` (one `JobArgs`
  type per kind, `Kind()` returning the name above); a breaking change to `args` is a new `version` value.
- Validator: none beyond the Go types — args are produced and consumed by the same module.
