---
id: T59
title: "Make the memberships lookup use an index on the lowered sign-in address"
layer: "migration"
deps: []
acs: ["AC-04d"]
files_hint:
  - "docs/features/association-onboarding/migrations/NNNNN_index_board_member_revisions_lower_sign_in_email.sql"
  - "docs/features/association-onboarding/migrations/MANIFEST.md"
  - "services/core/internal/governance/adapters/postgres/queries/roster.sql"
  - "docs/features/association-onboarding/data-model.md"
owner: "Andrii Opryshko"
estimate: "S"
status: "done"
---

# T59 — Make the memberships lookup use an index on the lowered sign-in address

## Why

[Review of 2026-11-10](../_review/review-2026-11-10.md), F20. [data-model](../data-model.md) query → index map: `MembershipsOf` is served by the index on `board_member_revisions(sign_in_email)`. The query and the RLS policy filter `lower(sign_in_email)`, which that index cannot serve.

**Satisfies:** [PRD §5](../PRD.md) AC-04d · **Blocked by:** —

## What

- First run `EXPLAIN` on a seeded database: the review did not verify this finding. If the plan already uses the index, correct only the data-model row.
- Otherwise stage `CREATE INDEX CONCURRENTLY IF NOT EXISTS` on `lower(sign_in_email)` (one statement, `-- +goose NO TRANSACTION`) and push the address predicate into the subquery of `MembershipsOf`.
- The old plain index is dropped in a later release, not here (migrations.md: drop is its own phase).

## Definition of Done

- [ ] Acceptance: `EXPLAIN` of `MembershipsOf` on a seeded roster uses an index for the address predicate, and the data model names that index.
- [ ] the migration applies and reverts cleanly (`/roundtrip`)
- [ ] a test or the PR text shows the plan
- [ ] `MembershipsOf` keeps returning only the latest revision of each entry
- [ ] `mise run lint` and `mise run test` green

## Notes

Promoted with `/promote-migration association-onboarding`.
