---
id: T55
title: "Bound the sign-in and storage-consent state"
layer: "app"
deps: ["T54"]
acs: ["AC-05", "AC-27"]
files_hint:
  - "services/core/internal/identity/app/signin.go"
  - "services/core/internal/identity/app/storage.go"
  - "services/core/internal/identity/adapters/http/handler.go"
  - "services/core/internal/identity/adapters/http/storage.go"
  - "services/core/internal/identity/signin_test.go"
  - "services/core/internal/identity/storage_test.go"
owner: "Andrii Opryshko"
estimate: "M"
status: "done"
---

# T55 — Bound the sign-in and storage-consent state

## Why

[Review of 2026-11-10](../_review/review-2026-11-10.md), B4. `StartSignIn` is public and adds an entry to an in-process map on every call, scanning the whole map under a mutex. The tracker's known gap «sign-in state lives in process memory» is the same state.

**Satisfies:** [PRD §5](../PRD.md) AC-05, AC-27 · **Blocked by:** [T54](./T54-commit-before-response.md)

## What

- Keep the state, the PKCE verifier and the nonce in a short-lived, HttpOnly, HMAC-signed cookie, or cap the map and sweep it on a timer; the cookie also closes the known gap for more than one API instance.
- The same for the storage consent state.

## Definition of Done

- [ ] Acceptance: Starting sign-in any number of times does not grow server memory, and a callback with a forged, expired or replayed state is still refused.
- [ ] test: 10 000 starts leave the state bounded
- [ ] the existing bad-state, expired-state and replay tests of sign-in and storage stay green
- [ ] `mise run lint` and `mise run test` green

## Notes

If the cookie is chosen, the signing key is a new secret: name it in README.md and `.env.example`, never in git.
