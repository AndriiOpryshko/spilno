---
status: Draft
owner: "Andrii Opryshko"
updated_at: "2026-10-27"
skill: ao-api-forge
---

# Async contract — board-protocol

Background work runs in `cmd/worker` on River (Postgres-backed queue, no broker; ADR 0015). The HTTP
operation commits the state change and inserts the job **in the same transaction** (`InsertTx`); the
job's outcome becomes visible through the resource's state, not through a callback:
`Protocol.assistant_status` for `assistant_draft`, `PrintCopy.status` for `render_pdf`. Derived from
sad.ao-architecture.md §6: every job here maps to an enqueue message in a flow.

## Rules for every job

- **Payload carries ids only:** association, protocol, revision, turn or file id. Never document
  text, the member's message, names, e-mails, amounts or other personal data; the handler re-reads
  them through its own ports (SAD §8 Background jobs, PRD §6.1).
- **Idempotency:** the handler checks the target state first and returns success if the work is
  already done (e.g. a print copy for this revision already exists); River `UniqueOpts` by the key
  below stops duplicate inserts while a job is queued or running.
- **Trace context:** the W3C `traceparent` of the enqueuing request goes into the job metadata, so
  the worker span joins the request trace (SAD §8 Observability).
- **Retry:** River retries with its backoff up to `MaxAttempts`; after the last attempt the job is
  discarded and an alert fires (SAD §7 "any job discarded").
- **Evolution:** args are additive only; removing or renaming a field is a new kind
  (`<kind>_v2`). The worker ignores unknown fields.

## Job: `assistant_draft`

- **Enqueued by:** `createProtocolDraft`, `addAgendaItem`, `askAssistant` in `governance`
  (sad.ao-architecture.md §6 Flow 1, Flow 11; AC-03, AC-06).
- **Worked by:** `cmd/worker`, `internal/governance/jobs/assistant_draft.go`.
- **Unique key:** `assistant_draft` + `protocol_id` + `revision` (River `UniqueOpts{ByArgs: true}`,
  ADR 0015 "unique per (protocol, revision)").
- **Max attempts / timeout:** not fixed by the SAD or an ADR; River defaults until
  `implement-tasks` sets them (Deferral D7). Target: ≤ 30 s p95 from request to filled draft (PRD §6).
- **On success:** a new protocol revision with the wording fragments and slot values of origin `ai`,
  unconfirmed, each with its source document and fragment; an assistant turn;
  `Protocol.assistant_status` becomes `idle`.
- **On discard:** `Protocol.assistant_status` becomes `failed`; the member is offered a retry
  (`askAssistant`) and may fill slots by hand (`addDecisionSlot`). No §5.1 code.

```json
{
  "association_id": "00000000-0000-7000-8000-000000000001",
  "protocol_id": "00000000-0000-7000-8000-000000000041",
  "revision": 1,
  "assistant_turn_id": "00000000-0000-7000-8000-000000000061"
}
```

| Field | Type | Origin |
|---|---|---|
| `association_id` | uuid | data-model.md -> `protocols.association_id` |
| `protocol_id` | uuid | data-model.md -> `protocols.id` |
| `revision` | int32 | data-model.md -> `protocols.revision`; the proposal is built on it (`assistant_turns.protocol_revision`) |
| `assistant_turn_id` | uuid | data-model.md -> `assistant_turns.id` (the member turn that asked) |

## Job: `render_pdf`

- **Enqueued by:** `markProtocolReadyForSigning` in `governance` (§6 Flow 5; AC-25, AC-25b, AC-45).
- **Worked by:** `cmd/worker`, `internal/governance/jobs/render_pdf.go` through the `Renderer` port
  (pinned Google Docs template, ADR 0018).
- **Unique key:** `render_pdf` + `protocol_id` + `revision` (ADR 0015); the handler also stops when
  `print_copies (protocol_id, protocol_revision)` already exists (data-model `uq_print_copies_protocol_revision`).
- **Max attempts / timeout:** not fixed by the SAD or an ADR (Deferral D7). Target: ≤ 5 s p95 from
  enqueue to issued (PRD §6, SAD §10 QG-4).
- **On success:** a `print_copies` row with the PDF `files` row (Drive file id, sha256) and the
  number taken at the transition; `PrintCopy.status` becomes `issued`, and `downloadFile` serves it.
  The copy carries the meeting form and signers from the settings, blank attendance and vote fields,
  and no poll result (AC-25, AC-46).
- **On discard:** the protocol stays `ready_for_signing` with `PrintCopy.status` `pending`; an alert
  fires. The member may return it to draft and render again (a new revision, a new job).

```json
{
  "association_id": "00000000-0000-7000-8000-000000000001",
  "protocol_id": "00000000-0000-7000-8000-000000000041",
  "revision": 7
}
```

| Field | Type | Origin |
|---|---|---|
| `association_id` | uuid | data-model.md -> `protocols.association_id` |
| `protocol_id` | uuid | data-model.md -> `protocols.id` |
| `revision` | int32 | data-model.md -> `print_copies.protocol_revision` |

## Job: `apply_drive_restrictions`

- **Enqueued by:** `signProtocol` (§6 Flow 6; AC-28) and `withdrawProtocol` (§6 Flow 7; AC-34,
  AC-34b, AC-34c) in `governance`.
- **Worked by:** `cmd/worker`, `internal/governance/jobs/apply_drive_restrictions.go` through the
  `Files` port (ADR 0004).
- **Unique key:** `apply_drive_restrictions` + `protocol_id` (a protocol reaches a final state once).
- **Max attempts / timeout:** not fixed by the SAD or an ADR; "retried until Drive accepts it"
  (ADR 0004, SAD §6 Flow 11) (Deferral D7).
- **On success:** Drive `contentRestrictions.readOnly` on the signed scan, the print copies and the
  справа documents of the protocol. Nothing changes on the wire.
- **On discard:** alert (SAD §7 Google API errors); the nightly `verify_integrity` still checks
  the files' sha256.

```json
{
  "association_id": "00000000-0000-7000-8000-000000000001",
  "protocol_id": "00000000-0000-7000-8000-000000000041"
}
```

| Field | Type | Origin |
|---|---|---|
| `association_id` | uuid | data-model.md -> `protocols.association_id` |
| `protocol_id` | uuid | data-model.md -> `protocol_snapshots.protocol_id` (the final protocol) |

## Periodic jobs

| Kind | Schedule | Reads | Writes | Alert |
|---|---|---|---|---|
| `verify_integrity` | nightly (ADR 0019, ADR 0004; exact time not fixed — Deferral D7) | `protocol_snapshots`, `files`, `case_documents` of open drafts, Drive metadata (`sha256Checksum`, `trashed`) | `integrity_runs`, `integrity_findings`; for a changed справа document of an open draft a new `case_documents` version, slots → `needs_review`, a new protocol revision (§6 Flow 9, AC-43) | any mismatch, trashed or missing locked file; no successful run in 26 h (SAD §7) |

Payload of `verify_integrity`: `{ "association_id": "00000000-0000-7000-8000-000000000001" }` (one
association in Stage 1, ADR 0016).

## Names against the data model

data-model.md (derived from the other SAD) names the jobs `propose_draft` (keyed by turn id) and
`render_print_copy`. This contract follows sad.ao-architecture.md / ADR 0015 (`assistant_draft`,
`render_pdf`), the SAD picked for this run. River state is not a table, so no column changes; the
names must agree in `implement-tasks` (report Deferral D6).
