# T13 — Serve the live contract's problem-type catalogue and check the registry against it

> Follow-up of [review 2026-11-24](../_review/review-2026-11-24.md).

- **Layer:** wiring
- **Depends on:** T09
- **AC:** AC-07

## Definition of Done

Review 2026-11-24 finding 2: `resolveContractUahEquivalent` in services/core/api/openapi.yaml carries the delta's `x-problem-types`; `TestRegistryStatusesMatchContractProblemTypes` reads the live spec; `mise run gen` leaves no diff beyond the embedded spec.

## Files

- `services/core/api/openapi.yaml`
- `services/core/internal/finance/domain/errors_test.go`
