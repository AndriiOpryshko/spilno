# T25 — Bound the rate lookup's HTTP requests so a hung source frees its socket

> Follow-up of [review 2026-11-25, round 3](../_review/review-2026-11-25-2.md).

- **Layer:** infra
- **Depends on:** T23
- **AC:** AC-04, AC-10

## Definition of Done

Review 2026-11-25 (round 3) finding 4: every HTTP request of a lookup (including the notifications/cancelled POST and the DELETE the go-sdk sends on a context of its own) is bounded by the caller's deadline, or 3 s when it has none; a test against a server that reads each request and hangs asserts Rate returns within the caller's deadline and no request stays open on the server after it. (Amended while implementing: a client Timeout above 3 s still let Rate wait 3.6 s on the cancel notice, against AC-04's 3 s bound.)

## Files

- `services/core/internal/finance/adapters/mcprates/mcprates.go`
- `services/core/internal/finance/adapters/mcprates/mcprates_test.go`
