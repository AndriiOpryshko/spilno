# spilno

Management app for an HOA (in Ukrainian, ОСББ): contracts, invoices, board meetings and the
documents that connect them.

The repo is also the proving ground for [petforge](https://github.com/AndriiOpryshko/petforge),
a Claude Code plugin that adds a vertical slice (migration, endpoint, tests, UI) in a few commands.
`LOG.md` and `CAPSTONE_LOG.md` record how long a feature takes by hand and with the plugin.

## Layout

```
services/core/   Go modular monolith: one bounded context per internal/<bc>/
apps/web/        React app (Vite, Tailwind, shadcn/ui, TanStack Router + Query)
infra/           docker compose: Postgres, grafana/otel-lgtm
docs/adr/        architecture decisions
```

## Getting started

Requirements: [mise](https://mise.jdx.dev) and Docker.

```sh
mise install          # Go, Node, pnpm, mprocs
mise run setup        # JS dependencies and the Playwright browser
mise run dev          # infra + core API (hot reload) + web, in mprocs panes
```

- Web: http://localhost:5180
- Core API: http://localhost:8080 (try `GET /system/health`, or the Bruno collection in
  `services/core/api/bruno`)
- Grafana: http://localhost:3000 (traces in Tempo, logs in Loki, metrics in Prometheus)

## Tasks

| Task | What it does |
| --- | --- |
| `mise run gen` | Regenerate Go server, sqlc code and TS client from the contract |
| `mise run lint` | golangci-lint with bounded-context rules, eslint, tsc |
| `mise run test` | Go unit + integration (testcontainers) and Vitest |
| `mise run e2e` | Playwright against the running stack |
| `mise run migrate:new <name>` | New goose migration |

## License

MIT, see [LICENSE](LICENSE).
