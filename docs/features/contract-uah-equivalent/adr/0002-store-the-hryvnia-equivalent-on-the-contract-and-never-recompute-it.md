---
status: accepted
date: 2026-11-20
feature: contract-uah-equivalent
sad_section: "§4"
deciders: ["Andrii Opryshko", "Claude (easy depth)"]
---

# 0002 Store the hryvnia equivalent on the contract and never recompute it

## Context and problem statement

The гривневий еквівалент could be computed on every read or stored once. The ревізійна комісія must be able
to trace it to a rate and a day (PRD AC-09), and reads must not depend on the rate source (AC-04).

## Decision drivers

- An official rate of a past day is final; the signing day of a contract does not change.
- Reads (get, list) must stay as fast and as available as today.
- `.claude/rules/migrations.md`: structured fields are first-class columns; no business `CHECK`.

## Blast radius

- Irreversible: partly — stored figures become the record; dropping the columns later loses them.
- Multi-context: no — `finance` only.
- Legitimate alternative: yes.
- Touches the database: yes — two nullable columns on `contracts`.

## Considered options

1. **Two nullable columns on `contracts`: `uah_amount_kopiyky BIGINT`, `uah_rate_micros BIGINT`; written
   once (at creation, or later only while `NULL`).**
2. Compute on read, cache rates in a `rates` table.
3. A separate `contract_equivalents` table.

## Decision outcome

**Chosen:** Option 1. Both columns are set together or not at all (enforced in the app and by the
conditional `UPDATE ... WHERE uah_amount_kopiyky IS NULL`). The дата курсу is the contract's `signed_at` and
the джерело курсу is always the NBU, so neither needs a column today.

- Option 2 makes every read depend on a rate being present and adds a table nobody else needs yet.
- Option 3 is a 1:1 table for two numbers.

### Consequences

- Good: reads are unchanged; the figure shown never moves.
- Good: additive, nullable, no backfill; existing rows get the equivalent on demand (AC-05).
- Bad: if a second source or another rate day is ever allowed, `rate_date` and `source` become columns.
- Bad: «both or neither» is an app invariant, not a database one.

## Migration

One goose migration adding the two columns (`IF NOT EXISTS`), Down drops them. No backfill, no index: no
query filters by them.

## Links

PRD AC-05, AC-06, AC-09, §9 · SAD §5, §8 · [data-model](../data-model.md)
