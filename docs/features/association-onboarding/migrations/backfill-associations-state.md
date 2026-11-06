---
table: associations
column: state
staged_file: NNNNN_backfill_associations_state.sql
phase: 2 of 3
owner: "Andrii Opryshko"
updated_at: "2026-11-05"
---

# Backfill - `associations.state`

## Why

PRD §9: the hand-made association becomes підключене ОСББ.

## Source of the value

The constant 'connected' — no other table is read.

## Size and ETA

| Rows to fill (estimate) | Batch size | Batches | ETA | Measured on |
|---|---|---|---|---|
| `SELECT count(*) FROM associations WHERE state IS NULL` -> 1 (the association set up by hand) | one statement | 1 | < 1 s | throwaway postgres:18.6, 2026-11-05 |

Small table: one `UPDATE` in goose's transaction, no batching and no `NO TRANSACTION`.

## Preconditions

- Phase 1 is deployed and the app writes `state` on every new row.

## Resumability

The `UPDATE` touches only rows where `state IS NULL`; re-running is safe.

## Verification (gate for phase 3)

```sql
SELECT count(*) FROM associations WHERE state IS NULL;   -- must be 0
```

## Rollback

Down is a no-op; rolling back phase 1 drops the column.
