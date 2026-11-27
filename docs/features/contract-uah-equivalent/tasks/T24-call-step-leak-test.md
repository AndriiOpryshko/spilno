# T24 — Pin the call step against an address leak

> Follow-up of [review 2026-11-25, round 3](../_review/review-2026-11-25-2.md).

- **Layer:** tests
- **Depends on:** T23
- **AC:** n/a

## Definition of Done

Review 2026-11-25 (round 3) finding 3: the vacuous "non-MCP answer" subtest of TestErrorKeepsCredentialsOutOfTheAddress is replaced by a stub that completes initialize and then drops the connection on tools/call (Post "<url>": EOF); with the scrubbing removed, that subtest fails.

## Files

- `services/core/internal/finance/adapters/mcprates/mcprates_test.go`
