---
id: T64
title: "Serve file downloads, and only signed copies while the association is read-only"
layer: "ports"
deps: ["T58", "T63"]
acs: ["AC-25", "AC-26"]
files_hint:
  - "services/core/internal/governance/app/files.go"
  - "services/core/internal/governance/adapters/postgres/files.go"
  - "services/core/internal/governance/adapters/postgres/queries/files.sql"
  - "services/core/internal/governance/adapters/http/files.go"
  - "services/core/internal/governance/adapters/http/stubs.go"
  - "services/core/internal/governance/module.go"
  - "services/core/cmd/api/main.go"
  - "services/core/internal/governance/adapters/filesfake/"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T64 — Serve file downloads, and only signed copies while the association is read-only

## Why

[Review of 2026-11-10](../_review/review-2026-11-10.md), F15. [ADR-0007](../adr/0007-enforce-a-disabled-association-as-read-only-in-the-identity-middleware.md) consequence: governance checks that a file downloaded with `read_only` set is pinned by a signed snapshot. `DownloadFile` answers 501 and nothing reads `Actor.ReadOnly`.

**Satisfies:** [PRD §5](../PRD.md) AC-25, AC-26 · **Blocked by:** [T58](./T58-storage-check-injection-and-pending-gauge.md), [T63](./T63-serve-registry-read.md)

## What

- `DownloadFile` streams through the `Files` port with `Content-Disposition: attachment` (`mime.FormatMediaType`) and `nosniff`.
- With `identity.ActorFrom(ctx).ReadOnly` set, a file without a signed snapshot answers 403 `identity.association_disabled`.
- `cmd/api` wires a `Files`: the in-memory fake under `SPILNO_DEV_STORAGE=true`; without it the operation answers 503 `governance.storage_unavailable` until the Drive adapter of board-protocol T13 exists.

## Definition of Done

- [ ] Acceptance: A member downloads the signed copy of a signed protocol as an attachment in an enabled and in a вимкнене ОСББ, and a file that no signed snapshot pins is refused with «Доступ призупинено» while the association is disabled.
- [ ] handler tests with the fake `Files`: signed copy served, sha256 checked, no storage id on the wire
- [ ] read-only actor and an unsigned file: 403, nothing streamed
- [ ] another association's file id answers 404
- [ ] `mise run lint` and `mise run test` green

## Notes

Carries the `downloadFile` part of board-protocol T21.
