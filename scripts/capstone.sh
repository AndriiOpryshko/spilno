#!/usr/bin/env bash
# Runs one capstone feature end to end with headless Claude Code and records metrics.
#
# Usage: scripts/capstone.sh f1|f2
#
#   f1  Contract in context finance, plain prompts, no plugin.
#   f2  Meeting in context governance, prompts that call petforge commands.
#       Loads the plugin from $PETFORGE_DIR (default: ../petforge next to this repo).
#
# What it does:
#   1. Creates a git worktree ../spilno-<f> on a new branch from origin/main.
#   2. Starts the timer in LOG.md, then sends every prompt from scripts/capstone/<f>.prompts
#      to one Claude session (claude -p, resumed by session id) and logs each prompt.
#   3. Runs the green check (mise run lint + mise run test). On failure it sends a fix prompt
#      with the failing output, up to $MAX_FIXES times; each counts as a prompt.
#   4. Stops the timer, writes a summary line, commits, pushes the branch and prints the
#      numbers for the capstone post.
#
# Environment: MAX_FIXES (default 3), CLAUDE_MODEL (default: CLI default),
# CAPSTONE_PERMISSION_MODE (default: auto), PETFORGE_DIR, CAPSTONE_PUSH=0 to skip the push.
set -euo pipefail

feature="${1:-}"
case "$feature" in
  f1) label=F1; branch="feature-1/contract"; entity=Contract ;;
  f2) label=F2; branch="feature-2/meeting"; entity=Meeting ;;
  *) echo "usage: scripts/capstone.sh f1|f2"; exit 2 ;;
esac

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
prompts="$root/scripts/capstone/$feature.prompts"
wt="$(dirname "$root")/spilno-$feature"
runs="$(dirname "$root")/spilno-$feature.runs"
max_fixes="${MAX_FIXES:-3}"
permission_mode="${CAPSTONE_PERMISSION_MODE:-auto}"

claude_args=(--output-format json --permission-mode "$permission_mode"
  --append-system-prompt "A runner script records prompts, checks and the timer in LOG.md. Do not edit LOG.md yourself. Do not commit or push; the runner does that.")
[ -n "${CLAUDE_MODEL:-}" ] && claude_args+=(--model "$CLAUDE_MODEL")
if [ "$feature" = f2 ]; then
  petforge="${PETFORGE_DIR:-$(dirname "$root")/petforge}"
  [ -d "$petforge/skills" ] || { echo "petforge plugin not found at $petforge (set PETFORGE_DIR)"; exit 1; }
  claude_args+=(--plugin-dir "$petforge")
fi

[ -f "$prompts" ] || { echo "missing $prompts"; exit 1; }
[ ! -e "$wt" ] || { echo "$wt already exists; remove it with: git worktree remove $wt"; exit 1; }
if git -C "$root" show-ref -q --verify "refs/heads/$branch"; then
  echo "branch $branch already exists; delete it (git branch -D $branch) if it holds nothing you need"; exit 1
fi

now() { date +%Y-%m-%dT%H:%M:%S%z; }
log() { echo "- $(now) | $label | $*" >> "$wt/LOG.md"; }

echo "== setup (not timed): worktree $wt on $branch"
git -C "$root" fetch -q origin main
git -C "$root" worktree add -q -b "$branch" "$wt" origin/main
mkdir -p "$runs"
cd "$wt"
mise trust > /dev/null
mise exec -- pnpm install --frozen-lockfile --silent

session=""
n_prompts=0
n_fixes=0
turns=0
cost=0

# send <prompt>: one turn of the shared session; updates the counters.
send() {
  local prompt="$1" out
  n_prompts=$((n_prompts + 1))
  out="$runs/$(printf '%02d' "$n_prompts").json"
  log "prompt | \"$(printf '%s' "$prompt" | head -c 300 | tr '\n' ' ')\""
  echo "== prompt $n_prompts: $(printf '%s' "$prompt" | head -c 100 | tr '\n' ' ')"
  local resume=()
  [ -n "$session" ] && resume=(--resume "$session")
  claude -p "$prompt" "${claude_args[@]}" ${resume[@]+"${resume[@]}"} < /dev/null > "$out" || true
  if ! jq -e . "$out" > /dev/null 2>&1; then
    log "stumble | claude returned no JSON for prompt $n_prompts (see $out)"
    return
  fi
  session="$(jq -r '.session_id // empty' "$out")"
  turns=$((turns + $(jq -r '.num_turns // 0' "$out")))
  cost="$(jq -n --argjson a "$cost" --argjson b "$(jq '.total_cost_usd // 0' "$out")" '$a + $b')"
  [ "$(jq -r '.is_error' "$out")" = "true" ] && log "stumble | claude reported an error on prompt $n_prompts"
  jq -r '.result // ""' "$out" | tail -5
}

green() {
  mise run lint > "$runs/check.log" 2>&1 && mise run test >> "$runs/check.log" 2>&1
}

scripts/log.sh start "$label"

while IFS= read -r line <&3 || [ -n "$line" ]; do
  case "$line" in ''|'#'*) continue ;; esac
  send "$line"
done 3< "$prompts"

until green; do
  log "check | mise run lint + mise run test -> fail"
  if [ "$n_fixes" -ge "$max_fixes" ]; then
    log "stumble | still failing after $max_fixes fix prompts; stopped"
    echo "still red after $max_fixes fix prompts; see $runs/check.log"
    exit 1
  fi
  n_fixes=$((n_fixes + 1))
  send "The green check failed. Fix the cause, then run mise run lint and mise run test until both pass. Last output:
$(tail -60 "$runs/check.log")"
done
log "check | mise run lint + mise run test -> pass"

minutes="$(scripts/log.sh stop "$label" | grep -Eo '[0-9]+ min' | grep -Eo '[0-9]+')"
hooks="$(grep -c "| $label | hook |" LOG.md || true)"
log "summary | minutes=$minutes prompts=$n_prompts fix_prompts=$n_fixes turns=$turns cost_usd=$(printf '%.2f' "$cost") hook_hits=$hooks"

git add -A
if [ "$feature" = f1 ]; then
  msg="F1: $entity vertical slice, built without the plugin"
else
  msg="F2: $entity vertical slice, built with petforge

Created through petforge; commands are listed in scripts/capstone/f2.prompts."
fi
git commit -q -m "$msg

Metrics: ${minutes} min, ${n_prompts} prompts (${n_fixes} fix), ${turns} turns, hook hits ${hooks}.

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
pushed="not pushed (CAPSTONE_PUSH=0)"
if [ "${CAPSTONE_PUSH:-1}" = 1 ]; then
  git push -q -u origin "$branch"
  pushed=pushed
fi

cat <<EOF

== $label done
Entity:     $entity
Time:       $minutes min (first prompt to green)
Prompts:    $n_prompts ($n_fixes fix prompts)
Turns:      $turns, cost \$$(printf '%.2f' "$cost")
Hook hits:  $hooks
Branch:     $branch ($pushed)
Open PR:    https://github.com/AndriiOpryshko/spilno/compare/main...$branch
Run files:  $runs
EOF
