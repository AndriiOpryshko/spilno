---
name: promote-migration
description: Promote a feature's staged migrations (docs/features/<slug>/migrations/<ts>_<name>.up.sql + .down.sql, written by sdlc:generate-data-model or ao-data-model) into goose files in services/core/migrations/, then lint, regenerate and roundtrip them. Use when implementing a feature whose data model is staged, or when asked to "promote migrations for <slug>", "move the staged migrations live", "/promote-migration <slug>", "перенеси міграції фічі в goose".
argument-hint: "<slug> [name or staged file to promote only that one]"
---

# promote-migration

Staged migrations are the reviewed spec of a feature's schema; the live tree is goose
(ADR 0004: one file per migration, `-- +goose Up` / `-- +goose Down`, sequential `NNNNN_`
names, embedded into the binary). This skill turns one into the other without losing the
review: the SQL moves as written, only goose framing is added.

## Steps

1. **Status.** From the repo root run
   `.claude/skills/promote-migration/scripts/promote.sh status <slug>`. It prints every staged
   pair and its live file with the same name, or `staged only`. Also compare by content: a live
   file under another name may already cover part of a staged one (see
   `references/conventions.md`, "Partial matches"). Show the plan (which pairs, in staged order)
   and confirm it with the user before writing anything.
2. **Promote, one pair at a time, in staged timestamp order.**
   `.claude/skills/promote-migration/scripts/promote.sh <dir>/<ts>_<name>.up.sql` writes
   `services/core/migrations/<next>_<name>.sql` (adds `-- +goose NO TRANSACTION` when the pair
   uses `CONCURRENTLY`) and runs `.claude/hooks/migration-lint.sh` on it. Exit 2 means lint
   findings: fix them in the new file (the same hook re-checks every edit).
3. **Adapt only what goose needs**, per `references/conventions.md`: StatementBegin/End around
   functions and `DO` blocks, one `CONCURRENTLY` statement per file, Down in reverse order.
   Any other change to the staged SQL (types, CHECKs, defaults) is a design change: ask, and
   if accepted, update the staged file and `data-model.md` too so they stay the spec.
4. **Regenerate and verify**, from the repo root:
   - `mise run gen` (sqlc reads the migrations' Up sections), then `mise run lint`;
   - `/roundtrip <N>` with N = the last migration version on `main`
     (`git ls-tree --name-only origin/main services/core/migrations/ | sort | tail -1`);
   - `mise run test` (testdb applies every migration to a fresh database).
5. **Report**: the files written (staged -> live), the adaptations made and why, the roundtrip
   and test results. Do not commit unless asked.

## Rules

- Never edit a live migration that is already on `main`; add a new one instead.
- Never rename staged files or renumber live ones to make the sequence fit.
- The `.down.sql` must fully reverse its `.up.sql`; a missing pair is a stop, not a TODO.
