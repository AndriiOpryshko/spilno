---
id: T81
title: "Prove the fixed flows end to end by clicking, not by typing addresses"
layer: "tests"
deps: ["T52", "T68", "T71", "T72", "T73"]
acs: ["AC-04c", "AC-05", "AC-06", "AC-08", "AC-11", "AC-21", "AC-29"]
files_hint:
  - "apps/web/e2e/us02-follow-and-refile.spec.ts"
  - "apps/web/e2e/us03-decide-application.spec.ts"
  - "apps/web/e2e/us06-disable-enable.spec.ts"
  - "apps/web/e2e/us09-association-card.spec.ts"
  - "apps/web/e2e/refusals.spec.ts"
  - "apps/web/e2e/support.ts"
owner: "Andrii Opryshko"
estimate: "M"
status: "done"
---

# T81 — Prove the fixed flows end to end by clicking, not by typing addresses

## Why

[Review of 2026-11-10](../_review/review-2026-11-10.md), F01, F02, F03, F04, F08, F11, F12 (e2e halves). The e2e suite passed over each of these defects: it signs in through the API, opens the card and the operator pages with `page.goto`, stops after a refile and asserts that «some» warning is visible.

**Satisfies:** [PRD §5](../PRD.md) AC-04c, AC-05, AC-06, AC-08, AC-11, AC-21, AC-29 · **Blocked by:** [T52](./T52-refuse-stale-filing-and-foreign-code.md), [T68](./T68-ui-association-shell-card-notice.md), [T71](./T71-ui-ukrainian-warnings-field-errors.md), [T72](./T72-card-ended-mark-only-when-a-term-ended.md), [T73](./T73-ui-operator-contact-address.md)

## What

- `us02-follow-and-refile.spec.ts` continues on the оператор's side through the approval.
- One spec signs in through the button; `us06` and `us09` reach the card and the associations list from the navigation.
- `refusals.spec.ts` opens the second competing заявка from the queue and asserts the warning text.

## Definition of Done

- [ ] Acceptance: `mise run e2e` passes with: sign-in by the «Увійти» button, the оператор checking and approving a refiled заявка after downloading its extract, the card and the operator pages reached by clicking, competing заявки opened from the queue, and the exact Ukrainian warning text.
- [ ] run on the isolated e2e stack, never on :5432, :5180 or :8080
- [ ] no remaining `page.goto` to a page that has a link in the navigation
- [ ] `mise run lint` and `mise run test` green; `mise run e2e` green
