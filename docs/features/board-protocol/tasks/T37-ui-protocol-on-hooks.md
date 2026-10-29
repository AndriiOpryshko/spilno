---
id: T37
title: "Move the protocol screen to generated hooks: agenda, slots and revision conflicts"
layer: "ui"
deps: ["T07"]
acs: ["AC-03", "AC-03c", "AC-21", "AC-28b", "AC-09", "AC-29", "AC-04", "AC-22", "AC-28"]
files_hint:
  - "apps/web/src/features/governance/protocol-screen.tsx"
  - "apps/web/src/features/governance/slot-row.tsx"
  - "apps/web/src/features/governance/api.ts"
  - "apps/web/src/features/governance/mock-data.ts"
  - "apps/web/src/features/governance/types.ts"
  - "apps/web/src/routes/protocols.$protocolId.tsx"
  - "apps/web/src/routes/cases.$caseId.protocol.tsx"
  - "apps/web/src/features/governance/meeting-facts.tsx"
owner: "Andrii Opryshko"
estimate: "L"
status: "todo"
---

# T37 — Move the protocol screen to generated hooks: agenda, slots and revision conflicts

## Why

design-spec (docs/design/design-spec.md) already built the lifecycle components on the in-memory api; [sad §4](../sad.md) UI architecture (generated hooks, polling).

**Satisfies:** [PRD §5](../PRD.md) AC-03, AC-03c, AC-21, AC-28b, AC-09, AC-29, AC-04, AC-22, AC-28 · **Blocked by:** [T07](./T07-merge-contract-identity-skeleton.md)

## What

- Replace the protocol part of `api.ts`/`mock-data.ts` with generated hooks; route `protocols/$protocolId` per [sad §5](../sad.md), old route redirects.
- Reuse SlotRow, OriginTag, StatusGlyph, ReadinessCard, LifecycleStepper, ProtocolStateBadge.
- Alternative values of a slot with their source documents (AC-22); the головуючий from the meeting record next to the meeting date (AC-28, review finding 2).

## Definition of Done

- [ ] Acceptance: Component tests show the screen renders meeting date, roster in force, documents and agenda items from getProtocol, confirm/correct slot call the API with `protocol_revision`, a 409 `draft_revision_conflict` reloads with a notice, and a slot with an alternative value shows both (AC-28b). The amount slot lists the other values with their documents next to the recommended one (AC-22), and the screen shows the головуючий from `meeting.chair_board_member_id`, «не вказано» when null (review findings 2 and 6).
- [ ] Vitest tests for load, confirm, correct, conflict
- [ ] existing stories updated to the generated types still pass `test:stories`
- [ ] component test: an amount slot with three values renders the recommended one plus the other two with document links (AC-22)
- [ ] component test: the головуючий name is rendered from the roster for `chair_board_member_id`; null renders «не вказано»
- [ ] `mise run lint` and `mise run test` green

## Notes

Shares `protocol-screen.tsx` with T38–T42 — they depend on this task.
