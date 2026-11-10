---
id: T33
title: "Connect the association's document storage from the web app"
layer: "ports"
deps: ["T14", "T19", "T20"]
acs: ["AC-27", "AC-28b"]
files_hint:
  - "services/core/internal/identity/app/"
  - "services/core/internal/identity/ports/ports.go"
  - "services/core/internal/identity/adapters/postgres/"
  - "services/core/internal/identity/adapters/http/"
  - "services/core/internal/identity/google.go"
  - "services/core/internal/identity/module.go"
  - "services/core/internal/platform/google/"
owner: "Andrii Opryshko"
estimate: "L"
status: "todo"
---

# T33 — Connect the association's document storage from the web app

## Why

[ADR-0008](../adr/0008-let-a-board-member-connect-the-association-storage-in-the-web-app.md): the board connects the storage itself; replaces the developer command of board-protocol T12. [sad](../sad.md) §6 Flow 10. Reuses what board-protocol built: the token cipher (T12a), the connection store (T12b), the consent exchange (T12c).

**Satisfies:** [PRD §5](../PRD.md) AC-27, AC-28b · **Blocked by:** [T14](./T14-roster-read-facade.md), [T19](./T19-association-and-operator-guards.md), [T20](./T20-promote-row-level-security.md)

## What

- `startStorageConnection`: 302 to `Consent.AuthURL` with a single-use `state` kept server-side for the session and the association.
- `completeStorageConnection`: check the state → `Consent.Complete(code)` → `Membership.MemberSignInAddresses` → an address equal to one of them is `identity.storage_account_personal` → otherwise `GoogleConnectionRepository.Replace` with the sealed token, the address and the acting member's account.
- `getStorageConnection`: connected or not, the address, who connected it. The provider is behind its port with an in-process fake.

## Definition of Done

- [ ] Acceptance: A current член правління starts the consent, completes it with the code and a state bound to their session and association, and the connection is stored with the account address and the member — unless the address is a member's sign-in address, which is refused and stores nothing.
- [ ] AC-27: the connection is stored with the address and the member; every current member of the association reads it
- [ ] AC-28b: a member's sign-in address in any letter case → 409, no row
- [ ] an unknown, foreign, expired or replayed `state`, and a code the provider refuses → nothing stored
- [ ] a later connection replaces the row and moves no file
- [ ] `mise run lint` and `mise run test` green

## Notes

Needs [board-protocol T12d](../../board-protocol/tasks/T12d-google-config.md) (Google configuration from the environment, `EnableGoogle`), in flight on `feat/core-google-config`. Open in api-sync-report §D: PRD §5.1 has no code for a bad `state` or a refused code — answer 422 `about:blank` naming `/state` or `/code` and invent no code; revisit after `/sdlc:clarify-prd`. Open in PRD §8: who may replace a connected storage. A disabled association refuses these operations through T31.
