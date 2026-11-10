---
id: T05
title: "Promote the expand migrations: application tables, association and roster columns, the app role"
layer: "migration"
deps: []
acs: ["AC-01", "AC-06", "AC-08", "AC-21", "AC-29"]
files_hint:
  - "docs/features/association-onboarding/migrations/NNNNN_create_applications.sql"
  - "docs/features/association-onboarding/migrations/NNNNN_add_associations_onboarding_columns.sql"
  - "docs/features/association-onboarding/migrations/NNNNN_index_associations_edrpou.sql"
  - "docs/features/association-onboarding/migrations/NNNNN_index_associations_confirmed_by_decision.sql"
  - "docs/features/association-onboarding/migrations/NNNNN_create_association_state_changes.sql"
  - "docs/features/association-onboarding/migrations/NNNNN_add_board_member_revisions_position.sql"
  - "docs/features/association-onboarding/migrations/NNNNN_index_board_member_revisions_sign_in_email.sql"
  - "docs/features/association-onboarding/migrations/NNNNN_add_contracts_association_id.sql"
  - "docs/features/association-onboarding/migrations/NNNNN_create_app_role.sql"
  - "docs/features/association-onboarding/migrations/MANIFEST.md"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T05 — Promote the expand migrations: application tables, association and roster columns, the app role

## Why

[data-model](../data-model.md) «Migration plan», group «PR 1» of [MANIFEST](../migrations/MANIFEST.md): five new identity tables, nullable `edrpou`, `address`, `state`, `confirmed_by_decision_id` on `associations`, `association_state_changes`, `position` and `decision_id` on `board_member_revisions`, `association_id` on `contracts`, and the `spilno_app` role with its insert-only revokes ([ADR-0003](../adr/0003-enforce-association-isolation-with-postgres-row-level-security.md), [ADR-0005](../adr/0005-check-the-signed-extract-manually-and-keep-the-uploaded-file-in-identity.md), [ADR-0007](../adr/0007-enforce-a-disabled-association-as-read-only-in-the-identity-middleware.md)).

**Satisfies:** [PRD §5](../PRD.md) AC-01, AC-06, AC-08, AC-21, AC-29 · **Blocked by:** —

## What

- Promote the nine staged files in manifest order into `services/core/migrations/` with `mise run migrate:new <name>`, moving the SQL as written; the three `CONCURRENTLY` files keep `-- +goose NO TRANSACTION` and one statement each.
- `mise run gen`, then `/roundtrip` from the last version on `main`.

## Definition of Done

- [ ] Acceptance: Staged files 1-9 are live goose files in manifest order that apply, revert and re-apply cleanly, and sqlc generates the new tables.
- [ ] nine live files, numbered by goose, in the order of MANIFEST.md rows 1–9
- [ ] `/roundtrip` (up → down to the previous head → up) leaves an empty schema diff
- [ ] `mise run gen:check` shows only the new sqlc models; `mise run test` green (`migrate_test.go`)
- [ ] `mise run lint` and `mise run test` green

## Notes

The staged files are goose single files with the `NNNNN_` placeholder, not `.up.sql` / `.down.sql` pairs, so `/promote-migration`'s `promote.sh` does not take them: follow the header of each file and MANIFEST.md. Every column added here is nullable and nothing reads it yet. Head of the migration lane: T05 → T06 → T16 → T17 → T20.
