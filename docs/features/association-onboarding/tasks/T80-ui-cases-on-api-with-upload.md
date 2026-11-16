---
id: T80
title: "Move the справи list and the case screen to the API, with document upload and the storage message"
layer: "ui"
deps: ["T65", "T68"]
acs: ["AC-25", "AC-28"]
files_hint:
  - "apps/web/src/features/governance/cases-screen.tsx"
  - "apps/web/src/features/governance/case-documents.tsx"
  - "apps/web/src/features/governance/api.ts"
  - "apps/web/src/routes/o.$associationId.cases.index.tsx"
  - "apps/web/src/routes/o.$associationId.cases.$caseId.index.tsx"
  - "apps/web/src/features/identity/association/suspended-notice.tsx"
owner: "Andrii Opryshko"
estimate: "L"
status: "todo"
---

# T80 — Move the справи list and the case screen to the API, with document upload and the storage message

## Why

[Review of 2026-11-10](../_review/review-2026-11-10.md), F15, F08 (the mock). [PRD](../PRD.md) AC-28: attaching a document without a connected storage says «Спершу підключіть сховище документів ОСББ». The справи screen is an in-memory mock, so nothing in the browser can trigger that message, and a member of a вимкнене ОСББ can «edit» in it.

**Satisfies:** [PRD §5](../PRD.md) AC-25, AC-28 · **Blocked by:** [T65](./T65-serve-cases-and-document-upload.md), [T68](./T68-ui-association-shell-card-notice.md)

## What

- `listCases` leaves `features/governance/api.ts`; the case screen with documents and their versions is [board-protocol T36](../../board-protocol/tasks/T36-ui-cases-documents.md).
- `governance.file_invalid` is shown inline; a client-side size and type check mirrors the server rule.
- The protocol draft screen keeps its mock until board-protocol builds drafts; it says so on screen, and it is not reachable in a вимкнене ОСББ.

## Definition of Done

- [ ] Acceptance: Component tests show the справи list and the case screen loaded from `listCases` and `getCase`, a document uploaded through `addCaseDocument`, the storage message with a link to the card on `governance.drive_not_connected`, and «Доступ призупинено» on a change in a вимкнене ОСББ.
- [ ] tests for the list, an upload, `file_invalid`, `drive_not_connected` and `association_disabled`
- [ ] stories pass a11y
- [ ] `mise run lint` and `mise run test` green

## Notes

Reuses `FileField`, `InlineAlert`, `Card`. Carries board-protocol T36. Dropped together with T65 if the follow-up must shrink.
