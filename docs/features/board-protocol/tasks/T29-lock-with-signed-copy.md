---
id: T29
title: "Lock with the signed scan and meeting facts, and apply Drive restrictions"
layer: "ports"
deps: ["T19", "T28"]
acs: ["AC-14", "AC-34", "AC-15", "AC-40", "AC-16", "AC-17", "AC-41"]
files_hint:
  - "services/core/internal/governance/app/signing.go"
  - "services/core/internal/governance/adapters/postgres/meeting_facts.go"
  - "services/core/internal/governance/adapters/postgres/snapshots.go"
  - "services/core/internal/governance/adapters/postgres/queries/meeting_facts.sql"
  - "services/core/internal/governance/adapters/postgres/queries/snapshots.sql"
  - "services/core/internal/governance/adapters/http/signed_copy.go"
  - "services/core/internal/governance/jobs/apply_drive_restrictions.go"
owner: "Andrii Opryshko"
estimate: "L"
status: "todo"
---

# T29 — Lock with the signed scan and meeting facts, and apply Drive restrictions

## Why

[ADR-0018](../adr/0018-enter-meeting-facts-with-the-signed-scan-and-refuse-lock-without-them.md), [ADR-0007](../adr/0007-lock-protocols-as-fingerprinted-immutable-snapshots.md), [sad §6](../sad.md) Flows 5 and 6; meeting facts never pre-filled from the poll (AC-41).

**Satisfies:** [PRD §5](../PRD.md) AC-14, AC-34, AC-15, AC-40, AC-16, AC-17, AC-41 · **Blocked by:** [T19](./T19-meeting-facts-snapshot-domain.md), [T28](./T28-ready-for-signing-print.md)

## What

- multipart scan + facts in one request; snapshot + registry entry insert; job sets Drive read-only.
- Ended term is refused by T11 middleware (AC-16) — covered by a test here.

## Definition of Done

- [ ] Acceptance: Handler tests show lockProtocolWithSignedCopy locks only from готовий до підпису (AC-40), refuses missing facts naming all (AC-34) or corrections without a note (AC-15), records author and time, writes one `signed` snapshot with fingerprint in the same tx and enqueues apply_drive_restrictions; any later change gets 409 `governance.protocol_locked` (AC-17).
- [ ] test: snapshot fingerprint recomputed from the row equals the stored one
- [ ] job test with fake Files: read-only set, idempotent
- [ ] `mise run lint` and `mise run test` green