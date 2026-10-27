#!/usr/bin/env bash
# Promotes staged migration pairs (docs/features/<slug>/migrations/<ts>_<name>.up.sql + .down.sql)
# into goose files in services/core/migrations/.
#   promote.sh status <slug>         staged pairs and their live counterpart (same <name>), if any
#   promote.sh <path/to/X.up.sql>    writes NNNNN_<name>.sql (next free number) and lints it
set -euo pipefail

root="$(git rev-parse --show-toplevel)"
live="$root/services/core/migrations"
lint="$root/.claude/hooks/migration-lint.sh"
name_of() { basename "$1" | sed -E 's/^[0-9]+_//; s/\.(up|down)\.sql$//; s/\.sql$//'; }

if [ "${1:-}" = status ]; then
  dir="$root/docs/features/${2:?usage: promote.sh status <slug>}/migrations"
  [ -d "$dir" ] || { echo "no staged migrations in ${dir#"$root"/}"; exit 1; }
  for up in "$dir"/*.up.sql; do
    n="$(name_of "$up")"
    hit="$(find "$live" -maxdepth 1 -name "[0-9][0-9][0-9][0-9][0-9]_${n}.sql" -exec basename {} \; | head -1)"
    miss=""; [ -f "${up%.up.sql}.down.sql" ] || miss=" (no .down.sql)"
    printf '%-64s %s%s\n' "$(basename "$up")" "${hit:-staged only}" "$miss"
  done
  exit 0
fi

up="${1:?usage: promote.sh <X.up.sql> | status <slug>}"
[[ "$up" == *.up.sql && -f "$up" ]] || { echo "not a staged .up.sql: $up"; exit 1; }
down="${up%.up.sql}.down.sql"
[ -f "$down" ] || { echo "missing pair: $down"; exit 1; }
name="$(name_of "$up")"
if found="$(find "$live" -maxdepth 1 -name "[0-9][0-9][0-9][0-9][0-9]_${name}.sql" | head -1)" && [ -n "$found" ]; then
  echo "already live: ${found#"$root"/}"; exit 1
fi
last="$(find "$live" -maxdepth 1 -name '[0-9][0-9][0-9][0-9][0-9]_*.sql' -exec basename {} \; | sort | tail -1 | cut -c1-5)"
target="$live/$(printf %05d $((10#${last:-0} + 1)))_${name}.sql"

{
  grep -qi concurrently "$up" "$down" && echo "-- +goose NO TRANSACTION"
  echo "-- +goose Up"
  echo "-- Promoted from ${up#"$root"/}."
  cat "$up"
  echo
  echo "-- +goose Down"
  cat "$down"
} > "$target"
echo "wrote ${target#"$root"/}"
"$lint" "$target" || { echo "lint findings above: fix them in the new file"; exit 2; }
