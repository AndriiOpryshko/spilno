---
id: T23
title: "Let the applicant follow and refile their application"
layer: "ports"
deps: ["T09", "T22"]
acs: ["AC-05", "AC-06", "AC-07", "AC-14"]
files_hint:
  - "services/core/internal/identity/app/"
  - "services/core/internal/identity/ports/ports.go"
  - "services/core/internal/identity/adapters/postgres/"
  - "services/core/internal/identity/adapters/http/"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T23 — Let the applicant follow and refile their application

## Why

[sad](../sad.md) §6 Flow 5; [PRD](../PRD.md) AC-05, AC-06; [api-sync-report](../contracts/api-sync-report.md): another person's application answers as not existing.

**Satisfies:** [PRD §5](../PRD.md) AC-05, AC-06, AC-07, AC-14 · **Blocked by:** [T09](./T09-application-lifecycle-domain.md), [T22](./T22-file-connection-application.md)

## What

- `listMyApplications` (cursor), `getMyApplication` (filings newest first, decisions with their reasons), `refileApplication` (T09 `Refile`, the checks of filing again, a new extract and filing, state back to `pending`), `downloadMyExtract` (attachment; the sha256 is compared with the stored one).
- Queries `ListApplicationsOfAccount`, `GetApplicationOfAccount`, `ListFilings`, `GetLatestFiling`, `GetExtract`, `UpdateApplicationState`.

## Definition of Done

- [ ] Acceptance: The applicant lists and opens only their own applications with status, filing date, history and rejection reason, refiles a rejected one into pending with the old filing kept, and downloads their own extract.
- [ ] AC-05: a pending заявка shows its state and filing date; no answer carries a field of the association it names
- [ ] AC-06: after a refile the same application is `pending` with the new filing date; the rejected filing, its extract and its reason are still in the history; the earlier check does not count for the new filing
- [ ] a refile of a pending or approved application → 409 `identity.application_not_pending`; a refile when the code was connected meanwhile → 409 and the application stays `rejected`
- [ ] another person's application or extract id → the same 404 as a missing id
- [ ] `mise run lint` and `mise run test` green

## Notes

The оператор's contact address is not in the API (api-sync-report §E); T37 shows it from a build setting.
