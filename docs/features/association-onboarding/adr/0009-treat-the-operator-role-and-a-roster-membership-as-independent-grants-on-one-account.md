---
status: accepted
date: 2026-11-10
feature: association-onboarding
sad_section: "§4"
deciders: ["Andrii Opryshko (product owner, clarify-prd 2026-11-10)", "Claude (ao-architecture, easy depth, reconcile pass)"]
---

# 0009 Treat the operator role and a roster membership as independent grants on one account

## Context and problem statement

The product owner is the one оператор and also heads the association set up by hand. [ADR-0001](./0001-serve-the-operator-back-office-as-an-operator-only-section-of-the-web-app.md) recognises the оператор by the session address equal to `SPILNO_OPERATOR_EMAIL` and gave the оператор «no association path at all»; SAD §8 refused the оператор on every association path. With one sign-in address that locks the оператор out of their own association, and letting both roles through without a rule lets the оператор approve their own запит на передачу. The PRD clarification of 2026-11-10 settled it: both roles may sit on one account, neither widens the other, and the оператор decides nothing they filed (PRD AC-16b, AC-16c, §10).

## Decision drivers

- PRD AC-16b: the оператор role opens no association's records; an оператор who is a current член правління works there as a member and nothing more.
- PRD AC-16c, §10 refused transitions: nobody decides their own proof of office.
- PRD §3: one оператор, no operator roles; PRD §6.1 abuse case «takeover of the оператор's account».
- SAD §1 quality goal 1: no path reads an association's rows without a roster entry in force.

## Blast radius

- Irreversible: no — moving to a dedicated operator address later is a configuration change plus removing one guard.
- Multi-context: yes — web and core: the identity middleware and the operator use cases, and the SPA shell that shows the association picker and the back office to the same person.
- Legitimate alternative: yes — a dedicated operator address that stands on no roster.
- Touches the database: no — applications already record the applicant's account and decisions record the оператор.

## Considered options

1. **Two independent grants on one account.** The association middleware asks only `governance.CurrentMember` and never looks at the operator role; operator paths ask only whether the session address equals `SPILNO_OPERATOR_EMAIL`; approve and reject refuse an application whose applicant account is the acting оператор's with `identity.operator_own_application`.
2. **A dedicated operator address** that stands on no roster and files nothing; an оператор who is also a член правління enters their association with another address.
3. **Both roles on one account without the own-filing rule.**

## Decision outcome

**Chosen:** Option 1 (product owner, 2026-11-10: «одна адреса, дві ролі»). One person keeps one account; isolation still rests on the roster alone, so the operator role cannot turn into a bypass; the own-filing refusal keeps the check of a proof of office independent of the person it confirms. Option 2 needs a second account and a rule the system would have to police on every roster edit and every filing. Option 3 lets the оператор confirm themself.

Shape:
- Association paths: session → `CurrentMember(association, email, today)` → association state → actor. An оператор off the roster gets `identity.not_current_member` like anyone, with the same body for a missing id.
- Operator paths (`/identity/operator/...`): the session address equals `SPILNO_OPERATOR_EMAIL`, else `identity.not_operator`. These use cases read applications, extracts and the list of associations, never an association-owned table, except the storage account address of [ADR-0003](./0003-enforce-association-isolation-with-postgres-row-level-security.md).
- Approve and reject compare the application's applicant account with the acting account and refuse a match; the filing stays очікує. Recording the check is not refused: it leads to no decision.
- Web: after sign-in the same person sees the association picker and the `/operator/` section.

### Consequences

- Good: no second account; the isolation test needs no operator special case — the оператор's address is seeded as an ordinary person (SAD §10 QG-1).
- Bad: with a single оператор their own filing is never decided, so the association set up by hand keeps «підключено вручну, без витягу» while the оператор heads it (AC-19c, PRD §7); one taken-over account holds both grants (SAD §11).
- Neutral: supersedes in part ADR-0001 (the sentence «the оператор gets no association path at all») and retires the denial reason `operator_on_association` of SAD §7; a second оператор later lifts the dead end without changing this rule.

### Migration and rollback

N/A — no stored data changes.

## Links

- PRD: [PRD.md](../PRD.md) — AC-16b, AC-16c, AC-19c, §10 Roles and actions, §6.1
- SAD: [sad.md](../sad.md) §4, §8
- Related ADR: [0001](./0001-serve-the-operator-back-office-as-an-operator-only-section-of-the-web-app.md) (superseded in part); [0004](./0004-let-any-verified-person-sign-in-and-derive-memberships-from-rosters.md)
