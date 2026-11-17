---
id: T108
title: "Prove the second review's fixes end to end"
layer: "tests"
deps: ["T88", "T89", "T99"]
acs: ["AC-01", "AC-04d", "AC-05", "AC-20", "AC-29"]
files_hint:
  - "apps/web/e2e/support.ts"
  - "apps/web/e2e/access.spec.ts"
  - "apps/web/e2e/us01-file-application.spec.ts"
  - "apps/web/e2e/us03-decide-application.spec.ts"
  - "apps/web/e2e/us05-decide-transfer.spec.ts"
  - "apps/web/e2e/us09-association-card.spec.ts"
  - "apps/web/e2e/transfer-support.ts"
owner: "Andrii Opryshko"
estimate: "M"
status: "done"
---

# T108 — Prove the second review's fixes end to end

## Why

[Review of 2026-11-17](../_review/review-2026-11-17.md), N03, N04, Q17. `fileConnection` types the form's address; the AC-04d test opens two sessions, so it would pass with the choice kept in the session; `today()` is the runner's day while the screens print the Kyiv day.

**Satisfies:** [PRD §5](../PRD.md) AC-01, AC-04d, AC-05, AC-20, AC-29 · **Blocked by:** [T88](./T88-ui-roster-marks-by-effect-and-kind.md), [T89](./T89-ui-entry-to-connection-form.md), [T99](./T99-sign-in-state-in-sealed-cookie.md)

## What

- `fileConnection` clicks from `/` to the form.
- AC-04d: `first.context().newPage()`, as the AC-15 test does.
- `today()` formats with `timeZone: 'Europe/Kyiv'`.
- `us03`: the mark of the голова of a connection. `us05`: edit the previous голова's name on the roster, then read the card.

## Definition of Done

- [ ] Acceptance: `mise run e2e` passes with: the connection form reached by clicking from `/`, AC-04d in two pages of one browser context, dates compared in Europe/Kyiv, the first голова's roster mark worded for a connection, and the card naming the previous голова as ended after their entry was edited.
- [ ] the suite passes with `TZ=America/Los_Angeles`
- [ ] no `page.goto` to `/applications/new` is left in `e2e/`
- [ ] `mise run lint` and `mise run test` green and `mise run e2e` green
