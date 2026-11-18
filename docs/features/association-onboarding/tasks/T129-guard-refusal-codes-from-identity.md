---
id: T129
title: "Take the guard's refusal codes from identity through cmd/api and drop the mirror"
layer: "wiring"
deps: ["T112"]
acs: ["AC-25", "AC-26"]
files_hint:
  - "services/core/cmd/api/main.go"
  - "services/core/internal/governance/module.go"
  - "services/core/internal/governance/domain/errors.go"
  - "services/core/internal/governance/adapters/http/files.go"
  - "services/core/internal/governance/adapters/http/handler.go"
  - "services/core/internal/governance/guard_codes_mirror_test.go"
  - "services/core/internal/governance/download_contract_test.go"
  - "services/core/internal/governance/file_download_actor_test.go"
owner: "Andrii Opryshko"
estimate: "M"
status: "done"
---

# T129 — Take the guard's refusal codes from identity through cmd/api and drop the mirror

## Why

[Second re-review of 2026-11-17](../_review/review-2026-11-17-2.md), R09. `identity.association_disabled` and `identity.not_current_member` are defined in identity and again as literals in governance; the mirror test only keeps the copies equal. The rule is one definition per code, and governance may not import identity.

**Satisfies:** [PRD §5](../PRD.md) AC-25, AC-26 · **Blocked by:** [T112](./T112-wire-consent-drive-address-and-test-api-wiring.md)

## What

- The access lookup `cmd/api` injects into governance carries the code to answer with (identity's constants are named only in `cmd/api`).
- The two constants and `TestGuardCodesMirrorTheIdentityModule` are removed; the literal scan of T90 keeps covering governance.

## Definition of Done

- [ ] Acceptance: No file under `internal/governance` outside its tests spells an `identity.*` code, and `downloadFile` still answers `identity.not_current_member` for a person off the roster or a request without an actor and `identity.association_disabled` for a file no signed snapshot pins in a вимкнене ОСББ.
- [ ] `download_contract_test.go` and `file_download_actor_test.go` pass with the codes read from the wire
- [ ] a scan test: no `"identity.` literal in governance sources
- [ ] the mirror test is gone; the tracker's known gap is removed
- [ ] `mise run lint` and `mise run test` green

## Notes

Shares `cmd/api/main.go` with [T112](./T112-wire-consent-drive-address-and-test-api-wiring.md), hence the order.
