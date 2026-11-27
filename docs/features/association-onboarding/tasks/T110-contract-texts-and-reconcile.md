---
id: T110
title: "Correct the contract's descriptions and reconcile the sync report"
layer: "wiring"
deps: ["T87", "T90", "T100"]
acs: ["AC-04c", "AC-26"]
files_hint:
  - "services/core/api/openapi.yaml"
  - "docs/features/association-onboarding/contracts/openapi.yaml"
  - "docs/features/association-onboarding/contracts/api-sync-report.md"
owner: "Andrii Opryshko"
estimate: "M"
status: "done"
---

# T110 — Correct the contract's descriptions and reconcile the sync report

## Why

[Review of 2026-11-17](../_review/review-2026-11-17.md), N07, N08. The description says `identity.association_already_connected` is answered only for a заявка на підключення; the report says run 4 changed no shape and that `downloadFile` keeps `NotCurrentMember` on purpose.

**Satisfies:** [PRD §5](../PRD.md) AC-04c, AC-26 · **Blocked by:** [T87](./T87-roster-decision-effect-and-no-reopen.md), [T90](./T90-one-definition-per-problem-code.md), [T100](./T100-competing-within-one-kind.md)

## What

- The two descriptions, in the delta and in the live contract.
- `/sdlc:api-forge association-onboarding --reconcile` after the contract changes of T87 and T90: `ended_head_full_name`, `decision_effect`, the 403 of `downloadFile`.

## Definition of Done

- [ ] Acceptance: The descriptions of `approveApplication` and `fileApplication` name the refusal of a transfer that claims another association's code, and `api-sync-report.md` of a new `--reconcile` run has a field-origin row for every response field, no claim the code contradicts, and `**Verdict:**` `PASS` or `PASS (with deferrals)`.
- [ ] `mise run gen` leaves no diff
- [ ] the report's run number, date and sources are of this run
- [ ] `mise run lint` and `mise run test` green

## Notes

Last in the contract lane: T87, T90 and this task edit `openapi.yaml` and are serialized by their shared `files_hint`.
