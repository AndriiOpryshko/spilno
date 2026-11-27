---
status: accepted
date: 2026-11-20
feature: contract-uah-equivalent
sad_section: "§4"
deciders: ["Andrii Opryshko", "Claude (easy depth)"]
---

# 0001 Read the NBU rate through an own MCP server behind a finance port

## Context and problem statement

The finance context needs the офіційний курс of a currency on a day (PRD AC-01, AC-10). The NBU serves it
from a public HTTP API. The same data is useful to an agent working in the repo or, later, in the product.
Where does the integration live, and how does the core API reach it?

## Decision drivers

- One place that knows the NBU API, its quirks (empty list for an unpublished day, rates as JSON numbers)
  and its cache.
- The core API must keep working when the source is down (AC-04).
- `domain/` and `app/` must not know the transport (repo ADR 0002).
- The course module asks for an MCP integration that a product feature really uses.

## Blast radius

- Irreversible: no — the port hides the choice; swapping the adapter is one package.
- Multi-context: yes — a new deployable (`mcp/nbu-rates`) and a new dependency of the core module.
- Legitimate alternative: yes — see below.
- Touches the database: no.

## Considered options

1. **An own MCP server (`mcp/nbu-rates`), reached from the core API as an MCP client over Streamable HTTP,
   behind `finance/ports.RateProvider`.**
2. A plain HTTP client to the NBU inside a finance adapter.
3. The MCP server spawned by the core API as a stdio subprocess.
4. A ready third-party currency MCP server.

## Decision outcome

**Chosen:** Option 1. The NBU knowledge sits in one small server usable by the core API and by agents alike;
the core API sees a typed tool with structured output and stable error codes. Streamable HTTP in stateless
mode suits a long-running service: no child process to supervise, and the server can be restarted on its own.

- Option 2 is simpler and has one hop less; it loses the reuse by agents and is the natural fallback: it
  would be another adapter of the same port.
- Option 3 ties the server's lifetime to the API process and needs the binary on the API's host.
- Option 4: none serves the NBU official rate, and a commercial rate is the wrong number (CONTEXT:
  офіційний курс).

### Consequences

- Good: one port, one adapter, a contract checked end to end against the real server binary (AC-11).
- Good: with `NBU_RATES_MCP_URL` unset the API runs and the feature degrades (AC-04, AC-07).
- Bad: one more process to run and one more hop (loopback) per lookup; a connect and a call per lookup,
  acceptable at a few contracts a month.
- Bad: the MCP endpoint has no authentication; it must stay on loopback or a private network.

## Migration

None. Enabling: start `nbu-rates -http`, set `NBU_RATES_MCP_URL`. Rollback: unset the variable.

## Links

PRD AC-04, AC-10, AC-11 · SAD §4, §5, §6 · [`mcp/nbu-rates/README.md`](../../../../mcp/nbu-rates/README.md)
