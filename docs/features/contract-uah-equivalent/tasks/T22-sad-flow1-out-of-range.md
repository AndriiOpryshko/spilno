# T22 — Show the out-of-range branch in SAD Flow 1 and count four codes in T03

> Follow-up of [review 2026-11-25, round 3](../_review/review-2026-11-25-2.md).

- **Layer:** docs
- **Depends on:** T19
- **AC:** AC-04, AC-07

## Definition of Done

Review 2026-11-25 (round 3) finding 1: sad.md §6 Flow 1 has an `equivalent out of int64 range → insert, equivalent NULL (logged)` branch inside `rate returned`, as amended AC-04 says; T03's DoD in tasks.json and T03-finance-problem-codes.md says four codes (extended by T14); sad.md updated_at is bumped.

## Files

- `docs/features/contract-uah-equivalent/sad.md`
- `docs/features/contract-uah-equivalent/tasks/tasks.json`
- `docs/features/contract-uah-equivalent/tasks/T03-finance-problem-codes.md`
