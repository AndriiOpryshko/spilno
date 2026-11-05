---
table: contracts
column: association_id
staged_file: NNNNN_backfill_contracts_association_id.sql
phase: 2 of 3
owner: "Andrii Opryshko"
updated_at: "2026-11-05"
---

# Backfill - `contracts.association_id`

## Why

SAD ADR-0003: every existing contract belongs to the one association seeded in 00004.

## Source of the value

The constant '00000000-0000-7000-8000-000000000001'::uuid — no other table is read.

## Size and ETA

| Rows to fill (estimate) | Batch size | Batches | ETA | Measured on |
|---|---|---|---|---|
| `SELECT count(*) FROM contracts WHERE association_id IS NULL` -> tens (finance contracts of one association) | one statement | 1 | < 1 s | throwaway postgres:18.6, 2026-11-05 |

Small table: one `UPDATE` in goose's transaction, no batching and no `NO TRANSACTION`.

## Preconditions

- Phase 1 is deployed and the app writes `association_id` on every new row.

## Resumability

The `UPDATE` touches only rows where `association_id IS NULL`; re-running is safe.

## Verification (gate for phase 3)

```sql
SELECT count(*) FROM contracts WHERE association_id IS NULL;   -- must be 0
```

## Rollback

Down is a no-op; rolling back phase 1 drops the column.
