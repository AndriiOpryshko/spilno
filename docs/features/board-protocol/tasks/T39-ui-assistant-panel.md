---
id: T39
title: "Add the помічник протоколу conversation panel with job polling"
layer: "ui"
deps: ["T38"]
acs: ["AC-26", "AC-03c"]
files_hint:
  - "apps/web/src/features/governance/assistant-panel.tsx"
  - "apps/web/src/features/governance/assistant-panel.stories.tsx"
  - "apps/web/src/features/governance/protocol-screen.tsx"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T39 — Add the помічник протоколу conversation panel with job polling

## Why

[sad §4](../sad.md) (polling, no server push), [sad §6](../sad.md) Flows 1 and 8.

**Satisfies:** [PRD §5](../PRD.md) AC-26, AC-03c · **Blocked by:** [T38](./T38-ui-text-editor-slot-chips.md)

## What

- Panel with turns from listAssistantTurns; TanStack Query refetchInterval while pending.

## Definition of Done

- [ ] Acceptance: Component tests show sending a request with attached documents calls createAssistantTurn, the panel polls getJob every ~2 s until it settles and then refetches the protocol, a failed job offers retryJob, and stories pass the a11y check.
- [ ] component tests for pending → done and failed → retry
- [ ] stories for empty, pending, failed
- [ ] `mise run lint` and `mise run test` green