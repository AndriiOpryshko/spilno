---
id: T18
title: "Sign in any verified person, keep the session in Postgres and tell the оператор"
layer: "app"
deps: ["T02", "T03", "T06"]
acs: ["AC-01", "AC-05", "AC-16"]
files_hint:
  - "services/core/internal/identity/domain/account.go"
  - "services/core/internal/identity/app/"
  - "services/core/internal/identity/ports/ports.go"
  - "services/core/internal/identity/adapters/postgres/"
  - "services/core/internal/identity/adapters/oidc/"
  - "services/core/internal/identity/adapters/http/"
  - "services/core/internal/identity/module.go"
  - "services/core/internal/platform/config/config.go"
  - "services/core/cmd/api/main.go"
owner: "Andrii Opryshko"
estimate: "L"
status: "todo"
---

# T18 — Sign in any verified person, keep the session in Postgres and tell the оператор

## Why

[ADR-0004](../adr/0004-let-any-verified-person-sign-in-and-derive-memberships-from-rosters.md): any verified person holds a session; an applicant has a session and no association. [ADR-0001](../adr/0001-serve-the-operator-back-office-as-an-operator-only-section-of-the-web-app.md), [ADR-0009](../adr/0009-treat-the-operator-role-and-a-roster-membership-as-independent-grants-on-one-account.md): the оператор is the session address equal to `SPILNO_OPERATOR_EMAIL`. Replaces [board-protocol T08](../../board-protocol/tasks/T08-identity-oidc-sessions.md) on the changed rule ([sad](../sad.md) §11); board-protocol ADR-0003 still gives the flow, cookies and sessions.

**Satisfies:** [PRD §5](../PRD.md) AC-01, AC-05, AC-16 · **Blocked by:** [T02](./T02-merge-contract-delta.md), [T03](./T03-identity-problem-codes.md), [T06](./T06-promote-drop-accounts-association.md)

## What

- Accounts (`oidc_subject`, `email`, no association) and sessions repositories with sqlc queries; issue, look up and revoke a session.
- `startSignIn`, `completeSignIn` with state and nonce, `signOut`, `getSession` → `Principal{account_id, email, is_operator}`; an address the provider does not mark verified is refused and creates no account.
- The provider behind a port, with an in-process fake; a development sign-in shim that only a dev or e2e configuration turns on.
- Configuration `SPILNO_OPERATOR_EMAIL` and the OIDC client.

## Definition of Done

- [ ] Acceptance: Against a fake sign-in provider a verified person off every roster gets an account and a session, getSession returns the principal with is_operator true only for SPILNO_OPERATOR_EMAIL, and an unverified address gets nothing.
- [ ] integration test: the callback with a verified address creates an account and an HttpOnly session cookie backed by a `sessions` row; `getSession` names the person; `signOut` revokes it
- [ ] no session → 401 `identity.session_required`; an unverified address → refused, `accounts` unchanged
- [ ] `is_operator` is true for the configured address in any letter case and false for any other
- [ ] a test asserts the shim is off in the production configuration; the four operations are no longer stubs
- [ ] `mise run lint` and `mise run test` green

## Notes

`bootstrap-member` of board-protocol T08 is not built: the first член правління of a new association comes from an approval (T26). Logs never carry the address together with a token.
