---
status: Accepted
owner: "Andrii Opryshko"
reviewers: []
updated_at: "2026-10-25"
feature_size: L
stage: "04-05"
ticket: ""
---

# 0009 — Track the protocol lifecycle with explicit states, poll rounds and a revision number

- **Status:** Accepted
- **Date:** 2026-10-25
- **Deciders:** Andrii Opryshko (Architect), with Claude during the architecture-design pass

> **Amended 2026-10-25 (Stage 1 rescope):** states follow the PRD §1 Stage 1 lifecycle — `знятий з розгляду` is in scope (AC-35 … AC-37); frozen poll wording is a numbered **round**; printing needs every proposed slot confirmed instead of resolved блокуючі перевірки (AC-21, AC-14c); the number is reserved at the first print and kept on return ([ADR-0017](./0017-reserve-the-protocol-number-at-first-print-and-never-reuse-it.md)); any current член правління edits (AC-29, AC-03b dropped). The decision — explicit states plus an optimistic revision — is unchanged.

## Context

A protocol moves through чернетка, готовий до голосування (wording frozen so members cast the informal poll on exactly that version), готовий до підпису (text frozen, print copy rendered) and підписаний (signed scan and meeting facts uploaded); a draft that will not be signed becomes знятий з розгляду. Any edit, or «повернути в чернетку», returns it to чернетка: the open round closes with its votes kept and no longer counted (AC-05d), and a print copy is voided (AC-14b). Several writers touch a draft: any current член правління (AC-29) and the помічник протоколу, whose proposal arrives from a worker job seconds later (AC-26, AC-28b). Adding agenda items while it is a draft is never blocked (AC-03c).

## Decision drivers

- AC-05d, AC-14b, AC-09: an edit or a newer document version must close the round and void the print copy mechanically, not by convention.
- AC-28b: a proposal must never overwrite a confirmed slot or a newer human edit.
- AC-30, AC-32, AC-40, AC-38, AC-39: each action is allowed only in the states the PRD names.
- QG-1: the snapshot (ADR-0007) records the rounds and the print copy that was signed.

## Considered options

1. **Explicit states + revision number (optimistic concurrency)** — every edit carries the revision the author saw; a stale write gets a conflict.
2. **Edit lease** — while one member edits, the draft is held (a lease row with ~10 min expiry); others and the assistant wait.

## Decision outcome

**Chosen:** Option 1. No lease to expire or renew among several editors, and AC-05d and AC-14b reduce to "the revision moved".

Shape:
- States: `чернетка` → `готовий до голосування` (AC-30) → `чернетка` (AC-05d); `чернетка` or `готовий до голосування` → `готовий до підпису` (agenda complete AC-04, every proposed slot confirmed AC-21; AC-14c, AC-33) → `чернетка` (AC-14b); `готовий до підпису` → `підписаний` (signed scan + meeting facts, AC-14, ADR-0018); `чернетка`, `готовий до голосування` or `готовий до підпису` → `знятий з розгляду` with a non-empty reason (AC-35, AC-36, AC-37). `підписаний` and `знятий з розгляду` are final (AC-17, AC-38, AC-39).
- `revision` is an integer incremented by every change to text, slots, agenda items or the documents they cite. Every mutating request carries the revision it was built on; a mismatch answers 409 problem+json `governance.draft_revision_conflict` and the SPA reloads.
- Starting the poll stores an immutable **poll round** N (agenda items, text fragments and slot values at that revision); a vote references the round and the member, and each member casts or replaces only their own (AC-05e). Votes are accepted only while the protocol is `готовий до голосування` (AC-32). A return to чернетка or marking готовий до підпису closes the round; the next start opens round N+1 with no votes (AC-05d, AC-33).
- Marking готовий до підпису records a **print copy** (revision, number, rendered PDF's Drive id and sha256 — ADR-0013, ADR-0017). An edit voids it and keeps it in the history; the signed scan is accepted only against the current print copy.
- The proposal is applied by the `propose_draft` job against the revision it was built on: it fills only unconfirmed slots and text; a value for a confirmed slot is stored beside it and marks the slot as needing review (AC-28b); a stale revision re-bases the proposal onto the current draft without overwriting anything newer.

## Consequences

**Positive**
- AC-05d, AC-14b and AC-28b become mechanical checks on revision and state.
- No silent lost updates between members' tabs or between a member and the assistant.

**Negative**
- Every mutating operation in the contract carries a revision; the UI must handle 409 by reloading.
- Poll rounds and print copies add rows per poll and per print.

**Neutral**
- Stage 2 checks plug into the `готовий до підпису` transition without changing the states.

## Links

- PRD: [PRD.md](../PRD.md) §1, AC-03c, AC-04, AC-05d, AC-05e, AC-09, AC-14, AC-14b, AC-14c, AC-21, AC-28b, AC-29, AC-30, AC-32, AC-33, AC-35, AC-36, AC-37, AC-38, AC-39, AC-40
- SAD: [sad.md](../sad.md) §5, §6, §8
- Related ADR: [ADR-0007](./0007-lock-protocols-as-fingerprinted-immutable-snapshots.md), [ADR-0013](./0013-render-signing-copies-from-google-docs-templates.md), [ADR-0017](./0017-reserve-the-protocol-number-at-first-print-and-never-reuse-it.md), [ADR-0018](./0018-enter-meeting-facts-with-the-signed-scan-and-refuse-lock-without-them.md)
