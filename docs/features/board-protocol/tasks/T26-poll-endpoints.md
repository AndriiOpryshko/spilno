---
id: T26
title: "Start the попереднє голосування, cast own votes and return to draft"
layer: "ports"
deps: ["T04", "T18", "T22"]
acs: ["AC-30", "AC-31", "AC-41", "AC-05d", "AC-05e", "AC-32"]
files_hint:
  - "services/core/internal/governance/app/poll.go"
  - "services/core/internal/governance/adapters/postgres/poll.go"
  - "services/core/internal/governance/adapters/postgres/queries/poll.sql"
  - "services/core/internal/governance/adapters/http/poll.go"
  - "services/core/internal/governance/adapters/http/return_to_draft.go"
owner: "Andrii Opryshko"
estimate: "L"
status: "todo"
---

# T26 — Start the попереднє голосування, cast own votes and return to draft

## Why

[sad §6](../sad.md) Flows 2 and 3; [ADR-0009](../adr/0009-track-draft-lifecycle-with-states-and-revision-numbers.md).

**Satisfies:** [PRD §5](../PRD.md) AC-30, AC-31, AC-41, AC-05d, AC-05e, AC-32 · **Blocked by:** [T04](./T04-stage-assistant-poll-river-migrations.md), [T18](./T18-poll-domain.md), [T22](./T22-open-draft-endpoints.md)

## What

- Poll repository; handlers for startPollRound, castPollVote, returnProtocolToDraft.

## Definition of Done

- [ ] Acceptance: Handler tests show startPollRound freezes round N (AC-30), castPollVote accepts only the caller's vote (403 `governance.vote_on_behalf`, AC-05e) and only while готовий до голосування (409 `governance.poll_not_open`, AC-32), the tally has no verdict and carries the informal label (AC-31, AC-41), and returnProtocolToDraft closes the round (AC-05d).
- [ ] test: round 3 after a return opens with no votes and round 2 votes remain readable
- [ ] `mise run lint` and `mise run test` green