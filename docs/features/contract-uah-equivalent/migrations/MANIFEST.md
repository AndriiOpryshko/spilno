# Staged migrations — contract-uah-equivalent

| Order | Staged file | Depends on | Expand/contract phase |
|---|---|---|---|
| 1 | `NNNNN_add_contracts_uah_equivalent.sql` | live `00002` (`contracts`) | expand (nullable, final) |

Promote with `mise run migrate:new add_contracts_uah_equivalent` and copy the body; then `mise run gen`.
