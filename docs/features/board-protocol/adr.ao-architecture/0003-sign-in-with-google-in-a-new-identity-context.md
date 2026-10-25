---
status: superseded
superseded_by: "0022"
date: 2026-10-25
feature: board-protocol
sad_section: "§4"
deciders: ["Andrii Opryshko"]
---

# 0003 Sign in with Google in a new identity context

> **Status:** superseded by [0022](./0022-sign-in-with-google-and-admit-only-current-board-members-from-the-governance-roster.md) (2026-10-25): no invitations — the verified Google e-mail is matched against the governance roster, and a membership check before any resource answers 403 `governance.not_board_member`. OIDC, sessions and the association record are restated there unchanged.

## Context and problem statement

Every right in PRD §10 belongs to "a член правління with a term in force", and AC-32 / AC-33 refuse people whose term ended and members of another association. The app has no sign-in and no association today (SAD §2); PRD §8 asks architecture to decide it as the Stage 0 prerequisite. The decision also fixes who the Google account behind Drive and Docs is (ADR 0004, 0018).

## Decision drivers

- QG-3: confidentiality within one association; "not found" for other associations (AC-33).
- PRD §6.1: first feature with authorization and personal data; security review required.
- One developer: no password storage, no mail delivery to run.

## Blast radius

- Irreversible: yes — accounts, sessions and the actor contract are used by every context.
- Multi-context: yes — `identity` gives the actor to `governance`, `finance` and the web client.
- Legitimate alternative: yes — an emailed one-time link, or a separate Stage 0 feature.
- Touches the database: yes — new tables for associations, accounts, invitations, sessions.

## Considered options

1. **Google sign-in (OIDC)** in a new `identity` context, server-side sessions.
2. **Emailed one-time link** in a new `identity` context, server-side sessions; needs an SMTP relay.
3. **Defer to a separate Stage 0 feature**, fixing only the actor contract here.

## Decision outcome

**Chosen:** option 1. OIDC (the standard "sign in with another account" protocol) with Google, PKCE, scopes `openid email profile`. A session row in Postgres and an HTTP-only, Secure, SameSite=Lax cookie; sessions can be revoked. A person may sign in only after an invitation to their e-mail for one association. For local development only, a sign-in shim behind a dev configuration flag replaces Google; a test asserts it is off in the production configuration (SAD §11). `identity/module.go` provides `RequireSession` middleware and `ActorFrom(ctx)` → (account id, association id, e-mail). `governance/app` maps the account to a board term and checks the term in force on the action date. `identity` also owns the association record: name, annual-meeting date setting (SAD §2) and the encrypted connection to the association's Google account (ADR 0004).

### Consequences

- Good: no passwords or mail infrastructure; Google 2-step verification protects accounts.
- Bad: a member without a Google account cannot sign in; one Google Cloud project with an OAuth consent screen has to be maintained.
- Neutral: new context wiring per CLAUDE.md (tag, oapi-codegen config, sqlc entry, `identity-private` depguard rule, `cmd/api/main.go`).

### Migration and rollback

- Migration: new tables `associations`, `accounts`, `invitations`, `sessions`; a seed creates the single association (ADR 0016).
- Rollback: drop the new tables while no protocol exists; afterwards accounts stay because protocols record who acted.

## Links

- PRD: [PRD.ao-prd.md](../PRD.ao-prd.md) — §8 sign-in OQ, §10 roles, AC-32, AC-33
- SAD: [sad.ao-architecture.md](../sad.ao-architecture.md) §4
- Related ADR: [0016 Single association](./0016-scope-governance-rows-to-the-single-association-with-an-explicit-filter-in-every-query.md) · repo [0002 Bounded contexts](../../../adr/0002-bounded-contexts.md)
