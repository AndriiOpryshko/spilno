# T25 — Bound the rate lookup's HTTP requests so a hung source frees its socket

> Follow-up of [review 2026-11-25, round 3](../_review/review-2026-11-25-2.md).

- **Layer:** infra
- **Depends on:** T23
- **AC:** AC-04, AC-10

## Definition of Done

Review 2026-11-25 (round 3) finding 4: the transport's http.Client has a Timeout a little above RateTimeout (or a ResponseHeaderTimeout); a test against a server that hangs asserts the server sees its request context cancelled shortly after Rate returns.

## Files

- `services/core/internal/finance/adapters/mcprates/mcprates.go`
- `services/core/internal/finance/adapters/mcprates/mcprates_test.go`
