# Changelog — contract-uah-equivalent

## contract-uah-equivalent — a foreign-currency contract carries its hryvnia equivalent at the NBU rate

**What:** a contract signed in EUR or USD now carries its гривневий еквівалент: the amount in kopiyky, the
офіційний курс it came from, the дата курсу (always the signing day) and the джерело курсу (`NBU`). It is
worked out when the contract is created, with nobody typing a rate. If the rate source cannot answer at that
moment the contract is still saved, without an equivalent, and a member can ask for it later with one request.
Once stored, the equivalent never changes. A UAH contract has no equivalent.

**Why:** everything the board decides and the ревізійна комісія checks is in hryvnias, but the hryvnia figure of
a euro or dollar contract was worked out by hand and its rate was written nowhere ([PRD](../PRD.md) §1, §2).
Two decisions carry the design:

- [ADR-0001](../adr/0001-read-the-nbu-rate-through-an-own-mcp-server-behind-a-finance-port.md): the finance
  context reads the rate through the project's own `nbu-rates` MCP server (`mcp/nbu-rates`, tool `get_rate`)
  behind a finance port. The source is optional: without it the API starts and every lookup is «unavailable».
- [ADR-0002](../adr/0002-store-the-hryvnia-equivalent-on-the-contract-and-never-recompute-it.md): the equivalent
  is stored on the contract and computed once, so a later rate never rewrites a figure a protocol already cites.

**How to use** ([contract delta](../contracts/openapi.yaml), merged into `services/core/api/openapi.yaml`):

- `POST /finance/associations/{association_id}/contracts` with `currency: EUR` or `USD` answers `201` with
  `uah_equivalent`, or with `uah_equivalent: null` when the rate could not be read (the wait is bounded by 3 s).
  Reading and listing contracts return the same field.

  ```json
  "uah_equivalent": { "amount_kopiyky": 6316388, "rate": "50.5311", "rate_date": "2026-10-20", "source": "NBU" }
  ```

  1 250.00 EUR × 50.5311 = 63 163.875 UAH → 6 316 388 kopiykas (rounded half up, integer arithmetic only).
- `POST /finance/associations/{association_id}/contracts/{id}/uah-equivalent` («get it later») answers `200`
  with the contract and its equivalent. It is safe to repeat: a contract that already has one is returned as
  stored and the rate source is not asked. Refusals are `application/problem+json`: `409
  finance.equivalent_not_applicable` (a UAH contract), `409 finance.rate_not_published` (no NBU rate for that
  day), `409 finance.equivalent_out_of_range` (the result does not fit 64 bits), `503 finance.rate_unavailable`
  (the source did not answer or is not configured; try again later).

**Operational notes:**

- Migration: `00050_add_contracts_uah_equivalent` adds two nullable columns to `contracts`
  (`uah_amount_kopiyky`, `uah_rate_micros`). It is applied on start-up, needs no backfill, and existing
  contracts keep `NULL` until someone asks for their equivalent. `goose down` drops both columns.
- Config: `NBU_RATES_MCP_URL` names the `nbu-rates` server's Streamable HTTP endpoint (locally `mise run mcp:nbu`
  → `http://127.0.0.1:3336/mcp`). Optional: unset, contracts are saved with `uah_equivalent: null` and «get it
  later» answers `503`. An address with credentials is logged redacted. An address `net/http` would print
  escaped is refused at start-up. The server reads `NBU_BASE_URL` (default: the NBU's public API).
- New process: the `nbu-rates` MCP server is its own Go module (`mcp/nbu-rates`) and has its own CI job (`mcp`).
- Rollback: revert the deploy, then `goose down` one step. The code that reads the columns has to be gone
  before they are dropped, because existing contract reads select them. Equivalents stored in the meantime are
  lost, which is safe: they can be worked out again from the NBU.

**Acceptance criteria delivered:** AC-01 (the equivalent at creation), AC-02 (half-up rounding in integers; an
out-of-range result is refused, never truncated), AC-03 (no equivalent for UAH, source not asked), AC-04 (the
contract is saved even when the source fails, within 3 s), AC-05 («get it later»), AC-06 (computed once, also
under concurrent calls), AC-07 (the refusals above), AC-08 (session, membership, disabled-association and
isolation guards), AC-09 (the rate is traceable to its day and source), AC-10 (the rate comes only through MCP),
AC-11 (end to end through the real `nbu-rates` binary).

**Known follow-ups** (deferred in [review round 6](../_review/review-2026-11-27-3.md), none blocking): the start-up
log line is misleading when the address is refused, a stale HTTP status can mislabel a timeout's class, and
the SAD and one test assertion are imprecise. Parked product items ([PRD](../PRD.md) §8): the equivalent in a
protocol slot (issue #65), and a retry job for contracts left without one.
