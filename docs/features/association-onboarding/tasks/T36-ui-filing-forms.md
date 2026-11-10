---
id: T36
title: "Build the filing forms for a заявка на підключення and a запит на передачу"
layer: "ui"
deps: ["T35"]
acs: ["AC-01", "AC-02", "AC-04", "AC-04b", "AC-17", "AC-18", "AC-30"]
files_hint:
  - "apps/web/src/routes/applications.new.tsx"
  - "apps/web/src/features/identity/applications/filing-form.tsx"
  - "apps/web/src/features/identity/applications/filing-form.test.tsx"
  - "apps/web/src/features/identity/applications/filing-form.stories.tsx"
  - "apps/web/src/components/ui/label.tsx"
  - "apps/web/src/components/ui/date-field.tsx"
  - "apps/web/src/components/ui/file-field.tsx"
  - "apps/web/src/components/ui/form-fields.stories.tsx"
owner: "Andrii Opryshko"
estimate: "L"
status: "todo"
---

# T36 — Build the filing forms for a заявка на підключення and a запит на передачу

## Why

[PRD](../PRD.md) US-01, US-04; [sad](../sad.md) §6 Flows 1 and 6; [test-plan](../test-plan.md) rows AC-01, AC-02, AC-04, AC-04b, AC-17, AC-18, AC-30 (component level).

**Satisfies:** [PRD §5](../PRD.md) AC-01, AC-02, AC-04, AC-04b, AC-17, AC-18, AC-30 · **Blocked by:** [T35](./T35-ui-shell-sign-in-picker.md)

## What

- `/applications/new?kind=connection|transfer` (and `association_id` when opened from the card of the association set up by hand): the fields of `NewApplication`; the position is the fixed text «голова правління»; the term end is optional; one extract file.
- Submits with the generated `fileApplication` hook; the `errors[]` pointers of a 422 mark every field at once; each 409 shows the message of [PRD](../PRD.md) §5.1 with its link — already connected → the запит form, already open → the open заявка, not connected → the заявка form, term ended.
- New primitives, because none exists (design-system «Gaps»: form set, file upload): `Label`, `DateField`, `FileField` in `src/components/ui`, in the shadcn style with stories. Reuses `Input`, `Button`, `Card`, `InlineAlert`.

## Definition of Done

- [ ] Acceptance: Component tests show the form without a position control, every broken field marked after one submit, each refusal with its message and link, and the success message of each kind.
- [ ] no control changes the position; the term end may stay empty
- [ ] one failed submit marks every broken field with its message; nothing else changes
- [ ] each 409 renders its message and its link; success shows «Заявку подано, очікує розгляду оператором» or «Запит на передачу подано, очікує розгляду оператором»
- [ ] stories (empty, with errors, already connected) pass the a11y check
- [ ] `mise run lint` and `mise run test` green

## Notes

The 5 MB limit is enforced by the API; the form only hints it. The visual baseline of this screen is T43.
