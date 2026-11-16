---
id: T61
title: "Record that the expand, backfill and contract migrations shipped in one release"
layer: "docs"
deps: ["T59"]
acs: ["AC-04d"]
files_hint:
  - "docs/features/association-onboarding/data-model.md"
  - "docs/features/association-onboarding/migrations/MANIFEST.md"
  - "docs/features/association-onboarding/tasks/_epic.md"
owner: "Andrii Opryshko"
estimate: "S"
status: "todo"
---

# T61 — Record that the expand, backfill and contract migrations shipped in one release

## Why

[Review of 2026-11-10](../_review/review-2026-11-10.md), F19. [data-model](../data-model.md) «Ships in PR 1–4», `migrations/MANIFEST.md` and [ADR-0004](../adr/0004-let-any-verified-person-sign-in-and-derive-memberships-from-rosters.md) describe four PRs and two releases; all 21 files reached `main` in PR #48.

**Satisfies:** [PRD §5](../PRD.md) AC-04d · **Blocked by:** [T59](./T59-index-lower-sign-in-email.md)

## What

- State the deviation: one PR, one goose run at startup, a stop-start deploy of a single instance with no old binary alive across it.
- Correct the «Ships in» columns and the PR numbers; note that the headers of the applied files keep their original comments, because applied migrations are not edited.
- State the rule going forward: the phases of a rename or drop are separate releases once more than one instance runs.

## Definition of Done

- [ ] Acceptance: `data-model.md` and `MANIFEST.md` state how the migrations were actually delivered, why that was safe, and the rule for the next expand/contract change.
- [ ] every link in the two files resolves
- [ ] the epic's «Risks» entry on the migration order points at the new text
- [ ] `mise run lint` and `mise run test` green
