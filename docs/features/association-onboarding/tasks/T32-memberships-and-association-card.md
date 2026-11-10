---
id: T32
title: "Serve my associations and the association card with the confirmation chain"
layer: "ports"
deps: ["T19", "T26"]
acs: ["AC-04d", "AC-19c", "AC-20", "AC-29"]
files_hint:
  - "services/core/internal/identity/app/"
  - "services/core/internal/identity/ports/ports.go"
  - "services/core/internal/identity/adapters/postgres/"
  - "services/core/internal/identity/adapters/http/"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T32 — Serve my associations and the association card with the confirmation chain

## Why

[sad](../sad.md) §6 Flow 12; [ADR-0004](../adr/0004-let-any-verified-person-sign-in-and-derive-memberships-from-rosters.md) (memberships from rosters); [PRD](../PRD.md) AC-29; [api-sync-report](../contracts/api-sync-report.md) §E «Roster marks»: the web matches `BoardMember.decision_id` to the card's history.

**Satisfies:** [PRD §5](../PRD.md) AC-04d, AC-19c, AC-20, AC-29 · **Blocked by:** [T19](./T19-association-and-operator-guards.md), [T26](./T26-approve-connection-application.md)

## What

- `listMemberships`: `Membership.MembershipsOf` joined with the association's name and state.
- `getAssociationCard`: the current голова, the confirmation (decision → check → extract reference), the history of approvals of запити and of state changes with reasons; no confirmation means «підключено вручну, без витягу».
- `downloadConfirmationExtract`: for members of that association, sha256 compared. Query `ListDecisionsOfAssociation`.

## Definition of Done

- [ ] Acceptance: listMemberships returns every association whose roster has the person today, and the card shows the current голова, the extract with its date and signer, the decision date and every later transfer and disabling, or the hand-made mark, with the extract downloadable by members only.
- [ ] a person on A and B gets both, each with their position and the association's state; a person on no roster gets an empty page
- [ ] AC-29: the card of an approved association has the голова, the extract date, the signer and the decision date; a transfer and a disabling appear in the history with their dates and reason
- [ ] AC-19c: the hand-made association has no confirmation until its запит is approved
- [ ] AC-20: a history event of a transfer carries the decision id, the filing date and the extract id
- [ ] a member of B asking for an extract of A → 403 as for any path of A
- [ ] `mise run lint` and `mise run test` green

## Notes

The roster list itself (`listBoardMembers`) is board-protocol T10; it only has to return `decision_id` and `position`.
