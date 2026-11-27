# T30 — Describe the structural redaction and its 6-byte floor in SAD §8

> Follow-up of [review 2026-11-27, round 5](../_review/review-2026-11-27-2.md).

- **Layer:** docs
- **Depends on:** T33
- **AC:** AC-10

## Definition of Done

Review 2026-11-27 (round 5) finding 1: sad.md §8 Logging no longer promises that no userinfo, query or fragment ever reaches the error. It says that the endpoint= field and every URL in the error are cut to scheme + host + path, that outside URLs only secret parts of at least 6 bytes of NBU_RATES_MCP_URL are masked (and not those already in the safe address, T33), and that an address net/http would print escaped is refused at start-up (T32). The SAD's updated_at is bumped.

## Files

- `docs/features/contract-uah-equivalent/sad.md`
