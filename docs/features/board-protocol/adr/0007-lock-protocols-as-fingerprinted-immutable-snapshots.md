---
status: Accepted
owner: "Andrii Opryshko"
reviewers: []
updated_at: "2026-10-25"
feature_size: L
stage: "04-05"
ticket: ""
---

# 0007 — Lock a protocol as an immutable, fingerprinted snapshot and never delete a protocol

- **Status:** Accepted
- **Date:** 2026-10-25
- **Deciders:** Andrii Opryshko (Architect), with Claude during the architecture-design pass

> **Amended 2026-10-25 (Stage 1 rescope):** the Stage 1 snapshot holds what Stage 1 records — frozen text, slots with their source document versions, the meeting facts entered at lock and the handwritten-corrections note ([ADR-0018](./0018-enter-meeting-facts-with-the-signed-scan-and-refuse-lock-without-them.md)), poll rounds as history, the print copy and the signed scan. Rule versions, computed quorum, checks and overrides join it in Stage 2 under a new `schema_version` (AC-17b). Added: a trigger on `protocols` rejects DELETE in any state (PRD §6 Retention). The snapshot decision is unchanged.

## Context

Once the signed scan and the meeting facts are uploaded, the protocol is locked (AC-14) and nothing about it may change (AC-17): not its text, its meeting facts or its signed copy. PRD §6 requires 0 changes after lock, checked by a periodic fingerprint comparison with an alert on mismatch, and requires that locked and withdrawn protocols, print copies and scans are kept ≥ 5 years and never hard-deleted. The registry and search read locked protocols.

## Decision drivers

- QG-1 (PRD §6 "Integrity of locked protocols").
- PRD §6 "Retention": nothing is ever hard-deleted, in any state (withdrawn drafts included, AC-35 … AC-37).
- AC-17: a locked protocol changes only through a new decision that references it.
- PRD §6: registry open and search p95 ≤ 200 ms.

## Considered options

1. **Immutable snapshot + fingerprint** — at lock, write one canonical JSON snapshot of everything the protocol shows, with its SHA-256; registry reads the snapshot.
2. **References to append-only versions** — the locked protocol keeps foreign keys to text, slot and file version rows; the fingerprint is computed over a join on demand.
3. **Event sourcing** — the protocol is its event log; the locked state is a projection.

## Decision outcome

**Chosen:** Option 1. The snapshot makes "nothing changes after lock" a property of one append-only row rather than of many tables that must never be updated.

Shape:
- Table `governance.locked_protocols` — protocol id, `association_id`, number, meeting date, `schema_version`, `snapshot jsonb`, `fingerprint` (SHA-256 of the canonical serialization: sorted keys, no insignificant whitespace, amounts as integer kopiyky, dates as ISO 8601), locked by, locked at.
- The Stage 1 snapshot (`schema_version` 1) holds the frozen text fragments, agenda items, слоти рішення with origin, confirmation and source document version ids, the meeting facts (members present, vote count per item, handwritten-corrections flag and note, who entered them and when), the poll rounds and votes labelled informal (AC-41), the print copy and the signed scan (Drive file id, revision id, sha256), and every cited document version (Drive file id, revision id, sha256).
- Two triggers, each wrapped in `-- +goose StatementBegin/End`: one rejects UPDATE and DELETE on `locked_protocols`; one rejects DELETE on `protocols` in any state. Changes to a `підписаний` or `знятий з розгляду` protocol are refused in `governance/app` (`governance.protocol_locked`, `governance.transition_not_allowed`); the registry shows the snapshot, never the working rows, so a change slipped past the app does not reach what members see and is caught by the fingerprint check.
- Lock is one transaction: store the scan metadata and meeting facts, build and serialize the snapshot, insert it, set the protocol `підписаний`.
- The nightly worker job recomputes each fingerprint from the stored snapshot and asks Drive for each pinned file's sha256 and revision; any mismatch is logged, counted in a metric and alerted.

## Consequences

**Positive**
- A replaced document or a code change cannot reach a locked protocol.
- No code path, bug or ad-hoc SQL can delete a protocol; withdrawn drafts survive with their history.
- Registry reads and search projections read a single row.
- Verification is mechanical and cheap at 15-20 protocols a year.

**Negative**
- Data is duplicated between the working tables and the snapshot.
- The snapshot format is a long-lived contract: Stage 2 bumps `schema_version`, and readers must handle every version ever written.
- A database owner can still drop a trigger; the fingerprint check and the backup detect but do not prevent tampering.
- A wrongly created draft cannot be deleted, only withdrawn with a reason.

**Neutral**
- Search indexes (counterparty, amount, subject, period — AC-18) are derived from snapshots and can be rebuilt at any time.

## Links

- PRD: [PRD.md](../PRD.md) §6, AC-14, AC-15, AC-17, AC-18, AC-35, AC-36, AC-37, AC-41
- SAD: [sad.md](../sad.md) §4, §6, §8, §10
- Related ADR: [ADR-0005](./0005-store-files-in-association-owned-google-drive-folder.md), [ADR-0009](./0009-track-draft-lifecycle-with-states-and-revision-numbers.md), [ADR-0018](./0018-enter-meeting-facts-with-the-signed-scan-and-refuse-lock-without-them.md), [ADR-0008](./0008-keep-charter-rules-as-typed-effective-dated-parameters.md) (Stage 2)
