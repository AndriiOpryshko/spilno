---
id: T17
title: "Promote the NOT NULL steps and the contracts index"
layer: "migration"
deps: ["T16"]
acs: ["AC-08", "AC-21"]
files_hint:
  - "docs/features/association-onboarding/migrations/NNNNN_contract_associations_state.sql"
  - "docs/features/association-onboarding/migrations/NNNNN_contract_board_member_revisions_position.sql"
  - "docs/features/association-onboarding/migrations/NNNNN_contract_contracts_association_id.sql"
  - "docs/features/association-onboarding/migrations/NNNNN_index_contracts_association_created.sql"
  - "docs/features/association-onboarding/migrations/MANIFEST.md"
owner: "Andrii Opryshko"
estimate: "S"
status: "todo"
---

# T17 — Promote the NOT NULL steps and the contracts index

## Why

Third step of the zero-downtime rule (`.claude/rules/migrations.md`) for the three columns; staged files 13–16 of [MANIFEST](../migrations/MANIFEST.md).

**Satisfies:** [PRD §5](../PRD.md) AC-08, AC-21 · **Blocked by:** [T16](./T16-promote-backfills.md)

## What

- Promote staged files 13–16 in manifest order; the index file keeps `-- +goose NO TRANSACTION`; `mise run gen`; `/roundtrip`.

## Definition of Done

- [ ] Acceptance: associations.state, board_member_revisions.position and contracts.association_id are NOT NULL, idx_contracts_association_created exists, and the whole test suite still writes every row.
- [ ] the three columns are `NOT NULL`; the helper `CHECK … NOT VALID` of each file is gone after it
- [ ] `/roundtrip` clean; `mise run test` green — every writer sets the columns
- [ ] `mise run lint` and `mise run test` green

## Notes

Staged files 17–19 of the same «PR 3» group are split into T20 because they need the middleware of T19; 13–16 need only the backfills.
