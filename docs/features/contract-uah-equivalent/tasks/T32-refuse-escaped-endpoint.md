# T32 — Refuse a rate-source address that net/http would print escaped

> Follow-up of [review 2026-11-27, round 5](../_review/review-2026-11-27-2.md).

- **Layer:** infra
- **Depends on:** T29
- **AC:** AC-10

## Definition of Done

Review 2026-11-27 (round 5) finding 3: an NBU_RATES_MCP_URL that fails url.Parse, or that holds any character strconv.Quote escapes (such as a double quote), is refused when the finance module is built: the rate source stays mcprates.Disabled, a warning names the reason and the safe part of the address only, and no lookup goes out. A test with http://127.0.0.1:1/mcp?token="quoted-secret" and with http://u:pa"ss-secret@127.0.0.1:1/mcp asserts that neither secret reaches the warning or any lookup error, and that the existing leak, redirect and wiring tests stay green.

## Files

- `services/core/internal/finance/adapters/mcprates/mcprates.go`
- `services/core/internal/finance/adapters/mcprates/unavailable_internal_test.go`
- `services/core/internal/finance/module.go`
- `services/core/internal/finance/rates_wiring_test.go`
