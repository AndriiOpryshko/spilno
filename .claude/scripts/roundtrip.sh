#!/usr/bin/env bash
# Schema roundtrip on the local dev database: baseline N -> up (latest) -> down-to N -> up.
# Proves every Down past N restores the N schema and Up rebuilds the same latest schema
# (pg_dump --schema-only diffs). Destroys data above version N in the local database.
# Usage: .claude/scripts/roundtrip.sh <N>
set -euo pipefail

base="${1:?usage: roundtrip.sh <baseline version N>}"
[[ "$base" =~ ^[0-9]+$ ]] || { echo "N must be a number, got: $base"; exit 1; }
cd "$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel)"
export DATABASE_URL="${DATABASE_URL:-postgres://spilno:spilno@localhost:5432/spilno?sslmode=disable}"

if (exec 3<> /dev/tcp/127.0.0.1/8080) 2> /dev/null; then
  echo "core API is listening on :8080; stop it (mprocs pane 'core') - it applies migrations on start"; exit 1
fi

out="$(mktemp -d)"
# pg_dump 18 prints \restrict / \unrestrict lines with a random key: drop them before diffing.
dump() { docker compose -f infra/compose.yaml exec -T postgres \
  pg_dump -U spilno -d spilno --schema-only --no-owner --no-privileges | grep -v '^\\\(un\)\?restrict' > "$out/$1.sql"; }
G() { (cd services/core && go tool -modfile=tools/go.mod goose -dir migrations postgres "$DATABASE_URL" "$@"); }

mise run infra:up > /dev/null
G down-to "$base" > /dev/null 2>&1 || true
G up-to "$base" > /dev/null
echo "== baseline v$base";   G version; dump base-a
echo "== up -> latest";      G up;            dump top-a
echo "== down-to $base";     G down-to "$base"; dump base-b
echo "== up -> latest again"; G up;           dump top-b
echo
rc=0
diff -u "$out/base-a.sql" "$out/base-b.sql" && echo "OK   down restores the v$base schema exactly" || { echo "FAIL down-to $base leaves a different schema (diff above)"; rc=1; }
diff -u "$out/top-a.sql" "$out/top-b.sql" && echo "OK   up rebuilds the latest schema exactly" || { echo "FAIL second up differs from the first (diff above)"; rc=1; }
echo "dumps: $out"
G status
exit $rc
