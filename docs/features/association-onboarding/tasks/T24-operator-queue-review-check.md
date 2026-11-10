---
id: T24
title: "Serve the оператор's queue and review page, and record the check of the extract"
layer: "ports"
deps: ["T10", "T19", "T22"]
acs: ["AC-04c", "AC-10", "AC-11", "AC-11b", "AC-16"]
files_hint:
  - "services/core/internal/identity/app/"
  - "services/core/internal/identity/ports/ports.go"
  - "services/core/internal/identity/adapters/postgres/"
  - "services/core/internal/identity/adapters/http/"
  - "services/core/internal/identity/identitytest/"
owner: "Andrii Opryshko"
estimate: "L"
status: "todo"
---

# T24 — Serve the оператор's queue and review page, and record the check of the extract

## Why

[sad](../sad.md) §6 Flows 2 and 7; [ADR-0005](../adr/0005-check-the-signed-extract-manually-and-keep-the-uploaded-file-in-identity.md): the оператор records what they checked outside Спільно; [PRD](../PRD.md) AC-10 (every value names its source).

**Satisfies:** [PRD §5](../PRD.md) AC-04c, AC-10, AC-11, AC-11b, AC-16 · **Blocked by:** [T10](./T10-name-match-extract-age.md), [T19](./T19-association-and-operator-guards.md), [T22](./T22-file-connection-application.md)

## What

- `listApplicationsForReview` (state filter, oldest first), `getApplicationForReview` (application, applicant's address, filings, checks, decisions, warnings from T10, ids of other pending applications for the same ЄДРПОУ), `downloadExtractForReview`, `recordExtractCheck` (insert-only; the latest check of a filing counts).
- Queries `ListPendingApplications`, `GetApplication`, `ListCompetingApplications`, `CreateCheck`, `GetLatestCheck`.
- Gauge `identity.applications.pending_age_days` (the oldest pending); fixtures `identitytest.PendingApplication` and `identitytest.Check`.

## Definition of Done

- [ ] Acceptance: The оператор lists pending applications, opens one with typed values, recorded checks, warnings and competing applications, downloads its extract and records a check as a new row; nobody else gets any of it.
- [ ] before a check `checks` is empty; after it the check carries the five recorded values, who checked and when
- [ ] AC-11: the warning names the differing values, on recording and on reopening, and blocks nothing; AC-11b: 31 days → warning, 30 → none
- [ ] AC-04c: two pending заявки for one code each list the other as competing
- [ ] AC-16: a член правління, an applicant and a person with nothing get 403 `identity.not_operator` on all four operations and learn nothing of the queue
- [ ] a corrected check is a second row; the first is unchanged
- [ ] `mise run lint` and `mise run test` green

## Notes

Recording a check of one's own filing is allowed (ADR-0009); deciding it is not (T25, T26). Operator use cases read no association-owned table.
