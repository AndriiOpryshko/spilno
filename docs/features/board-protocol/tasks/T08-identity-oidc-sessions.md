---
id: T08
title: "Sign in with Google OIDC and keep sessions in Postgres"
layer: "app"
deps: ["T01", "T06", "T07"]
acs: ["AC-19", "AC-16"]
files_hint:
  - "services/core/internal/identity/domain/"
  - "services/core/internal/identity/app/"
  - "services/core/internal/identity/ports/"
  - "services/core/internal/identity/adapters/postgres/"
  - "services/core/internal/identity/adapters/oidc/"
  - "services/core/internal/identity/adapters/http/handlers.go"
  - "services/core/internal/platform/config/"
owner: "Andrii Opryshko"
estimate: "L"
status: "todo"
---

# T08 — Sign in with Google OIDC and keep sessions in Postgres

## Why

[ADR-0003](../adr/0003-sign-in-with-oidc-and-postgres-sessions.md): verified Google email → account; the session is a Postgres row; membership is decided by governance on every request (T11). [sad §6 Flow 7](../sad.md).

**Satisfies:** [PRD §5](../PRD.md) AC-19, AC-16 · **Blocked by:** [T01](./T01-stage-identity-migrations.md), [T06](./T06-problem-codes-registry.md), [T07](./T07-merge-contract-identity-skeleton.md)

## What

- Accounts and sessions repositories (sqlc queries), session issue / lookup / revoke.
- OIDC flow with state + nonce; `identity.session_required` / `identity.not_current_member` in the identity error registry.
- Dev sign-in shim enabled only by dev config.

## Definition of Done

- [ ] Acceptance: startSignIn / completeSignIn / signOut / getSession work against a fake OIDC provider in integration tests, the session is an HttpOnly cookie backed by a `sessions` row, and a test asserts the dev sign-in shim is off in the production config.
- [ ] integration test: callback with a verified email creates an account and a session; sign-out revokes it
- [ ] test: production config has the shim disabled
- [ ] contract operations `startSignIn`, `completeSignIn`, `signOut`, `getSession` no longer stubbed
- [ ] `mise run lint` and `mise run test` green

## Notes

No PII in fixtures: `user-<uuid>@example.test`.
