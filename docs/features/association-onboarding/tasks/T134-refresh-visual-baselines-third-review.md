---
id: T134
title: "Refresh the visual baseline of the association card"
layer: "tests"
deps: ["T116", "T117", "T118", "T121"]
acs: ["AC-17", "AC-29"]
files_hint:
  - "apps/web/src/features/identity/__screenshots__/"
  - "apps/web/src/features/identity/association/association-card.stories.tsx"
owner: "Andrii Opryshko"
estimate: "S"
status: "todo"
---

# T134 — Refresh the visual baseline of the association card

## Why

[Second re-review of 2026-11-17](../_review/review-2026-11-17-2.md), T117 adds a link to the association card, one of the three baselined stories; T118 and T121 touch the filing form and the review page, whose default look should not change.

**Satisfies:** [PRD §5](../PRD.md) AC-17, AC-29 · **Blocked by:** [T116](./T116-ui-lasting-links-to-applications.md), [T117](./T117-ui-transfer-request-link-for-every-member.md), [T118](./T118-ui-extract-refusal-with-every-other-field.md), [T121](./T121-ui-operator-loading-states-and-way-back.md)

## What

- The card's baseline is updated on linux-arm64 here and on linux-x64 in the pinned Playwright image; the other two are checked, not updated.

## Definition of Done

- [ ] Acceptance: `pnpm --filter @spilno/web test:stories` passes on linux-arm64 and in the `web` job on linux-x64, with a card baseline that shows the link and with the filing-form and review-page baselines unchanged.
- [ ] both card baselines committed together
- [ ] a diff of the other two baselines would be a finding, not an update
- [ ] `mise run lint` and `mise run test` green

## Notes

Story runs are serial (one fixed port) and the first run after a change may fail while collecting: rerun once before reading a failure.
