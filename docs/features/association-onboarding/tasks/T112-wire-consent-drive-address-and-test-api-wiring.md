---
id: T112
title: "Give the production storage consent its Drive address and test the wiring cmd/api runs"
layer: "wiring"
deps: []
acs: ["AC-26", "AC-27", "AC-28", "AC-28b"]
files_hint:
  - "services/core/cmd/api/main.go"
  - "services/core/cmd/api/google_test.go"
  - "services/core/internal/governance/module.go"
  - "services/core/internal/governance/drive_wiring_test.go"
  - "services/core/internal/platform/google/consent.go"
owner: "Andrii Opryshko"
estimate: "M"
status: "done"
---

# T112 — Give the production storage consent its Drive address and test the wiring cmd/api runs

## Why

[Second re-review of 2026-11-17](../_review/review-2026-11-17-2.md), R01, R02. `wireGoogle` builds `google.Consent` with no `DriveURL`, so a real consent fails after the code exchange and no association ever gets a connection with a folder; the resolver `cmd/api` runs (`driveFiles`) is under no test, `drive_wiring_test.go` re-types its body.

**Satisfies:** [PRD §5](../PRD.md) AC-26, AC-27, AC-28, AC-28b · **Blocked by:** —

## What

- `wireGoogle` passes the configured Drive address into the consent; the consent refuses to be built without one.
- The composition of the resolver (token source, folder id, Drive client) lives in one exported constructor that `cmd/api` and the module test both call.
- A `cmd/api` test completes a consent built by `wireGoogle` against a fake Google (token endpoint, `/about`, `/files`), then uploads and downloads through the resolver `cmd/api` builds.

## Definition of Done

- [ ] Acceptance: A consent built by `wireGoogle` completes against a fake Google and writes a connection with a folder, and a test fails when `cmd/api` builds the consent or the Drive resolver differently from what the test runs.
- [ ] `cmd/api` test: consent built by `wireGoogle` completes; the connection row has the folder id
- [ ] the same test uploads a case document and downloads it back through the resolver of `cmd/api`
- [ ] `drive_wiring_test.go` calls the shared constructor, no copy of its body
- [ ] the tracker's known gap about SEC-13 is reworded: what remains is a run against real Google
- [ ] `mise run lint` and `mise run test` green

## Notes

Still not a run against real Google (no test account, manual consent): that stays board-protocol T13 and a PRD §8 entry of [T132](./T132-record-open-gaps-in-prd.md).
