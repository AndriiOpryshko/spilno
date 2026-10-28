---
id: T19
title: "Model meeting facts and the canonical fingerprinted snapshot"
layer: "domain"
deps: ["T16", "T18"]
acs: ["AC-14", "AC-34", "AC-15", "AC-17"]
files_hint:
  - "services/core/internal/governance/domain/meeting_facts.go"
  - "services/core/internal/governance/domain/snapshot.go"
  - "services/core/internal/governance/domain/snapshot_test.go"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T19 — Model meeting facts and the canonical fingerprinted snapshot

## Why

[ADR-0018](../adr/0018-enter-meeting-facts-with-the-signed-scan-and-refuse-lock-without-them.md) (facts entered with the scan, never pre-filled from the poll), [ADR-0007](../adr/0007-lock-protocols-as-fingerprinted-immutable-snapshots.md) (canonical jsonb, `schema_version`, fingerprint).

**Satisfies:** [PRD §5](../PRD.md) AC-14, AC-34, AC-15, AC-17 · **Blocked by:** [T16](./T16-agenda-slots-domain.md), [T18](./T18-poll-domain.md)

## What

- Meeting facts: members present, vote count per agenda item, corrections flag + note.
- Snapshot kinds `signed` / `withdrawn`; canonical serialization (sorted keys) + SHA-256.

## Definition of Done

- [ ] Acceptance: Unit tests prove the lock refuses missing members present or a missing vote count per item naming all of them, requires a correction note when corrections are reported, and the canonical snapshot form yields a stable SHA-256 fingerprint for equal input.
- [ ] unit tests for AC-34 and AC-15
- [ ] golden test: fingerprint is stable across map ordering and changes when any field changes
- [ ] `mise run lint` and `mise run test` green