---
status: Accepted
owner: "Andrii Opryshko"
reviewers: []
updated_at: "2026-10-25"
feature_size: L
stage: "04-05"
ticket: ""
---

# 0016 — Back up Postgres nightly off the VM with quarterly restore drills

- **Status:** Accepted
- **Date:** 2026-10-25
- **Deciders:** Andrii Opryshko (Architect), with Claude during the Stage 1 rescope

> Replaces the earlier SAD decision override "no database backup in the first release" (SAD §1, §7, §10, §11 before 2026-10-25), which was never an ADR.

## Context

PRD §6 "Durability of the registry": RPO ≤ 24 h (nightly backup), RTO ≤ 4 h; restore verified quarterly, measured by the nightly backup log and a quarterly restore drill record. Everything runs on one VM with docker compose (SAD §7); losing its disk loses every record, lock snapshot, session and file reference. Files themselves live on the association's Google Drive ([ADR-0005](./0005-store-files-in-association-owned-google-drive-folder.md)). The dump contains confidential personal data (PRD §6.1).

## Decision drivers

- PRD §6: RPO ≤ 24 h, RTO ≤ 4 h, quarterly verified restore.
- PRD §6.1: confidential data must stay encrypted outside the VM.
- Independence: a backup must survive the loss of the VM and of the VM provider account.
- One developer: the simplest mechanism that can be drilled.

## Considered options

1. **Nightly `pg_dump` in a compose `backup` service, encrypted with `age`, pushed to S3-compatible storage outside the VM provider account.**
2. **Provider VM or volume snapshots** — no app work, but tied to the same provider account and not a database-consistent restore by default.
3. **Continuous WAL archiving (pgBackRest or WAL-G)** — RPO of minutes; more moving parts than RPO ≤ 24 h needs.
4. **Dump into the association's Google Drive** — reuses `drive.file`, but puts the database and the files in one account, so losing it loses both.

## Decision outcome

**Chosen:** Option 1. A `backup` service in the production compose file (same `postgres:18.6` image, so `pg_dump` matches the server) runs once a night: `pg_dump -Fc` → encrypt with `age` to a public key whose private key is kept offline by two board members → upload to an S3-compatible bucket in a different provider account → prune to 30 daily and 12 monthly dumps. It logs size and exit code and pushes a success timestamp metric; §7 alerts on a failed or missing dump. Restore: a fresh VM, the compose file, the latest dump (`pg_restore`), then the nightly integrity check run once to confirm every locked snapshot and pinned Drive file still agree. The drill runs quarterly, timed against 4 h, with a record kept per run.

## Consequences

**Positive**
- PRD §6 durability is met from the first release; no SAD decision override remains.
- The backup survives the loss of the VM, its provider account and the association's Google account.
- A restore also proves the lock fingerprints (QG-1).

**Negative**
- One more container, bucket, credential and encryption key to operate.
- Files are not in the dump; their protection is Drive itself plus the integrity check (SAD §11 accepted debt).
- The runbook depends on one developer until a second person has run a drill (SAD §11).

**Neutral**
- Moving to WAL archiving later changes only the `backup` service.

## Links

- PRD: [PRD.md](../PRD.md) §6, §6.1
- SAD: [sad.md](../sad.md) §1, §4, §6, §7, §10, §11
- Related ADR: [ADR-0005](./0005-store-files-in-association-owned-google-drive-folder.md), [ADR-0007](./0007-lock-protocols-as-fingerprinted-immutable-snapshots.md), [ADR-0001](./0001-build-backend-web-and-worker-surfaces.md)
