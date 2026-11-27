# T17 — Pin the NBU client's body cap and deadline with tests

> Follow-up of [review 2026-11-24](../_review/review-2026-11-24.md).

- **Layer:** tests
- **Depends on:** —
- **AC:** AC-10

## Definition of Done

Review 2026-11-24 finding 5: TestRateErrors has a body over 1 MiB and a blocking upstream with a short deadline; both end in ErrUnavailable, the second promptly.

## Files

- `mcp/nbu-rates/internal/nbu/nbu_test.go`
