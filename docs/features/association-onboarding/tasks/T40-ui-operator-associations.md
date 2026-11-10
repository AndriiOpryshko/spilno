---
id: T40
title: "Build the оператор's associations list with disabling and enabling"
layer: "ui"
deps: ["T35"]
acs: ["AC-21", "AC-22", "AC-23", "AC-24"]
files_hint:
  - "apps/web/src/routes/operator.associations.tsx"
  - "apps/web/src/features/identity/operator/associations.tsx"
  - "apps/web/src/features/identity/operator/associations.test.tsx"
  - "apps/web/src/features/identity/operator/associations.stories.tsx"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T40 — Build the оператор's associations list with disabling and enabling

## Why

[PRD](../PRD.md) US-06; [sad](../sad.md) §6 Flow 9.

**Satisfies:** [PRD §5](../PRD.md) AC-21, AC-22, AC-23, AC-24 · **Blocked by:** [T35](./T35-ui-shell-sign-in-picker.md)

## What

- `/operator/associations`: name, ЄДРПОУ, state as a `Badge`; «Вимкнути» opens an inline reason field (`Textarea`) and posts `disableAssociation`; «Увімкнути» posts `enableAssociation`.
- Refusals: «Вкажіть причину — її побачать члени правління», «ОСББ уже в цьому стані».

## Definition of Done

- [ ] Acceptance: Component tests show an association disabled with a reason and enabled again, and the messages for an empty reason and for the same state.
- [ ] a disabling with a reason flips the badge to вимкнене; an enabling flips it back
- [ ] both refusal messages render from their codes
- [ ] the story passes a11y
- [ ] `mise run lint` and `mise run test` green

## Notes

Inline form on purpose: no dependency on the dialog of T39, so this runs next to T36 and T41. Reuses `Card`, `Badge`, `Button`, `Textarea`, `InlineAlert`.
