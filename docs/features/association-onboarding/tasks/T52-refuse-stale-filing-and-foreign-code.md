---
id: T52
title: "Refuse a decision on a stale filing and a card transfer that claims another association's code"
layer: "ports"
deps: []
acs: ["AC-15", "AC-19c"]
files_hint:
  - "services/core/internal/identity/adapters/postgres/rejection.go"
  - "services/core/internal/identity/adapters/postgres/queries/rejection.sql"
  - "services/core/internal/identity/adapters/postgres/applications.go"
  - "services/core/internal/identity/adapters/postgres/approval.go"
  - "services/core/internal/identity/rejection_test.go"
  - "services/core/internal/identity/applications_transfer_test.go"
  - "services/core/internal/identity/approval_transfer_test.go"
owner: "Andrii Opryshko"
estimate: "M"
status: "done"
---

# T52 — Refuse a decision on a stale filing and a card transfer that claims another association's code

## Why

[Review of 2026-11-10](../_review/review-2026-11-10.md), F01 (backend half), B3. [PRD](../PRD.md) AC-15 and §10 «Refused transitions»: a refused decision answers a 409 with its code, never a 500. A reject of an earlier filing hits `uq_application_decisions_filing_id`, and a card transfer with another association's ЄДРПОУ hits `uq_associations_edrpou` on approval.

**Satisfies:** [PRD §5](../PRD.md) AC-15, AC-19c · **Blocked by:** —

## What

- Reject reads the latest filing under the lock and refuses a stale `filing_id` the way approve does (`FilingIsLatest`).
- `fileFromCard` refuses a typed code that another association holds; `SetAssociationEdrpou` maps a unique violation to `ports.ErrAssociationConnected` as a second line.
- A unique violation on the decision insert maps to `application_not_pending`.

## Definition of Done

- [ ] Acceptance: A reject that names an earlier filing answers 409 `identity.application_not_pending`, and a card transfer that names a code held by another association is refused at filing with `identity.association_already_connected`; neither path can answer 500.
- [ ] reject with the first filing's id after a refile: 409, nothing recorded
- [ ] card transfer with a foreign code: 409 at filing; an approval that still meets the unique key answers 409
- [ ] the B3 scenario is first reproduced by a failing test, since the review did not verify it
- [ ] `mise run lint` and `mise run test` green

## Notes

The web half of F01 is T69.
