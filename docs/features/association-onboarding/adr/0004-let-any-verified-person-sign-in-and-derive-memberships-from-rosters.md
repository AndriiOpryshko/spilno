---
status: accepted
date: 2026-11-05
feature: association-onboarding
sad_section: "§4"
deciders: ["Andrii Opryshko", "Claude (ao-architecture)"]
---

# 0004 Let any verified person sign in and derive memberships from rosters

## Context and problem statement

Board-protocol [ADR-0003](../../board-protocol/adr/0003-sign-in-with-oidc-and-postgres-sessions.md) refuses sign-in to an account whose address matches no roster person, and `accounts.association_id` is `NOT NULL`. An applicant has no association yet (PRD AC-01, AC-05), and a person may sit on several rosters (AC-04d).

## Decision drivers

- PRD AC-01, AC-05, AC-07: an applicant signs in and sees only their own заявка.
- PRD AC-04d: one person, several associations.
- PRD §6: a term end or a transfer is in force from the next request.

## Blast radius

- Irreversible: yes — the account shape and the sign-in rule.
- Multi-context: yes — identity and governance.
- Legitimate alternative: yes — a separate «applicant» account type and a membership table owned by identity.
- Touches the database: yes — `accounts.association_id` is dropped.

## Considered options

1. **Any verified OIDC address gets an account and a session; associations come from `governance.MembershipsOf(email, today)`; per request `CurrentMember(association, email, today)`.**
2. **Applicant accounts separate from member accounts**, plus an identity-owned membership table synced from the roster.

## Decision outcome

**Chosen:** Option 1. The roster stays the single source of membership, as in board-protocol ADR-0003; only the «no roster, no sign-in» rule is dropped. Option 2 keeps a second copy of membership that can drift.

### Consequences

- Good: no sync; the previous голова loses access on the first request after the term end.
- Bad: anyone with a Google account can create an account row; harmless, it reaches nothing but `/identity/applications`.
- Neutral: the OIDC flow, cookies and sessions of board-protocol ADR-0003 are unchanged.

### Migration and rollback

- Migration: release N stops reading `accounts.association_id`; release N+1 drops the column and its index (3-step rule for a dropped column).
- Rollback: before the drop, revert the code; after the drop, re-add the column nullable and backfill from the roster of the one association.

## Links

- PRD: [PRD.md](../PRD.md) — AC-01, AC-04d, AC-05, AC-07
- SAD: [sad.md](../sad.md) §4, §8
- Related ADR: board-protocol [ADR-0003](../../board-protocol/adr/0003-sign-in-with-oidc-and-postgres-sessions.md); [0002](./0002-put-the-association-id-in-every-association-scoped-path.md)
