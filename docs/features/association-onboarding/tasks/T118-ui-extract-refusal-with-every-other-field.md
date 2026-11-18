---
id: T118
title: "Name a refused extract together with every other broken field"
layer: "ui"
deps: []
acs: ["AC-02"]
files_hint:
  - "apps/web/src/features/identity/applications/filing-form.tsx"
  - "apps/web/src/features/identity/applications/filing-form.test.tsx"
owner: "Andrii Opryshko"
estimate: "S"
status: "todo"
---

# T118 — Name a refused extract together with every other broken field

## Why

[Second re-review of 2026-11-17](../_review/review-2026-11-17-2.md), R05. Since T106 a wrong or oversized extract returns from `submit` before any request, so the other broken fields appear only on the next submit. AC-02: every missing or wrong field at once.

**Satisfies:** [PRD §5](../PRD.md) AC-02 · **Blocked by:** —

## What

- When the extract is refused locally the form still sends the other fields without the file, takes the server's field errors and puts its own message under «Витяг».
- A filing without a file cannot be recorded: the extract is required, so the answer is the 422 that names the fields.

## Definition of Done

- [ ] Acceptance: A component test files with an empty ЄДРПОУ and a `.png` extract and sees both fields named after one submit, with the extract's own message, and no request that carries the refused file.
- [ ] test: a bad extract plus a broken text field, both named at once
- [ ] test: a bad extract with valid fields names only «Витяг» and records nothing
- [ ] the tests of T106 still pass
- [ ] `mise run lint` and `mise run test` green

## Notes

The alternative, repeating the server's rules for every field on the client, was not taken: two copies of the validation drift.
