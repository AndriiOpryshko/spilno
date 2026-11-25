---
status: Approved
owner: "Andrii Opryshko"
reviewers: ["Andrii Opryshko"]
approved_by: "Andrii Opryshko"
approved_at: "2026-11-20"
updated_at: "2026-11-20"
feature_size: "S"
stage: "03"
depth: "easy"
ticket: ""
---

# PRD — contract-uah-equivalent

> **Inputs:** [idea brief](./idea-brief.md) · [CONTEXT](./CONTEXT.md)
> **Reference context:** `services/core/internal/finance` (contract: create, get, list; `domain/contract.go`,
> `app/contracts.go`), `mcp/nbu-rates` (the MCP server with the офіційний курс: tool `get_rate`,
> [README](../../../mcp/nbu-rates/README.md)).
> **Depth:** easy; written in one pass from the owner's pick of the idea, no Socratic walk. The test plan is
> inline in §7 (size S).

## 1. Context

A contract in the finance context keeps its amount in its own currency: UAH, EUR or USD. Everything the board
decides and the ревізійна комісія checks is in hryvnias. Today the hryvnia figure of a euro or dollar contract
is worked out by hand, and the rate behind it is written nowhere.

The product now has its first external data source: the `nbu-rates` MCP server relays the офіційний курс of
the National Bank of Ukraine for a currency and a day. This feature is its first consumer.

## 2. Goals

- Every contract signed in EUR or USD shows its гривневий еквівалент with the rate, the дата курсу and the
  джерело курсу, without anyone typing a rate.
- A contract is never lost or delayed because the rate source did not answer.

## 3. Non-goals

- Revaluation at another day, exchange differences, payments.
- Rates other than the NBU official one; currencies other than EUR and USD.
- A web screen: the finance context has no pages yet; the API and its generated client carry the data.
- Editing or deleting a stored equivalent.

## 4. User stories

### US-01: See the hryvnia equivalent of a foreign-currency contract

As the головуючий who records a contract in euros or dollars, I want its гривневий еквівалент worked out
from the офіційний курс of the signing day, so that the figure in a protocol does not depend on my arithmetic.

### US-02: Trace the equivalent to its rate

As a member of the ревізійна комісія, I want to see which rate, of which day and from which source an
equivalent came, so that I can check it against the NBU's own page.

### US-03: Get the equivalent later

As the головуючий, I want to ask for the equivalent again when it could not be worked out at creation, so
that a short outage of the rate source leaves no permanent gap.

## 5. Acceptance criteria

Amounts are integer minor units (`amount_cents`, `amount_kopiyky`). The example used throughout: a contract of
1 250.00 EUR (`amount_cents` 125000) signed on 2026-10-20, офіційний курс `50.5311` → 6 316 387.5 kopiykas →
**`amount_kopiyky` 6316388**.

### AC-01 (US-01) — happy path

**Given** the rate source answers `50.5311` for EUR on 2026-10-20,
**When** a current член правління creates a contract of 125000 EUR cents signed on 2026-10-20,
**Then** the answer is `201` and the contract carries `uah_equivalent` = `{amount_kopiyky: 6316388, rate: "50.5311", rate_date: "2026-10-20", source: "NBU"}`, and reading the contract and listing contracts return the same `uah_equivalent`.

### AC-02 (US-01) — domain invariant (rounding)

**Given** an amount in minor units and a rate as a decimal string,
**When** the equivalent is computed,
**Then** it equals amount × rate rounded half up to the kopiyka, using integer arithmetic only: 125000 × `50.5311` → 6316388; 1 × `45.051` → 45; 1 × `45.055` → 45 (45.055 → 45); 10 × `45.055` → 451 (450.55 → 451); 9 000 000 000 000 000 × `50.5311` does not overflow. A result that does not fit a signed 64-bit number of kopiyky is refused, never truncated (200 000 000 000 000 000 × `50.5311`). A rate that is not a positive decimal with at most six fraction digits is refused.

### AC-03 (US-01) — domain invariant (hryvnia contract)

**Given** a contract in UAH,
**When** it is created,
**Then** `uah_equivalent` is `null` and the rate source is not asked.

### AC-04 (US-01) — error (rate source fails at creation)

**Given** the rate source is unreachable, slow, not configured, or has no rate for that day, or the equivalent is out of range (AC-02),
**When** a foreign-currency contract is created,
**Then** the answer is still `201` with `uah_equivalent: null`, the contract is stored, and the wait for the rate source is at most 3 seconds.

### AC-05 (US-03) — happy path (get it later)

**Given** a stored EUR contract without an equivalent and a rate source that now answers,
**When** a current член правління calls `POST /finance/associations/{association_id}/contracts/{id}/uah-equivalent`,
**Then** the answer is `200` with the contract and its `uah_equivalent`, and later reads return the same.

