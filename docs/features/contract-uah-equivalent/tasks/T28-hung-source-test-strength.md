# T28 — Make the hung-source test fail without the deadline transport

> Follow-up of [review 2026-11-27, round 4](../_review/review-2026-11-27.md).

- **Layer:** tests
- **Depends on:** T25
- **AC:** AC-04, AC-10

## Definition of Done

Review 2026-11-27 (round 4) finding 3: the hung-source stub completes initialize and hangs on tools/call and notifications/cancelled; the test asserts Rate returns within the caller's deadline and no request stays open on the server after it; a case with context.Background() asserts Rate returns within about 3 s (fallbackBudget). Against the pre-T25 mcprates.go, or with deadlineTransport replaced by http.DefaultTransport, the test fails on every run (checked with -count=20).

_Amended while implementing: go-sdk v1.8.0 sends no notifications/cancelled after an expired call; the request that hangs is the DELETE of session.Close, which the stub hangs too, and the test watches the server for 800 ms after Rate returns because that DELETE goes out after it._

## Files

- `services/core/internal/finance/adapters/mcprates/mcprates_test.go`
