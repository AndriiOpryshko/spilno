---
id: T109
title: "Refresh the visual baselines the second review's fixes changed"
layer: "tests"
deps: ["T88", "T89", "T102", "T104", "T105", "T106"]
acs: ["AC-01", "AC-20", "AC-29"]
files_hint:
  - "apps/web/src/features/identity/__screenshots__/"
  - "apps/web/src/features/governance/"
owner: "Andrii Opryshko"
estimate: "S"
status: "done"
---

# T109 — Refresh the visual baselines the second review's fixes changed

## Why

[Review of 2026-11-17](../_review/review-2026-11-17.md). The ui tasks change what the baselined stories show.

**Satisfies:** [PRD §5](../PRD.md) AC-01, AC-20, AC-29 · **Blocked by:** [T88](./T88-ui-roster-marks-by-effect-and-kind.md), [T89](./T89-ui-entry-to-connection-form.md), [T102](./T102-ui-operator-fallback-and-focus.md), [T104](./T104-ui-page-roster-and-memberships.md), [T105](./T105-ui-page-operator-lists.md), [T106](./T106-ui-extract-file-check-before-upload.md)

## What

- `test:stories -u` on linux-arm64; the linux-x64 files from the CI Playwright image.

## Definition of Done

- [ ] Acceptance: `pnpm --filter @spilno/web test:stories` passes on linux-arm64 and in CI on linux-x64 with baselines that show the new links, marks and «Показати ще».
- [ ] only the stories named by the ui tasks changed
- [ ] `mise run lint` and `mise run test` green
