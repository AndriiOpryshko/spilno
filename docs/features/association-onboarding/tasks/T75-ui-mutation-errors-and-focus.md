---
id: T75
title: "Report every failed operator action and keep focus after a decision"
layer: "ui"
deps: ["T67"]
acs: ["AC-13", "AC-15", "AC-22"]
files_hint:
  - "apps/web/src/features/identity/operator/associations.tsx"
  - "apps/web/src/features/identity/operator/associations.test.tsx"
  - "apps/web/src/features/identity/operator/check-form.tsx"
  - "apps/web/src/features/identity/operator/decision-panel.tsx"
  - "apps/web/src/features/identity/operator/decision-panel.test.tsx"
  - "apps/web/src/features/identity/operator/decision-panel.stories.tsx"
  - "apps/web/src/i18n/uk.ts"
owner: "Andrii Opryshko"
estimate: "M"
status: "done"
---

# T75 — Report every failed operator action and keep focus after a decision

## Why

[Review of 2026-11-10](../_review/review-2026-11-10.md), W3, W6. [PRD](../PRD.md) §5.1: every refusal has a message. Disable and enable show nothing for an unexpected code, the check form shows one generic line and ignores the 422 fields, a 409 on a decision leaves stale buttons, and focus drops to the page body after a rejection.

**Satisfies:** [PRD §5](../PRD.md) AC-13, AC-15, AC-22 · **Blocked by:** [T67](./T67-ui-operator-layout-entry-guard.md)

## What

- `messageFor` falls back to a generic message; the check form maps `errors[].pointer` to its fields as the filing form does.
- `onError` of approve and reject refetches the review on 409.
- The outcome renders as `role="status"` and receives focus when the dialog closes.

## Definition of Done

- [ ] Acceptance: Component tests show a generic message for an unexpected failure of disable and enable, each broken field of the check form named, the review refetched after a 409 on a decision, and focus on the announced outcome after a rejection.
- [ ] tests for a 500 and a network error on «Вимкнути»
- [ ] check form with a bad ЄДРПОУ and a bad date at once
- [ ] a11y story for the decided state
- [ ] `mise run lint` and `mise run test` green
