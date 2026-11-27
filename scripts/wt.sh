#!/usr/bin/env bash
# Gives a git worktree its own environment: API and web ports, its own database, its own deps.
#
#   scripts/wt.sh new <slug> [branch]  sibling worktree ../spilno-<slug> off origin/main (branch feat/<slug>),
#                                      copies the .worktreeinclude files, then runs `init` in it
#   scripts/wt.sh init                 inside any linked worktree (also `claude --worktree` and
#                                      `isolation: worktree` ones under .claude/worktrees/): ports, database, deps
#   scripts/wt.sh info                 what this worktree got
#
# Everything lands in the worktree's mise.local.toml (git-ignored, never in .worktreeinclude): mise loads it over
# mise.toml, so `mise run dev:core` / `mise run dev:web` / `mise run e2e` pick the worktree's values. Ports are
# reserved, not only free: a port another worktree's mise.local.toml already holds is skipped even if nothing
# listens on it yet. The database spilno_wt_<slug> lives in the shared Postgres (`mise run infra:up`, :5432).
set -euo pipefail

PG_ADMIN_URL="${WT_PG_ADMIN_URL:-postgres://spilno:spilno@localhost:5432/postgres?sslmode=disable}"
API_PORTS=(18080 18099)
WEB_PORTS=(15180 15199)

die() { echo "wt: $*" >&2; exit 1; }

is_main_checkout() {
  [ "$(git rev-parse --path-format=absolute --git-dir)" = "$(git rev-parse --path-format=absolute --git-common-dir)" ]
}

# Ports written into any worktree's mise.local.toml.
reserved_ports() {
  git worktree list --porcelain | sed -n 's/^worktree //p' | while read -r dir; do
    [ -f "$dir/mise.local.toml" ] && grep -oE '^(WT_API_PORT|WT_WEB_PORT) = "[0-9]+"' "$dir/mise.local.toml" | grep -oE '[0-9]+'
  done || true
}

pick_port() {
  local lo="$1" hi="$2" taken p
  taken="$(reserved_ports) $(ss -Hltn | awk '{print $4}' | grep -oE '[0-9]+$')"
  for ((p = lo; p <= hi; p++)); do
    grep -qw "$p" <<<"$taken" || { echo "$p"; return; }
  done
  die "no free port in $lo-$hi"
}

cmd_init() {
  is_main_checkout && die "this is the main checkout; init runs inside a linked worktree"
  local root slug db api web
  root="$(git rev-parse --show-toplevel)"
  cd "$root"
  slug="$(basename "$root" | tr '[:upper:]' '[:lower:]' | sed 's/^spilno-//; s/[^a-z0-9]/_/g')"
  db="spilno_wt_${slug}"
  db="${db:0:63}"

  if [ -f mise.local.toml ] && grep -q '^WT_API_PORT' mise.local.toml; then
    api="$(sed -n 's/^WT_API_PORT = "\(.*\)"/\1/p' mise.local.toml)"
    web="$(sed -n 's/^WT_WEB_PORT = "\(.*\)"/\1/p' mise.local.toml)"
  else
    api="$(pick_port "${API_PORTS[@]}")"
    web="$(pick_port "${WEB_PORTS[@]}")"
  fi

  cat >mise.local.toml <<EOF
# Written by scripts/wt.sh init for worktree $(basename "$root"). Git-ignored; never copy it to another worktree.
[env]
WT_API_PORT = "$api"
WT_WEB_PORT = "$web"
HTTP_ADDR = ":$api"
CORE_API_URL = "http://localhost:$api"
WEB_PORT = "$web"
E2E_WEB_PORT = "$web"
SPILNO_WEB_ORIGIN = "http://localhost:$web"
OTEL_SERVICE_NAME = "spilno-core-$slug"
MIGRATIONS_DATABASE_URL = "postgres://spilno:spilno@localhost:5432/$db?sslmode=disable"
DATABASE_URL = "postgres://spilno_app_login:spilno_app_login@localhost:5432/$db?sslmode=disable"
EOF
  mise trust -q mise.toml && mise trust -q mise.local.toml

  psql "$PG_ADMIN_URL" -qtAc 'select 1' >/dev/null 2>&1 || die "Postgres on :5432 is down; run 'mise run infra:up' in the main checkout"
  # The login role is cluster-wide; an old volume may predate it (CLAUDE.md gotcha).
  if [ -z "$(psql "$PG_ADMIN_URL" -qtAc "select 1 from pg_roles where rolname = 'spilno_app_login'")" ]; then
    psql "$PG_ADMIN_URL" -q -v ON_ERROR_STOP=1 -f infra/postgres-init/01-app-role.sql
  fi
  if [ -z "$(psql "$PG_ADMIN_URL" -qtAc "select 1 from pg_database where datname = '$db'")" ]; then
    psql "$PG_ADMIN_URL" -qc "create database \"$db\" owner spilno"
  fi

  echo "wt: installing dependencies in $root"
  pnpm install --frozen-lockfile --silent
  (cd services/core && go mod download)
  [ ! -d mcp/nbu-rates ] || (cd mcp/nbu-rates && go mod download)

  cmd_info
}

cmd_new() {
  local slug="${1:?usage: wt.sh new <slug> [branch]}" branch="${2:-feat/$1}" main here dir pattern
  # The ignored files live in the main checkout, whichever worktree this runs from.
  main="$(dirname "$(git rev-parse --path-format=absolute --git-common-dir)")"
  here="$(git rev-parse --show-toplevel)"
  dir="$(dirname "$main")/spilno-$slug"
  git fetch -q origin
  git worktree add -q "$dir" -b "$branch" origin/main
  # `git worktree add` ignores .worktreeinclude (only `claude --worktree` reads it): copy its ignored files by hand.
  if [ -f "$here/.worktreeinclude" ]; then
    grep -vE '^\s*(#|$)' "$here/.worktreeinclude" | while read -r pattern; do
      (cd "$main" && git ls-files --others --ignored --exclude-standard -- "$pattern") | while read -r f; do
        mkdir -p "$dir/$(dirname "$f")" && cp -P "$main/$f" "$dir/$f"
      done
    done
  fi
  (cd "$dir" && "$here/scripts/wt.sh" init)
}

cmd_info() {
  local root
  root="$(git rev-parse --show-toplevel)"
  [ -f "$root/mise.local.toml" ] || die "no mise.local.toml here; run scripts/wt.sh init"
  echo "worktree $(basename "$root") on $(git branch --show-current)"
  grep -E '^(HTTP_ADDR|WEB_PORT|DATABASE_URL)' "$root/mise.local.toml" | sed 's/^/  /'
  echo "  run: mise run dev:core   # API"
  echo "       mise run dev:web    # web, proxies /api to this worktree's API"
}

case "${1:-}" in
  new) shift; cmd_new "$@" ;;
  init) cmd_init ;;
  info) cmd_info ;;
  *) sed -n '2,9p' "$0"; exit 1 ;;
esac
