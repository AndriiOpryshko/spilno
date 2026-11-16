---
id: T82
title: "Prove the roster, the registry and the document refusal end to end"
layer: "tests"
deps: ["T78", "T79", "T80"]
acs: ["AC-09", "AC-16b", "AC-20", "AC-26", "AC-28"]
files_hint:
  - "apps/web/e2e/access.spec.ts"
  - "apps/web/e2e/us03-decide-application.spec.ts"
  - "apps/web/e2e/us07-read-only-association.spec.ts"
  - "apps/web/e2e/us08-connect-storage.spec.ts"
  - "apps/web/e2e/us09-association-card.spec.ts"
  - "apps/web/e2e/association-support.ts"
  - "apps/web/playwright.config.ts"
owner: "Andrii Opryshko"
estimate: "L"
status: "todo"
---

# T82 — Prove the roster, the registry and the document refusal end to end

## Why

[Review of 2026-11-10](../_review/review-2026-11-10.md), F15. Four e2e cases are `test.fixme`, and the AC-26 case accepts a 501 as «not refused».

**Satisfies:** [PRD §5](../PRD.md) AC-09, AC-16b, AC-20, AC-26, AC-28 · **Blocked by:** [T78](./T78-ui-roster-screen.md), [T79](./T79-ui-registry-screens.md), [T80](./T80-ui-cases-on-api-with-upload.md)

## What

- Replace the four `test.fixme` cases; the AC-26 case asserts 200 and the file bytes.
- A signed protocol is seeded on the isolated stack through its owner connection, with its file in the dev `Files` fake.
- `us09` asserts the marks on the roster entries, not only in the card history.

## Definition of Done

- [ ] Acceptance: `mise run e2e` passes with no `test.fixme` for this feature: the approved голова adds a член правління, an оператор on a roster works as a member there, a member of a вимкнене ОСББ opens the registry and downloads a signed copy whose bytes match, the roster shows the transfer marks, and attaching a document without a storage shows the message.
- [ ] `grep test.fixme apps/web/e2e` is empty
- [ ] the seeding helper cannot run against a database that is not the isolated e2e one
- [ ] `mise run lint` and `mise run test` green; `mise run e2e` green

## Notes

How the fake `Files` is seeded from outside the API process is decided here; a dev-only seed command under `services/core/cmd/` is the fallback.
