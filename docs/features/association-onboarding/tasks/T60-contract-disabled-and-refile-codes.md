---
id: T60
title: "Advertise the two missing problem codes in the contract and reconcile the sync report"
layer: "wiring"
deps: []
acs: ["AC-06", "AC-25"]
files_hint:
  - "services/core/api/openapi.yaml"
  - "docs/features/association-onboarding/contracts/openapi.yaml"
  - "docs/features/association-onboarding/contracts/api-sync-report.md"
  - "services/core/internal/identity/applications_follow_test.go"
  - "services/core/internal/governance/adapters/http/gen/"
  - "services/core/internal/identity/adapters/http/gen/"
  - "apps/web/src/api/gen/"
owner: "Andrii Opryshko"
estimate: "M"
status: "done"
---

# T60 — Advertise the two missing problem codes in the contract and reconcile the sync report

## Why

[Review of 2026-11-10](../_review/review-2026-11-10.md), F17, F18, F21. `.claude/rules/openapi.md`: every operation outside the registry reads answers the shared `MemberOrDisabled` 403, and one code in the contract is one code in the registry. `castPollVote` lists no `identity.association_disabled`; `refileApplication` can answer `identity.application_already_open`, which it does not list.

**Satisfies:** [PRD §5](../PRD.md) AC-06, AC-25 · **Blocked by:** —

## What

- Add the example and the description to the inline 403 of `castPollVote`.
- Reproduce the refile case with a test (a new заявка for the same code filed after a rejection, then a refile of the old one); add the 409 example to the delta and the live contract.
- `mise run gen`, then `/sdlc:api-forge association-onboarding --reconcile`.

## Definition of Done

- [ ] Acceptance: The live contract and the feature delta list `identity.association_disabled` on `castPollVote` and `identity.application_already_open` on `refileApplication`, a test shows the second one answered, and `api-sync-report.md` is re-run with a PASS verdict.
- [ ] `npx @stoplight/spectral-cli lint` on the delta
- [ ] `mise run gen` leaves no diff after a second run
- [ ] the report's base commit, its «sentinels not written yet» deferral and §F are corrected
- [ ] `mise run lint` and `mise run test` green

## Notes

If the refile test shows another code or no conflict, record that in the review record instead of changing the contract.
