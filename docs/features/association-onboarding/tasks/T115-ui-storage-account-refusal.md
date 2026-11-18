---
id: T115
title: "Tell the член правління why another storage account was refused"
layer: "ui"
deps: ["T114"]
acs: ["AC-27"]
files_hint:
  - "apps/web/src/routes/storage.callback.tsx"
  - "apps/web/src/features/identity/storage/storage-line.tsx"
  - "apps/web/src/features/identity/storage/storage-line.test.tsx"
  - "apps/web/src/features/identity/storage/storage-line.stories.tsx"
  - "apps/web/src/i18n/uk.ts"
owner: "Andrii Opryshko"
estimate: "S"
status: "todo"
---

# T115 — Tell the член правління why another storage account was refused

## Why

[Second re-review of 2026-11-17](../_review/review-2026-11-17-2.md), R07. The refusal of [T114](./T114-refuse-another-storage-account-while-files-exist.md) reaches the storage callback page, which today shows the generic failure.

**Satisfies:** [PRD §5](../PRD.md) AC-27 · **Blocked by:** [T114](./T114-refuse-another-storage-account-while-files-exist.md)

## What

- The callback page maps the new code to a Ukrainian message that names the connected account's address and says the files stay there.
- Reuses `InlineAlert` and the existing way back to the card; the string lives in `uk.ts`.

## Definition of Done

- [ ] Acceptance: A component test shows the storage callback answering the new 409 with a Ukrainian message that names the reason and a link back to the association card, and no English detail.
- [ ] component test: the 409 with the code shows the message and the link
- [ ] story of the refused state
- [ ] `mise run lint` and `mise run test` green

## Notes

No e2e: the development consent signs every association in with one fake account.
