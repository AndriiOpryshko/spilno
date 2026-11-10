---
id: T31
title: "Make a disabled association read-only in the middleware"
layer: "wiring"
deps: ["T19", "T30"]
acs: ["AC-21", "AC-23", "AC-25", "AC-26"]
files_hint:
  - "services/core/internal/identity/adapters/http/middleware.go"
  - "services/core/internal/identity/adapters/http/middleware_test.go"
  - "services/core/internal/identity/actor.go"
  - "services/core/internal/identity/adapters/postgres/"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T31 — Make a disabled association read-only in the middleware

## Why

[ADR-0007](../adr/0007-enforce-a-disabled-association-as-read-only-in-the-identity-middleware.md): one place, an allow-list, a new operation refused by default; [sad](../sad.md) §6 Flow 4, §10 QG-2; [api-sync-report](../contracts/api-sync-report.md) §E «Disabled association».

**Satisfies:** [PRD §5](../PRD.md) AC-21, AC-23, AC-25, AC-26 · **Blocked by:** [T19](./T19-association-and-operator-guards.md), [T30](./T30-disable-enable-association.md)

## What

- The guard of T19 reads the association's state inside the request's transaction; `disabled` sets `read_only` on the actor and applies the allow-list of T11 by operation id; the refusal carries the reason of the latest disabling in `detail`.
- `identity.access.denied` gains the reason `disabled_write`.

## Definition of Done

- [ ] Acceptance: A member of a disabled association passes the guard only on searchRegistry, getSignedProtocol and downloadFile, every other association-scoped operation of the contract answers identity.association_disabled with the reason, and this holds from the first request after the disabling.
- [ ] a test walks every operation of the spec with `{association_id}` as a member of a disabled association: the three allow-listed pass the guard, each other answers 403 `identity.association_disabled` with the reason
- [ ] a person off that roster still gets `identity.not_current_member` (AC-07 comes first)
- [ ] a session opened before the disabling is refused on its first request after it; after the enabling it is served again
- [ ] `mise run lint` and `mise run test` green

## Notes

governance must still check, when `read_only` is set, that a downloaded file is pinned by a signed snapshot (ADR-0007) — that check belongs to board-protocol's `downloadFile` handler (T21 there).
