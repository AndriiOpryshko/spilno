# T16 — Log the rate-source address without credentials

> Follow-up of [review 2026-11-24](../_review/review-2026-11-24.md).

- **Layer:** wiring
- **Depends on:** T10
- **AC:** n/a

## Definition of Done

Review 2026-11-24 finding 4: the startup log line carries the URL with userinfo redacted; a test pins the redaction.

## Files

- `services/core/cmd/api/main.go`
