---
id: T45
title: "Prove a transfer, the takeover and the association card end to end (US-04, US-05, US-09)"
layer: "tests"
deps: ["T29", "T41", "T44"]
acs: ["AC-17", "AC-18", "AC-18b", "AC-19", "AC-20", "AC-29"]
files_hint:
  - "apps/web/e2e/"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T45 — Prove a transfer, the takeover and the association card end to end (US-04, US-05, US-09)

## Why

[test-plan](../test-plan.md) rows of US-04, US-05, US-09; [sad](../sad.md) §10 QG-2 seen from the browser.

**Satisfies:** [PRD §5](../PRD.md) AC-17, AC-18, AC-18b, AC-19, AC-20, AC-29 · **Blocked by:** [T29](./T29-approve-transfer-request.md), [T41](./T41-ui-association-card.md), [T44](./T44-e2e-connection-flows.md)

## What

- Flow US-04: file a запит. Flow US-05: approve it; the new голова signs in and opens the association; the previous голова's open window gets the no-access page. Flow US-09: open the card, read the chain, download the extract.
- Refusals for AC-18 and AC-18b; the appointed and ended notes of AC-20 in the card's history.

## Definition of Done

- [ ] Acceptance: mise run e2e passes the flows of US-04, US-05 and US-09: a запит is filed and approved, the new голова opens the association, the previous one gets no access, and the card shows the chain with a downloadable extract.
- [ ] three flow specs green on the harness of T44
- [ ] the previous голова's already open page is refused on its next action, without a new sign-in
- [ ] `mise run lint` and `mise run test` green

## Notes

Shares `apps/web/e2e/` with T44 and T46, so the three run one after another.
