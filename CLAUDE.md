# spilno

HOA (condo association) management app. Monorepo: `services/core` (Go modular monolith),
`apps/web` (React), `infra/` (local dependencies). Decisions live in `docs/adr/`; read them
before changing structure.

## Commands

All tasks run from the repo root through mise (`mise tasks` lists them):

- `mise run gen` - regenerate Go server, sqlc code and TS client. Run after editing
  `services/core/api/openapi.yaml`, any `queries/*.sql` or a migration.
- `mise run test` - Go unit + integration (needs Docker for testcontainers) and Vitest.
- `mise run lint` - golangci-lint (includes the depguard context boundaries), eslint, tsc.
- `mise run e2e` - Playwright; the core API must be running (`mise run dev`).
- `mise run storybook` - component workshop on :6006. Stories live next to components
  (`*.stories.tsx`), mirror the Figma component variants and run as tests in Chromium with
  a11y checks (`pnpm --filter @spilno/web test:stories`, part of `mise run test`).
- `mise run migrate:new <name>` - new goose migration.

"Green" means `mise run lint`, `mise run test` and `mise run e2e` all pass.

## Layout rules

- One OpenAPI contract: `services/core/api/openapi.yaml` (3.1). Every operation has exactly one
  tag, the name of its bounded context. Paths start with `/<context>/`.
- A bounded context is `services/core/internal/<bc>/`:
  `domain/` (stdlib only), `app/` (use cases), `ports/` (interfaces app needs),
  `adapters/http/` (implements the generated `StrictServerInterface`),
  `adapters/http/gen/` (oapi-codegen config + output), `adapters/postgres/` (ports on Postgres),
  `adapters/postgres/queries/` (sqlc input), `adapters/postgres/sqlcgen/` (sqlc output),
  `module.go` (the only package anything outside the context may import).
- A new context needs: its tag in the spec, `adapters/http/gen/oapi-codegen.yaml`, an entry in
  `sqlc.yaml`, a `<bc>-private` depguard rule in `.golangci.yml`, and a line in `cmd/api/main.go`.
- Never edit generated files (`*.gen.go`, `sqlcgen/`, `apps/web/src/api/gen/`, `routeTree.gen.ts`).
- Errors on the wire are RFC 9457 `application/problem+json` (`internal/platform/httpapi`).
- Integration tests use `internal/platform/testdb.New(t)`: a fresh migrated database per test.
- Web: generated hooks from `@/api/gen/endpoints/<bc>/<bc>`; pages under `src/routes/`,
  feature components under `src/features/<bc>/`; UI primitives from shadcn (`src/components/ui`).

## Gotchas

- goose splits statements on `;`. Wrap functions and `DO` blocks in
  `-- +goose StatementBegin` / `-- +goose StatementEnd`.
- Do not use `r.Mount` for a context: generated handlers register absolute paths. Use the
  `r.Group` pattern from `internal/system/module.go`.
- The web dev server runs on :5180 and proxies `/api/*` to the core API on :8080.

## Metrics log

This repo is a capstone measured in `LOG.md` (raw) and `CAPSTONE_LOG.md` (summary).
On branches `feature-1/*` and `feature-2/*`:

- After every user prompt, append `- <ISO timestamp> | <F1|F2> | prompt | "<prompt verbatim>"`.
- After every manual check (tests, lint, migration, browser), append `... | check | <what> -> <result>`.
- When something goes wrong and is redone, append `... | stumble | <what happened>`.
- After every commit, append `... | commit | <short hash> <subject> | <metric: files changed, tests>`.
- Timer marks come from `mise run log:start <F>` and `mise run log:stop <F>`; do not write them by hand.
- Timestamps come from `date +%Y-%m-%dT%H:%M:%S%z`, never from memory.
