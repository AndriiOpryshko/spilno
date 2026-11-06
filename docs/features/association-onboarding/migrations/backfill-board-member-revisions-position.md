---
table: board_member_revisions
column: position
staged_file: NNNNN_backfill_board_member_revisions_position.sql
phase: 2 of 3
owner: "Andrii Opryshko"
updated_at: "2026-11-05"
---

# Backfill - `board_member_revisions.position`

## Why

PRD §9: existing roster entries are члени правління; nobody is голова until AC-19c.

## Source of the value

The constant 'member' — no other table is read.

## Size and ETA

| Rows to fill (estimate) | Batch size | Batches | ETA | Measured on |
|---|---|---|---|---|
| `SELECT count(*) FROM board_member_revisions WHERE position IS NULL` -> ≈ 7–20 (one association's roster revisions) | one statement | 1 | < 1 s | throwaway postgres:18.6, 2026-11-05 |

Small table: one `UPDATE` in goose's transaction, no batching and no `NO TRANSACTION`.

## Preconditions

- Phase 1 is deployed and the app writes `position` on every new row.

## Resumability

The `UPDATE` touches only rows where `position IS NULL`; re-running is safe.

## Verification (gate for phase 3)

```sql
SELECT count(*) FROM board_member_revisions WHERE position IS NULL;   -- must be 0
```

## Rollback

Down is a no-op; rolling back phase 1 drops the column.
