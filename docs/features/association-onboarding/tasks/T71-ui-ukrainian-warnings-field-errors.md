---
id: T71
title: "Write the warnings and the field errors in Ukrainian from structured data"
layer: "ui"
deps: ["T70"]
acs: ["AC-02", "AC-11", "AC-11b"]
files_hint:
  - "apps/web/src/features/identity/operator/review-page.tsx"
  - "apps/web/src/features/identity/operator/review-page.test.tsx"
  - "apps/web/src/features/identity/operator/review-page.stories.tsx"
  - "apps/web/src/features/identity/applications/filing-form.tsx"
  - "apps/web/src/features/identity/applications/filing-form.test.tsx"
  - "apps/web/src/features/identity/applications/filing-form.stories.tsx"
  - "apps/web/src/i18n/uk.ts"
owner: "Andrii Opryshko"
estimate: "M"
status: "done"
---

# T71 — Write the warnings and the field errors in Ukrainian from structured data

## Why

[Review of 2026-11-10](../_review/review-2026-11-10.md), F12. [PRD](../PRD.md) AC-11, AC-11b and §5.1: the wording is the product's. The page prints the server's English `detail`, cuts a date out of it with a regex, and the filing form prints `errors[].detail`; the tests feed a detail the backend never sends.

**Satisfies:** [PRD §5](../PRD.md) AC-02, AC-11, AC-11b · **Blocked by:** [T70](./T70-ui-review-source-labels-competing.md)

## What

- Build each warning from `kind` plus the values the page already has (typed name, signer, голова from the check, extract date); `detail` is not rendered.
- Map each `errors[].pointer` of the filing to a `uk` message; `identity.application_invalid` shows the §5.1 summary «Перевірте поля заявки: …»; an unknown code shows `t.failed`.
- Fixtures and stories are copied from real responses of the integration tests.

## Definition of Done

- [ ] Acceptance: Component tests fed with the exact payloads the backend sends show «розбіжність імен» with the three names, the extract-age warning with a `дд.мм.рррр` date, and a Ukrainian message next to each broken field of the filing form, with no English server text on either screen.
- [ ] `filing-form.test.tsx` no longer locks in English strings
- [ ] the e2e assertion of the real text is in T81
- [ ] `mise run lint` and `mise run test` green

## Notes

No contract change: the client composes the text. If a value is missing on the page, add a structured field to `ReviewWarning` rather than parsing `detail`.
