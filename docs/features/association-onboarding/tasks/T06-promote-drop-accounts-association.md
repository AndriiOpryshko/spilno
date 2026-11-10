---
id: T06
title: "Promote the drop of accounts.association_id ahead of the manifest order"
layer: "migration"
deps: ["T05"]
acs: ["AC-04d", "AC-05"]
files_hint:
  - "docs/features/association-onboarding/migrations/NNNNN_drop_accounts_association_id.sql"
  - "docs/features/association-onboarding/migrations/MANIFEST.md"
  - "docs/features/association-onboarding/data-model.md"
owner: "Andrii Opryshko"
estimate: "S"
status: "todo"
---

# T06 — Promote the drop of accounts.association_id ahead of the manifest order

## Why

[ADR-0004](../adr/0004-let-any-verified-person-sign-in-and-derive-memberships-from-rosters.md): an account no longer carries one association. `accounts.association_id` is `NOT NULL` (live `00009`), so a person off every roster cannot get an account row (AC-05) while it exists. [MANIFEST](../migrations/MANIFEST.md) puts this file last, «after the release that stops reading it» — no Go code reads or writes `accounts` today (`identity.sql` has only `CountAccounts`), so that release is already out. Kept last, the order would also loop: row-level security (T20) needs the middleware (T19), which needs sign-in (T18), which needs this drop.

**Satisfies:** [PRD §5](../PRD.md) AC-04d, AC-05 · **Blocked by:** [T05](./T05-promote-expand-migrations.md)

## What

- Check that nothing reads the column: `rg -n association_id services/core/internal/identity` shows only `google_connections`.
- Promote `NNNNN_drop_accounts_association_id.sql` right after the files of T05; `mise run gen`; `/roundtrip`.
- Move row 20 in `MANIFEST.md` and in the «Migration plan» table of `data-model.md` to its new place, with one line of reason.

## Definition of Done

- [ ] Acceptance: accounts has no association_id column, the migration reverts and re-applies cleanly, and MANIFEST.md and data-model.md record the changed order with its reason.
- [ ] the live file drops `idx_accounts_association_id` and the column; Down restores both as staged
- [ ] `/roundtrip` clean; `mise run test` green
- [ ] MANIFEST.md and data-model.md name the new order and why
- [ ] `mise run lint` and `mise run test` green

## Notes

Deviation from the staged order, decided in this breakdown — confirm it at review. Down is exact only while the seeded association is the only one an account could have belonged to (header of the staged file).
