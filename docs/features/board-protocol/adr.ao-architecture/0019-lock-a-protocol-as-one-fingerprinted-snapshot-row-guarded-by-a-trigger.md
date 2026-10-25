---
status: accepted
date: 2026-10-25
feature: board-protocol
sad_section: "§4"
deciders: ["Andrii Opryshko"]
---

# 0019 Lock a protocol as one fingerprinted snapshot row guarded by a trigger, and never delete a protocol

## Context and problem statement

Uploading the signed scan with the meeting facts locks the protocol (AC-28, AC-44, AC-47); a підписаний protocol changes only through a new referencing protocol (AC-30, AC-30b); a знятий з розгляду draft is final (AC-35); signed and withdrawn protocols are never hard-deleted (PRD §6 Retention), with 0 changes to text, meeting facts or scan after lock (PRD §6 Integrity). ADR 0009 put a trigger on about nine child tables and recomputed a hash over them every night, which ties the guard to every future table. This ADR supersedes ADR 0009.

## Decision drivers

- QG-1: integrity of signed protocols; PRD §6 Integrity "integrity check job".
- PRD §6 Retention: ≥ 5 years, never hard-deleted.
- AC-47: lock refused without members present and a vote count for every item (`governance.meeting_facts_required`).
- ADR 0004: files live on Drive, so their sha256 must be part of what is locked.

## Blast radius

- Irreversible: yes — the snapshot format is a long-lived legal record; loosening the guard later is itself a legal question.
- Multi-context: no — `governance` only (core and worker run the same context).
- Legitimate alternative: yes — triggers on every child table (ADR 0009), or a check in `app/` only.
- Touches the database: yes — snapshot table, triggers, retention.

## Considered options

1. **One snapshot row per locked or withdrawn protocol** (jsonb + fingerprint), a trigger that refuses UPDATE and DELETE on that table, and a trigger that refuses DELETE on `protocols` in any state.
2. **Triggers on every child table + hash recomputed from live rows** (ADR 0009).
3. **`app/` check + hash only**, no database guard.

## Decision outcome

**Chosen:** option 1. The legal record is one row, so the guard does not grow with the schema.

- Lock (готовий до підпису → підписаний) runs in one transaction: `app/` refuses without the meeting facts (`governance.meeting_facts_required`, AC-47) or without a corrections note when corrections are marked (`governance.corrections_note_required`, AC-44); writes `protocol_meeting_facts` (members present, vote count per item, corrections flag and note, entered by, at); builds the canonical snapshot — frozen revision text and agenda, slots with sources and confirmations, members, meeting facts, poll rounds as history, print copy and signed scan with their Drive file ids and sha256, settings applied (meeting form, signers) — with a `schema_version`; stores it in `protocol_snapshots` with `fingerprint = sha256(canonical JSON)`; sets the state; enqueues `apply_drive_restrictions` (ADR 0004, 0015).
- Withdrawal (any non-final state → знятий з розгляду) writes the same kind of row with kind `withdrawn` and the reason (AC-34, AC-34b, AC-34c).
- Trigger `governance_snapshot_immutable()` raises on any UPDATE or DELETE of `protocol_snapshots`. Trigger `governance_protocol_never_deleted()` raises on DELETE of `protocols` in every state; drafts are withdrawn, not deleted. `app/` refuses every change to a final protocol with `governance.protocol_locked` / `governance.transition_not_allowed` (AC-30, AC-30b, AC-35).
- The worker's nightly `verify_integrity` recomputes each fingerprint from the stored jsonb and compares every file's sha256 with Drive metadata (`files.get` fields `sha256Checksum, trashed`), without downloading files; it restores a trashed file with `untrash` and writes the result to `protocol_integrity_checks` (append-only); any mismatch raises the alert (SAD §6 Flow 9, §7).
- Corrections and revocations are new protocols with `corrects_protocol_id` (AC-30, AC-30b).

### Consequences

- Good: the registry reads a signed protocol from one row; no code path, migration or manual SQL changes it without dropping a trigger, which is visible in review.
- Good: the integrity check is cheap (one hash per protocol and one metadata call per file).
- Bad: live child rows of a final protocol are guarded by `app/` only; a bug there does not change the legal record, but the screen could drift from the snapshot — the registry renders final protocols from the snapshot.
- Neutral: the snapshot format is versioned by `schema_version`; a new version never rewrites old rows.

### Migration and rollback

- Migration: tables `protocol_meeting_facts`, `protocol_snapshots` (protocol id PK, association id, kind locked / withdrawn, schema version, snapshot jsonb, fingerprint, created by, created at), `protocol_integrity_checks`; `protocols.corrects_protocol_id`; both trigger functions inside `-- +goose StatementBegin/End`.
- Rollback: the down migration drops the triggers only; snapshot rows are kept and never deleted.

## Links

- PRD: [PRD.ao-prd.md](../PRD.ao-prd.md) — US-07, US-08; AC-28, AC-30, AC-30b, AC-34c, AC-35, AC-44, AC-47; §6 Integrity, Retention
- SAD: [sad.ao-architecture.md](../sad.ao-architecture.md) §4, §6, §8, §10
- Related ADR: supersedes [0009 Trigger and snapshot hash](./0009-guard-signed-protocols-with-a-database-trigger-and-a-snapshot-hash.md) · [0004 Files on Google Drive](./0004-store-case-files-and-signed-scans-on-the-associations-google-drive-with-hashes-in-postgres.md) · repo [0004 Migrations and queries](../../../adr/0004-migrations-and-queries.md)
