# T34 — Name the rate lookup's failure by class instead of redacting its text

> Follow-up of [review 2026-11-27, round 5](../_review/review-2026-11-27-2.md): ends the redaction edge cases of T16, T18, T23, T29, T32 and T33 by removing what had to be redacted.

- **Layer:** infra
- **Depends on:** T33
- **AC:** AC-10

## Definition of Done

The mcprates error is `rate unavailable: <step>: endpoint=<scheme://host/path>: <class>`, where the class is a fixed word or two: `http <status>` of the last answer that was not a success, `timeout`, `canceled`, `connection refused`, `connection closed`, `network error` or `protocol error`. No text of net/http or the go-sdk reaches the error, so the substring masking (`addressForms`, `maskSecrets`, `redactURL`, the 6-byte floor) is deleted along with its tests. TestUnavailableNamesTheErrorClassOnly pins every class against a cause that carries a quoted secret URL; TestErrorNamesTheFailureClass pins `http 503` and `connection refused` end to end; the existing leak, redirect, hung-source and wiring tests stay green. SAD §8 Logging states the rule, not the mechanism, in the same commit.

## Files

- `services/core/internal/finance/adapters/mcprates/mcprates.go`
- `services/core/internal/finance/adapters/mcprates/mcprates_test.go`
- `services/core/internal/finance/adapters/mcprates/unavailable_internal_test.go`
- `docs/features/contract-uah-equivalent/sad.md`
