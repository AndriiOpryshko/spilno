---
status: accepted
date: 2026-10-25
feature: board-protocol
sad_section: "§4"
superseded_by: ""
deciders: ["Andrii Opryshko"]
---

# 0022 Sign in with Google and admit only a person the governance roster names as a current член правління

## Context and problem statement

Every right in PRD §10 belongs to "a член правління with a term in force"; AC-32 refuses a person whose term ended and AC-33 answers "not found" for an identifier not stored for the association. ADR 0003 chose Google sign-in in a new `identity` context and admitted a person only from a per-e-mail admission list kept by `identity`, a second list of people next to the board roster that `governance` already keeps (ADR 0020). The cross-check with the other SAD (`sad.md`, its ADR 0003) agreed on one list: the roster. A person is admitted when the verified e-mail of their Google account is the sign-in e-mail of a current член правління in the roster. This flips ADR 0003's admission model, so this ADR supersedes it; the rest of ADR 0003 (OIDC with Google, server-side sessions, the association record and its Google connection in `identity`) is restated here unchanged.

## Decision drivers

- QG-3: only a current член правління acts on or reads the records; a refused person learns nothing about what exists (AC-32, PRD §6.1 "Former member access").
- One source of who is on the board: the roster with terms already drives the members slot and authorization (AC-01, AC-08, ADR 0020); an admission list would be a second one to keep in sync.
- One developer: no passwords, no mail delivery, no second list to keep in sync.
- Repo ADR 0002: contexts talk only through `module.go`, no import cycle.

## Blast radius

- Irreversible: yes — the session, the actor contract and the admission rule are used by every protected route; accounts are referenced by every recorded action.
- Multi-context: yes — `identity` (sessions, middleware), `governance` (roster with sign-in e-mail, the membership query), `cmd/api` wiring, and the web sign-in page.
- Legitimate alternative: yes — a per-e-mail admission list in `identity` (ADR 0003), an emailed one-time link, or checking membership only inside `governance/app` per action.
- Touches the database: yes — `associations`, `accounts`, `sessions`; `board_terms.sign_in_email` instead of `account_id`; ADR 0003's admission-list table is not created.

## Considered options

1. **Roster match** — the verified Google e-mail is matched against the governance roster; an `identity` middleware asks `governance` on every request whether that e-mail is a current член правління today.
2. **Admission list** (ADR 0003) — `identity` keeps admitted e-mails per association; `governance/app` checks the term per action.
3. **Check only inside `governance/app`** — any Google account gets a session; each use case refuses a non-member.

## Decision outcome

**Chosen:** option 1.

- OIDC with Google, PKCE, scopes `openid email profile`; only an ID token with `email_verified = true` is accepted. A session row in Postgres and an HTTP-only, Secure, SameSite=Lax cookie; sessions can be revoked. For local development only, a sign-in shim behind a dev configuration flag replaces Google; a test asserts it is off in the production configuration (SAD §11).
- The roster is `governance`'s: each `board_terms` row carries the member's `sign_in_email` (lower-cased), recorded with the term (ADR 0020, amended). "Current член правління" is a term with `term_from ≤ today ≤ term_to` in Europe/Kyiv that no later row supersedes.
- `governance/module.go` exposes `CurrentMember(ctx, association, email, date) (Member, bool, error)`. `identity/ports` declares a `MembershipChecker` interface with that method, and `cmd/api` passes the governance module to `identity.New(pool, governance)`; `identity` never imports `governance`, so the existing `governance → identity` dependency (`ActorFrom`, `GoogleTokenSource`) stays acyclic.
- At the OIDC callback, an e-mail that is not a current член правління gets no session; the sign-in page shows `governance.not_board_member`. On every request to a protected group, `identity.RequireMember` (session + membership check) runs **before** routing reaches any handler: no session → 401 `identity.session_required`; a session whose e-mail is no longer a current член правління today → 403 `governance.not_board_member`, the same answer for every path and identifier, so a refused person never learns whether a protocol or справа exists. The session is not deleted, so it starts working again if a new term is recorded.
- `ActorFrom(ctx)` → (account id, association id, e-mail, roster person id). Inside `governance`, a member who passed the gate and opens an identifier not stored for the association gets 404 `governance.protocol_not_found` (AC-33, ADR 0016); with one association there is no foreign association whose ids could leak.
- `identity` also owns the association record: name, annual-meeting date setting (SAD §2) and the encrypted connection to the association's Google account (ADR 0004).
- Bootstrap: the setup command that creates the association (ADR 0016) records the first член правління's term with their sign-in e-mail; every later member is added through the board settings (US-01).

### Consequences

- Good: one list of who may act; ending a term in the roster cuts access on the next request, with nothing to revoke elsewhere.
- Good: refusal happens before any resource is read, so AC-32 holds for every operation and existence is never revealed to a non-member.
- Bad: a roster lookup on every request (one indexed query on `board_terms (association_id, sign_in_email)`); a member who changes their Google e-mail is locked out until the roster is corrected.
- Bad: a член правління without a Google account cannot sign in (SAD §11).
- Neutral: new context wiring per CLAUDE.md (tag, oapi-codegen config, sqlc entry, `identity-private` depguard rule, `cmd/api/main.go`).

### Migration and rollback

- Migration: new tables `associations`, `accounts` (Google subject, e-mail), `sessions`; no admission-list table; `board_terms.sign_in_email text NOT NULL` with an index on `(association_id, sign_in_email)` (ADR 0020). A seed creates the single association and its first roster entry (ADR 0016).
- Rollback: drop the identity tables while no protocol exists; afterwards accounts stay because protocols record who acted. Returning to an admission list later is additive (a new table checked before the roster), not a data migration.

## Links

- PRD: [PRD.ao-prd.md](../PRD.ao-prd.md) — §8 sign-in OQ, §10 roles, AC-01, AC-32, AC-33; §5.1 `identity.session_required`, `governance.not_board_member`
- SAD: [sad.ao-architecture.md](../sad.ao-architecture.md) §4, §5, §6 Flow 10, §8 Authorization
- Related ADR: supersedes [0003 Sign in with Google](./0003-sign-in-with-google-in-a-new-identity-context.md) · [0020 Board terms](./0020-keep-board-terms-and-building-shares-as-dated-rows-and-charter-rules-as-current-settings.md) · [0016 Single association](./0016-scope-governance-rows-to-the-single-association-with-an-explicit-filter-in-every-query.md) · idea from [other SAD ADR 0003](../adr/0003-sign-in-with-oidc-and-postgres-sessions.md)
