---
status: accepted
date: 2026-10-25
feature: board-protocol
sad_section: "§4"
deciders: ["Andrii Opryshko"]
---

# 0009 Guard signed and withdrawn protocols with a database trigger and a snapshot hash

## Context and problem statement

A підписаний protocol changes only through a new protocol that references it (AC-30, AC-30b), a знятий з розгляду draft is final (AC-35), and signed and withdrawn protocols are never hard-deleted (PRD §6 Retention). PRD §6 also asks for an integrity check job with 0 mismatches between printed numbers and confirmed slots. A bug, a hand-written SQL fix or a later migration must not be able to change such a record silently.

## Decision drivers

- QG-1: integrity of signed protocols.
- PRD §6: Retention ≥ 5 years, never hard-deleted; Integrity check job.
- ADR 0004: files live on Drive, so their hashes must be part of what is verified.

## Blast radius

- Irreversible: yes — once records are locked, loosening the guard is itself a legal question.
- Multi-context: no — `governance` tables only.
- Legitimate alternative: yes — enforce in `app/` only, plus the hash.
- Touches the database: yes — trigger functions on every protocol table; hash column.

## Considered options

1. **Check in `app/` + Postgres trigger + snapshot hash verified daily.**
2. **Check in `app/` + snapshot hash only.**

## Decision outcome

**Chosen:** option 1. `app/` refuses changes with `governance.protocol_locked` and `governance.transition_not_allowed`. A trigger function `governance_reject_locked_change()` runs BEFORE UPDATE OR DELETE on `protocols` and on every child table (`protocol_versions`, slots, slot confirmations, votes, meeting facts, overrides, check snapshots, `governance_files` rows of the protocol) and raises when the protocol is `signed` or `withdrawn`; DELETE of a protocol is refused in every state (drafts are withdrawn, not deleted). On lock, `app/` computes sha256 over a canonical JSON snapshot (frozen text, slots with sources and confirmations, votes, meeting facts, overrides, check snapshot, applied rule ids, file sha256s) and stores it in `protocols.locked_hash` in the same transaction that sets `signed`. The worker job `verify_integrity` recomputes it daily, checks the Drive files (ADR 0004) and writes the result to `protocol_integrity_checks` (append-only), alerting on a mismatch. A correction or revocation is a new protocol with `corrects_protocol_id`.

### Consequences

- Good: no code path, migration or manual SQL can change a locked record without dropping the trigger, which is visible in review.
- Good: the integrity KPI is a query over `protocol_integrity_checks`.
- Bad: logic in the database; it is tested through `testdb.New(t)` and written inside goose `StatementBegin/End`.
- Neutral: the only update allowed on a signed protocol row is none; verification results live in their own table.

### Migration and rollback

- Migration: one goose migration with the trigger function and triggers, `protocols.locked_hash`, `protocols.corrects_protocol_id`, table `protocol_integrity_checks`.
- Rollback: the down migration drops the triggers only; locked rows and hashes are kept and never deleted.

## Links

- PRD: [PRD.ao-prd.md](../PRD.ao-prd.md) — US-07, US-08, AC-30, AC-30b, AC-35; §6 Integrity, Retention
- SAD: [sad.ao-architecture.md](../sad.ao-architecture.md) §4, §8
- Related ADR: [0004 Files on Google Drive](./0004-store-case-files-and-signed-scans-on-the-associations-google-drive-with-hashes-in-postgres.md) · repo [0004 Migrations and queries](../../../adr/0004-migrations-and-queries.md)
