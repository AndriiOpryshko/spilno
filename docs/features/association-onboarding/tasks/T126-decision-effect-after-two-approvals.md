---
id: T126
title: "Prove each roster entry's decision effect after two transfers in a row"
layer: "tests"
deps: []
acs: ["AC-19b", "AC-20"]
files_hint:
  - "services/core/internal/identity/approval_transfer_test.go"
  - "services/core/internal/governance/roster_endpoints_test.go"
owner: "Andrii Opryshko"
estimate: "S"
status: "done"
---

# T126 — Prove each roster entry's decision effect after two transfers in a row

## Why

[Second re-review of 2026-11-17](../_review/review-2026-11-17-2.md), R13. `decision_effect` is asserted only on hand-seeded rows of one decision. The code is right by trace; nothing proves it through the approval path.

**Satisfies:** [PRD §5](../PRD.md) AC-19b, AC-20 · **Blocked by:** —

## What

- One integration test approves a connection and two transfers through `approveApplication`, then reads `listBoardMembers`.
- A second case re-seats an applicant who already has an entry.

## Definition of Done

- [ ] Acceptance: After a connection and two approved transfers the roster read through the API reports the first голова `ended`, the second `ended` and the third `seated`, each with the decision that did it, and an applicant already on the roster is re-seated on their existing entry with `seated`.
- [ ] test: three entries, three effects, three decision ids
- [ ] test: the re-seated applicant has one entry
- [ ] `mise run lint` and `mise run test` green

## Notes

If a case fails, the finding becomes a defect of T86 or T87 and goes back to the review record.
