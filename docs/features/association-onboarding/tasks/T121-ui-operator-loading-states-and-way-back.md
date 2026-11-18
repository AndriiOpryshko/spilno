---
id: T121
title: "Show a loading state in the оператор's queue and a way back from a missing protocol or справа"
layer: "ui"
deps: []
acs: ["AC-04c", "AC-26"]
files_hint:
  - "apps/web/src/features/identity/operator/queue.tsx"
  - "apps/web/src/features/identity/operator/review-page.tsx"
  - "apps/web/src/features/identity/shell/page-states.test.tsx"
  - "apps/web/src/features/identity/operator/review-page.test.tsx"
  - "apps/web/src/features/governance/registry/signed-protocol-screen.tsx"
  - "apps/web/src/features/governance/cases-screen.tsx"
  - "apps/web/src/features/governance/registry/registry.test.tsx"
  - "apps/web/src/features/governance/cases.test.tsx"
  - "apps/web/src/i18n/uk.ts"
owner: "Andrii Opryshko"
estimate: "S"
status: "todo"
---

# T121 — Show a loading state in the оператор's queue and a way back from a missing protocol or справа

## Why

[Second re-review of 2026-11-17](../_review/review-2026-11-17-2.md), S09, S10. The queue renders nothing while it loads; a competing заявка that is still loading puts a page-sized skeleton `<div>` inside the `<ul>`; a 404 for a protocol or a справа offers only «Спробувати ще раз», and the strings that led back have no reader.

**Satisfies:** [PRD §5](../PRD.md) AC-04c, AC-26 · **Blocked by:** —

## What

- The queue returns `PageSkeleton` while pending; the competing item renders its skeleton inside an `<li>`.
- The detail screens pass a link back to the registry or the list as the fallback's action on a 404; strings without a reader are removed from `uk.ts`.

## Definition of Done

- [ ] Acceptance: Component tests show the queue's skeleton while it loads, list markup that holds only `<li>` children while a competing заявка loads, and a link back to the реєстр or the справи list on a 404.
- [ ] `page-states.test.tsx`: the queue row
- [ ] `review-page.test.tsx`: the pending competing item is an `<li>`
- [ ] `registry.test.tsx` and `cases.test.tsx`: the 404 shows the way back
- [ ] no key of `uk.ts` without a reader among the ones the review named
- [ ] `mise run lint` and `mise run test` green
