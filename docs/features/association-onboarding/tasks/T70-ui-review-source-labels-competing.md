---
id: T70
title: "Label the source of every value on the review page and show competing заявки side by side"
layer: "ui"
deps: ["T60", "T69"]
acs: ["AC-04c", "AC-10", "AC-19b"]
files_hint:
  - "apps/web/src/features/identity/operator/review-page.tsx"
  - "apps/web/src/features/identity/operator/review-page.test.tsx"
  - "apps/web/src/features/identity/operator/review-page.stories.tsx"
  - "apps/web/src/features/identity/operator/source-label.tsx"
  - "apps/web/src/features/identity/operator/queue.tsx"
  - "apps/web/src/i18n/uk.ts"
  - "services/core/api/openapi.yaml"
  - "docs/features/association-onboarding/contracts/openapi.yaml"
owner: "Andrii Opryshko"
estimate: "M"
status: "done"
---

# T70 — Label the source of every value on the review page and show competing заявки side by side

## Why

[Review of 2026-11-10](../_review/review-2026-11-10.md), F10, F11. [PRD](../PRD.md) AC-10 lists every value and its source; AC-04c and AC-19b ask for competing заявки and запити «side by side». The page labels two of six typed values and three of five checked ones, and a competitor is only a badge.

**Satisfies:** [PRD §5](../PRD.md) AC-04c, AC-10, AC-19b · **Blocked by:** [T60](./T60-contract-disabled-and-refile-codes.md), [T69](./T69-ui-review-latest-filing-extract-download.md)

## What

- Typed rows for the ЄДРПОУ, the address, the position and the term; checked rows for the signing date and the extract date, each with `SignatureSource` / `ExtractSource`.
- The competing applications are listed with applicant, filing date and a link; the queue row carries the marker.
- If the queue item has no competing marker in the contract, add it in the delta and the live contract in this task (`mise run gen`).

## Definition of Done

- [ ] Acceptance: Component tests assert each value of AC-10 with its source label one by one, and a review of a заявка with a competitor lists the other заявки with links, with the marker in the queue as well.
- [ ] one assertion per label/value pair, no `length > 0`
- [ ] test and story for two competing заявки and for two competing запити
- [ ] `mise run lint` and `mise run test` green

## Notes

Reuses `Badge`, `Card`.
