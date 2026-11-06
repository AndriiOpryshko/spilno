---
status: accepted
date: 2026-11-05
feature: association-onboarding
sad_section: "§4"
deciders: ["Andrii Opryshko", "Claude (ao-architecture)"]
---

# 0002 Put the association id in every association-scoped path

## Context and problem statement

A person may be on the rosters of several associations (PRD §1 ¶4, AC-04d). Today paths are `/<context>/...` and the association comes from the account (one per account, board-protocol ADR-0015). Every request needs an unambiguous association, taken from neither the body nor a guess.

## Decision drivers

- PRD AC-04d, AC-07: records of each association only within that association and only to its roster.
- Repo ADR 0003 and `.claude/rules/openapi.md`: one contract, paths start with `/<context>/`.
- Product owner 2026-11-05: «хоч рут різний з айді осбб».

## Blast radius

- Irreversible: yes — every association-scoped operation, the generated client and the web routes change.
- Multi-context: yes — identity, governance, finance and the web app.
- Legitimate alternative: yes — the active association stored in the session.
- Touches the database: no.

## Considered options

1. **Path parameter after the context prefix: `/<context>/associations/{association_id}/...`; web routes `/o/{association_id}/...`.**
2. **Active association stored in the session**, switched by an identity operation; paths unchanged.
3. **A request header** carrying the association id.

## Decision outcome

**Chosen:** Option 1. The association is visible in every URL and every log line, tabs are independent, and the middleware reads one path parameter. Option 2 makes two tabs fight over one session; option 3 hides the scope from the URL.

### Consequences

- Good: links to a protocol carry their association; the isolation test can iterate every operation with `{association_id}`.
- Bad: ~30 existing `governance` operations and the `finance` contracts operations move; board-protocol tasks building against the old paths must follow (SAD §11).
- Neutral: unscoped operations stay as they are: `/identity/session`, `/identity/sign-in*`, `/identity/memberships`, `/identity/applications...`, `/identity/operator/...`, `/system/health`.

### Migration and rollback

N/A — no stored data changes. The contract change and its generated code land in one commit (CLAUDE.md); no external clients exist, so no old paths are kept.

## Links

- PRD: [PRD.md](../PRD.md) — AC-04d, AC-07
- SAD: [sad.md](../sad.md) §4, §8
- Related ADR: [0003](./0003-enforce-association-isolation-with-postgres-row-level-security.md); repo [0003 Spec-first codegen](../../../adr/0003-spec-first-codegen.md)
