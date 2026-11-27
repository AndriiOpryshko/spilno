# T21 — Log an out-of-range equivalent without the amount

> Follow-up of [review 2026-11-25](../_review/review-2026-11-25.md).

- **Layer:** app
- **Depends on:** T14
- **AC:** AC-04

## Definition of Done

Review 2026-11-25 finding 4: the out-of-range warning on create logs currency, day and error, no amount, as SAD §8 says; a test pins the fields.

## Files

- `services/core/internal/finance/app/contracts.go`
- `services/core/internal/finance/app/contracts_test.go`
