---
id: T41
title: "Wire printing and withdrawal: ready for signing, job polling, void warning, withdraw and copy"
layer: "ui"
deps: ["T40"]
acs: ["AC-14c", "AC-33", "AC-14b", "AC-04", "AC-21", "AC-35", "AC-36", "AC-37", "AC-39", "AC-38", "AC-17"]
files_hint:
  - "apps/web/src/features/governance/readiness-card.tsx"
  - "apps/web/src/features/governance/transition-actions.tsx"
  - "apps/web/src/features/governance/withdraw-dialog.tsx"
  - "apps/web/src/features/governance/protocol-screen.tsx"
owner: "Andrii Opryshko"
estimate: "L"
status: "todo"
---

# T41 — Wire printing and withdrawal: ready for signing, job polling, void warning, withdraw and copy

## Why

[sad §6](../sad.md) Flows 4 and 9; [ADR-0017](../adr/0017-reserve-the-protocol-number-at-first-print-and-never-reuse-it.md).

**Satisfies:** [PRD §5](../PRD.md) AC-14c, AC-33, AC-14b, AC-04, AC-21, AC-35, AC-36, AC-37, AC-39, AC-38, AC-17 · **Blocked by:** [T40](./T40-ui-live-poll.md)

## What

- ReadinessCard on markProtocolReadyForSigning + getJob; WithdrawDialog on withdrawProtocol; copy on copyWithdrawnProtocol.
- Replace `window.print()` with the rendered PDF download.
- Locked state (review finding 6): hide withdraw and edit actions; render `protocol_locked` refusals as the explanation and the «нове рішення правління» offer (AC-17, AC-38).

## Definition of Done

- [ ] Acceptance: Component tests show the print action lists every blocker at once (AC-04, AC-21), otherwise polls the render job and offers the PDF with its number, a return to draft warns that earlier copies must not be signed (AC-14b), withdraw requires a reason, and a withdrawn draft offers «copy as new draft» (AC-39). On a locked protocol the screen offers no withdraw or edit action; a refused withdraw shows the offer of a new board decision that revokes or corrects it (AC-38), and a refused change shows that a locked protocol changes only through a new referencing decision (AC-17) (review finding 6).
- [ ] stories for ready, rendering, voided, withdrawn
- [ ] component tests: signed protocol → no withdraw / edit actions; a 409 `governance.protocol_locked` on withdraw renders the new-decision offer (AC-38), on edit the explanation (AC-17)
- [ ] `mise run lint` and `mise run test` green
