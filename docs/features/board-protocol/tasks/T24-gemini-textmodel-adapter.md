---
id: T24
title: "Implement the Gemini TextModel adapter as the only importer of the SDK"
layer: "infra"
deps: ["T07", "T17"]
acs: ["AC-26", "AC-21", "AC-22"]
files_hint:
  - "services/core/internal/governance/ports/textmodel.go"
  - "services/core/internal/governance/adapters/gemini/"
  - "services/core/.golangci.yml"
  - "services/core/go.mod"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T24 — Implement the Gemini TextModel adapter as the only importer of the SDK

## Why

[ADR-0004](../adr/0004-use-gemini-api-for-document-reading-and-drafting.md) (paid tier, structured output, people confirm), [sad §5](../sad.md) depguard rule, [sad §11](../sad.md) risks (inline when it fits, delete Files API uploads right after, no content in logs).

**Satisfies:** [PRD §5](../PRD.md) AC-26, AC-21, AC-22 · **Blocked by:** [T07](./T07-merge-contract-identity-skeleton.md), [T17](./T17-text-fragments-proposal-rules.md)

## What

- `TextModel` port; Gemini adapter with a response schema; inline vs Files API with immediate delete.
- Fake `TextModel` returning a scripted proposal for T25 and e2e.

## Definition of Done

- [ ] Acceptance: A unit test maps a recorded structured Gemini answer into a domain proposal (slot values with fragment and document version, text fragments), malformed answers are an error, and the `gemini-sdk-in-adapter` depguard rule fails lint for any other importer.
- [ ] unit test on recorded responses (no network)
- [ ] lint: an import of `google.golang.org/genai` outside `adapters/gemini` fails depguard
- [ ] no document text in logs
- [ ] `mise run lint` and `mise run test` green

## Notes

Shares `.golangci.yml` with T07 (serialized by dep).
