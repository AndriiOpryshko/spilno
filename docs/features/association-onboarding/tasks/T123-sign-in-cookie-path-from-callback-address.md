---
id: T123
title: "Scope the pending sign-in cookie by the path of the callback address"
layer: "app"
deps: ["T122"]
acs: ["AC-05"]
files_hint:
  - "services/core/internal/identity/app/signin.go"
  - "services/core/internal/identity/signin.go"
  - "services/core/internal/identity/signin_test.go"
  - "services/core/cmd/api/main.go"
owner: "Andrii Opryshko"
estimate: "S"
status: "done"
---

# T123 — Scope the pending sign-in cookie by the path of the callback address

## Why

[Second re-review of 2026-11-17](../_review/review-2026-11-17-2.md), S05. The cookie defaults to `/identity/sign-in/callback`, while the browser reaches the callback under `/api`; only the development shim passes the path. A real provider wired with the defaults would have every sign-in refused, and the path and the redirect address are two hand-kept copies of one fact.

**Satisfies:** [PRD §5](../PRD.md) AC-05 · **Blocked by:** [T122](./T122-bound-sign-in-replay-set.md)

## What

- The cookie's path is taken from the provider's redirect address; `CallbackPath` and its default go.
- `cmd/api` states the address once.

## Definition of Done

- [ ] Acceptance: A test starts a sign-in with a provider whose redirect address is `https://app.example.test/api/identity/sign-in/callback` and reads a cookie scoped to `/api/identity/sign-in/callback`, with no path passed beside the address.
- [ ] `signin_test.go`: the cookie path follows the redirect address, with and without a prefix
- [ ] the e2e sign-in still passes (T133 runs the suite)
- [ ] `mise run lint` and `mise run test` green
