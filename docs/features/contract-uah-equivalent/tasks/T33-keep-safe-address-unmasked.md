# T33 — Do not mask a secret value that is already part of the safe address

> Follow-up of [review 2026-11-27, round 5](../_review/review-2026-11-27-2.md).

- **Layer:** infra
- **Depends on:** T32
- **AC:** AC-10

## Definition of Done

Review 2026-11-27 (round 5) finding 4: a secret part of the endpoint that is a substring of its scheme + host + path is not masked outside URLs, since the log line prints that address anyway. TestUnavailableShortSecretsDoNotGarbleTheMessage gains the case http://127.0.0.1:<port>/mcp?h=127.0.0.1 and asserts that the dial error keeps 127.0.0.1:<port> readable; the long-secret masking test stays green.

## Files

- `services/core/internal/finance/adapters/mcprates/mcprates.go`
- `services/core/internal/finance/adapters/mcprates/unavailable_internal_test.go`
