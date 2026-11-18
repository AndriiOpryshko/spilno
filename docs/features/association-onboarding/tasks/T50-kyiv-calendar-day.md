---
id: T50
title: "Take «today» for every roster date from the Europe/Kyiv calendar"
layer: "infra"
deps: []
acs: ["AC-09", "AC-19", "AC-29"]
files_hint:
  - "services/core/internal/platform/kyivday/"
  - "services/core/internal/identity/guard.go"
  - "services/core/internal/identity/adapters/postgres/applications.go"
  - "services/core/internal/identity/app/card.go"
  - "services/core/internal/identity/app/memberships.go"
  - "services/core/internal/identity/app/storage.go"
  - "services/core/internal/identity/app/approval.go"
  - "services/core/internal/identity/app/review.go"
  - "services/core/internal/governance/app/roster.go"
  - "services/core/internal/identity/approval_transfer_test.go"
owner: "Andrii Opryshko"
estimate: "M"
status: "done"
---

# T50 — Take «today» for every roster date from the Europe/Kyiv calendar

## Why

[Review of 2026-11-10](../_review/review-2026-11-10.md), F07. [sad](../sad.md) §8 «Authorization» and «Dates»: the roster is asked for «today Europe/Kyiv». The guard, the card filing, the card, the memberships and the storage consent use the UTC day, so around Kyiv midnight an approval takes effect two to three hours late.

**Satisfies:** [PRD §5](../PRD.md) AC-09, AC-19, AC-29 · **Blocked by:** —

## What

- One helper for «today in Europe/Kyiv» with `time/tzdata` embedded, taking the clock as an argument so tests can pin it.
- Use it at every roster date: the guard, `fileFromCard`, the card, the memberships, the storage consent, `SeatHead`, and the two existing `LoadLocation` calls in `app/approval.go` and `app/review.go`.
- `domain/filing.go` keeps its own zone lookup: `domain/` is stdlib only and may not import `platform`.

## Definition of Done

- [ ] Acceptance: A test pinned at 00:40 Europe/Kyiv on the first day of a new term shows the new голова admitted and the previous one refused, and no roster date in identity or governance is taken from the UTC day.
- [ ] unit test of the helper on both sides of Kyiv midnight, in winter and in summer time
- [ ] integration test: an approval at 00:40 Kyiv with term start that day admits the new голова on the next request
- [ ] `grep` finds no `time.Now().UTC()` feeding a roster date
- [ ] `mise run lint` and `mise run test` green

## Notes

The web side of the same rule is T76.
