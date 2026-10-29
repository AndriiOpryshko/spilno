---
id: T36
title: "Move the cases list and case screen to generated hooks with document upload"
layer: "ui"
deps: ["T07"]
acs: ["AC-26", "AC-09"]
files_hint:
  - "apps/web/src/features/governance/cases-screen.tsx"
  - "apps/web/src/routes/cases.index.tsx"
  - "apps/web/src/routes/cases.$caseId.index.tsx"
  - "apps/web/src/features/governance/case-documents.tsx"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T36 — Move the cases list and case screen to generated hooks with document upload

## Why

[ADR-0005](../adr/0005-store-files-in-association-owned-google-drive-folder.md); PRD §6 upload limits (20 MB, PDF/JPEG/PNG).

**Satisfies:** [PRD §5](../PRD.md) AC-26, AC-09 · **Blocked by:** [T07](./T07-merge-contract-identity-skeleton.md)

## What

- Replace `listCases` in `api.ts`; case screen with documents and their versions.
- Downloads through downloadFile (attachment).

## Definition of Done

- [ ] Acceptance: Component tests show the case list and case screen load from listCases / getCase, uploading a document (or a newer version) calls addCaseDocument and shows `governance.file_invalid` inline, and stories pass the a11y check.
- [ ] component tests for list, upload success and `file_invalid`
- [ ] client-side size/type pre-check mirrors the server rule
- [ ] `mise run lint` and `mise run test` green