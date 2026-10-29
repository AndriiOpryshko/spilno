---
id: T22
title: "Open a draft for a meeting from a case, get and list protocols"
layer: "ports"
deps: ["T10", "T20", "T21"]
acs: ["AC-42", "AC-03", "AC-03c", "AC-04", "AC-28", "AC-29"]
files_hint:
  - "services/core/internal/governance/app/drafts.go"
  - "services/core/internal/governance/adapters/http/protocols.go"
  - "services/core/internal/governance/adapters/http/agenda_items.go"
  - "services/core/internal/governance/domain/meeting.go"
  - "services/core/internal/governance/adapters/http/handler.go"
  - "services/core/internal/governance/adapters/postgres/queries/meetings.sql"
  - "services/core/internal/governance/adapters/postgres/meetings.go"
owner: "Andrii Opryshko"
estimate: "L"
status: "todo"
---

# T22 — Open a draft for a meeting from a case, get and list protocols

## Why

[sad §6 Flow 1](../sad.md) (load, settings check, one draft per meeting); [ADR-0009](../adr/0009-track-draft-lifecycle-with-states-and-revision-numbers.md).

**Satisfies:** [PRD §5](../PRD.md) AC-42, AC-03, AC-03c, AC-04, AC-28, AC-29 · **Blocked by:** [T10](./T10-settings-roster-endpoints.md), [T20](./T20-protocol-repository.md), [T21](./T21-cases-endpoints.md)

## What

- Use case + handlers for addAgendaItemFromCase, getProtocol, listProtocols (cursor `(created_at, id)`).
- Meeting date and головуючий come from the meeting record.
- Meeting chair path (review finding 2): `domain.Meeting.ChairBoardMemberID`, CreateMeeting / Get / List sqlc queries, `toMeeting` and `getProtocol.meeting.chair_board_member_id` read it.

## Definition of Done

- [ ] Acceptance: Handler tests show addAgendaItemFromCase opens a draft or adds the справа as a new item to the meeting's existing draft (AC-03c), refuses with 409 `governance.settings_missing` naming every missing setting (AC-42), and getProtocol shows meeting date, roster in force, case documents and items (AC-03). createMeeting stores `chair_board_member_id` and getMeeting / listMeetings / getProtocol return it instead of null (review finding 2); an id that is no board member of the association is refused with 422 naming `/chair_board_member_id`.
- [ ] handler tests for AC-42, AC-03, AC-03c
- [ ] p95 read ≤ 200 ms is measured by the request-duration metric (no test gate)
- [ ] handler tests: createMeeting with a chair → getMeeting and getProtocol return it; without one → null; an unknown member id → 422 `/chair_board_member_id` (AC-28, AC-29)
- [ ] `mise run lint` and `mise run test` green
