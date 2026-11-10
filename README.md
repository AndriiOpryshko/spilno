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

## Environment

| Variable | Meaning |
| --- | --- |
| `DATABASE_URL` | The API's connection, as `spilno_app_login` (no `BYPASSRLS`) |
| `MIGRATIONS_DATABASE_URL` | Migrations, as the database owner |
| `SPILNO_DEV_SIGN_IN` | `true` turns on the development sign-in shim, the only sign-in for now |
| `SPILNO_OPERATOR_EMAIL` | Address of the оператор who reviews applications |
| `SPILNO_WEB_ORIGIN` | Where the web app is served (default `http://localhost:5180`); the storage connection returns to `<origin>/storage/callback`, register that address with the Google OAuth client |
| `GOOGLE_OAUTH_CLIENT_ID`, `GOOGLE_OAUTH_CLIENT_SECRET`, `GOOGLE_TOKEN_KEY_ID`, `GOOGLE_TOKEN_KEY` | Set all four to turn the storage connection on; without them it answers 501 |
| `VITE_OPERATOR_CONTACT_EMAIL` | Web build: the address an applicant is told to write to |

An existing local Postgres volume needs `infra/postgres-init/01-app-role.sql` run once (or recreate the
volume) so that `spilno_app_login` exists. The оператор's procedure is in
[the runbook](docs/runbooks/operator-extract-check.md).

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
