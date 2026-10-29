---
id: T30
title: "Withdraw a protocol with a reason and copy a withdrawn draft"
layer: "ports"
deps: ["T29"]
acs: ["AC-35", "AC-36", "AC-37", "AC-38", "AC-39"]
files_hint:
  - "services/core/internal/governance/app/withdrawal.go"
  - "services/core/internal/governance/adapters/http/withdrawal.go"
  - "services/core/internal/governance/adapters/http/copy.go"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T30 — Withdraw a protocol with a reason and copy a withdrawn draft

## Why

[sad §6 Flow 9](../sad.md); [ADR-0007](../adr/0007-lock-protocols-as-fingerprinted-immutable-snapshots.md), [ADR-0017](../adr/0017-reserve-the-protocol-number-at-first-print-and-never-reuse-it.md).

**Satisfies:** [PRD §5](../PRD.md) AC-35, AC-36, AC-37, AC-38, AC-39 · **Blocked by:** [T29](./T29-lock-with-signed-copy.md)

## What

- Handlers for withdrawProtocol and copyWithdrawnProtocol; enqueue apply_drive_restrictions for withdrawn files.

## Definition of Done

- [ ] Acceptance: Handler tests show withdrawProtocol from чернетка, готовий до голосування or готовий до підпису writes a `withdrawn` snapshot with the reason and keeps справи, versions, rounds and any number (AC-35–37), refuses a locked protocol (AC-38) and a missing reason (`governance.reason_required`), and copyWithdrawnProtocol starts a new draft from a withdrawn one (AC-39).
- [ ] test: a withdrawn protocol with No 14 leaves a gap marked «знято з розгляду» in the register
- [ ] `mise run lint` and `mise run test` green