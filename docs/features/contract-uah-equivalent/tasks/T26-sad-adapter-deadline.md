# T26 — Describe the adapter's own deadline and redirect refusal in the SAD

> Follow-up of [review 2026-11-27, round 4](../_review/review-2026-11-27.md).

- **Layer:** docs
- **Depends on:** T25
- **AC:** AC-04, AC-10

## Definition of Done

Review 2026-11-27 (round 4) finding 1: sad.md §8 Timeouts says the app sets the 3 s budget and the mcprates adapter bounds every HTTP request (including those the go-sdk sends on its own context) by the caller's deadline, with a 3 s fallback, and refuses redirects; the mcprates row of §5 says it refuses redirects; tasks/T05-mcp-rate-provider.md no longer says the adapter sets no timeout; sad.md updated_at is bumped.

## Files

- `docs/features/contract-uah-equivalent/sad.md`
- `docs/features/contract-uah-equivalent/tasks/T05-mcp-rate-provider.md`
