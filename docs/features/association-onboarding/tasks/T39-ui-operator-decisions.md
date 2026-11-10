---
id: T39
title: "Add the оператор's decisions: approve, reject and their refusals"
layer: "ui"
deps: ["T38"]
acs: ["AC-03", "AC-03b", "AC-08", "AC-12", "AC-13", "AC-14", "AC-15", "AC-16c", "AC-18b"]
files_hint:
  - "apps/web/src/features/identity/operator/decision-panel.tsx"
  - "apps/web/src/features/identity/operator/decision-panel.test.tsx"
  - "apps/web/src/features/identity/operator/decision-panel.stories.tsx"
  - "apps/web/src/components/ui/dialog.tsx"
  - "apps/web/src/components/ui/dialog.stories.tsx"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T39 — Add the оператор's decisions: approve, reject and their refusals

## Why

[PRD](../PRD.md) US-03, US-05; [sad](../sad.md) §6 Flows 2, 3, 8; [ADR-0009](../adr/0009-treat-the-operator-role-and-a-roster-membership-as-independent-grants-on-one-account.md): the web may hide the buttons on the own filing, the 409 stays the rule.

**Satisfies:** [PRD §5](../PRD.md) AC-03, AC-03b, AC-08, AC-12, AC-13, AC-14, AC-15, AC-16c, AC-18b · **Blocked by:** [T38](./T38-ui-operator-queue-review.md)

## What

- On the review page: «Схвалити» with the field «документ, на якому ґрунтується рішення» (up to 500 characters) and «Відхилити» with a reason in a dialog; after a decision the state and its date.
- Refusals by `code`, verbatim from [PRD](../PRD.md) §5.1: `signature_not_checked`, `extract_code_mismatch`, `mismatch_needs_document`, `reason_required`, `application_not_pending`, `operator_own_application`, `association_already_connected`.
- New primitive `Dialog` over the installed `radix-ui` (design-system «Gaps»), with a story; both buttons are hidden when `applicant_email` equals the principal's address.

## Definition of Done

- [ ] Acceptance: Component tests show an approval and a rejection through to the new state, each server refusal with its PRD §5.1 message, and no decision buttons on the оператор's own filing.
- [ ] one test per refusal message; an approval after typing the document succeeds
- [ ] the own filing shows «Це ваша власна заявка — оператор не ухвалює рішення щодо себе» and no buttons
- [ ] a rejection with an empty reason shows «Вкажіть причину — її побачить заявник»
- [ ] the dialog story passes a11y (focus trap, Esc, labelled)
- [ ] `mise run lint` and `mise run test` green

## Notes

Reuses `Button`, `Textarea`, `InlineAlert`.
