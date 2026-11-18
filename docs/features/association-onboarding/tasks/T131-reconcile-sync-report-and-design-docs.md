---
id: T131
title: "Reconcile the sync report and write the roster rules into ADR-0006, the SAD and the data model"
layer: "docs"
deps: ["T124", "T129", "T130"]
acs: ["AC-04c", "AC-19", "AC-20", "AC-26"]
files_hint:
  - "docs/features/association-onboarding/contracts/api-sync-report.md"
  - "docs/features/association-onboarding/adr/0006-apply-approvals-to-the-roster-through-a-governance-facade-in-one-transaction.md"
  - "docs/features/association-onboarding/sad.md"
  - "docs/features/association-onboarding/data-model.md"
  - "docs/features/association-onboarding/PRD.md"
  - "services/core/internal/governance/domain/errors.go"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T131 — Reconcile the sync report and write the roster rules into ADR-0006, the SAD and the data model

## Why

[Second re-review of 2026-11-17](../_review/review-2026-11-17-2.md), R08, R11. Run 5 of `api-sync-report.md` predates T98 and T99 and will predate T114 and T130. ADR-0006 and the SAD lifecycle row lack the two roster rules of T86 and T87; `data-model.md` lacks the rows of `RosterDecisionEffects` and `GetEntryOfDecision` and shows `ListCompetingApplications` without `kind`; `governance.drive_not_connected` is still marked «planned»; a comment names a code constant that does not exist.

**Satisfies:** [PRD §5](../PRD.md) AC-04c, AC-19, AC-20, AC-26 · **Blocked by:** [T124](./T124-multipart-operations-outside-guard-transaction.md), [T129](./T129-guard-refusal-codes-from-identity.md), [T130](./T130-contract-cookie-413-term-end-and-decision-effect.md)

## What

- `/sdlc:api-forge association-onboarding --reconcile` (run 6) after the contract tasks.
- ADR-0006: an «Amended 2026-11-17» paragraph for the first-revision rule and the no-reopen rule; one clause in the SAD §8 lifecycle row.
- `data-model.md`: the three query rows; PRD §5.1 and SAD §8: the «planned» markers removed; the comment in `governance/domain/errors.go` corrected.

## Definition of Done

- [ ] Acceptance: `api-sync-report.md` of run 6 has a row for every response field and every problem code at HEAD, no claim the code contradicts and `**Verdict:**` `PASS` or `PASS (with deferrals)`; ADR-0006, SAD §8 and `data-model.md` state the rules and the queries the code runs.
- [ ] run 6 header names the commit it was run on
- [ ] ADR-0006 amendment, SAD §8 clause
- [ ] three query rows in `data-model.md`, each with the index that serves it
- [ ] no «planned» marker for a registered code
- [ ] `mise run lint` and `mise run test` green

## Notes

Last of the contract lane, as T110 was.
