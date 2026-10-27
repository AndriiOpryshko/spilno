---
status: Accepted
owner: "Andrii Opryshko"
reviewers: []
updated_at: "2026-10-25"
feature_size: L
stage: "04-05"
ticket: ""
---

# 0017 — Reserve the protocol number at the first print copy and never reuse it

- **Status:** Accepted
- **Date:** 2026-10-25
- **Deciders:** Andrii Opryshko (Architect), with Claude during the Stage 1 rescope

> **Amended 2026-10-25 (cross-check with the ao-architecture SAD):** the register start value may be changed only until the first number is issued; after that a change is refused with 409 `governance.register_start_locked` (the earlier "only upward" rule is dropped). The реєстр протоколів lists signed protocols only; a number taken by a знятий з розгляду protocol shows there as a gap marked «знято з розгляду», and the withdrawn protocol itself stays an internal record ([ADR-0007](./0007-lock-protocols-as-fingerprinted-immutable-snapshots.md)). The decision is unchanged.

## Context

The number is printed on the paper copy that the board signs, so it becomes part of a legal record the moment it is printed. PRD: a protocol without a number takes the next number of the реєстр протоколів at готовий до підпису (one sequence, no duplicates — AC-14c, AC-33); on return to чернетка it keeps its number for the next print copy, and a number once taken is never given to another protocol, so the register may have gaps (AC-14b); a printed but unsigned protocol withdrawn keeps its number (AC-37); a draft withdrawn before printing takes none (AC-35, AC-36). The register continues the association's paper numbering from a start value in the settings (AC-42).

## Decision drivers

- AC-14b, AC-37: a number that may already be on paper must never appear on another protocol.
- AC-14c: one sequence, no duplicates, even with two members printing at once.
- AC-42: the first number comes from the settings.
- A start value changed after numbers were printed would put two numbering schemes on paper.

## Considered options

1. **Reserve at the first print from a per-association counter; never reuse; gaps allowed.**
2. **Reserve the lowest free number** at готовий до підпису and release it on return or withdrawal — no gaps, but a number already printed could be reused.
3. **Assign at lock** — no gaps, but the print copy cannot carry the number, contradicting AC-14c.
4. **A Postgres `SEQUENCE`** — no duplicates, but sequences cannot be restarted from the settings per association transactionally and skip values on rollback without a record.

## Decision outcome

**Chosen:** Option 1. `association_settings.next_protocol_number` holds the next free number, initialised from the start value the members enter (AC-42). In the transaction that marks a protocol готовий до підпису, if `protocols.number` is null, `UPDATE association_settings SET next_protocol_number = next_protocol_number + 1 WHERE association_id = $1 RETURNING next_protocol_number - 1` takes the row lock, and the value is written to the protocol; a unique index on `(association_id, number)` backs it. The number is never cleared: returning to чернетка, re-printing and withdrawing keep it. The start value may be changed while no protocol has a number; once the first number is issued (`next_protocol_number` moved from the start value, checked in the same row lock), a change is refused with 409 `governance.register_start_locked`.

## Consequences

**Positive**
- No number that might be on paper is ever reused; duplicates are impossible under concurrency.
- The register continues the paper numbering.

**Negative**
- The register can have gaps (a printed protocol withdrawn, AC-37); the реєстр протоколів lists signed protocols only and shows such a number as a gap marked «знято з розгляду», so the gap is explained without listing the withdrawn protocol, which stays an internal record.
- A wrong start value noticed after the first print cannot be fixed in settings; the numbering continues from it.
- The counter row is a write hotspot — irrelevant at 15-20 protocols a year.

**Neutral**
- Stage 2 numbering per calendar year, if the charter asks, is a change to the counter key only.

## Links

- PRD: [PRD.md](../PRD.md) AC-14b, AC-14c, AC-33, AC-35, AC-36, AC-37, AC-42
- SAD: [sad.md](../sad.md) §5, §6, §8
- Related ADR: [ADR-0009](./0009-track-draft-lifecycle-with-states-and-revision-numbers.md), [ADR-0013](./0013-render-signing-copies-from-google-docs-templates.md), [ADR-0015](./0015-scope-owned-rows-to-one-association-with-explicit-query-filters.md), [ADR-0007](./0007-lock-protocols-as-fingerprinted-immutable-snapshots.md)
