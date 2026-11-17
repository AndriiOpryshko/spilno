---
id: T99
title: "Keep the pending sign-in in a sealed cookie of the browser that started it"
layer: "app"
deps: []
acs: ["AC-05", "AC-27"]
files_hint:
  - "services/core/internal/identity/app/signin.go"
  - "services/core/internal/identity/app/storage.go"
  - "services/core/internal/identity/adapters/http/handler.go"
  - "services/core/internal/identity/signin.go"
  - "services/core/internal/identity/signin_test.go"
  - "services/core/internal/identity/state_bounded_test.go"
  - "tests/load/client.go"
  - "apps/web/e2e/support.ts"
  - "docs/features/association-onboarding/sad.md"
  - "docs/features/association-onboarding/tasks/tracker.md"
owner: "Andrii Opryshko"
estimate: "L"
status: "done"
---

# T99 — Keep the pending sign-in in a sealed cookie of the browser that started it

## Why

[Review of 2026-11-17](../_review/review-2026-11-17.md), Q03. The cap of T55 bounds memory, but about a thousand requests to the public `GET /identity/sign-in` push out the state of everyone who is signing in. [Security review](../_review/security-2026-11-10.md) SEC-02 asks for the same change for another reason: the `state` is not bound to the browser.

**Satisfies:** [PRD §5](../PRD.md) AC-05, AC-27 · **Blocked by:** —

## What

- `startSignIn` seals the state, the PKCE verifier, the nonce and the expiry into an `HttpOnly`, `SameSite=Lax` cookie scoped to the callback path; the server keeps nothing. The key is random per process, as the in-memory map was.
- `completeSignIn` opens the cookie, compares the `state`, clears the cookie. A replay is refused because the cookie is gone.
- The storage consent keeps its server-side state (it needs a session) with one pending entry per account.
- SAD §8 and the tracker's known gap «sign-in state lives in process memory» are reworded; the load client and the e2e support keep the cookie between start and callback.

## Definition of Done

- [ ] Acceptance: Tests show that any number of sign-in starts leaves a sign-in already in progress able to finish, a callback without the cookie of its `state` is refused, a forged, expired or replayed state is refused, and one account holds at most one pending storage consent.
- [ ] `signin_test.go`: no cookie, another browser's cookie, expired, replayed, tampered
- [ ] `state_bounded_test.go`: a flood of starts, then a callback of an earlier start succeeds
- [ ] cookie flags asserted: `HttpOnly`, `SameSite=Lax`, `Secure` outside the development shim, path
- [ ] `mise run e2e` and `mise run load` still sign in
- [ ] `mise run lint` and `mise run test` green

## Notes

The alternative, a rate limit per client address, needs a trusted proxy header the deployment does not define yet. More than one API instance needs a shared key, the same limit the in-memory map had.
