# Tracker — contract-uah-equivalent

> Status of every task in the epic. `implement-tasks` updates `done` as it commits each task.
> States: `todo` · `in_progress` · `blocked` · `review` · `done`. Estimates: S = 2 h, M = half a day, L = a day.

| # | Task | Layer | Owner | Estimate | Blocked by | Status |
|---|---|---|---|---|---|---|
| [T01](./T01-promote-uah-equivalent-migration.md) | Promote the migration that adds the hryvnia equivalent columns to contracts | migration | Andrii Opryshko | S | — | done |
| [T02](./T02-rate-and-equivalent-domain.md) | Model the official rate and compute the hryvnia equivalent with half-up rounding | domain | Andrii Opryshko | M | — | done |
| [T03](./T03-finance-problem-codes.md) | Register the finance problem codes of PRD §5.1 | domain | Andrii Opryshko | S | — | done |
| [T04](./T04-merge-contract-delta.md) | Merge the contract delta and stub the new finance operation | wiring | Andrii Opryshko | M | — | done |
| [T05](./T05-mcp-rate-provider.md) | Read the official rate through the nbu-rates MCP server behind a finance port | infra | Andrii Opryshko | L | T02 | done |
| [T06](./T06-store-uah-equivalent.md) | Store the hryvnia equivalent with a contract and set it only once | infra | Andrii Opryshko | M | T01, T02 | done |
| [T07](./T07-resolve-equivalent-on-create.md) | Work out the equivalent when a foreign-currency contract is created, and never fail the creation over it | app | Andrii Opryshko | M | T05, T06 | done |
| [T08](./T08-resolve-equivalent-later.md) | Add the use case that works out a missing equivalent later, once | app | Andrii Opryshko | M | T03, T07 | done |
| [T09](./T09-uah-equivalent-http.md) | Return the equivalent on every contract answer and serve the «get it later» operation | ports | Andrii Opryshko | L | T04, T08 | done |
| [T10](./T10-wire-rates-config.md) | Wire NBU_RATES_MCP_URL from the environment to the finance module | wiring | Andrii Opryshko | S | T09 | done |
| [T11](./T11-end-to-end-real-mcp-server.md) | Prove the chain end to end against the real nbu-rates binary and a fake NBU | tests | Andrii Opryshko | M | T10 | todo |
| [T12](./T12-docs-and-bruno.md) | Document the feature: Bruno request, environment variable, CLAUDE.md gotcha | docs | Andrii Opryshko | S | T09 | todo |
