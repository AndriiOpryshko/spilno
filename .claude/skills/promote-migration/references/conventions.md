# Staged -> goose: what changes and what does not

Sources: `docs/adr/0004-migrations-and-queries.md`, `.claude/rules/migrations.md`
("spilno specifics"), and the already promoted 00004..00008 (board-protocol, Stage 1).

## Framing (always)

| Staged | Live goose file |
| --- | --- |
| `<YYYYMMDDhhmmss>_<name>.up.sql` + `.down.sql` | `NNNNN_<name>.sql`, next free number |
| Up file body | after `-- +goose Up` |
| Down file body | after `-- +goose Down`, same file |
| `CREATE INDEX CONCURRENTLY` | first line `-- +goose NO TRANSACTION`, one statement per file |
| `CREATE FUNCTION`, `DO $$ ... $$` | inside `-- +goose StatementBegin` / `-- +goose StatementEnd` |

Goose splits statements on `;` and still parses an annotation inside a comment
(`-- -- +goose Up`), so a commented-out annotation must be deleted, not commented.

## Content (only with the user's yes)

The staged files follow the sdlc defaults (`VARCHAR(N)`, `IF NOT EXISTS`, no `CHECK`); 00004
went live with `text` + `CHECK (char_length(...) BETWEEN ...)`, no `IF NOT EXISTS`, and folded
the bootstrap seed in. Such a change is a design decision: ask, then sync the staged file and
`data-model.md`. Triggers are allowed only for the board-protocol ADR-0007 tables.

## Partial matches

A live file can cover part of a staged one: `00005_add_association_to_meetings.sql` adds
`association_id`, staged `..._add_association_and_chair_to_meetings` also adds the chair column.
Promote only the remainder as a new migration (e.g. `<next>_add_chair_to_meetings.sql`, made by
hand with `mise run migrate:new add_chair_to_meetings`), never by editing 00005.

## Expand / backfill / contract

Keep the staged split: a nullable column, the backfill and `SET NOT NULL` are three files and
three deploys (00005, 00006, 00007). A backfill note next to the staged files (`*.md`) explains
the data part; link it from the backfill migration's comment.
