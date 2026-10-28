---
id: T28
title: "Mark ready for signing: reserve the number and render the print copy in a job"
layer: "ports"
deps: ["T05", "T14", "T23", "T26", "T27"]
acs: ["AC-04", "AC-21", "AC-14c", "AC-33", "AC-14b", "AC-41", "AC-42"]
files_hint:
  - "services/core/internal/governance/app/printing.go"
  - "services/core/internal/governance/jobs/render_print_copy.go"
  - "services/core/internal/governance/adapters/postgres/print_copies.go"
  - "services/core/internal/governance/adapters/postgres/queries/print_copies.sql"
  - "services/core/internal/governance/adapters/postgres/queries/protocol_registers.sql"
  - "services/core/internal/governance/adapters/http/ready_for_signing.go"
owner: "Andrii Opryshko"
estimate: "L"
status: "todo"
---

# T28 — Mark ready for signing: reserve the number and render the print copy in a job

## Why

[ADR-0017](../adr/0017-reserve-the-protocol-number-at-first-print-and-never-reuse-it.md) (number at first print, never reused), [ADR-0013](../adr/0013-render-signing-copies-from-google-docs-templates.md), [sad §6 Flow 4](../sad.md), [events.md](../contracts/events.md) `governance.render_print_copy.v1`.

**Satisfies:** [PRD §5](../PRD.md) AC-04, AC-21, AC-14c, AC-33, AC-14b, AC-41, AC-42 · **Blocked by:** [T05](./T05-stage-lock-integrity-migrations.md), [T14](./T14-river-worker-jobs.md), [T23](./T23-wording-slot-endpoints.md), [T26](./T26-poll-endpoints.md), [T27](./T27-gdocs-renderer-adapter.md)

## What

- Number from `protocol_registers` counter in the transition tx (`SELECT … FOR UPDATE`).
- Job renders via Renderer, stores `print_copies` row with sha256; print-duration metric (p95 ≤ 5 s).
- Void current copy on any return to чернетка (hook into T23/T26 transitions).

## Definition of Done

- [ ] Acceptance: Handler and job tests show markProtocolReadyForSigning refuses an incomplete agenda or unconfirmed slot naming all (AC-04, AC-21), otherwise freezes the text, takes the next register number once (concurrent calls give no duplicates), closes an open round (AC-33) and renders the copy; a return to draft voids the copy and keeps the number (AC-14b).
- [ ] job test with fake Renderer: copy stored, idempotent on retry
- [ ] `mise run lint` and `mise run test` green