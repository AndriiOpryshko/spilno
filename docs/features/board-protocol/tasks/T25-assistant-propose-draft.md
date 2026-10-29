---
id: T25
title: "Take conversation turns and propose a draft in the propose_draft job"
layer: "app"
deps: ["T04", "T14", "T23", "T24"]
acs: ["AC-26", "AC-21", "AC-22", "AC-28", "AC-28b", "AC-03c"]
files_hint:
  - "services/core/internal/governance/app/assistant.go"
  - "services/core/internal/governance/jobs/propose_draft.go"
  - "services/core/internal/governance/adapters/postgres/assistant_turns.go"
  - "services/core/internal/governance/adapters/postgres/queries/assistant_turns.sql"
  - "services/core/internal/governance/adapters/http/assistant_turns.go"
owner: "Andrii Opryshko"
estimate: "L"
status: "todo"
---

# T25 — Take conversation turns and propose a draft in the propose_draft job

## Why

[sad §6](../sad.md) Flows 1 and 8; [events.md](../contracts/events.md) `governance.propose_draft.v1`; [ADR-0004](../adr/0004-use-gemini-api-for-document-reading-and-drafting.md), [ADR-0006](../adr/0006-run-background-work-in-worker-via-river-queue.md).

**Satisfies:** [PRD §5](../PRD.md) AC-26, AC-21, AC-22, AC-28, AC-28b, AC-03c · **Blocked by:** [T04](./T04-stage-assistant-poll-river-migrations.md), [T14](./T14-river-worker-jobs.md), [T23](./T23-wording-slot-endpoints.md), [T24](./T24-gemini-textmodel-adapter.md)

## What

- createAssistantTurn: store the turn + attached documents, open or reuse the meeting's draft (AC-03c), enqueue in the same tx.
- Job: fetch docs via Files, call TextModel, run fact validator + amount rule + MergeProposal, save with revision guard.
- Proposal-duration metric (p95 ≤ 30 s).

## Definition of Done

- [ ] Acceptance: With the fake TextModel and Files, createAssistantTurn returns a job, the propose_draft handler fills the draft's slots as unconfirmed with fragments and document links (AC-26, AC-21), keeps confirmed slots (AC-28b), is idempotent on retry, and listAssistantTurns returns the conversation.
- [ ] integration test: turn → job → draft with unconfirmed slots
- [ ] test: running the job twice gives the same draft
- [ ] failed TextModel → job `failed`, slots can still be filled by hand (Flow 8)
- [ ] `mise run lint` and `mise run test` green