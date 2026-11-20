# Epic — contract-uah-equivalent

> **PRD:** [PRD.md](../PRD.md) · **Design:** [sad.md](../sad.md) · **Data model:** [data-model.md](../data-model.md) · **API:** [openapi.yaml](../contracts/openapi.yaml) · **ADRs:** [adr/](../adr/) · **Migrations:** [MANIFEST](../migrations/MANIFEST.md) · **MCP server:** [`mcp/nbu-rates`](../../../../mcp/nbu-rates/README.md)

## Goal

A contract signed in EUR or USD carries its гривневий еквівалент at the офіційний курс of the signing day, read by the finance context from the `nbu-rates` MCP server, stored once with the rate behind it; the rate source may fail without a contract being lost (PRD §2).

## Scope

- **In:** surface `[backend-service]`; context `finance` (domain, port, MCP adapter, storage, use cases, HTTP), `platform/config`, `cmd/api`; one staged migration; the contract delta.
- **Out:** the MCP server itself (already in `mcp/nbu-rates`), any web page, governance, a retry job (PRD §3, §8).

## Task map

Same DAG as [tasks.json](./tasks.json). Day-one parallel starts: T01, T02, T03, T04. Longest chain: 7 (T02 → T05 → T07 → T08 → T09 → T10 → T11). Tasks sharing files (`app/contracts.go`, `handler.go`, `module.go`) are serialized by their `deps`.

```mermaid
flowchart LR
    T01["T01 migration: promote-uah-equivalent-migration"]:::mig
    T02["T02 domain: rate-and-equivalent-domain"]:::dom
    T03["T03 domain: finance-problem-codes"]:::dom
    T04["T04 wiring: merge-contract-delta"]:::wir
    T05["T05 infra: mcp-rate-provider"]:::inf
    T06["T06 infra: store-uah-equivalent"]:::inf
    T07["T07 app: resolve-equivalent-on-create"]:::app
    T08["T08 app: resolve-equivalent-later"]:::app
    T09["T09 ports: uah-equivalent-http"]:::prt
    T10["T10 wiring: wire-rates-config"]:::wir
    T11["T11 tests: end-to-end-real-mcp-server"]:::tst
    T12["T12 docs: docs-and-bruno"]:::doc
    T02 --> T05
    T01 --> T06
    T02 --> T06
    T05 --> T07
    T06 --> T07
    T03 --> T08
    T07 --> T08
    T04 --> T09
    T08 --> T09
    T09 --> T10
    T10 --> T11
    T09 --> T12
    classDef mig fill:#fde68a,stroke:#92400e,color:#111
    classDef dom fill:#bbf7d0,stroke:#166534,color:#111
    classDef inf fill:#bfdbfe,stroke:#1e40af,color:#111
    classDef app fill:#ddd6fe,stroke:#5b21b6,color:#111
    classDef prt fill:#fecaca,stroke:#991b1b,color:#111
    classDef wir fill:#e5e7eb,stroke:#374151,color:#111
    classDef tst fill:#fbcfe8,stroke:#9d174d,color:#111
    classDef doc fill:#f5f5f4,stroke:#57534e,color:#111
```

## AC coverage

| AC | Tasks |
|---|---|
| AC-01 | T01, T06, T07, T09, T11 |
| AC-02 | T02 |
| AC-03 | T07 |
| AC-04 | T07, T10 |
| AC-05 | T04, T08, T09, T12 |
| AC-06 | T06, T08, T09 |
| AC-07 | T03, T08, T09 |
| AC-08 | T09 |
| AC-09 | T02, T04, T09 |
| AC-10 | T05, T10, T11 |
| AC-11 | T11 |
