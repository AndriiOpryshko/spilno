---
status: accepted
date: 2026-10-21
---

# 0001 Stack for the core service and web app

## Context and problem statement

spilno is a pet project and the proving ground for the petforge plugin, which automates adding
a vertical slice (migration, endpoint, tests, UI). The stack has to be boring to operate, fast to
generate code for, and observable locally.

## Decision outcome

| Concern | Choice |
| --- | --- |
| HTTP | Go 1.27, chi v5 (plain `net/http`, so `otelhttp` and stdlib middleware work) |
| Contract | OpenAPI 3.1, oapi-codegen v2.8 strict server (see ADR 0003) |
| Data | Postgres 18, pgx/v5 + sqlc, otelpgx tracing, goose migrations (see ADR 0004) |
| Logs | `log/slog` only: JSON to stdout and `otelslog` via `slog.NewMultiHandler` |
| Web | Vite, React 19, TypeScript 6, Tailwind v4, shadcn/ui, TanStack Router + Query, orval |
| Tests | testcontainers-go, Vitest + Testing Library, Playwright, Bruno for manual API calls |
| Local dev | mise (toolchain + tasks), docker compose, mprocs, air, grafana/otel-lgtm |

### Consequences

- Good: every layer is plain and replaceable; one trace covers HTTP and SQL in Grafana.
- Good: zap was considered but its API takes no `context.Context`, so trace ids would not reach
  logs; slog backends can still be swapped without touching call sites.
- Bad: TypeScript stays on 6.0 because typescript-eslint supports `<6.1`.
- Bad: fasthttp/fiber were rejected (no HTTP/2, non-`net/http` handlers); throughput is not the
  bottleneck for a CRUD app.

## More information

Research and a verification spike were run on 2026-10-21 before this decision.
