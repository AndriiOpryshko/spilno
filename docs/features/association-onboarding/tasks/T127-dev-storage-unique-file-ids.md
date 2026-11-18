---
id: T127
title: "Give the development storage file ids that are unique across associations"
layer: "infra"
deps: []
acs: ["AC-26"]
files_hint:
  - "services/core/internal/governance/adapters/filesfake/fake.go"
  - "services/core/internal/governance/adapters/filesfake/fake_test.go"
  - "services/core/internal/governance/module.go"
  - "services/core/internal/governance/files_per_association_test.go"
  - "apps/web/e2e/association-support.ts"
owner: "Andrii Opryshko"
estimate: "S"
status: "todo"
---

# T127 — Give the development storage file ids that are unique across associations

## Why

[Second re-review of 2026-11-17](../_review/review-2026-11-17-2.md), S07. Since T92 every association has its own in-memory storage that numbers files from 1, while `files.drive_file_id` is unique across associations: the second association that uploads under `SPILNO_DEV_STORAGE` collides and the upload is answered as a failure. This is the AC-26 e2e failure on a reused database that the previous record left without a cause.

**Satisfies:** [PRD §5](../PRD.md) AC-26 · **Blocked by:** —

## What

- The fake's ids are unique across instances (a UUID).
- `attachDocument` in the e2e waits for the upload's answer and asserts 201, so a failure names its status.

## Definition of Done

- [ ] Acceptance: Two associations that each upload their first document under the development storage both get 201 and download their own bytes, and the e2e helper fails with the response status when an upload is refused.
- [ ] `files_per_association_test.go`: two associations, first upload each, both recorded
- [ ] the fake's own test: ids differ across instances
- [ ] `association-support.ts`: the helper asserts the 201
- [ ] `mise run lint` and `mise run test` green

## Notes

Production ids come from Drive and are unique. The second e2e run on one database is in [T133](./T133-e2e-third-review-fixes.md).
