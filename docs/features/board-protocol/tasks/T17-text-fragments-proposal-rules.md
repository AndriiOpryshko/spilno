---
id: T17
title: "Add text fragments, the fact validator and the amount recommendation rule"
layer: "domain"
deps: ["T16"]
acs: ["AC-28", "AC-22", "AC-21", "AC-28b"]
files_hint:
  - "services/core/internal/governance/domain/text.go"
  - "services/core/internal/governance/domain/factcheck.go"
  - "services/core/internal/governance/domain/proposal.go"
  - "services/core/internal/governance/domain/factcheck_test.go"
  - "services/core/internal/governance/domain/proposal_test.go"
owner: "Andrii Opryshko"
estimate: "L"
status: "todo"
---

# T17 — Add text fragments, the fact validator and the amount recommendation rule

## Why

[ADR-0011](../adr/0011-store-decision-text-as-fragments-with-slot-placeholders.md) (fragments + slot placeholders, deterministic validator); AC-28, AC-22; merging a proposal never overwrites a confirmed slot (AC-28b).

**Satisfies:** [PRD §5](../PRD.md) AC-28, AC-22, AC-21, AC-28b · **Blocked by:** [T16](./T16-agenda-slots-domain.md)

## What

- Fragments = plain text + slot refs + meeting-fact refs (meeting date, головуючий).
- Fact validator: assistant text → extract facts into proposed slots; human text → highlight list.
- Amount rule over document candidates (VAT, estimate vs invoice, advance vs full, doc date not upload date).
- `MergeProposal(protocol, proposal)` keeps confirmed slots and puts new values beside them.

## Definition of Done

- [ ] Acceptance: Unit tests prove an amount, date, document reference, name or count written by the assistant in plain text becomes a new unconfirmed slot, a person-typed fact is only highlighted, and the amount rule picks the latest-dated full-cost document per AC-22 while listing the alternatives.
- [ ] table tests for AC-22 cases (with/without VAT, superseded estimate, advance vs full, no invoice yet)
- [ ] tests for AC-28 (assistant vs human text) and AC-28b (merge)
- [ ] `mise run lint` and `mise run test` green