---
status: accepted
date: 2026-10-25
feature: board-protocol
sad_section: "§5"
deciders: ["Andrii Opryshko"]
---

# 0021 Reserve the protocol number on the first print copy and never reuse it

## Context and problem statement

A draft has no number (AC-08). The first render of the print copy gives a protocol without a number the next free number of the association's register (one sequence, no duplicates, AC-25); the protocol keeps it on a return to чернетка and prints it again (AC-26) and keeps it when withdrawn (AC-34c); a number once taken is never given to another protocol, so the register may have gaps (AC-26). The register continues from a number in the settings (AC-45, PRD §3). ADR 0013 chose "lowest free number" with release on return and withdrawal, which the rescoped PRD reverses. This ADR supersedes ADR 0013.

## Decision drivers

- AC-25, AC-26, AC-34c: no duplicates, no reuse, gaps allowed.
- AC-45: the register continues from the settings' start number.
- QG-1: a printed and signed number never changes.
- Two members may render two drafts at the same moment.

## Blast radius

- Irreversible: yes — numbers are printed on signed paper.
- Multi-context: no — inside `governance`.
- Legitimate alternative: yes — the lowest free number (ADR 0013), or `max + 1` computed per call.
- Touches the database: yes — counter row, unique index.

## Considered options

1. **Per-association counter row, incremented once per protocol on its first render.**
2. **Lowest free number, released on return and withdrawal** (ADR 0013).
3. **`max(number) + 1` under an advisory lock** — no counter, but a withdrawn top number could be recomputed after a data fix.

## Decision outcome

**Chosen:** option 1. `protocol_number_counters` (association PK, `next_number`) is created with `register_start_number` from the settings. In the transition to готовий до підпису, when `protocols.number` is null, `app/` runs `UPDATE protocol_number_counters SET next_number = next_number + 1 WHERE association_id = $1 RETURNING next_number - 1` (the row lock serializes concurrent renders until commit) and writes the number to the protocol in the same transaction; a protocol that has a number keeps it. Nothing ever sets the number back to null. A unique index `(association_id, number) WHERE number IS NOT NULL` is the final guard. `register_start_number` may be changed only while the counter has issued no number (`governance.register_start_locked`, added to PRD §5.1 via SAD §11).

### Consequences

- Good: no duplicates, no reuse, no search for holes; numbers follow the order of first render.
- Bad: the register has gaps for withdrawn printed protocols; the registry shows them as знятий з розгляду so the ревізійна комісія sees why.
- Neutral: a number reserved in a transaction that rolls back is not consumed (the counter update rolls back too).

### Migration and rollback

- Migration: `protocols.number int NULL` with the partial unique index; table `protocol_number_counters`, row created with the association settings.
- Rollback: drop the counter table; numbers on protocols stay.

## Links

- PRD: [PRD.ao-prd.md](../PRD.ao-prd.md) — AC-08, AC-25, AC-26, AC-34c, AC-45
- SAD: [sad.ao-architecture.md](../sad.ao-architecture.md) §5, §6
- Related ADR: supersedes [0013 Lowest free number](./0013-reserve-the-lowest-free-protocol-number-per-association-on-ready-to-sign.md) · [0018 Render the print copy](./0018-render-the-print-copy-from-a-pinned-google-docs-template-in-a-worker-job.md)
