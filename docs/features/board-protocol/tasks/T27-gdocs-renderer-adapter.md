---
id: T27
title: "Render the print copy from a pinned Google Docs template, with a fake Renderer"
layer: "infra"
deps: ["T12"]
acs: ["AC-14c", "AC-41", "AC-42"]
files_hint:
  - "services/core/internal/governance/ports/renderer.go"
  - "services/core/internal/governance/adapters/gdocs/"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T27 — Render the print copy from a pinned Google Docs template, with a fake Renderer

## Why

[ADR-0013](../adr/0013-render-signing-copies-from-google-docs-templates.md) (Docs renders only, never edits; PDF exported to Drive with sha256); [sad §11](../sad.md) (fake Renderer for tests, one smoke test against a test account).

**Satisfies:** [PRD §5](../PRD.md) AC-14c, AC-41, AC-42 · **Blocked by:** [T12](./T12-google-connection-token-source.md)

## What

- `Renderer` port; Docs adapter: copy template → batchUpdate → export PDF → upload via Files.
- Fake Renderer for tests and e2e.

## Definition of Done

- [ ] Acceptance: A unit test checks the Docs batchUpdate requests fill the frozen text, meeting date, meeting form, signers and number and leave blank fields for members present and vote counts with no poll result, and a fake Renderer produces a deterministic PDF for tests.
- [ ] unit test on the generated batchUpdate requests
- [ ] smoke test against a test Google account behind an env flag
- [ ] `mise run lint` and `mise run test` green