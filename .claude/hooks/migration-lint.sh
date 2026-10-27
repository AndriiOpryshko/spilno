#!/usr/bin/env bash
# Lints goose migrations in services/core/migrations/ (ADR 0004, .claude/rules/migrations.md).
#   PostToolUse (Edit|Write|MultiEdit): reads the hook JSON on stdin; findings go to stderr with
#   exit 2, so Claude sees them and fixes the file. Hits are logged to LOG.md on feature-<n>/*.
#   Direct: migration-lint.sh <file.sql>... (exit 1 on findings), used by promote-migration.
# Checks: NNNNN_<name>.sql with a unique, gap-free number; exactly one live `-- +goose Up` and
# `-- +goose Down`; no commented-out annotations (goose still parses `-- -- +goose Up`);
# functions and DO blocks inside balanced StatementBegin/End; CONCURRENTLY only with
# `-- +goose NO TRANSACTION`.
set -uo pipefail

lint() {
  local f="$1" dir base n
  dir="$(dirname "$f")"; base="$(basename "$f")"
  [ -f "$f" ] || return 0
  if [[ ! "$base" =~ ^([0-9]{5})_[a-z0-9_]+\.sql$ ]]; then
    echo "$base: name must be NNNNN_<snake_name>.sql (mise run migrate:new <name>)"
  else
    n="${BASH_REMATCH[1]}"
    local dup; dup="$(find "$dir" -maxdepth 1 -name "${n}_*.sql" | wc -l | tr -d ' ')"
    [ "$dup" -gt 1 ] && echo "$base: number $n is used by $dup files; take the next free one"
    local max i; max="$(find "$dir" -maxdepth 1 -name '[0-9][0-9][0-9][0-9][0-9]_*.sql' -exec basename {} \; | sort | tail -1 | cut -c1-5)"
    for ((i = 1; i < 10#$max; i++)); do
      ls "$dir"/"$(printf %05d "$i")"_*.sql > /dev/null 2>&1 || echo "$base: gap in the sequence, $(printf %05d "$i") is missing"
    done
  fi
  awk -v f="$base" '
    function say(m) { print f ":" NR ": " m }
    /^[[:space:]]*--[[:space:]]*--+[[:space:]]*\+goose/ { say("commented-out goose annotation is still parsed by goose; delete the line"); next }
    /^-- \+goose Up[[:space:]]*$/            { up++ }
    /^-- \+goose Down[[:space:]]*$/          { down++ }
    /^-- \+goose NO TRANSACTION[[:space:]]*$/ { notx = 1 }
    /^-- \+goose StatementBegin/ { if (inb) say("nested StatementBegin"); inb = 1; next }
    /^-- \+goose StatementEnd/   { if (!inb) say("StatementEnd without StatementBegin"); inb = 0; next }
    /^[[:space:]]*--/ { next }
    toupper($0) ~ /(^|[^A-Z_])DO[[:space:]]+(\$|LANGUAGE)|CREATE[[:space:]]+(OR[[:space:]]+REPLACE[[:space:]]+)?(FUNCTION|PROCEDURE)/ {
      if (!inb) say("function / DO block outside -- +goose StatementBegin/StatementEnd (goose splits on ;)")
    }
    toupper($0) ~ /CONCURRENTLY/ { conc = NR }
    END {
      if (up != 1)   print f ": needs exactly one `-- +goose Up` (found " up+0 ")"
      if (down != 1) print f ": needs exactly one `-- +goose Down` (found " down+0 "); ADR 0004 wants Up and Down in one file"
      if (inb)       print f ": StatementBegin is never closed"
      if (conc && !notx) print f ":" conc ": CONCURRENTLY needs `-- +goose NO TRANSACTION` on the first line and one statement per file"
    }' "$f"
}

if [ $# -gt 0 ]; then
  rc=0
  for f in "$@"; do out="$(lint "$f")"; [ -n "$out" ] && { echo "$out"; rc=1; }; done
  exit $rc
fi

command -v jq > /dev/null || exit 0
input="$(cat)"
file="$(jq -r '.tool_input.file_path // empty' <<< "$input")"
case "$file" in */services/core/migrations/*.sql | services/core/migrations/*.sql) ;; *) exit 0 ;; esac
cwd="$(jq -r '.cwd // empty' <<< "$input")"
[[ "$file" = /* ]] || file="${cwd:-.}/$file"
out="$(lint "$file")"
[ -z "$out" ] && exit 0

{ echo "migration-lint: fix these before moving on"; echo "$out"; } >&2
root="$(git -C "$(dirname "$file")" rev-parse --show-toplevel 2> /dev/null || true)"
branch="$(git -C "${root:-.}" branch --show-current 2> /dev/null || true)"
if [ -f "$root/LOG.md" ] && [[ "$branch" =~ ^feature-([0-9]+)/ ]]; then
  echo "- $(date +%Y-%m-%dT%H:%M:%S%z) | F${BASH_REMATCH[1]} | hook | migration-lint ${file#"$root"/}: $(head -1 <<< "$out" | cut -d: -f2- | sed 's/^ *//')" >> "$root/LOG.md"
fi
exit 2
