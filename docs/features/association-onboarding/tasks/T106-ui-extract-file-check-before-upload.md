---
id: T106
title: "Check the extract's size and type before the upload and name the field on a 413"
layer: "ui"
deps: []
acs: ["AC-02"]
files_hint:
  - "apps/web/src/features/identity/applications/filing-form.tsx"
  - "apps/web/src/features/identity/applications/filing-form.test.tsx"
  - "apps/web/src/i18n/uk.ts"
owner: "Andrii Opryshko"
estimate: "S"
status: "todo"
---

# T106 — Check the extract's size and type before the upload and name the field on a 413

## Why

[Review of 2026-11-17](../_review/review-2026-11-17.md), Q16. [PRD](../PRD.md) AC-02 names every broken field; an oversized extract answers 413 with no `errors[]`, so the form shows the generic failure.

**Satisfies:** [PRD §5](../PRD.md) AC-02 · **Blocked by:** —

## What

- The check that `case-documents.tsx` already does, moved to a shared helper and used by the filing form.
- A 413 sets the `extract` field error.

## Definition of Done

- [ ] Acceptance: Component tests show an extract over 5 MB or of another type refused under the field without a request, and a 413 answer shown as the same field message.
- [ ] `filing-form.test.tsx`: too large, wrong type, 413
- [ ] `mise run lint` and `mise run test` green
