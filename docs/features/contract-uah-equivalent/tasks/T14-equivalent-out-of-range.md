# T14 — Refuse a hryvnia equivalent that does not fit int64 instead of storing a wrong one

> Follow-up of [review 2026-11-24](../_review/review-2026-11-24.md).

- **Layer:** domain
- **Depends on:** T13
- **AC:** AC-02, AC-04, AC-07

## Definition of Done

Review 2026-11-24 finding 1: `Rate.Convert` returns an error when the product does not fit int64; Create stores the contract with `uah_equivalent: null` and logs it; the get-it-later operation answers `409 finance.equivalent_out_of_range` and leaves the contract unchanged. PRD AC-02/AC-07 + §5.1, the delta and the live contract name the new code; registry and catalogue tests cover 4 codes; a boundary test row exists.

## Files

- `services/core/internal/finance/domain/rate.go`
- `services/core/internal/finance/domain/errors.go`
- `services/core/internal/finance/app/contracts.go`
- `services/core/internal/finance/adapters/http/handler.go`
- `services/core/api/openapi.yaml`
- `docs/features/contract-uah-equivalent/PRD.md`
- `docs/features/contract-uah-equivalent/contracts/openapi.yaml`
