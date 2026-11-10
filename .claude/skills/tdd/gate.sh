#!/usr/bin/env bash
# Deterministic gates of the /tdd pipeline. Usage:
#   gate.sh preflight <story-id>
#   gate.sh red       <story-id> <base-sha>
#   gate.sh green     <story-id> <red-sha>
#   gate.sh refactor  <story-id> <green-sha>
# Prints "GATE <phase>: PASS ..." and exits 0, or "GATE <phase>: FAIL <reason>" and exits 1.
#
# Story front matter: test_cmd, package, and the files RED may touch: test_file or a test_files
# list, stub or a stubs list (a story that spans packages lists them; go.mod and go.sum go under
# stubs when the stub needs a new module).
set -uo pipefail
phase=${1:?phase}; id=${2:?story-id}; ref=${3:-}
root=$(git rev-parse --show-toplevel) || exit 1
cd "$root" || exit 1

fail() { echo "GATE $phase: FAIL $*"; exit 1; }
story=$(ls docs/features/*/tasks/"$id"-*.md docs/features/*/tasks/"$id".md 2>/dev/null | head -1)
[ -n "$story" ] || fail "story file docs/features/*/tasks/$id*.md not found"
front() { sed -n '1{/^---$/!q;};1,/^---$/{1!p;}' "$story" | sed '$d'; }
fm() { front | sed -n "s/^$1: *//p" | head -1 | sed 's/^"\(.*\)"$/\1/'; }
# fm_list prints the items of a block list ("key:" followed by "  - item" lines).
fm_list() {
  front | awk -v key="$1:" '
    $0 == key { on = 1; next }
    on && /^ +- / { sub(/^ +- +/, ""); gsub(/^"|"$/, ""); print; next }
    on { exit }'
}
test_cmd=$(fm test_cmd); pkg=$(fm package); scope=$(fm scope)
tests=$({ fm test_file; fm_list test_files; } | sed '/^$/d' | sort -u)
stubs=$({ fm stub; fm_list stubs; } | sed '/^$/d' | sort -u)
[ -n "$test_cmd" ] && [ -n "$pkg" ] && [ -n "$tests" ] && [ -n "$stubs" ] ||
  fail "story front matter needs test_cmd, package, test_file or test_files, stub or stubs"
subject=$(git log -1 --pretty=%s)
run_tests() { out=$(bash -c "$test_cmd" 2>&1); code=$?; }
# go_lines sums the lines of the .go files in a newline-separated list.
go_lines() {
  local n=0 f
  while IFS= read -r f; do
    case $f in *.go) [ -f "$f" ] && n=$((n + $(wc -l <"$f"))) ;; esac
  done <<<"$1"
  echo "$n"
}
vet() { (cd services/core && go vet "./${pkg#services/core/}/...") >/dev/null 2>&1; }

case $phase in
preflight)
  [ -z "$(git status --porcelain)" ] || fail "working tree not clean; commit or stash before /tdd $id"
  echo "GATE preflight: PASS story=$story scope=$scope base=$(git rev-parse HEAD) test_cmd=\"$test_cmd\""
  ;;
red)
  [ -n "$ref" ] || fail "need base sha"
  [[ $subject == test\(* ]] || fail "last commit subject does not start with test(: $subject"
  extra=$(git diff --name-only "$ref" HEAD | grep -v -x -F -f <(printf '%s\n%s\n' "$tests" "$stubs") | tr '\n' ' ')
  [ -z "$extra" ] || fail "RED commit touched files the story does not list as tests or stubs: $extra"
  [ -z "$(git status --porcelain)" ] || fail "uncommitted changes after RED"
  run_tests
  [ $code -ne 0 ] || fail "tests pass after RED: they do not test new behaviour"
  grep -qE '\[build failed\]|\[setup failed\]|undefined: ' <<<"$out" && fail "build error, not a failing assertion: $(grep -m1 -E 'undefined: |build failed' <<<"$out")"
  n=$(grep -c -- '^--- FAIL' <<<"$out")
  [ "$n" -ge 1 ] || fail "no '--- FAIL' line in the test output"
  echo "GATE red: PASS failing_tests=$n test_lines=$(go_lines "$tests")"
  ;;
green)
  [ -n "$ref" ] || fail "need red sha"
  [[ $subject == feat\(* ]] || fail "last commit subject does not start with feat(: $subject"
  touched=$(git diff --name-only "$ref" HEAD -- '*_test.go')
  [ -z "$touched" ] || fail "implementer changed tests: $touched"
  [ -z "$(git status --porcelain)" ] || fail "uncommitted changes after GREEN"
  run_tests
  [ $code -eq 0 ] || fail "tests still red: $(grep -m1 -- '--- FAIL' <<<"$out")"
  vet || fail "go vet reports problems"
  echo "GATE green: PASS impl_lines=$(go_lines "$stubs")"
  ;;
refactor)
  [ -n "$ref" ] || fail "need green sha"
  [ -z "$(git status --porcelain)" ] || fail "uncommitted changes after REFACTOR"
  if [ "$(git rev-parse HEAD)" = "$(git rev-parse "$ref")" ]; then
    run_tests; [ $code -eq 0 ] || fail "tests red"; echo "GATE refactor: PASS no change"; exit 0
  fi
  [[ $subject == refactor\(* ]] || fail "last commit subject does not start with refactor(: $subject"
  touched=$(git diff --name-only "$ref" HEAD -- '*_test.go')
  [ -z "$touched" ] || fail "refactorer changed tests: $touched"
  run_tests
  [ $code -eq 0 ] || fail "refactor broke tests: $(grep -m1 -- '--- FAIL' <<<"$out")"
  [ -z "$(gofmt -l "$pkg")" ] || fail "gofmt: $(gofmt -l "$pkg")"
  vet || fail "go vet reports problems"
  echo "GATE refactor: PASS impl_lines=$(go_lines "$stubs")"
  ;;
*) fail "unknown phase" ;;
esac
