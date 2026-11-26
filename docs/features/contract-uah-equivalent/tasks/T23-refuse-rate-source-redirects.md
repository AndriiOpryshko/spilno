# T23 — Refuse redirects from the rate source and scrub each secret part of its address

> Follow-up of [review 2026-11-25, round 3](../_review/review-2026-11-25-2.md).

- **Layer:** infra
- **Depends on:** T18
- **AC:** n/a

## Definition of Done

Review 2026-11-25 (round 3) finding 2: the MCP transport uses its own http.Client whose CheckRedirect returns http.ErrUseLastResponse; unavailable() removes the password, the username, every query value and the fragment as substrings, not only whole-URL spellings; a test against a server that redirects (e.g. /mcp → /mcp/) asserts the error text has no token, username, password or fragment.

## Files

- `services/core/internal/finance/adapters/mcprates/mcprates.go`
- `services/core/internal/finance/adapters/mcprates/mcprates_test.go`
