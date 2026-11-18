---
id: T69
title: "Review the latest filing with its latest check, and let the оператор download the extract"
layer: "ui"
deps: ["T67"]
acs: ["AC-03", "AC-06", "AC-08", "AC-15"]
files_hint:
  - "apps/web/src/features/identity/operator/review-page.tsx"
  - "apps/web/src/features/identity/operator/review-page.test.tsx"
  - "apps/web/src/features/identity/operator/review-page.stories.tsx"
  - "apps/web/src/features/identity/operator/check-form.tsx"
  - "apps/web/src/i18n/uk.ts"
owner: "Andrii Opryshko"
estimate: "M"
status: "done"
---

# T69 — Review the latest filing with its latest check, and let the оператор download the extract

## Why

[Review of 2026-11-10](../_review/review-2026-11-10.md), F01 (web half), F04. [sad](../sad.md) §6 Flow 2 starts with the check of the КЕП and the extract outside Спільно, and T38 listed the download. The page takes the oldest filing of a newest-first list and never links to the extract, so a refiled заявка cannot be checked or decided.

**Satisfies:** [PRD §5](../PRD.md) AC-03, AC-06, AC-08, AC-15 · **Blocked by:** [T67](./T67-ui-operator-layout-entry-guard.md)

## What

- `filings[0]` (or a sort by `created_at`, as `application-detail.tsx` does) and the last check of that filing.
- A download link per filing built from the generated URL helper of `downloadExtractForReview`.
- Earlier filings are listed below as history, without the check form.

## Definition of Done

- [ ] Acceptance: Component tests with two filings show the page presenting, checking and deciding the newest one with its latest check, and a download link to the extract of each filing.
- [ ] the fixture gains a second filing and a second check
- [ ] the check form and the decision panel receive the newest `filing_id`
- [ ] story for a refiled заявка
- [ ] `mise run lint` and `mise run test` green

## Notes

The server refuses a stale filing in T52.
