# T29 — Redact the rate-source address structurally instead of masking substrings

> Follow-up of [review 2026-11-27, round 4](../_review/review-2026-11-27.md).

- **Layer:** infra
- **Depends on:** T23
- **AC:** AC-10

## Definition of Done

Review 2026-11-27 (round 4) finding 4: the lookup error no longer replaces secret substrings anywhere in the message; each URL in it is parsed and printed as scheme + host + path (no userinfo, query or fragment), with a placeholder that cannot collide with message text. A test with a one-character username, ?v=1 and ?x=end asserts the host (127.0.0.1:<port>) and the step stay readable and no NUL byte or secret reaches the error; the existing leak and redirect tests stay green.

## Files

- `services/core/internal/finance/adapters/mcprates/mcprates.go`
- `services/core/internal/finance/adapters/mcprates/unavailable_internal_test.go`
