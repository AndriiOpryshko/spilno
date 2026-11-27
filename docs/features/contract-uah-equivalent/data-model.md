---
status: Accepted
owner: "Andrii Opryshko"
updated_at: "2026-11-25"
feature_size: "S"
prd: "PRD.md"
sad: "sad.md"
mode: "brownfield"
stage: "07"
contexts: [finance]
---

# Data model — contract-uah-equivalent

Brownfield delta on one table owned by `finance`. Decision: [ADR-0002](./adr/0002-store-the-hryvnia-equivalent-on-the-contract-and-never-recompute-it.md).

## Entities

### `contracts` (existing) — added columns

| Column | Type | Null | Meaning | Origin |
|---|---|---|---|---|
| `uah_amount_kopiyky` | `BIGINT` | yes | the гривневий еквівалент in kopiykas | PRD AC-01, AC-02 |
| `uah_rate_micros` | `BIGINT` | yes | the офіційний курс of `signed_at`, hryvnia × 10⁶ per unit of `currency` | PRD AC-09 |

Invariant (app, not `CHECK`): both `NULL` or both set; set only when `currency` is not `UAH`; once set, never
changed. `rate_date` = `signed_at` and `source` = `NBU` are derived, not stored.

## Indexes

None: no query filters or orders by the new columns.

## Query → index map

| Query (sqlc) | Change | Index |
|---|---|---|
| `CreateContract` | inserts the two columns (may be `NULL`) | — |
| `GetContract`, `ListContracts` | `SELECT *` already returns them | existing `idx_contracts_association_created` |
| `SetContractUAHEquivalent` (new) | `UPDATE contracts SET ... WHERE id = $1 AND association_id = $2 AND uah_amount_kopiyky IS NULL RETURNING *` | primary key |

## Error catalogue → constraints

No database constraint backs the `finance.*` codes of PRD §5.1; all four are decided in `app`.

## Migration plan

| Order | Staged file | Phase |
|---|---|---|
| 1 | [`NNNNN_add_contracts_uah_equivalent.sql`](./migrations/NNNNN_add_contracts_uah_equivalent.sql) | expand only: nullable columns, no backfill, no contract step |

Row-level security on `contracts` (live `00046`) is per row and needs no change.

## Deviations from the rules

None. The table predates the rules (`CHECK`s, `updated_at` trigger stay); the new columns follow them.

## Seeds / fixtures

None. Tests create contracts through the API.
