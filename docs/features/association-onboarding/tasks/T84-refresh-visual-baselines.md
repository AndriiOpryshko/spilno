---
id: T84
title: "Refresh the visual baselines of the review page and the association card"
layer: "tests"
deps: ["T71", "T72", "T77"]
acs: ["AC-10", "AC-29"]
files_hint:
  - "apps/web/src/features/identity/__screenshots__/"
owner: "Andrii Opryshko"
estimate: "S"
status: "done"
---

# T84 — Refresh the visual baselines of the review page and the association card

## Why

[Review of 2026-11-10](../_review/review-2026-11-10.md), follow-up of F09, F10, F11, F12. The review page gains rows and a competitor list and the card changes its history marks, so the committed screenshots no longer match.

**Satisfies:** [PRD §5](../PRD.md) AC-10, AC-29 · **Blocked by:** [T71](./T71-ui-ukrainian-warnings-field-errors.md), [T72](./T72-card-ended-mark-only-when-a-term-ended.md), [T77](./T77-ui-reuse-inline-alert-hooks-links.md)

## What

- Update the arm64 baselines locally with `test:stories -u`; take the x64 ones from the CI Playwright image.
- Look at each new image before committing it.

## Definition of Done

- [ ] Acceptance: `pnpm --filter @spilno/web test:stories` passes on linux-arm64 and in CI on linux-x64 with baselines that show the new rows.
- [ ] story tests run serially and are rerun once on a cold-start flake
- [ ] only `review-page-*` and `association-card-*` change
- [ ] `mise run lint` and `mise run test` green