### AC-06 (US-03) — domain invariant (computed once)

**Given** a contract that already has an equivalent,
**When** the same operation is called again, even if the rate source would now answer another rate,
**Then** the answer is `200` with the stored `uah_equivalent` unchanged and the rate source is not asked. Two concurrent calls store one equivalent.

### AC-07 (US-03) — error (refused)

**When** the operation is called for a UAH contract, **Then** `409 finance.equivalent_not_applicable`.
**When** the equivalent of the amount is out of range (AC-02), **Then** `409 finance.equivalent_out_of_range`.
**When** the NBU has no rate for the signing day, **Then** `409 finance.rate_not_published`.
**When** the rate source is unreachable, slow or not configured, **Then** `503 finance.rate_unavailable`.
**When** no contract has the id in this association, **Then** `404`.
In every case the contract is unchanged and the body is `application/problem+json` with that `code`.

### AC-08 (US-03) — authorization

**Given** the shared guards of association-scoped operations,
**When** the operation is called without a session, by a person off the roster, for a вимкнене ОСББ, or with the id of another association's contract,
**Then** the answers are `401 identity.session_required`, `403 identity.not_current_member`, `403 identity.association_disabled` and `404` respectively, and the repo's isolation test covers the new operation.

### AC-09 (US-02) — traceability

**Given** a stored equivalent,
**Then** `rate` is the decimal the source published, digit for digit (trailing fraction zeros dropped: `45.0510` is returned as `45.051`), `rate_date` equals the contract's `signed_at`, and `source` is `NBU`; the three are returned wherever `amount_kopiyky` is.

### AC-10 (US-01) — cross-context (the rate comes through MCP)

**Given** `NBU_RATES_MCP_URL` names the `nbu-rates` server,
**When** the finance context needs a rate,
**Then** it calls the `get_rate` tool over MCP (Streamable HTTP) and nothing else; a tool error starting with `rate_not_published` means «not published», any other failure means «unavailable»; an answer whose currency or date differs from the question, or whose rate is malformed, is «unavailable» and is never stored. With `NBU_RATES_MCP_URL` unset the API starts and every lookup is «unavailable».

### AC-11 (US-01) — happy path (end to end)

**Given** the real `nbu-rates` server process in front of a fake NBU upstream,
**When** a contract is created through the core API's HTTP surface,
**Then** AC-01 holds end to end, with no call to the real NBU.

## 5.1 Error catalogue

| Code | Status | Raised when |
|---|---|---|
| `finance.equivalent_not_applicable` | 409 | the contract is in UAH |
| `finance.equivalent_out_of_range` | 409 | the hryvnia equivalent of the amount does not fit a signed 64-bit number of kopiyky |
| `finance.rate_not_published` | 409 | the NBU has no rate of that currency for the signing day |
| `finance.rate_unavailable` | 503 | the rate source did not answer, answered nonsense, or is not configured |

Reused: `identity.session_required` (401), `identity.not_current_member` / `identity.association_disabled`
(403), the uncoded `404` and `422` of the existing contract operations.

## 6. Non-functional requirements

- **Latency:** a rate lookup is bounded by 3 s (AC-04); contract creation adds no more than that. No retries
  inside a request.
- **Money:** no floating point between the published decimal and the stored kopiykas.
- **Availability:** the rate source is optional for every existing operation; only the «get it later»
  operation needs it.
- **Security:** the rate source is public data over a private network address; no secret, no personal data
  leaves the core API (only a currency code and a date).
- **Observability:** a failed lookup is logged with the currency, the day and the reason class.

## 7. Test plan (inline)

| AC | Level | Where |
|---|---|---|
| AC-02, AC-09 (format) | unit | `finance/domain` |
| AC-03, AC-04, AC-06 (source not asked) | unit | `finance/app` with a fake rate port |
| AC-10 | integration | `finance/adapters/mcprates` against an in-process MCP server (go-sdk) |
| AC-01, AC-05, AC-06, AC-07, AC-08 | integration | `finance` module tests through the router on `testdb`, fake rate port |
| AC-06 (concurrency) | integration | two parallel calls, one stored value |
| AC-08 (isolation) | integration | `internal/platform/isolation_test.go` |
| AC-11 | end to end | built `nbu-rates` binary + `httptest` NBU + finance module |

No web or Playwright level: the feature has no UI surface.

## 8. Open questions

None blocking. Parked: showing the equivalent in a protocol slot (governance), a retry job for contracts left
without an equivalent.

## 9. Migration plan

Two nullable columns on `contracts` (`uah_amount_kopiyky`, `uah_rate_micros`), one migration, no backfill:
existing contracts keep `NULL` and get the equivalent through AC-05 on demand.

## 10. RBAC and state

Any current член правління of the association may create contracts and ask for the equivalent, as for the
existing contract operations. The equivalent has two states: absent → present; there is no way back.
