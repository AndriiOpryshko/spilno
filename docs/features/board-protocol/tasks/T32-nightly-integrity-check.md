---
id: T32
title: "Run the nightly integrity check of snapshots and clean up orphan files"
layer: "app"
deps: ["T14", "T29", "T30"]
acs: ["AC-17"]
files_hint:
  - "services/core/internal/governance/jobs/verify_protocol_snapshots.go"
  - "services/core/internal/governance/jobs/remove_orphan_file.go"
  - "services/core/internal/governance/adapters/postgres/integrity_runs.go"
  - "services/core/internal/governance/adapters/postgres/queries/integrity_runs.sql"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T32 — Run the nightly integrity check of snapshots and clean up orphan files

## Why

PRD §6 Integrity of locked protocols; [ADR-0007](../adr/0007-lock-protocols-as-fingerprinted-immutable-snapshots.md), [ADR-0005](../adr/0005-store-files-in-association-owned-google-drive-folder.md); [sad §6 Flow 10](../sad.md); [events.md](../contracts/events.md) periodic job.

**Satisfies:** [PRD §5](../PRD.md) AC-17 · **Blocked by:** [T14](./T14-river-worker-jobs.md), [T29](./T29-lock-with-signed-copy.md), [T30](./T30-withdraw-and-copy.md)

## What

- River periodic job, unique per period; remove_orphan_file for uploads with no row.

## Definition of Done

- [ ] Acceptance: Job tests with the fake Files show the nightly run recomputes every signed and withdrawn fingerprint and pinned-file sha256, records an integrity run with findings, untrashes a trashed file and logs an alert on any mismatch.
- [ ] job tests: clean run, fingerprint mismatch, trashed file restored
- [ ] alert log line is structured (sloglint)
- [ ] `mise run lint` and `mise run test` green