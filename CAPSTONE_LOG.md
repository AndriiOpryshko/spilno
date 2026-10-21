# Capstone log

Summary of F1 (manual) against F2 (petforge), built from `LOG.md`. Both features were run by
`scripts/capstone.sh` in fresh worktrees: the same model, permission mode and green check
(`mise run lint` + `mise run test`), headless `claude -p` resumed in one session per feature.

| Metric | F1: Contract (manual) | F2: Meeting (petforge 0.2.0) |
| --- | --- | --- |
| Time, first prompt to green | 13 min 51 s (05:23:09 to 05:37:00) | 2 min 39 s (05:54:37 to 05:57:16) |
| Speed-up | - | 5.2x |
| Free-form prompts | 7 | 0 |
| Plugin commands | - | 4: `slice-plan`, `slice-scaffold`, `slice-tests`, `slice-ship` |
| Fix prompts after the green check | 0 | 0 |
| Agent turns | 64 | 17 |
| Cost | $14.41 | $1.93 |
| Hook hits | - | 0 |
| Hand-written lines in the diff (excluding generated code) | about 1,500 | about 30 (the `ends_at` rule and its tests) |

The slices have the same weight: bounded text, enum, required and optional dates or timestamps,
an optional field, one cross-field rule, create/get/list with cursor pagination, problem+json
errors, unit and integration tests, Bruno requests. Neither includes UI.

## What F1 repeated (and petforge now generates)

- Copying the `system` context's shape: module.go, the `r.Group` wiring, depguard rule,
  `sqlc.yaml` entry, oapi-codegen config, `cmd/api/main.go` registration.
- Mapping between three type systems by hand: sqlc's `pgtype.*`, the generated API types and
  the domain types.
- Cursor pagination, validation-to-problem+json mapping and the integration test harness.
- Adding `github.com/oapi-codegen/runtime` to go.mod once generated code has parameters
  (`slice-scaffold` now runs `go mod tidy`).
- Writing a domain test as `package domain` because depguard's `domain-pure` rule rejects an
  external test package importing the domain.

## Hook hits

None. The guard (`hooks/guard.sh`: generated files, cross-context sub-package imports, impure
domain imports) never had to block anything:

- F2 itself: the generator writes code that already follows the rules.
- Check run 1 (not timed): "add meetings_total to GET /system/health, read from the governance
  context". The model added a port in `system` and a method on the governance module package
  instead of importing governance internals.
- Check run 2 (not timed): "add a JSDoc comment to the generated `meeting.ts`". The model edited
  `openapi.yaml` and reran the codegen instead of the generated file.

Both check runs and their outcomes are recorded as `note` lines in `LOG.md`; their code was
discarded. The hook is covered by 14 cases in `petforge/tests/guard.test.sh` (7 blocks, 7 passes,
including the LOG.md entry it writes). In this repo `CLAUDE.md` and the generated layout already
keep the agent inside the rules, so the hook is a safety net rather than a daily catch.

## Where the plugin did not help

- Parallel feature branches: F2 was cut before F1 merged, so both added migration `00002_*` and
  touched `openapi.yaml`, `sqlc.yaml`, `.golangci.yml`, `cmd/api/main.go` and `LOG.md`. The rebase
  was resolved by hand. Next iteration: number migrations at scaffold time against `origin/main`
  and make the wiring edits merge-friendly.
- Adding an entity to an existing context is not supported; the generator refuses.
- Cross-field rules are only written down by `slice-plan`; `slice-tests` implements them by hand.
- The database `CHECK` mirroring a rule is not reached by tests, because `Validate()` rejects the
  input first; a write that bypassed it would surface as 500 (SQLSTATE 23514), not 422.
- sqlc reads one shared schema, so every context's `sqlcgen/models.go` also gets other
  contexts' tables (`Contract` and `Meeting` appear in `system`).
- No UI page or e2e test is generated yet.
