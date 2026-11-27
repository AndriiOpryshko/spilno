# T15 — Drop the staging header from the promoted migration

> Follow-up of [review 2026-11-24](../_review/review-2026-11-24.md).

- **Layer:** migration
- **Depends on:** T01
- **AC:** n/a

## Definition of Done

Review 2026-11-24 finding 3: services/core/migrations/00050_add_contracts_uah_equivalent.sql no longer says it is staged; schema unchanged.

## Files

- `services/core/migrations/00050_add_contracts_uah_equivalent.sql`
