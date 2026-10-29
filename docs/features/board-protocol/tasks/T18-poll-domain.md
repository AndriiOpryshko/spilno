---
id: T18
title: "Model poll rounds and the own-vote rule"
layer: "domain"
deps: ["T15"]
acs: ["AC-30", "AC-31", "AC-05e", "AC-32", "AC-41", "AC-05d"]
files_hint:
  - "services/core/internal/governance/domain/poll.go"
  - "services/core/internal/governance/domain/poll_test.go"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T18 — Model poll rounds and the own-vote rule

## Why

AC-30, AC-31, AC-05e, AC-32, AC-05d; the poll is informal (AC-41). [sad §6](../sad.md) Flows 2, 3.

**Satisfies:** [PRD §5](../PRD.md) AC-30, AC-31, AC-05e, AC-32, AC-41, AC-05d · **Blocked by:** [T15](./T15-protocol-lifecycle-domain.md)

## What

- Round N with frozen revision; vote = member, choice (for/against/abstain), time; one per member, changeable.
- Tally per agenda item with members and times, a fixed informal label constant.

## Definition of Done

- [ ] Acceptance: Unit tests prove a round freezes the wording at its revision, a member casts and changes only their own vote, voting is refused unless готовий до голосування, a closed round keeps its votes uncounted, and the tally carries no quorum or adoption verdict.
- [ ] domain-pure depguard clean
- [ ] `mise run lint` and `mise run test` green