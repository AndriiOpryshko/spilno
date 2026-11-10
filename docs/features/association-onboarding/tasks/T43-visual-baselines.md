---
id: T43
title: "Add visual-regression checks for the filing form, the review page and the association card"
layer: "tests"
deps: ["T36", "T38", "T41"]
acs: ["AC-01", "AC-10", "AC-29"]
files_hint:
  - "apps/web/vite.config.ts"
  - "apps/web/src/features/identity/applications/filing-form.stories.tsx"
  - "apps/web/src/features/identity/operator/review-page.stories.tsx"
  - "apps/web/src/features/identity/association/association-card.stories.tsx"
  - "apps/web/src/features/identity/__screenshots__/"
owner: "Andrii Opryshko"
estimate: "S"
status: "todo"
---

# T43 — Add visual-regression checks for the filing form, the review page and the association card

## Why

[test-plan](../test-plan.md) «Level decisions»: visual-regression is on for three screens (product owner, 2026-11-10). The repo has no such check yet.

**Satisfies:** [PRD §5](../PRD.md) AC-01, AC-10, AC-29 · **Blocked by:** [T36](./T36-ui-filing-forms.md), [T38](./T38-ui-operator-queue-review.md), [T41](./T41-ui-association-card.md)

## What

- Screenshot assertions on the three stories in the browser test runner the story tests already use — no second runner; baselines committed and produced in one fixed environment.
- A documented command to update a baseline deliberately.

## Definition of Done

- [ ] Acceptance: The story tests of the three screens compare against committed baselines and fail on an unintended visual change.
- [ ] `mise run test` passes on the baselines and fails when a spacing token of one of the three stories is changed (shown once in the PR, then reverted)
- [ ] the baselines are in the repo; the update command is in the PR description for T48 to document
- [ ] `mise run lint` and `mise run test` green

## Notes

Fonts and the viewport are pinned so the CI image and a laptop render the same pixels.
