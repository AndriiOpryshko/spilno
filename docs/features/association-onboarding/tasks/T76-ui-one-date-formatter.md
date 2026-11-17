---
id: T76
title: "Format every date with one Europe/Kyiv formatter"
layer: "ui"
deps: []
acs: ["AC-05", "AC-29"]
files_hint:
  - "apps/web/src/i18n/uk.ts"
  - "apps/web/src/features/identity/applications/format-date.ts"
  - "apps/web/src/features/identity/operator/source-label.tsx"
  - "apps/web/src/features/identity/applications/application-detail.tsx"
  - "apps/web/src/features/identity/operator/decision-panel.tsx"
  - "apps/web/src/features/identity/association/association-card.tsx"
owner: "Andrii Opryshko"
estimate: "S"
status: "done"
---

# T76 — Format every date with one Europe/Kyiv formatter

## Why

[Review of 2026-11-10](../_review/review-2026-11-10.md), W4. [sad](../sad.md) §8 «Dates»: dates are Europe/Kyiv calendar dates. Two formatters take the UTC day and one the browser's zone, so one decision shows two dates.

**Satisfies:** [PRD §5](../PRD.md) AC-05, AC-29 · **Blocked by:** —

## What

- One formatter in `i18n/uk.ts` with `timeZone: 'Europe/Kyiv'`; `format-date.ts` and the one in `source-label.tsx` are removed.
- Calendar dates (terms, extract date) are formatted without a zone shift.

## Definition of Done

- [ ] Acceptance: A test shows a decision at 2026-11-10T22:30Z rendered as 11.11.2026 on the applicant's page, the operator panel and the card, and only one date formatter exists in the web app.
- [ ] unit test on both sides of Kyiv midnight
- [ ] `grep` finds no other `toLocaleDateString` or ISO slicing
- [ ] `mise run lint` and `mise run test` green

## Notes

The server side of the same rule is T50.
