---
status: accepted
date: 2026-10-25
feature: board-protocol
sad_section: "§4"
deciders: ["Andrii Opryshko"]
---

# 0004 Store case files and signed scans on the association's Google Drive with hashes in Postgres

## Context and problem statement

A справа holds source documents (estimates, invoices, contracts, owner requests, service notes); a protocol gets a printable PDF and a signed scan (AC-28, AC-29). PRD §6 limits files to 20 MB, PDF / JPEG / PNG, and keeps signed and withdrawn protocols ≥ 5 years, never hard-deleted. The user decided that all files live on Google Drive and Postgres keeps only metadata. Drive by itself lets anyone with access edit or delete a file, and its trash is emptied after 30 days, so the decision must say how QG-1 still holds.

## Decision drivers

- QG-1: nothing in a locked protocol changes, including its signed copy.
- PRD §6: ≤ 20 MB, PDF / JPEG / PNG; retention ≥ 5 years; upload ≤ 3 s p95 for 20 MB.
- QG-3: only current members of the board see the files; the registry belongs to the association, not to a member (CONTEXT "реєстр протоколів").
- ADR 0003: the product already depends on Google for sign-in.

## Blast radius

- Irreversible: yes — moving years of signed scans to another store is a migration of legal records.
- Multi-context: yes — `governance` (core and worker) and `identity` (the association's Google connection), and the web client's upload and download.
- Legitimate alternative: yes — a `bytea` table in Postgres, or Drive with a Postgres copy of each signed scan.
- Touches the database: yes — file metadata table, sha256 per file, retention.

## Considered options

1. **Drive of one association-owned Google account, metadata + sha256 in Postgres** — folder not shared; read-only content restriction on lock; daily hash check.
2. **Drive + an immutable Postgres copy of each signed scan** — the legal copy is also in the database backup.
3. **Drive of each member who uploads** — rejected by the glossary: documents would stay in a former member's personal Drive.

## Decision outcome

**Chosen:** option 1. At association setup a member connects one Google account owned by the association; `identity` stores its refresh token encrypted (AES-GCM, key from the environment). spilno uses scope `drive.file` (only files the app created), creates `spilno/<association>/` with sub-folders per справа and per protocol, and shares nothing in it except the wording Doc of ADR 0006. Uploads go web → `core` → Drive in one request, hashed with sha256 while streaming; `governance_files` stores `drive_file_id`, sha256, size, media type, kind, uploader and time. Members download through `core`, which streams from Drive. On lock, the job `apply_drive_restrictions` (ADR 0008) sets `contentRestrictions.readOnly` on the scan, the PDF and the справа documents, retried until Drive accepts it (SAD §6 Flow 8); the worker re-reads every locked file daily, compares sha256, restores a trashed file with `untrash` and raises an alert on a mismatch or a permanently deleted file.

### Consequences

- Good: no file store to run; the association sees its documents in its own Drive.
- Good: tampering is detected within a day, and trashed files are restored before Drive empties the trash.
- Bad: whoever signs in to the association's Google account can delete a file permanently; protection is detection only (SAD §11, High).
- Bad: Drive files are not in the Postgres backup; RPO/RTO for files depend on Google (SAD §7, §11).
- Neutral: upload latency includes the Drive upload; the 3 s target is watched per route (SAD §10).

### Migration and rollback

- Migration: new table `governance_files`; no existing rows. Association setup adds the Google connection (ADR 0003).
- Rollback: metadata rows of signed protocols are never deleted (ADR 0009); files stay on Drive. Moving to another store later is a copy job keyed by `drive_file_id` and verified by sha256.

## Links

- PRD: [PRD.ao-prd.md](../PRD.ao-prd.md) — US-02, US-07, AC-28, AC-29; §6 File size, Retention
- SAD: [sad.ao-architecture.md](../sad.ao-architecture.md) §4, §7
- Related ADR: [0009 Guard signed protocols](./0009-guard-signed-protocols-with-a-database-trigger-and-a-snapshot-hash.md) · [0003 Sign in with Google](./0003-sign-in-with-google-in-a-new-identity-context.md)
