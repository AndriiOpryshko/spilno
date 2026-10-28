---
id: T40
title: "Make the попереднє голосування live: start, own vote, tally, informal label, return to draft"
layer: "ui"
deps: ["T39"]
acs: ["AC-30", "AC-31", "AC-41", "AC-05d", "AC-05e", "AC-32"]
files_hint:
  - "apps/web/src/features/governance/poll-round-panel.tsx"
  - "apps/web/src/features/governance/transition-actions.tsx"
  - "apps/web/src/features/governance/lifecycle.ts"
  - "apps/web/src/features/governance/protocol-screen.tsx"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T40 — Make the попереднє голосування live: start, own vote, tally, informal label, return to draft

## Why

[sad §6](../sad.md) Flows 2 and 3; reuses PollRoundPanel and TransitionActions from the design-spec work.

**Satisfies:** [PRD §5](../PRD.md) AC-30, AC-31, AC-41, AC-05d, AC-05e, AC-32 · **Blocked by:** [T39](./T39-ui-assistant-panel.md)

## What

- Wire PollRoundPanel and TransitionActions to startPollRound, castPollVote, returnProtocolToDraft.

## Definition of Done

- [ ] Acceptance: Component tests show start poll, cast and change own vote, the tally with members and times, the exact informal label of AC-41 on every poll view, polling of other members' votes, and «повернути в чернетку» closing the round.
- [ ] updated stories pass `test:stories`
- [ ] `mise run lint` and `mise run test` green