#!/usr/bin/env bash
# Regenerates all code derived from api/openapi.yaml and the SQL queries.
# Convention: every bounded context keeps its oapi-codegen config at
# internal/<bc>/adapters/http/gen/oapi-codegen.yaml and its sqlc entry in sqlc.yaml.
set -euo pipefail

cd "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

tool() { go tool -modfile=tools/go.mod "$@"; }

for cfg in internal/*/adapters/http/gen/oapi-codegen.yaml; do
  echo "oapi-codegen $cfg"
  tool oapi-codegen -config "$cfg" api/openapi.yaml
done

echo "sqlc generate"
tool sqlc generate
