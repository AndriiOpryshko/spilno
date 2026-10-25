---
status: Accepted
owner: "Andrii Opryshko"
reviewers: []
updated_at: "2026-10-25"
feature_size: L
stage: "04-05"
ticket: ""
---

# 0005 — Store files in an association-owned Google Drive folder, with metadata in Postgres

- **Status:** Accepted
- **Date:** 2026-10-25
- **Deciders:** Andrii Opryshko (Architect), with Claude during the architecture-design pass

> **Amended 2026-10-25 (Stage 1 rescope):** spilno uses only the `drive.file` scope, so it sees only files it created; "attach an existing Drive file by copying it" is dropped from Stage 1 (it needs a wider scope or a Picker flow — SAD §11). Limits restated against PRD §6 (≤ 20 MB, PDF / JPEG / PNG, upload p95 ≤ 3 s). The refresh token is encrypted with AES-GCM. The database is now backed up ([ADR-0016](./0016-back-up-postgres-nightly-off-the-vm-with-quarterly-restore-drills.md)); files are not. Export packages are Stage 2. The choice of the association's Drive is unchanged.

> **Amended 2026-10-25 (cross-check with the ao-architecture SAD):** two protections added inside the same `drive.file` scope. On lock and on withdrawal a River job `governance.apply_drive_restrictions` sets `contentRestrictions.readOnly` (with the reason «protocol locked» or «protocol withdrawn») on the protocol's files — its print copies, the signed scan and the case document versions it cites; `drive.file` allows it because spilno created those files. The nightly integrity check restores a trashed pinned file with `files.untrash`, records it and alerts; a file deleted for good stays a mismatch with an alert. Kept as before: a newer version of a case document is a new file and marks the citing slots of open drafts `needs_review` (AC-09). The choice of the association's Drive is unchanged.

## Context

The feature stores case documents and their versions, the PDF print copies and the signed scans, files up to 20 MB (PRD §6), at 15-20 protocols a year. The association keeps its documents in its Google Drive; the author's requirement is that all files stay on the association's Drive, with only metadata in Postgres. A locked protocol must keep the exact document versions it cited and its signed scan (AC-14, AC-17), a newer document version must flag the slots confirmed against it (AC-09), and a former member must see nothing (AC-19).

## Decision drivers

- QG-1, AC-17: cited and signed files never change after lock; the periodic integrity check must cover files of signed and withdrawn protocols.
- QG-3, AC-19: access ends on the term end date, with no hint that records exist.
- PRD §6: files up to 20 MB, PDF / JPEG / PNG, upload p95 ≤ 3 s for 20 MB; retention ≥ 5 years, never hard-deleted.
- PRD §6.1 abuse case "Malicious upload": files served for download, never rendered as active content.
- Least privilege: the smallest Google scope that does the job.

## Considered options

1. **Immutable content-addressed blobs in Postgres** behind a port — one backup covers data and files.
2. **S3-compatible object storage** (MinIO locally) with metadata in Postgres.
3. **Association's Google Drive, folder created by spilno under `drive.file` and shared with nobody** — members reach files only through spilno.
4. **Association's Google Drive with the full `drive` scope, folder shared with current members**, permissions synced daily with the roster terms.

## Decision outcome

**Chosen:** Option 3. Files stay where the association keeps its records, access control stays in one place — spilno's per-request term check — and `drive.file` limits what a leaked token can reach to the files spilno itself created.

Shape:
- spilno is authorized once against the association-owned Google account (OAuth consent for `drive.file` with an offline refresh token). The token is stored by `identity` encrypted with AES-GCM under a key from the environment and handed to `governance` as a token source ([ADR-0014](./0014-keep-board-protocol-in-governance-and-add-only-an-identity-context.md)). `internal/platform/google` builds the Drive and Docs clients; `governance/adapters/gdrive` implements the `Files` port.
- Layout: `spilno/cases/<case-id>/`, `spilno/protocols/<protocol-id>/` (print copies, signed scan), `spilno/templates/` (layout templates, ADR-0013). spilno creates the folders; they are not shared with anyone.
- Per file version Postgres keeps: Drive file id, head revision id, sha256 (computed while streaming and compared with Drive's `sha256Checksum`), size, media type, original name, author, time, `association_id`. A new version of a document is a new Drive file, never an overwrite.
- Uploads go browser → API (content-sniffed type and size check, ≤ 20 MB) → Drive upload. Downloads are proxied by the API after the term check and served with `Content-Disposition: attachment`; members never receive Drive links.
- On lock and on withdrawal the `governance.apply_drive_restrictions` job (enqueued in the lock or withdrawal transaction, idempotent per file) sets `contentRestrictions.readOnly` on every file the snapshot pins: print copies, the signed scan, the cited case document versions ([ADR-0007](./0007-lock-protocols-as-fingerprinted-immutable-snapshots.md)).
- The nightly integrity job compares the sha256 and revision Drive reports for every file pinned by a signed or withdrawn snapshot with the stored ones; a pinned file found in the trash is restored with `files.untrash`, recorded and alerted; a missing file or a different sha256 is a mismatch with an alert. For documents of open drafts a change marks the citing slots `needs_review` (AC-09).

## Consequences

**Positive**
- The association's records stay in the association's account, independent of spilno's hosting.
- One authorization path: the roster term check, enforced by the API for every file read.
- `drive.file` is a non-sensitive scope: a smaller consent screen and a smaller blast radius for the token.
- Drive reports a sha256 per binary file, so the integrity check needs no download.
- `contentRestrictions.readOnly` and `untrash` need no wider scope than `drive.file`.

**Negative**
- A file read or write depends on the Drive API being reachable; uploads go through the API process, adding a hop to the 3 s upload budget (SAD §11).
- The database backup does not cover files: restoring needs the dump and the Drive folder, and the restore drill checks they agree (ADR-0016).
- A file already elsewhere in the association's Drive cannot be attached directly; a member downloads and uploads it (SAD §11).
- Whoever holds the association Google account can still lift the read-only restriction, or empty the trash, outside spilno; the restriction stops casual edits, `untrash` undoes a move to the trash, and the integrity check detects the rest but cannot prevent it (SAD §11).
- If the refresh token is revoked, file access stops until someone re-authorizes; the account's storage quota is shared with its mail.

**Neutral**
- The `Files` port keeps a later move to object storage an adapter plus a copy script.

## Links

- PRD: [PRD.md](../PRD.md) §6, §6.1, AC-09, AC-14, AC-17, AC-19
- SAD: [sad.md](../sad.md) §4, §5, §6, §7, §8, §11
- Related ADR: [ADR-0004](./0004-use-gemini-api-for-document-reading-and-drafting.md), [ADR-0006](./0006-run-background-work-in-worker-via-river-queue.md), [ADR-0007](./0007-lock-protocols-as-fingerprinted-immutable-snapshots.md), [ADR-0014](./0014-keep-board-protocol-in-governance-and-add-only-an-identity-context.md), [ADR-0016](./0016-back-up-postgres-nightly-off-the-vm-with-quarterly-restore-drills.md)
