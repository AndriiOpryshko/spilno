---
id: T72
title: "Show «повноваження завершено» in the card history only when the decision ended a term"
layer: "ports"
deps: ["T51", "T60"]
acs: ["AC-19c", "AC-20"]
files_hint:
  - "services/core/api/openapi.yaml"
  - "docs/features/association-onboarding/contracts/openapi.yaml"
  - "services/core/internal/identity/adapters/postgres/queries/card.sql"
  - "services/core/internal/identity/adapters/postgres/card.go"
  - "services/core/internal/identity/adapters/http/card.go"
  - "services/core/internal/identity/card_test.go"
  - "apps/web/src/features/identity/association/association-card.tsx"
  - "apps/web/src/features/identity/association/association-card.test.tsx"
  - "apps/web/src/features/identity/association/association-card.stories.tsx"
owner: "Andrii Opryshko"
estimate: "M"
status: "done"
---

# T72 — Show «повноваження завершено» in the card history only when the decision ended a term

## Why

[Review of 2026-11-10](../_review/review-2026-11-10.md), F09. [PRD](../PRD.md) AC-19c: on a hand-made association «no other term ends». The card prints both marks on every approved transfer and does not say whose entry each belongs to.

**Satisfies:** [PRD §5](../PRD.md) AC-19c, AC-20 · **Blocked by:** [T51](./T51-carry-term-end-end-previous-head.md), [T60](./T60-contract-disabled-and-refile-codes.md)

## What

- `AssociationEvent` gains the name of the голова whose entry the decision ended, null when none (delta and live contract, `mise run gen`).
- `ListCardHistory` reads it from the roster revision that carries the decision.
- The card renders each mark with its person.

## Definition of Done

- [ ] Acceptance: The card of an association whose transfer ended a голова shows «повноваження завершено» with that person's name, and the card of the AC-19c association shows «призначено» only; a backend test and a component test assert both.
- [ ] `card_test.go`: a transfer with and without an ended entry
- [ ] `association-card.test.tsx`: a `transfer_approved` event of each kind
- [ ] story for the AC-19c card
- [ ] `mise run lint` and `mise run test` green

## Notes

One vertical change across the contract, the query and the card, so the contract and its generated code land together.
