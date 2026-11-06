#!/usr/bin/env bash
# /goal run for T13a: one headless `claude -p "/goal …"` invocation, same
# permission allowlist, mode and budget as scripts/ralph.sh, so the two
# methods differ only in what drives the next turn (a fresh evaluator in one
# warm session vs. a cold process per iteration).
#
#   docs/features/board-protocol/execution/T13a/goal.sh
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"

DIR=docs/features/board-protocol/execution/T13a
CHECK="$DIR/check.sh"
MAX_USD="${MAX_USD:-15}"
OUT="$DIR/goal.jsonl"

read -r -d '' CONDITION <<'EOF' || true
T13a is implemented and committed in this repo (spilno). Context: read CLAUDE.md (Layout rules, Git), the story docs/features/board-protocol/tasks/T13a-files-port-fake.md, and the acceptance tests services/core/internal/governance/adapters/filestest/contract.go and services/core/internal/governance/adapters/filesfake/fake_test.go, whose "Expected API" headers are the contract: create governance/ports/files.go (the Files port) and governance/adapters/filesfake/fake.go (the in-memory fake), add CodeStorageUnavailable to domain/errors.go and use it in ErrorRegistry, mark the T13a row of docs/features/board-protocol/tasks/tracker.md done, and commit only those files with one commit "feat(governance): <imperative summary>" whose body says why and ends with the trailers "SDLC-Task: T13a", one "SDLC-AC: <id>" per story AC and "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>" (git add <path> only, never -A; no push, no amend). Done means: the command docs/features/board-protocol/execution/T13a/check.sh (run exactly by that path) prints PASS and exits 0 in the transcript after the commit, the two acceptance test files are unchanged, and no file outside governance/ports, governance/adapters/filesfake, governance/domain/errors.go and the tracker is changed; or stop after 15 turns.
EOF

ALLOWED_TOOLS=(
  Read Edit Write Glob Grep
  "Bash(go:*)" "Bash(gofmt:*)" "Bash($CHECK)" "Bash(./$CHECK)" "Bash(bash $CHECK)"
  "Bash(git status:*)" "Bash(git diff:*)" "Bash(git log:*)" "Bash(git show:*)"
  "Bash(git add:*)" "Bash(git commit:*)" "Bash(ls:*)" "Bash(cat:*)"
)
DISALLOWED_TOOLS=("Bash(git push:*)" "Bash(git reset:*)" "Bash(git rebase:*)" "Bash(gh:*)" WebFetch WebSearch)

start=$(date +%s)
echo "$(date -Iseconds) goal start head=$(git rev-parse --short HEAD)" | tee "$DIR/goal.log"
set +e
claude -p "/goal $CONDITION" \
  --output-format stream-json --verbose --no-session-persistence \
  --permission-mode acceptEdits --max-budget-usd "$MAX_USD" \
  --allowedTools "${ALLOWED_TOOLS[@]}" --disallowedTools "${DISALLOWED_TOOLS[@]}" \
  >"$OUT" 2>"$DIR/goal.err"
rc=$?
set -e
end=$(date +%s)
result=$(grep '"type":"result"' "$OUT" | tail -1)
cost=$(jq -r '.total_cost_usd // 0' <<<"$result" 2>/dev/null || echo "?")
turns=$(jq -r '.num_turns // "?"' <<<"$result" 2>/dev/null || echo "?")
echo "$(date -Iseconds) goal end rc=$rc wall_s=$((end - start)) cost_usd=$cost turns=$turns head=$(git rev-parse --short HEAD)" | tee -a "$DIR/goal.log"
echo "independent check: $("$CHECK" 2>&1 | tail -1)" | tee -a "$DIR/goal.log"
