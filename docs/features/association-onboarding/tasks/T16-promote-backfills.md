---
id: T16
title: "Promote the backfills: association state, roster positions and the contracts' association"
layer: "migration"
deps: ["T06", "T13", "T14"]
acs: ["AC-19c", "AC-29"]
files_hint:
  - "docs/features/association-onboarding/migrations/NNNNN_backfill_associations_state.sql"
  - "docs/features/association-onboarding/migrations/NNNNN_backfill_board_member_revisions_position.sql"
  - "docs/features/association-onboarding/migrations/NNNNN_backfill_contracts_association_id.sql"
  - "docs/features/association-onboarding/migrations/backfill-associations-state.md"
  - "docs/features/association-onboarding/migrations/backfill-board-member-revisions-position.md"
  - "docs/features/association-onboarding/migrations/backfill-contracts-association-id.md"
  - "docs/features/association-onboarding/migrations/MANIFEST.md"
owner: "Andrii Opryshko"
estimate: "S"
status: "todo"
---

# T16 — Promote the backfills: association state, roster positions and the contracts' association

## Why

[PRD](../PRD.md) §9 «Existing records»: the association set up by hand becomes `connected`, its roster entries get `member`, its contracts get the seeded association. Group «PR 2» of [MANIFEST](../migrations/MANIFEST.md); the app already writes the three columns (T13, T14).

**Satisfies:** [PRD §5](../PRD.md) AC-19c, AC-29 · **Blocked by:** [T06](./T06-promote-drop-accounts-association.md), [T13](./T13-finance-contracts-association.md), [T14](./T14-roster-read-facade.md)

## What

- Promote staged files 10–12 in manifest order with `mise run migrate:new <name>`; `/roundtrip`.
- Follow the three backfill notes for the checks before and after.

## Definition of Done

- [ ] Acceptance: After the three backfills no row has a NULL in associations.state, board_member_revisions.position or contracts.association_id, and the migrations revert and re-apply cleanly.
- [ ] migration test: rows inserted with NULLs before the backfill are filled after it (`connected`, `member`, the seeded association id)
- [ ] Down is a no-op; `/roundtrip` clean; `mise run test` green
- [ ] `mise run lint` and `mise run test` green

## Notes

Nobody is marked `head` by the backfill: the hand-made association gets its голова only through AC-19c (T29).
