---
id: T37
title: "Move the protocol screen to generated hooks: agenda, slots and revision conflicts"
layer: "ui"
deps: ["T07"]
acs: ["AC-03", "AC-03c", "AC-21", "AC-28b", "AC-09", "AC-29", "AC-04"]
files_hint:
  - "apps/web/src/features/governance/protocol-screen.tsx"
  - "apps/web/src/features/governance/slot-row.tsx"
  - "apps/web/src/features/governance/api.ts"
  - "apps/web/src/features/governance/mock-data.ts"
  - "apps/web/src/features/governance/types.ts"
  - "apps/web/src/routes/protocols.$protocolId.tsx"
  - "apps/web/src/routes/cases.$caseId.protocol.tsx"
owner: "Andrii Opryshko"
estimate: "L"
status: "todo"
---

# T37 — Move the protocol screen to generated hooks: agenda, slots and revision conflicts

## Why

design-spec (docs/design/design-spec.md) already built the lifecycle components on the in-memory api; [sad §4](../sad.md) UI architecture (generated hooks, polling).

**Satisfies:** [PRD §5](../PRD.md) AC-03, AC-03c, AC-21, AC-28b, AC-09, AC-29, AC-04 · **Blocked by:** [T07](./T07-merge-contract-identity-skeleton.md)

## What

- Replace the protocol part of `api.ts`/`mock-data.ts` with generated hooks; route `protocols/$protocolId` per [sad §5](../sad.md), old route redirects.
- Reuse SlotRow, OriginTag, StatusGlyph, ReadinessCard, LifecycleStepper, ProtocolStateBadge.

## Definition of Done

- [ ] Acceptance: Component tests show the screen renders meeting date, roster in force, documents and agenda items from getProtocol, confirm/correct slot call the API with `protocol_revision`, a 409 `draft_revision_conflict` reloads with a notice, and a slot with an alternative value shows both (AC-28b).
- [ ] Vitest tests for load, confirm, correct, conflict
- [ ] existing stories updated to the generated types still pass `test:stories`
- [ ] `mise run lint` and `mise run test` green

## Notes

Shares `protocol-screen.tsx` with T38–T42 — they depend on this task.
