---
status: Accepted
owner: "Andrii Opryshko"
reviewers: []
updated_at: "2026-10-25"
feature_size: L
stage: "04-05"
ticket: ""
---

# 0003 — Sign in through OIDC (Google first) with server-side sessions in Postgres

- **Status:** Accepted
- **Date:** 2026-10-25
- **Deciders:** Andrii Opryshko (Architect), with Claude during the architecture-design pass

> **Amended 2026-10-25 (Stage 1 rescope):** AC-03b (only the головуючий edits) was dropped — any current член правління edits a draft (AC-29); the `identity` context is fixed by [ADR-0014](./0014-keep-board-protocol-in-governance-and-add-only-an-identity-context.md); the membership question goes to `governance.CurrentMember` on today's Europe/Kyiv date. The decision is unchanged.

## Context

The app has no sign-in today (PRD §6.1, §8: "How does a член правління sign in … decided in architecture as a prerequisite"). The feature needs to know who casts each попереднє голосування (AC-05e), who edits a draft (AC-29 — any current член правління, author recorded), who uploads the signed scan and locks a protocol (AC-16), and must deny a person whose term has ended without revealing that records exist (AC-19). The board is 5-7 volunteers; the association's documents already live in a Google account (ADR-0005).

## Decision drivers

- QG-3: only current члени правління see cases, drafts and the registry; access ends on the term end date.
- PRD §6.1: authors of every edit, vote, print, upload, withdrawal and settings change are recorded.
- §3: no email-delivery system in the context — sign-in must not require one.
- One developer; security review required — fewer secrets kept by us is better.

## Considered options

1. **OIDC sign-in (Google first, provider configurable) + server-side sessions in Postgres** — the provider proves who the person is; spilno links the verified email to a roster person and keeps the session.
2. **Email magic link + sessions** — a one-time link sent by email; needs an email-delivery service.
3. **Passwords (argon2id) + sessions** — spilno stores credentials; resets without email only by another member.

## Decision outcome

**Chosen:** Option 1. No passwords or email delivery to operate, and the provider never decides membership: on every request the session's person is checked against the roster for a term covering today, so a term end cuts access immediately even if the Google session lives on.

Shape: authorization-code flow with PKCE in `identity`; an account row links a provider subject and verified email to a roster person; a session row (random id, person, created, last seen, expires, revoked) behind an HttpOnly, Secure, SameSite=Lax cookie; sign-out and revocation delete the row. An HTTP middleware resolves the session and asks `governance` (through its `module.go`, `CurrentMember(ctx, email, date)`) whether the person is a current член правління on today's Europe/Kyiv date; if not, every context answers as if nothing exists (AC-19). An account whose email matches no roster person cannot sign in.

## Consequences

**Positive**
- No credentials stored by spilno; revocation is one row delete.
- Membership is evaluated per request from the roster, matching AC-16 and AC-19 exactly.
- The same identity provider later serves other associations by configuration.

**Negative**
- Every член правління needs a Google account (or another configured OIDC provider) whose email is recorded on their roster entry.
- Needs an OAuth client registered with the provider and an HTTPS origin in production; local dev needs a test provider or a dev-only login shim.
- Every existing and future endpoint and e2e test now runs behind sign-in; `system/health` stays public.

**Neutral**
- Adding magic-link sign-in later is a second `identity` flow producing the same session rows.

## Links

- PRD: [PRD.md](../PRD.md) §6.1, §8, AC-29, AC-05e, AC-16, AC-19
- SAD: [sad.md](../sad.md) §4, §8
- Related ADR: [ADR-0014](./0014-keep-board-protocol-in-governance-and-add-only-an-identity-context.md), [ADR-0005](./0005-store-files-in-association-owned-google-drive-folder.md)
