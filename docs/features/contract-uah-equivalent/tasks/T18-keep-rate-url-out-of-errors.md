# T18 — Keep the rate-source address out of the lookup error

> Follow-up of [review 2026-11-25](../_review/review-2026-11-25.md).

- **Layer:** infra
- **Depends on:** T16
- **AC:** n/a

## Definition of Done

Review 2026-11-25 finding 1: an error from the MCP client carries the endpoint as scheme, host and path only (no userinfo, query or fragment); a test with ?token= in the URL asserts the returned error text has no query value.

## Files

- `services/core/internal/finance/adapters/mcprates/mcprates.go`
- `services/core/internal/finance/adapters/mcprates/mcprates_test.go`
