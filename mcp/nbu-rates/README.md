# nbu-rates

An MCP server with the official hryvnia exchange rates of the National Bank of Ukraine (public API, no key).
The finance context uses it to store the hryvnia equivalent of a contract signed in EUR or USD; an agent can
use the same server to answer "what was the rate on the day we signed".

| Primitive | Name | What it gives |
|---|---|---|
| tool | `get_rate(currency, date)` | `{currency, date, rate, source}`; `rate` is a decimal string, hryvnias per one unit |
| resource | `nbu://about` | what the rate means, when it is published, how to convert, the error codes |
| prompt | `uah_equivalent(amount, currency, date)` | asks for the equivalent through `get_rate`, no guessing |

A failed `get_rate` is a tool error whose text starts with `invalid_argument`, `rate_not_published` or
`upstream_unavailable`. The upstream detail stays in the server log (stderr).

## Run

```bash
mise run mcp:nbu            # Streamable HTTP on 127.0.0.1:3336, endpoint /mcp, GET /healthz
go run -C mcp/nbu-rates .   # stdio, for an MCP client that spawns it
mise run mcp:nbu:inspect    # MCP Inspector CLI: tools/list and one live get_rate
npx @modelcontextprotocol/inspector go run -C mcp/nbu-rates .   # Inspector UI
```

`mise run dev` starts the HTTP variant in its own pane. For Claude Code, copy the entry from
`.mcp.json.example` into the repo's `.mcp.json`.

| Variable | Default | Meaning |
|---|---|---|
| `NBU_RATES_HTTP_ADDR` | empty (stdio) | same as `-http` |
| `NBU_BASE_URL` | the NBU statistics endpoint | override for tests |

## Consumer

`services/core/internal/finance/adapters/mcprates` calls `get_rate` over Streamable HTTP and relies on the
structured output (`currency`, `date`, `rate`, `source`) and on the `rate_not_published` error prefix. Change
either and `finance/rates_e2e_test.go` fails: it runs this binary in front of a fake NBU.

## Notes

- The HTTP transport is stateless and has no authentication: the data is public and read-only. Keep it on
  loopback or a private network; do not publish the port.
- A published rate never changes, so it is cached in memory (up to 4096 entries). Failures are not cached.
- Tests never reach the NBU: `internal/nbu` runs against an `httptest` upstream, `internal/server` against a
  fake source over the in-memory and the Streamable HTTP transports.
