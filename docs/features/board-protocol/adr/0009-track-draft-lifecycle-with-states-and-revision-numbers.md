---
status: Accepted
owner: "Andrii Opryshko"
reviewers: []
updated_at: "2026-10-25"
feature_size: L
stage: "04-05"
ticket: ""
---

# 0009 — Track the draft lifecycle with explicit states and a revision number

- **Status:** Accepted
- **Date:** 2026-10-25
- **Deciders:** Andrii Opryshko (Architect), with Claude during the architecture-design pass

## Context

A protocol draft moves through готовий до голосування (wording frozen so members vote on exactly that version), готовий до підпису (frozen for printing after the blocking checks) and locked (signed scan uploaded). Any edit returns it to draft: votes stay with the old wording version and no longer count (AC-05d), and a copy printed earlier must be re-issued (AC-14b). Two writers touch a draft: the головуючий (AC-03b — the only member who edits) and the помічник протоколу, whose proposal arrives from a worker job seconds later (AC-26, AC-28b). The PRD keeps the lifecycle naive — no lifecycle rule blocks adding items while it is a draft (AC-03c).

## Decision drivers

- AC-05d, AC-14b: an edit must invalidate frozen wording and issued copies mechanically, not by convention.
- AC-28b: a proposal must never overwrite a confirmed slot or a newer human edit.
- AC-03b, AC-05e: authorship rules per action.
- QG-1: the snapshot (ADR-0007) records which wording version the votes were cast on.

## Considered options

1. **Explicit states + revision number (optimistic concurrency)** — every edit carries the revision the author saw; a stale write gets a conflict.
2. **Edit lease** — while the головуючий edits, the draft is held (a lease row with ~10 min expiry); the assistant waits.

## Decision outcome

**Chosen:** Option 1. No lease to expire or renew, and both AC-05d and AC-14b reduce to "the revision moved".

Shape:
- States: `draft` → `ready_for_vote` → `draft` (on edit) …; `draft` or `ready_for_vote` → `ready_for_signing` (all блокуючі перевірки resolved or overridden, number chosen, AC-14c) → `draft` (on edit) …; `ready_for_signing` → `locked` (signed scan + confirmation, AC-14); a report of handwritten corrections keeps it at `ready_for_signing` (AC-15). `знятий з розгляду` is not in the first release (PRD §3).
- `revision` is an integer incremented by every change to text, slots, agenda items, attendance or неучасть у голосуванні. Every mutating request carries the revision it was built on; a mismatch answers 409 problem+json `governance.draft_revision_conflict` and the SPA reloads.
- Freezing for votes stores an immutable **wording version** (agenda items, decision text fragments and slot values at that revision); a попереднє голосування references the wording version, so votes on an older version are kept but not counted.
- Marking готовий до підпису records an **issued copy** (revision, rendered PDF's Drive id and sha256 — ADR-0013). An edit voids it; the signed scan is accepted only against the current issued copy.
- The assistant's proposal is applied through `governance.ApplyProposal` against the revision it was built on: it fills only unconfirmed slots and text; a value for a confirmed slot is stored beside it and marks the slot as needing review (AC-28b); a stale revision re-bases the proposal onto the current draft without overwriting anything newer.

## Consequences

**Positive**
- AC-05d, AC-14b and AC-28b become mechanical checks on revision and state.
- No silent lost updates between the головуючий's tabs or between the головуючий and the assistant.

**Negative**
- Every mutating operation in the contract carries a revision; the UI must handle 409 by reloading.
- Wording versions and issued copies add rows per freeze and per print.

**Neutral**
- A future `знятий з розгляду` state is one more terminal transition from `draft`.

## Links

- PRD: [[../PRD.md]] §1, AC-03b, AC-03c, AC-05d, AC-14, AC-14b, AC-14c, AC-15, AC-28b
- SAD: [[../sad.md]] §5, §6
- Related ADR: [[0007-lock-protocols-as-fingerprinted-immutable-snapshots]], [[0013-render-signing-copies-from-google-docs-templates]]
