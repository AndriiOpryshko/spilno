---
id: T38
title: "Build the оператор's queue and the review page with sources, warnings and the check form"
layer: "ui"
deps: ["T36"]
acs: ["AC-04c", "AC-10", "AC-11", "AC-11b", "AC-16", "AC-19b", "AC-28b"]
files_hint:
  - "apps/web/src/routes/operator.applications.index.tsx"
  - "apps/web/src/routes/operator.applications.$applicationId.tsx"
  - "apps/web/src/features/identity/operator/review-page.tsx"
  - "apps/web/src/features/identity/operator/review-page.test.tsx"
  - "apps/web/src/features/identity/operator/review-page.stories.tsx"
  - "apps/web/src/features/identity/operator/source-label.tsx"
  - "apps/web/src/features/identity/operator/check-form.tsx"
  - "apps/web/src/features/identity/operator/queue.tsx"
owner: "Andrii Opryshko"
estimate: "L"
status: "todo"
---

# T38 — Build the оператор's queue and the review page with sources, warnings and the check form

## Why

[ADR-0001](../adr/0001-serve-the-operator-back-office-as-an-operator-only-section-of-the-web-app.md): the back office is a section of the same SPA; [PRD](../PRD.md) US-03, US-05, AC-10; [sad](../sad.md) §6 Flow 7.

**Satisfies:** [PRD §5](../PRD.md) AC-04c, AC-10, AC-11, AC-11b, AC-16, AC-19b, AC-28b · **Blocked by:** [T36](./T36-ui-filing-forms.md)

## What

- `/operator/applications` (state filter; «конкуруюча заявка» / «конкуруючий запит» with `Badge`) and `/operator/applications/$applicationId`: every value with `SourceLabel` — «введено заявником», «з витягу з ЄДР, внесено оператором <дата>», «з КЕП, перевірено оператором <дата>», or «не перевірено».
- Warnings with `InlineAlert`: the розбіжність імен naming the values, «Витяг сформовано <дата> — старший за 30 днів»; competing applications side by side; for a запит the account that holds the files; the extract download.
- The check form (signer, signing date, керівник, the extract's ЄДРПОУ and date, the finding on the signature) posts `recordExtractCheck`; reuses `Label`, `DateField` of T36.

## Definition of Done

- [ ] Acceptance: Component tests show the three source labels and «не перевірено», both advisory warnings without a blocked decision, the competing marker for both kinds, the storage account of a запит and no queue for a person who is not the оператор.
- [ ] before a check the checked values read «не перевірено»; after it they carry the check date
- [ ] the name warning names the differing values; the age warning leaves «Схвалити» available
- [ ] the competing marker renders for a заявка and for a запит
- [ ] a person who is not the оператор gets the no-access page and no request for the queue is shown; stories (unchecked, clean, mismatch, old extract, competing) pass a11y
- [ ] `mise run lint` and `mise run test` green

## Notes

`SourceLabel` is new: `OriginTag` of `features/governance` has other origins (system, ai, human). The Figma back-office frames of BR-21 (`docs/design/stage1-screens-spec.md` §13) predate this PRD — take the layout, not the fields. Visual baseline in T43.
