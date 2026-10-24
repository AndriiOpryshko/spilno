---
status: accepted
date: 2026-10-25
feature: board-protocol
sad_section: "§5"
deciders: ["Andrii Opryshko"]
---

# 0012 Add a funding source to finance contracts and read contracts in force through finance module.go

## Context and problem statement

A decision's джерело коштів links to a contract in force on the meeting date; an ended contract is refused with its end date (AC-10, PRD §9). `finance.contracts` has counterparty, amount, currency and dates but no fund line. PRD §8 asked architecture whether the fund is a field of the contract or a separate fund line; the PRD default is a contract in force with a fund line added.

## Decision drivers

- AC-10: only contracts in force on the meeting date are offered.
- Repo ADR 0002: `governance` must not read `contracts` directly.
- QG-1: a signed protocol must not change when a contract is edited later.

## Blast radius

- Irreversible: no — a column and a read method; a funds table can replace it later.
- Multi-context: yes — `finance` changes its table and its `module.go`; `governance` consumes it.
- Legitimate alternative: yes — a separate `funds` table with a contract-to-fund link.
- Touches the database: yes — new column on an existing table, filled by hand for contracts in force.

## Considered options

1. **`contracts.funding_source`** + `ContractsInForce(ctx, association, date)` in `finance/module.go`.
2. **Separate `funds` table** with a many-to-many link to contracts.

## Decision outcome

**Chosen:** option 1. `funding_source text NULL CHECK (funding_source IN ('MAINTENANCE','REPAIR_FUND','RESERVE_FUND','RENTAL_INCOME'))`. `finance/module.go` exports `ContractsInForce` returning id, counterparty, funding source, `signed_at`, `ends_at`; a contract is in force on D when `signed_at ≤ D` and (`ends_at` is null or `ends_at ≥ D`). The governance slot "джерело коштів" stores the contract id together with a copy of counterparty, funding source and dates taken at confirmation, so the signed protocol does not depend on later contract edits. A contract in force without a funding source is offered with a prompt to fill it in `finance`.

### Consequences

- Good: closes both PRD §8 funding OQs with the PRD default; smallest change to `finance` for Stage 1.
- Bad: a contract paid from two funds cannot be described; Stage 2 expense closure (US-12) may need option 2.

### Migration and rollback

- Migration: `ALTER TABLE contracts ADD COLUMN funding_source text NULL` with the CHECK; no backfill script — a член правління fills it for contracts in force before the first protocol (PRD §9); ended contracts stay empty.
- Rollback: drop the column; protocols keep their own copy in the slot.

## Links

- PRD: [PRD.ao-prd.md](../PRD.ao-prd.md) — AC-10; §8 funding-source OQs; §9 Existing contracts
- SAD: [sad.ao-architecture.md](../sad.ao-architecture.md) §5
- Related ADR: [0002 Governance owns protocols](./0002-keep-protocols-cases-roster-and-charter-rules-in-the-governance-context.md) · repo [0002 Bounded contexts](../../../adr/0002-bounded-contexts.md)
