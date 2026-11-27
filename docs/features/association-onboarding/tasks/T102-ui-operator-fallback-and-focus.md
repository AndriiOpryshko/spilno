---
id: T102
title: "Show the shared fallback on the оператор's review and associations pages and keep focus after an approval"
layer: "ui"
deps: []
acs: ["AC-10", "AC-15", "AC-21"]
files_hint:
  - "apps/web/src/features/identity/operator/review-page.tsx"
  - "apps/web/src/features/identity/operator/associations.tsx"
  - "apps/web/src/features/identity/operator/decision-panel.tsx"
  - "apps/web/src/features/identity/operator/decision-panel.test.tsx"
  - "apps/web/src/features/identity/operator/review-page.test.tsx"
  - "apps/web/src/features/identity/operator/associations.test.tsx"
  - "apps/web/src/features/identity/shell/page-states.test.tsx"
owner: "Andrii Opryshko"
estimate: "M"
status: "done"
---

# T102 — Show the shared fallback on the оператор's review and associations pages and keep focus after an approval

## Why

[Review of 2026-11-17](../_review/review-2026-11-17.md), Q11, Q14. Two pages return `null` on a failed or pending read, so a mistyped id shows an empty page; approve does not move focus as reject does.

**Satisfies:** [PRD §5](../PRD.md) AC-10, AC-15, AC-21 · **Blocked by:** —

## What

- `RequestFallback` and `PageSkeleton`, as on `application-detail.tsx`.
- Approve `onSuccess` sets the focus flag, as reject does.

## Definition of Done

- [ ] Acceptance: Component tests show a skeleton while loading and `RequestFallback` with a retry on a 404, a 5xx and a network error on the review page, the competing item and the associations list, and focus on the announced outcome after an approval.
- [ ] the two addresses are rows of the `pages` table in `page-states.test.tsx`
- [ ] `decision-panel.test.tsx`: `toHaveFocus()` after approval
- [ ] `mise run lint` and `mise run test` green
