---
id: T16
title: "Model agenda items and слоти рішення: confirm, correct, needs review and readiness"
layer: "domain"
deps: ["T15"]
acs: ["AC-04", "AC-21", "AC-28b", "AC-09", "AC-29", "AC-03c"]
files_hint:
  - "services/core/internal/governance/domain/agenda_item.go"
  - "services/core/internal/governance/domain/slot.go"
  - "services/core/internal/governance/domain/readiness.go"
  - "services/core/internal/governance/domain/slot_test.go"
  - "services/core/internal/governance/domain/readiness_test.go"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T16 — Model agenda items and слоти рішення: confirm, correct, needs review and readiness

## Why

[sad §5](../sad.md) «слот рішення» (typed value, origin ai/document/human, status, source version, fragment); AC-21, AC-28b, AC-09, AC-04.

**Satisfies:** [PRD §5](../PRD.md) AC-04, AC-21, AC-28b, AC-09, AC-29, AC-03c · **Blocked by:** [T15](./T15-protocol-lifecycle-domain.md)

## What

- Slot kinds: amount (kopiyky + currency), counterparty, джерело коштів, document reference, date/term, person.
- Agenda item with versioned wording (author, time — AC-29); items may be added while чернетка (AC-03c).
- `Readiness()` → all blockers at once (AC-04, AC-21).

## Definition of Done

- [ ] Acceptance: Unit tests prove a proposed slot stays unconfirmed until a person confirms or corrects it, a later value never overwrites a confirmed one (stored beside it, needs review), a newer document version marks citing slots for review, and readiness lists every missing agenda part and unconfirmed slot.
- [ ] unit tests for each AC above
- [ ] domain-pure depguard clean
- [ ] `mise run lint` and `mise run test` green