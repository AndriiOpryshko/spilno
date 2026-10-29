#!/usr/bin/env bash
# PreToolUse (Bash): blocks `git add` / `git commit` that would put a secret file into the index:
# .env, .env.* (except .env.example), *.pem, *.key, id_rsa*. Same idea as the course's
# block-env.sh, but it checks what git would actually stage instead of grepping the command:
#   git add   - explicit paths by name, plus `git add --dry-run <same args>` (catches -f, -A, ., dirs);
#   git commit - what is already staged, tracked changes for -a, explicit pathspecs.
# Follows `cd <dir>` and `git -C <dir>` inside compound commands. Exit 2 + stderr blocks the call.
# Hits are logged to LOG.md on feature-<n>/*. Needs jq; without it the hook does nothing.
set -uo pipefail

command -v jq > /dev/null || exit 0
input="$(cat)"
cmd="$(jq -r '.tool_input.command // empty' <<< "$input")"
[ -z "$cmd" ] && exit 0
grep -Eq '(^|[^[:alnum:]_-])git([[:space:]]|$)' <<< "$cmd" || exit 0
dir="$(jq -r '.cwd // empty' <<< "$input")"
dir="${dir:-$PWD}"

is_secret() {
  case "$(basename "$1")" in
    .env.example) return 1 ;;
    .env | .env.* | *.pem | *.key | id_rsa*) return 0 ;;
  esac
  return 1
}

hits=()
check() { local p; for p in "$@"; do is_secret "$p" && hits+=("$p"); done; }

# Shell-ish tokenizer: one simple command per line, words separated by \037. Honours quotes and
# backslashes, splits on unquoted ; & | and newlines (not >&), skips heredoc bodies.
split_cmd() {
  awk 'BEGIN { RS = "\001" }
  function emit() { if (have) { seg = seg (nw++ ? "\037" : "") tok }
    if (want) { hd = tok; want = 0 } else if (tok ~ /^<<-?[^<]/) hd = tok
    if (hd ~ /^<</) { sub(/^<<-?/, "", hd) }
    if (tok == "<<" || tok == "<<-") want = 1
    tok = ""; have = 0 }
  function brk() { emit(); if (nw) print seg; seg = ""; nw = 0 }
  { s = $0; n = length(s)
    for (i = 1; i <= n; i++) {
      c = substr(s, i, 1)
      if (q == "\047") { if (c == q) q = ""; else tok = tok c; continue }
      if (q == "\"") { if (c == "\\") { tok = tok substr(s, ++i, 1); continue } if (c == q) q = ""; else tok = tok c; continue }
      if (c == "\\") { tok = tok substr(s, ++i, 1); have = 1; continue }
      if (c == "\047" || c == "\"") { q = c; have = 1; continue }
      if (c == " " || c == "\t") { emit(); continue }
      if (c == "&" && (substr(s, i - 1, 1) == ">" || substr(s, i - 1, 1) == "<")) { tok = tok c; continue }
      if (c == ";" || c == "&" || c == "|") { brk(); continue }
      if (c == "\n") { brk()
        if (hd != "") { rest = substr(s, i + 1); m = split(rest, ln, "\n"); skip = 0
          for (k = 1; k <= m; k++) { skip += length(ln[k]) + 1; t = ln[k]; sub(/^\t+/, "", t); if (t == hd) break }
          i += skip; hd = "" }
        continue }
      tok = tok c; have = 1
    }
    brk() }'
}

while IFS= read -r seg; do
  IFS=$'\037' read -ra raw <<< "$seg"
  w=(); skipnext=0
  for t in "${raw[@]}"; do
    if [ $skipnext = 1 ]; then skipnext=0; continue; fi
    [[ "$t" =~ ^[0-9]*(\<|\>|\>\>|\>\&|\<\<\<)$ ]] && { skipnext=1; continue; }
    [[ "$t" =~ ^[0-9]*[\<\>] ]] && continue
    w+=("$t")
  done
  [ ${#w[@]} -eq 0 ] && continue
  if [ "${w[0]}" = cd ] && [ ${#w[@]} -gt 1 ]; then
    case "${w[1]}" in /*) dir="${w[1]}" ;; *) dir="$dir/${w[1]}" ;; esac
    continue
  fi
  i=0
  while [ $i -lt ${#w[@]} ] && [[ "${w[$i]}" =~ ^[A-Za-z_][A-Za-z0-9_]*= || "${w[$i]}" = sudo ]]; do i=$((i + 1)); done
  [ "${w[$i]:-}" = git ] || continue
  i=$((i + 1))
  gdir="$dir"
  while [[ "${w[$i]:-}" = -* ]]; do
    case "${w[$i]}" in
      -C) case "${w[$((i + 1))]:-}" in /*) gdir="${w[$((i + 1))]}" ;; *) gdir="$gdir/${w[$((i + 1))]:-}" ;; esac; i=$((i + 2)) ;;
      -c) i=$((i + 2)) ;;
      *) i=$((i + 1)) ;;
    esac
  done
  sub="${w[$i]:-}"
  args=("${w[@]:$((i + 1))}")
  case "$sub" in
    add)
      paths=()
      for a in "${args[@]}"; do [[ "$a" = -* ]] || paths+=("$a"); done
      check "${paths[@]}"
      opts=()
      for a in "${args[@]}"; do case "$a" in -p | --patch | -i | --interactive | -e | --edit) ;; *) opts+=("$a") ;; esac; done
      while IFS= read -r f; do check "$f"; done < <(git -C "$gdir" add --dry-run --ignore-missing "${opts[@]}" 2> /dev/null | sed -n "s/^add '\(.*\)'$/\1/p")
      ;;
    commit)
      while IFS= read -r f; do check "$f"; done < <(git -C "$gdir" diff --cached --name-only --diff-filter=ACMR 2> /dev/null)
      all=0; skip=0; paths=()
      for a in "${args[@]}"; do
        if [ $skip = 1 ]; then skip=0; continue; fi
        case "$a" in
          -a | --all) all=1 ;;
          -[!-]*a*) [[ "$a" =~ ^-[A-Za-z]+$ ]] && all=1; [[ "$a" =~ [mFCc]$ ]] && skip=1 ;;
          -m | -F | -C | -c | --author | --date | --fixup | --squash | --trailer) skip=1 ;;
          -*) ;;
          *) paths+=("$a") ;;
        esac
      done
      check "${paths[@]}"
      [ $all = 1 ] && while IFS= read -r f; do check "$f"; done < <(git -C "$gdir" diff --name-only --diff-filter=ACMR 2> /dev/null)
      ;;
  esac
done < <(split_cmd <<< "$cmd")

[ ${#hits[@]} -eq 0 ] && exit 0
mapfile -t hits < <(printf '%s\n' "${hits[@]}" | sort -u)
{
  echo "block-secrets: refusing to stage or commit secret files: ${hits[*]}"
  echo "Secrets (.env, .env.*, *.pem, *.key, id_rsa*) never go into git; only .env.example with placeholder values does."
  echo "If one is already staged: git restore --staged <file>. Do not retry with -f or a different path."
} >&2
root="$(git -C "$dir" rev-parse --show-toplevel 2> /dev/null || true)"
branch="$(git -C "${root:-.}" branch --show-current 2> /dev/null || true)"
if [ -f "$root/LOG.md" ] && [[ "$branch" =~ ^feature-([0-9]+)/ ]]; then
  echo "- $(date +%Y-%m-%dT%H:%M:%S%z) | F${BASH_REMATCH[1]} | hook | block-secrets: ${hits[*]}" >> "$root/LOG.md"
fi
exit 2
