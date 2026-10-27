#!/usr/bin/env bash
# Stop hook: blocks the end of a turn once when codegen inputs changed in the working tree
# (api/openapi.yaml, queries/*.sql, migrations/*.sql) but no generated file is newer than them,
# i.e. `mise run gen` was not run after the edit. CI's `mise run gen:check` would fail later.
# Fast (git status + find -newer, no codegen). Needs jq; without it the hook does nothing.
set -uo pipefail

command -v jq > /dev/null || exit 0
input="$(cat)"
[ "$(jq -r '.stop_hook_active // false' <<< "$input")" = "true" ] && exit 0
cwd="$(jq -r '.cwd // empty' <<< "$input")"
root="$(git -C "${cwd:-.}" rev-parse --show-toplevel 2> /dev/null)" || exit 0
cd "$root" || exit 0

core=services/core
api_out=("$core"/internal/*/adapters/http/gen/*.gen.go apps/web/src/api/gen)
sql_out=("$core"/internal/*/adapters/postgres/sqlcgen)

stale=()
while IFS= read -r src; do
  [ -f "$src" ] || continue
  case "$src" in
    "$core"/api/openapi.yaml) out=("${api_out[@]}") ;;
    "$core"/internal/*/adapters/postgres/queries/*.sql | "$core"/migrations/*.sql) out=("${sql_out[@]}") ;;
    *) continue ;;
  esac
  [ -n "$(find "${out[@]}" -type f -newer "$src" -print -quit 2> /dev/null)" ] || stale+=("$src")
done < <(git status --porcelain --untracked-files=all -- "$core/api/openapi.yaml" "$core/internal" "$core/migrations" | cut -c4- | sed 's/.* -> //')

[ ${#stale[@]} -eq 0 ] && exit 0
reason="gen-drift: generated code is older than ${stale[*]}. Run \`mise run gen\` (then \`go build ./...\` in services/core) before finishing; never hand-edit the generated files."
jq -n --arg r "$reason" '{decision: "block", reason: $r}'
