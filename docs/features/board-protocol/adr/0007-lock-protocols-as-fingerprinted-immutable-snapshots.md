---
status: Accepted
owner: "Andrii Opryshko"
reviewers: []
updated_at: "2026-10-25"
feature_size: L
stage: "04-05"
ticket: ""
---

# 0007 — Lock a protocol as an immutable, fingerprinted snapshot

- **Status:** Accepted
- **Date:** 2026-10-25
- **Deciders:** Andrii Opryshko (Architect), with Claude during the architecture-design pass

## Context

Once the signed scan is uploaded, the protocol is locked (AC-14) and nothing about it may change (AC-17): not its text, votes or signed copy, and not what it applied — a правило статуту or roster term back-dated before its meeting, or a cited document later replaced, must not alter it (AC-17b). PRD §6 requires 0 changes after lock, checked by a periodic fingerprint comparison with an alert on mismatch. The registry, search and export read locked protocols.

## Decision drivers

- QG-1 (PRD §6 "Integrity of locked protocols").
- AC-17b: the locked protocol keeps rule versions, roster, computed quorum and majority, and document versions.
- QG-2: a later bug fix in quorum code must not silently change an old protocol's recorded result.
- PRD §6: registry open and search p95 ≤ 1 s with ≤ 500 locked protocols.

## Considered options

1. **Immutable snapshot + fingerprint** — at lock, write one canonical JSON snapshot of everything the protocol shows and applied, with its SHA-256; registry and export read the snapshot.
2. **References to append-only versions** — the locked protocol keeps foreign keys to rule, roster and document version rows; the fingerprint is computed over a join on demand.
3. **Event sourcing** — the protocol is its event log; the locked state is a projection.

## Decision outcome

**Chosen:** Option 1. The snapshot makes "nothing changes after lock" a property of one append-only row rather than of eight tables that must never be updated, and it records computed results instead of recomputing them with today's code.

Shape: table `governance.locked_protocols` — protocol id, number, meeting date, `schema_version`, `snapshot jsonb`, `fingerprint` (SHA-256 of the canonical serialization: sorted keys, no insignificant whitespace, amounts as integer kopiyky, dates as ISO 8601), locked by, locked at, the uploader's confirmation (AC-14). The snapshot holds text, agenda items, слоти рішення with source document version ids, attendance, неучасть у голосуванні with reasons, the wording version and votes, the rule versions and roster entries applied, computed quorum, кворум пункту and majority per item, every check with its override (reason, author, time), decision links (revokes / replaces / refers to), and the Drive file id, revision id and sha256 of the signed scan and every cited document version. A trigger (wrapped in `-- +goose StatementBegin/End`) rejects UPDATE and DELETE on the table. Lock is one transaction: compute, serialize, insert, mark the draft locked, apply decision-link effects (AC-17c). The nightly worker job recomputes each fingerprint from the stored snapshot and asks Drive for each pinned file's sha256; any mismatch is logged, counted in a metric and alerted.

## Consequences

**Positive**
- Back-dated rules, replaced documents and code changes cannot reach a locked protocol.
- Registry reads, search projections and exports read a single row.
- Verification is mechanical and cheap at ≤ 500 protocols.

**Negative**
- Data is duplicated between the working tables and the snapshot.
- The snapshot format is a long-lived contract: changes bump `schema_version`, and readers must handle every version ever written.
- A database owner can still drop the trigger; the fingerprint check detects but does not prevent tampering.

**Neutral**
- Search indexes (counterparty, amount, subject, period — AC-18) are derived from snapshots and can be rebuilt at any time.

## Links

- PRD: [[../PRD.md]] §6, AC-14, AC-17, AC-17b, AC-17c, AC-18, AC-20
- SAD: [[../sad.md]] §4, §5, §10
- Related ADR: [[0005-store-files-in-association-owned-google-drive-folder]], [[0008-keep-charter-rules-as-typed-effective-dated-parameters]]
