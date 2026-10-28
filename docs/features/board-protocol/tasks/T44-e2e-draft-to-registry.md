---
id: T44
title: "Prove the flow end to end: draft to registry through the UI with fake Google and Gemini"
layer: "tests"
deps: ["T25", "T28", "T29", "T30", "T31", "T32", "T34", "T35", "T36", "T42", "T43"]
acs: ["AC-26", "AC-30", "AC-14c", "AC-14", "AC-18", "AC-19", "AC-35"]
files_hint:
  - "apps/web/e2e/board-protocol.spec.ts"
  - "services/core/internal/platform/config/"
  - "mprocs.yaml"
owner: "Andrii Opryshko"
estimate: "L"
status: "todo"
---

# T44 — Prove the flow end to end: draft to registry through the UI with fake Google and Gemini

## Why

Closes the Stage 1 lifecycle (PRD §1); [sad §6](../sad.md) Flows 1–9, 13.

**Satisfies:** [PRD §5](../PRD.md) AC-26, AC-30, AC-14c, AC-14, AC-18, AC-19, AC-35 · **Blocked by:** [T25](./T25-assistant-propose-draft.md), [T28](./T28-ready-for-signing-print.md), [T29](./T29-lock-with-signed-copy.md), [T30](./T30-withdraw-and-copy.md), [T31](./T31-registry-search.md), [T32](./T32-nightly-integrity-check.md), [T34](./T34-ui-sign-in-session.md), [T35](./T35-ui-settings-roster.md), [T36](./T36-ui-cases-documents.md), [T42](./T42-ui-signed-copy-form.md), [T43](./T43-ui-registry.md)

## What

- Dev config selects the fake TextModel, Files and Renderer; one Playwright spec over the whole path.

## Definition of Done

- [ ] Acceptance: `mise run e2e` signs in with the dev shim, fills settings, opens a case, gets a proposed draft, confirms slots, polls, prints, uploads the signed scan, finds the protocol in the registry, withdraws a second draft, and a signed-out visit is refused.
- [ ] `mise run e2e` green; the run is part of «green» (CLAUDE.md)
- [ ] `mise run lint` and `mise run test` green

## Notes

Fakes are dev-only; a test asserts production config selects the real adapters.
