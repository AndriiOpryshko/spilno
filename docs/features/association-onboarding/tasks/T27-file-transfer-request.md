---
id: T27
title: "File a запит на передачу for a connected association"
layer: "ports"
deps: ["T14", "T22"]
acs: ["AC-02", "AC-17", "AC-18", "AC-19c", "AC-24b", "AC-30"]
files_hint:
  - "services/core/internal/identity/app/"
  - "services/core/internal/identity/ports/ports.go"
  - "services/core/internal/identity/adapters/postgres/"
  - "services/core/internal/identity/adapters/http/"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T27 — File a запит на передачу for a connected association

## Why

[sad](../sad.md) §6 Flow 6; [PRD](../PRD.md) AC-17 as clarified 2026-11-10: any signed-in person may file, on the roster or not.

**Satisfies:** [PRD §5](../PRD.md) AC-02, AC-17, AC-18, AC-19c, AC-24b, AC-30 · **Blocked by:** [T14](./T14-roster-read-facade.md), [T22](./T22-file-connection-application.md)

## What

- `fileApplication` with `kind = transfer`: the validation of T08 for its fields → the association by ЄДРПОУ (`connected` or `disabled`), or by `association_id` for the association set up by hand, whose member files from the card with the code typed in the запит → none is `identity.association_not_connected` → insert the extract, the application with its association and the filing.
- Query `ListTransferRequestsOfAssociation`; the open-application rule of AC-04b applies per account and code.

## Definition of Done

- [ ] Acceptance: fileApplication with kind transfer records a pending запит attached to the association found by its code or by the card's id, reveals nobody of its roster, and refuses a code that is not connected.
- [ ] AC-17: filed by a person off the roster and by a current member → 201 `pending`; the body names no roster member
- [ ] AC-18: an unknown code → 409 `identity.association_not_connected`; nothing stored
- [ ] AC-19c: with the card's `association_id` and a code the association does not have yet → attached to that association
- [ ] AC-24b: a disabled association is accepted; AC-02 and AC-30 hold for the запит's own fields
- [ ] `mise run lint` and `mise run test` green

## Notes

Default taken here, not in the PRD: an `association_id` sent by a person who is not a current member of that association answers `identity.association_not_connected`, so the id reveals nothing.
