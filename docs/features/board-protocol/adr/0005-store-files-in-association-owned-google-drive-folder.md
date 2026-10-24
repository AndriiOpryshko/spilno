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

## Context

The feature stores case documents and their versions, signed scans up to 50 MB, the copy printed for signing and export packages — roughly up to 1.2 GB a year at 15-20 protocols. The association keeps its documents in its Google Drive; the author's requirement is that all files stay on the association's Drive, with only metadata in Postgres. A locked protocol must keep the exact document versions it cited (AC-17b), a changed document must flag the slots confirmed against it (AC-09), and a former member must see nothing (AC-19).

## Decision drivers

- QG-1, AC-17b: cited and signed files never change after lock; the periodic integrity check must cover files.
- QG-3, AC-19: access ends on the term end date, with no hint that records exist.
- PRD §6: signed-copy upload p95 ≤ 20 s. (Durability RPO ≤ 24 h with quarterly restore is deliberately not met in the first release — no database backup; SAD §1 decision override, §7, §11.)
- PRD §6.1 abuse case "Malicious upload": files served for download, never rendered as active content.

## Considered options

1. **Immutable content-addressed blobs in Postgres** behind a port — one backup covers data and files.
2. **S3-compatible object storage** (MinIO locally) with metadata in Postgres.
3. **Association's Google Drive, folder shared with nobody but the association account** — members reach files only through spilno.
4. **Association's Google Drive, folder shared with current members**, permissions synced daily with the roster terms.

## Decision outcome

**Chosen:** Option 3. Files stay where the association keeps its records, and access control stays in one place — spilno's per-request term check — instead of a daily permission sync that would let a former member keep Drive access for up to a day and edit cited files.

Shape:
- spilno is authorized once against the association-owned Google account (OAuth consent with an offline refresh token, kept encrypted as a secret). A shared infrastructure package `internal/platform/google` wraps the Drive and Docs APIs; `cases` and `governance` each implement their own file port on it.
- Layout: `spilno/cases/<case-id>/`, `spilno/protocols/<protocol-id>/` (copies for signing, signed scan), `spilno/exports/`. The folder is not shared with anyone.
- Per file version Postgres keeps: Drive file id, head revision id, sha256 (from Drive's `sha256Checksum`, verified on upload), size, media type, original name, author, time. A new version of a document is a new Drive file, never an overwrite.
- An existing file in the association's Drive is attached by **copying** it into the case folder through spilno's own file browser, so editing the original later never touches the cited copy.
- Uploads go browser → API (type and size check, ≤ 50 MB) → Drive resumable upload. Downloads are proxied by the API after the term check and served with `Content-Disposition: attachment`; members never receive Drive links.
- The nightly integrity job compares the sha256 Drive reports for every pinned file with the stored one.

## Consequences

**Positive**
- The association's records stay in the association's account, independent of spilno's hosting.
- One authorization path: the roster term check, enforced by the API for every file read.
- Drive reports a sha256 per binary file, so the integrity check needs no download.

**Negative**
- A file read or write depends on the Drive API being reachable; uploads and downloads go through the API process, adding a hop to the 20 s upload budget.
- A database backup, once added, will not cover files: restoring will need both the database dump and the Drive folder, and the restore drill must check they agree. The first release has no database backup at all (SAD §1 decision override, §11).
- Whoever holds the association Google account can edit or delete files outside spilno; the integrity check detects it but cannot prevent it (tracked in §11).
- If the refresh token is revoked, file access stops until someone re-authorizes; the account's storage quota is shared with its mail.

**Neutral**
- The file ports keep a later move to object storage an adapter plus a copy script.

## Links

- PRD: [[../PRD.md]] §6, §6.1, AC-09, AC-17b, AC-19, AC-20
- SAD: [[../sad.md]] §4, §5, §7, §11
- Related ADR: [[0004-use-gemini-api-for-document-reading-and-drafting]], [[0007-lock-protocols-as-fingerprinted-immutable-snapshots]]
