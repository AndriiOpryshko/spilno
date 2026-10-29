# Project Claude Code tooling

Committed tooling for this repo. Everything here is project-specific: it knows the goose
migrations, the mise tasks, the codegen outputs and the `LOG.md` protocol from `CLAUDE.md`.

| Piece | What it does |
| --- | --- |
| `hooks/migration-lint.sh` | PostToolUse on Edit/Write of `services/core/migrations/*.sql`: sequential gap-free `NNNNN_` name, one `-- +goose Up` and `Down`, no commented-out annotations, functions / `DO` inside StatementBegin/End, `CONCURRENTLY` only with `NO TRANSACTION`. Findings go back to Claude (exit 2) and to `LOG.md` as a `hook` line on `feature-<n>/*`. Also runs on files: `migration-lint.sh <file>...`. |
| `hooks/gen-drift.sh` | Stop: when `openapi.yaml`, `queries/*.sql` or a migration changed but no generated file is newer, asks Claude once to run `mise run gen` before finishing (what CI's `gen:check` would catch later). |
| `hooks/block-secrets.sh` | PreToolUse on Bash: blocks `git add` / `git commit` that would stage `.env`, `.env.*` (not `.env.example`), `*.pem`, `*.key` or `id_rsa*`. Checks explicit paths, what `git add --dry-run` with the same args would add (so `-f`, `-A`, `.`), the staged set and `-a` for commits; follows `cd` and `git -C`, ignores quoted text and heredocs. Exit 2 with the reason; logs a `hook` line on `feature-<n>/*`. |
| `settings.json` `permissions.deny` | `git reset --hard` and force-push are denied outright (see `## Git` in `CLAUDE.md`). |
| `commands/green.md` | `/green [step]`: `mise run lint` -> `test` -> `e2e`, stops on the first failure, summarizes it, logs a `check` line. |
| `commands/roundtrip.md` | `/roundtrip <N>`: up -> down-to N -> up with `pg_dump --schema-only` diffs, via `scripts/roundtrip.sh`. |
| `skills/promote-migration/` | Moves staged `docs/features/<slug>/migrations/*.up.sql` + `.down.sql` into goose files (next number, NO TRANSACTION, lint), then gen, lint, roundtrip, test. |
| `rules/` | Path rules for migrations and the OpenAPI contract (bootstrapped by the sdlc skills). |

Hooks are registered in `settings.json` and need only bash and jq; without jq they do nothing.

## Relation to the plugins

- **petforge** (`/petforge:slice*`, `hooks/guard.sh`, `pattern-scout`): scaffolds a whole new
  context and blocks edits to generated code, cross-context imports and impure domain packages.
  Nothing here repeats that: these hooks check migration files and codegen freshness, which the
  guard does not look at.
- **sdlc** (`/sdlc:*`) and the **ao-*** skills (symlinked from petforge, not committed): write
  the feature artifacts in `docs/features/<slug>/`, including the staged migrations that
  `promote-migration` takes live and `/roundtrip` verifies.
