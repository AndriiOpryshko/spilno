#!/usr/bin/env bash
# Ralph loop for one story: re-run one PROMPT.md through a cold `claude -p`
# until the story's own check passes. Course lesson 7.2 pattern plus extra nets.
#
#   scripts/ralph.sh docs/features/<slug>/ralph/<story>
#
# The story dir holds PROMPT.md and check.sh (binary DoD, exit 0 = done); the
# loop writes cost.log, progress.md (the model's notes), DONE / BLOCKED there.
#
# Safety nets:
#   1. MAX_ITER       hard iteration ceiling (default 5)
#   2. cost.log       per-iteration cost, turns, duration; MAX_USD caps the run
#   3. trap INT/TERM  Ctrl-C stops the loop, logs where, keeps the tree as is
#   4. ITER_TIMEOUT   wall clock per iteration; ITER_USD budget per iteration
#   5. DONE is verified: the loop runs check.sh itself and ignores a false DONE
#   6. tools are an allowlist in acceptEdits mode, never skip-permissions
set -euo pipefail

STORY_DIR="${1:?usage: scripts/ralph.sh <story-dir with PROMPT.md and check.sh>}"
MAX_ITER="${MAX_ITER:-5}"
MAX_USD="${MAX_USD:-15}"
ITER_USD="${ITER_USD:-5}"
ITER_TIMEOUT="${ITER_TIMEOUT:-20m}"
CLAUDE_BIN="${CLAUDE_BIN:-claude}"

cd "$(git rev-parse --show-toplevel)"
PROMPT="$STORY_DIR/PROMPT.md"
CHECK="$STORY_DIR/check.sh"
LOG="$STORY_DIR/cost.log"
DONE="$STORY_DIR/DONE"
BLOCKED="$STORY_DIR/BLOCKED"
for f in "$PROMPT" "$CHECK"; do
  [ -f "$f" ] || { echo "missing $f" >&2; exit 2; }
done

ALLOWED_TOOLS=(
  Read Edit Write Glob Grep
  "Bash(go:*)" "Bash(gofmt:*)" "Bash($CHECK)" "Bash(./$CHECK)" "Bash(bash $CHECK)"
  "Bash(git status:*)" "Bash(git diff:*)" "Bash(git log:*)" "Bash(git show:*)"
  "Bash(git add:*)" "Bash(git commit:*)"
  "Bash(touch $DONE)" "Bash(ls:*)" "Bash(cat:*)"
)
DISALLOWED_TOOLS=("Bash(git push:*)" "Bash(git reset:*)" "Bash(git rebase:*)" "Bash(gh:*)" WebFetch WebSearch)

ITER=0
SPENT=0
log() { echo "$(date -Iseconds) $*" | tee -a "$LOG"; }

on_interrupt() {
  log "INTERRUPTED at iter=$ITER spent_usd=$SPENT; tree left as is, see git status"
  exit 130
}
trap on_interrupt INT TERM

rm -f "$DONE"
log "=== run start story=$STORY_DIR max_iter=$MAX_ITER max_usd=$MAX_USD head=$(git rev-parse --short HEAD)"

while :; do
  if [ -f "$BLOCKED" ]; then
    log "=== BLOCKED after iter=$ITER: $(head -c 300 "$BLOCKED")"
    exit 3
  fi
  ITER=$((ITER + 1))
  if [ "$ITER" -gt "$MAX_ITER" ]; then
    log "=== MAX_ITER=$MAX_ITER reached without a passing check; spent_usd=$SPENT"
    exit 1
  fi

  log "--- iter=$ITER start"
  out="$STORY_DIR/iter-$ITER.json"
  # Cold start: a fresh session each turn, nothing carried over but the disk.
  set +e
  timeout --foreground "$ITER_TIMEOUT" "$CLAUDE_BIN" -p "$(cat "$PROMPT")" \
    --output-format json --no-session-persistence \
    --permission-mode acceptEdits --max-budget-usd "$ITER_USD" \
    --allowedTools "${ALLOWED_TOOLS[@]}" --disallowedTools "${DISALLOWED_TOOLS[@]}" \
    >"$out" 2>"$STORY_DIR/iter-$ITER.err"
  rc=$?
  set -e
  # timeout (and claude) catch SIGINT and exit 130 instead of dying from it, so
  # bash never runs the INT trap on its own: treat 130 as the interrupt.
  [ "$rc" -eq 130 ] && on_interrupt
  [ "$rc" -eq 124 ] && log "--- iter=$ITER killed by ITER_TIMEOUT=$ITER_TIMEOUT"

  cost=$(jq -r '.total_cost_usd // 0' "$out" 2>/dev/null || true)
  turns=$(jq -r '.num_turns // "?"' "$out" 2>/dev/null || true)
  secs=$(jq -r '((.duration_ms // 0) / 1000 | floor)' "$out" 2>/dev/null || true)
  cost=${cost:-0} turns=${turns:-?} secs=${secs:-?}
  SPENT=$(awk -v a="$SPENT" -v b="$cost" 'BEGIN { printf "%.4f", a + b }')
  log "--- iter=$ITER end rc=$rc cost_usd=$cost turns=$turns secs=$secs spent_usd=$SPENT"
  jq -r '.result // empty' "$out" 2>/dev/null | tail -5 | sed 's/^/    | /'

  if [ -f "$DONE" ]; then
    if verdict=$("$CHECK" 2>&1 | tail -1) && [ "$verdict" = PASS ]; then
      log "=== DONE verified by check.sh after iter=$ITER; spent_usd=$SPENT head=$(git rev-parse --short HEAD)"
      exit 0
    fi
    log "--- false DONE at iter=$ITER: $verdict; removed, looping on"
    rm -f "$DONE"
  fi
  if awk -v s="$SPENT" -v m="$MAX_USD" 'BEGIN { exit !(s >= m) }'; then
    log "=== MAX_USD=$MAX_USD reached after iter=$ITER; spent_usd=$SPENT"
    exit 4
  fi
  sleep 1 # a window for Ctrl-C between turns
done
