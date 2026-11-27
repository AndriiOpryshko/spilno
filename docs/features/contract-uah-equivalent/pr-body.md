## Why

A contract in EUR or USD kept only its own-currency amount. The board and the ревізійна комісія work in hryvnias,
so the hryvnia figure was worked out by hand and its rate was written nowhere
([PRD](docs/features/contract-uah-equivalent/PRD.md) §1). The contract now carries its гривневий еквівалент at
the NBU's офіційний курс of the signing day, read through the project's own `nbu-rates` MCP server
([ADR-0001](docs/features/contract-uah-equivalent/adr/0001-read-the-nbu-rate-through-an-own-mcp-server-behind-a-finance-port.md)).
The equivalent is stored once and never recomputed
([ADR-0002](docs/features/contract-uah-equivalent/adr/0002-store-the-hryvnia-equivalent-on-the-contract-and-never-recompute-it.md)).

## What

- `mcp/nbu-rates`: a new Go module, an MCP server with the official NBU rates (tool `get_rate`, resource
  `nbu://about`, prompt `uah_equivalent`; stdio and stateless Streamable HTTP), plus its own `mcp` CI job.
- `services/core/migrations/00050_add_contracts_uah_equivalent.sql`: two nullable columns on `contracts`, no backfill.
- `internal/finance/domain`: the official rate and the hryvnia equivalent. Integer half-up rounding; an
  out-of-range result is refused.
- `internal/finance/adapters/mcprates`: the rate port over MCP. Every request is bounded by the caller's
  deadline, redirects are refused, a published row for another day is refused, and the rate-source address is
  redacted by structure.
- `internal/finance/app` and `adapters/postgres`: the equivalent is worked out at creation (a failure leaves
  `null`, the contract is still saved), and «get it later» stores it once, also under concurrent calls.
- `services/core/api/openapi.yaml` and generated code: `Contract.uah_equivalent`, the new
  `POST .../contracts/{id}/uah-equivalent` and four problem codes, registered in the finance error registry.
- `cmd/api`, `internal/platform/config`: `NBU_RATES_MCP_URL`. It is optional, the address is logged redacted,
  and an address `net/http` would print escaped is refused at start-up.
- Docs: the feature's PRD, SAD, ADRs, data model, contract delta, tasks, the six review rounds and
  [CHANGELOG](docs/features/contract-uah-equivalent/CHANGELOG.md); CLAUDE.md and the MCP README.

## Acceptance criteria

- AC-01: a 1 250.00 EUR contract of 2026-10-20 gets `{amount_kopiyky: 6316388, rate: "50.5311", rate_date, source: NBU}` ✓
- AC-02: amount × rate is rounded half up in integers; a result that does not fit int64 is refused ✓
- AC-03: a UAH contract has `uah_equivalent: null` and the source is not asked ✓
- AC-04: if the source fails, the contract is still `201` with `null`, within 3 s ✓
- AC-05: «get it later» answers `200` with the equivalent ✓
- AC-06: computed once; a repeat does not ask the source, and concurrent calls store one value ✓
- AC-07: `409 equivalent_not_applicable` / `equivalent_out_of_range` / `rate_not_published`, `503 rate_unavailable`, `404` ✓
- AC-08: `401` / `403 not_current_member` / `403 association_disabled` / `404`, and the isolation test covers the operation ✓
- AC-09: rate, rate date and source are traceable wherever the amount is ✓
- AC-10: the rate comes only through MCP `get_rate`; with the address unset every lookup is «unavailable» ✓
- AC-11: end to end through the real `nbu-rates` binary in front of an `httptest` NBU ✓

## Checks

Run on 2026-11-27 at `5135177` (linux arm64):

- [x] `mise run lint`: golangci-lint (core and `mcp/nbu-rates`), eslint and tsc are clean.
- [x] `mise run test`: Go unit and integration tests (testcontainers) and Vitest (373) pass. Story tests failed
  on the first run with the known cold-start flake (an iframe or dynamic import that was not ready, no
  assertion failed). They passed on an immediate rerun (28 files, 93 tests).
- [x] `mise run e2e`: Playwright on an isolated throwaway stack, 34 passed and 4 skipped. The 4 are existing
  skips that wait for board-protocol tasks.
- [x] I ran the feature by hand on a throwaway Postgres, the built core API and the real `nbu-rates` against
  the live NBU API:
  - AC-01, AC-02, AC-09: EUR `50.5311` → 6316388 and USD 1 cent × `45.0013` → 45. Both rates equal the NBU's
    own page for 2026-10-20.
  - AC-03: UAH gives `null`.
  - AC-04: a contract created while the MCP server was down, and another with `NBU_RATES_MCP_URL` unset, both
    answered `201` with `null`.
  - AC-07: `503` while down. After restart, «get it later» answered `200` (AC-05).
  - AC-06: a repeat returned the stored value unchanged, and did so with the server stopped.
  - AC-07 refusals: `409` for UAH, for a 2031 date and for 9·10¹⁸ cents; `404` for a missing id.
  - AC-08: `401` without a session; `403 not_current_member` for an existing id and for a missing id alike.
- [x] `/sdlc:review-feature`: six rounds, the last [APPROVED](docs/features/contract-uah-equivalent/_review/review-2026-11-27-3.md).
  It deferred 5 non-blocking findings.
- [ ] `/security-review`: not run as a separate pass in this session. The security findings (redaction,
  redirects, deadline, escaped address) came from review rounds 2–5 and are fixed.
- Not run by hand: `403 association_disabled` (covered by the shared guard's tests). The `mcp` CI job and the
  isolation test ran in CI and `mise run test`.

## Operational notes

- Migration `00050` is applied on start-up. Rollback: revert the deploy first, because contract reads select
  the columns, then run `goose down` one step. Stored equivalents can be worked out again from the NBU.
- Config: `NBU_RATES_MCP_URL` (optional; locally `mise run mcp:nbu` → `http://127.0.0.1:3336/mcp`). The new
  process `nbu-rates` reads `NBU_BASE_URL`, which defaults to the NBU's public API.

## Links

- No issue for this feature. The follow-up is #65 (the equivalent in a protocol slot).
- Artifacts: `docs/features/contract-uah-equivalent/`:
  [PRD](docs/features/contract-uah-equivalent/PRD.md),
  [SAD](docs/features/contract-uah-equivalent/sad.md), [ADRs](docs/features/contract-uah-equivalent/adr/),
  [data model](docs/features/contract-uah-equivalent/data-model.md),
  [contract delta](docs/features/contract-uah-equivalent/contracts/openapi.yaml),
  [tasks](docs/features/contract-uah-equivalent/tasks/),
  [reviews](docs/features/contract-uah-equivalent/_review/).
- SDLC-Task commits: T01 `37b8fe9` · T02 `ca42535` · T03 `5ca89e1` · T04 `0f90371` · T05 `6c9c459` ·
  T06 `600dcc0` · T07 `22d2ffc` · T08 `13ad560` · T09 `2606d2a` · T10 `52f54c8` · T11 `d1ec21c` · T12 `ce2a21a`.
  Review follow-ups T13–T34: `git log --grep SDLC-Task main..HEAD`.

🤖 Generated with [Claude Code](https://claude.com/claude-code)
