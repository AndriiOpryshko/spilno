---
id: T06
title: "Store the hryvnia equivalent with a contract and set it only once"
layer: "infra"
deps: ["T01", "T02"]
acs: ["AC-01", "AC-06"]
files_hint:
  - "services/core/internal/finance/adapters/postgres/queries/finance.sql"
  - "services/core/internal/finance/adapters/postgres/sqlcgen/"
  - "services/core/internal/finance/adapters/postgres/contracts.go"
  - "services/core/internal/finance/adapters/postgres/contracts_test.go"
  - "services/core/internal/finance/ports/ports.go"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T06 — Store the hryvnia equivalent with a contract and set it only once

## Why

[data-model](../data-model.md) «Query → index map»; [SAD](../sad.md) Flow 2 (conditional update).

**Satisfies:** [PRD §5](../PRD.md) AC-01, AC-06 · **Blocked by:** T01, T02

## What

- `CreateContract` inserts `uah_amount_kopiyky`, `uah_rate_micros` (nullable).
- New `SetContractUAHEquivalent :one` — `UPDATE … WHERE id AND association_id AND uah_amount_kopiyky IS NULL RETURNING *`; `mise run gen`.
- `ports.ContractRepository`: `Create(ctx, terms, *domain.UAHEquivalent)`; `SetUAHEquivalent(ctx, id, eq) (domain.Contract, bool, error)` — `false` when nothing was updated.
- Map the columns into `domain.Contract.UAHEquivalent` on every read (both NULL → nil).

## Definition of Done

- [ ] Acceptance: Repository tests on a real database show: a contract created with an equivalent reads back with the same kopiykas and rate; one created without reads back with none; `SetUAHEquivalent` sets it on a contract that has none and changes nothing on one that has; another association's contract is not found.
- [ ] integration (`testdb.NewApp`): the four cases of the DoD
- [ ] integration: `testdb.NewTwoAssociations` — set on the other association's id updates nothing
- [ ] `mise run lint` and `mise run test` green
