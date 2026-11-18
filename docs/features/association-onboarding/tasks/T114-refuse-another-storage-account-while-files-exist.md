---
id: T114
title: "Refuse to connect another storage account while the association has files"
layer: "ports"
deps: []
acs: ["AC-26", "AC-27"]
files_hint:
  - "services/core/internal/identity/app/storage.go"
  - "services/core/internal/identity/storage.go"
  - "services/core/internal/identity/storage_test.go"
  - "services/core/internal/identity/domain/errors.go"
  - "services/core/internal/identity/module.go"
  - "services/core/internal/governance/module.go"
  - "services/core/internal/governance/adapters/postgres/queries/files.sql"
  - "services/core/internal/governance/adapters/postgres/sqlcgen/"
  - "services/core/cmd/api/main.go"
  - "services/core/api/openapi.yaml"
  - "services/core/internal/identity/adapters/http/gen/"
  - "apps/web/src/api/gen/"
  - "docs/features/association-onboarding/contracts/openapi.yaml"
  - "docs/features/association-onboarding/PRD.md"
  - "docs/features/association-onboarding/adr/0008-let-a-board-member-connect-the-association-storage-in-the-web-app.md"
  - "docs/features/association-onboarding/data-model.md"
owner: "Andrii Opryshko"
estimate: "L"
status: "done"
---

# T114 — Refuse to connect another storage account while the association has files

## Why

[Second re-review of 2026-11-17](../_review/review-2026-11-17-2.md), R07. A connection with another Google account replaces the row and moves no files, so every earlier file, the signed copies of AC-26 included, answers 404. The product owner chose on 2026-11-17 to refuse such a connection while the association has files; this closes the open question of PRD §8.

**Satisfies:** [PRD §5](../PRD.md) AC-26, AC-27 · **Blocked by:** —

## What

- `completeStorageConnection` compares the account of the new grant with the account of the current connection; when they differ and the association has at least one file, it answers 409 with a new identity code and writes nothing. The same account may connect again.
- Governance exposes «has files» to identity through its module facade, injected in `cmd/api` as the roster lookup is.
- The code is registered once in identity, listed in PRD §5.1 and in the contract (delta and live) with an example; `mise run gen`.
- PRD §8: the question is closed with the decision; PRD §5 gets the edge case next to AC-27; ADR-0008 gets a dated amendment in place of its «Open» line; `data-model.md` gets the query row.

## Definition of Done

- [ ] Acceptance: With a file recorded in the association, completing a consent with another Google account answers 409 with the new code and leaves the connection row and the file's download unchanged; the same account connects again; an association without files connects another account.
- [ ] module test through HTTP: another account with files refused, row unchanged, the file still downloads
- [ ] module test: the same account with files accepted; another account without files accepted
- [ ] the «has files» read runs under row-level security in the association's transaction, before the code exchange result is written
- [ ] contract: the 409 lists the code with an example; the registry and the contract both hold 18 identity codes
- [ ] PRD §8 question closed, PRD §5.1 row, ADR-0008 amended
- [ ] `mise run lint` and `mise run test` green

## Notes

Default name of the code: `identity.storage_has_files`. The alternatives the product owner did not take: keep the default with the consequence written down, or a distinct answer in place of the 404. The check reads `files` by `association_id`; name the index it uses in `data-model.md` and add one under the migration rules only if none serves it. The screen's message is [T115](./T115-ui-storage-account-refusal.md); the sync report is reconciled in [T131](./T131-reconcile-sync-report-and-design-docs.md).
