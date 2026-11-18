---
id: T133
title: "Prove the third review's fixes end to end by clicking"
layer: "tests"
deps: ["T116", "T117", "T127"]
acs: ["AC-01", "AC-04d", "AC-05", "AC-16c", "AC-17", "AC-26"]
files_hint:
  - "apps/web/e2e/access.spec.ts"
  - "apps/web/e2e/support.ts"
  - "apps/web/e2e/transfer-support.ts"
  - "apps/web/e2e/us02-follow-and-refile.spec.ts"
  - "apps/web/e2e/us04-file-transfer.spec.ts"
  - "apps/web/e2e/us07-read-only-association.spec.ts"
  - "apps/web/e2e/association-support.ts"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T133 — Prove the third review's fixes end to end by clicking

## Why

[Second re-review of 2026-11-17](../_review/review-2026-11-17-2.md), R03, R04, S07. The AC-04d spec files both заявки before either is approved, so a current член of one association never files for a second; the AC-16c spec holds only while the оператор is on zero or on several rosters; no spec reaches the запит form as a person on the roster.

**Satisfies:** [PRD §5](../PRD.md) AC-01, AC-04d, AC-05, AC-16c, AC-17, AC-26 · **Blocked by:** [T116](./T116-ui-lasting-links-to-applications.md), [T117](./T117-ui-transfer-request-link-for-every-member.md), [T127](./T127-dev-storage-unique-file-ids.md)

## What

- AC-04d in the AC's order: the first заявка is approved, the second is filed from inside the first association by clicking.
- `openConnectionForm` works from the picker, the list of заявки and the association sidebar, so AC-16c no longer depends on the оператор's rosters.
- A person on the roster opens the запит form from the card and files it (AC-17); a member follows their заявка from inside the association (AC-05).

## Definition of Done

- [ ] Acceptance: `mise run e2e` passes on a fresh database with the AC-04d spec filing the second заявка as a current член by clicking and an on-roster запит filed from the card, and the AC-16c and AC-26 specs pass again when run a second time on the same database.
- [ ] the full suite on a fresh throwaway database
- [ ] `--grep` AC-16c and AC-26 a second time on the same database
- [ ] no new `page.goto` except setup; no fixed sleep
- [ ] `mise run lint` and `mise run test` green

## Notes

The suite as a whole keeps its fresh database: the specs read only the first page of the оператор's queue (PRD §8, [T132](./T132-record-open-gaps-in-prd.md)).
