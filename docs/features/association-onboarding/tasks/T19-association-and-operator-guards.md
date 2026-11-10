---
id: T19
title: "Guard association paths with the session, the roster and a scoped transaction, and operator paths with the оператор flag"
layer: "wiring"
deps: ["T04", "T11", "T14", "T18"]
acs: ["AC-04d", "AC-07", "AC-16", "AC-16b"]
files_hint:
  - "services/core/internal/identity/adapters/http/middleware.go"
  - "services/core/internal/identity/adapters/http/middleware_test.go"
  - "services/core/internal/identity/actor.go"
  - "services/core/internal/identity/module.go"
  - "services/core/internal/identity/ports/ports.go"
  - "services/core/internal/identity/adapters/postgres/google_connections.go"
  - "services/core/internal/governance/module.go"
  - "services/core/internal/finance/module.go"
  - "services/core/cmd/api/main.go"
owner: "Andrii Opryshko"
estimate: "L"
status: "todo"
---

# T19 — Guard association paths with the session, the roster and a scoped transaction, and operator paths with the оператор flag

## Why

[sad](../sad.md) §6 Flow 4, §8 «Authorization», «Association scope»; [ADR-0002](../adr/0002-put-the-association-id-in-every-association-scoped-path.md), [ADR-0003](../adr/0003-enforce-association-isolation-with-postgres-row-level-security.md), [ADR-0009](../adr/0009-treat-the-operator-role-and-a-roster-membership-as-independent-grants-on-one-account.md). Supersedes [board-protocol T11](../../board-protocol/tasks/T11-require-session-current-member.md), which guarded unscoped paths.

**Satisfies:** [PRD §5](../PRD.md) AC-04d, AC-07, AC-16, AC-16b · **Blocked by:** [T04](./T04-in-association-transaction.md), [T11](./T11-association-state-access-matrix.md), [T14](./T14-roster-read-facade.md), [T18](./T18-sign-in-any-verified-person.md)

## What

- Middleware exported from `identity/module.go`: session cookie → session; on a path with `{association_id}`: `db.InAssociation(association, email)` → `Membership.CurrentMember` (today, Europe/Kyiv) → the decision of T11 → `identity.ActorFrom(ctx)`; on `/identity/operator/...`: the оператор flag; on `/identity/applications...` and `/identity/memberships`: the session only.
- `Membership` is a port of identity that governance implements; `cmd/api/main.go` passes the governance module in, so identity does not import governance (governance imports identity for `ActorFrom`).
- Mounted with `r.Group` in the governance, finance and identity modules; the `google_connections` repository takes its runner from `db.From(ctx)`.
- Counter `identity.access.denied` by reason (`not_member`, `not_operator`).

## Definition of Done

- [ ] Acceptance: Every association-scoped route of governance, finance and identity runs inside InAssociation after a roster check that answers a non-member and an unknown id with the same body, and every /identity/operator route refuses a person who is not the оператор.
- [ ] HTTP test on two associations: a member of A is served under A and gets 403 `identity.not_current_member` under B, byte-equal to the answer for an unknown association id
- [ ] an applicant with a pending заявка, a person with nothing and a member whose term ended yesterday get the same 403; no session → 401
- [ ] the оператор off the roster gets that 403 on an association path; on A's roster the оператор is served under A as a member and refused under B
- [ ] a person who is not the оператор gets 403 `identity.not_operator` on every `/identity/operator/...` operation of the spec
- [ ] inside a handler `ActorFrom` returns the member and `current_setting('app.association_id', true)` is the path id
- [ ] `mise run lint` and `mise run test` green

## Notes

Never `r.Mount` (CLAUDE.md gotcha). The disabled state is added by T31. board-protocol handlers written before this task read the association from the path; after it they read `ActorFrom`.
