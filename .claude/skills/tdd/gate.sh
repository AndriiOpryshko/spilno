#!/usr/bin/env bash
# Deterministic gates of the /tdd pipeline. Usage:
#   gate.sh preflight <story-id>
#   gate.sh red       <story-id> <base-sha>
#   gate.sh green     <story-id> <red-sha>
#   gate.sh refactor  <story-id> <green-sha>
# Prints "GATE <phase>: PASS ..." and exits 0, or "GATE <phase>: FAIL <reason>" and exits 1.
set -uo pipefail
phase=${1:?phase}; id=${2:?story-id}; ref=${3:-}
root=$(git rev-parse --show-toplevel) || exit 1
cd "$root" || exit 1

fail() { echo "GATE $phase: FAIL $*"; exit 1; }
story=$(ls docs/features/*/tasks/"$id"-*.md docs/features/*/tasks/"$id".md 2>/dev/null | head -1)
[ -n "$story" ] || fail "story file docs/features/*/tasks/$id*.md not found"
fm() { sed -n '/^---$/,/^---$/p' "$story" | sed -n "s/^$1: *//p" | head -1 | sed 's/^"\(.*\)"$/\1/'; }
test_cmd=$(fm test_cmd); stub=$(fm stub); test_file=$(fm test_file); pkg=$(fm package); scope=$(fm scope)
[ -n "$test_cmd" ] && [ -n "$stub" ] && [ -n "$test_file" ] && [ -n "$pkg" ] || fail "story front matter needs test_cmd, stub, test_file, package"
subject=$(git log -1 --pretty=%s)
run_tests() { out=$(bash -c "$test_cmd" 2>&1); code=$?; }

case $phase in
preflight)
  [ -z "$(git status --porcelain)" ] || fail "working tree not clean; commit or stash before /tdd $id"
  echo "GATE preflight: PASS story=$story scope=$scope base=$(git rev-parse HEAD) test_cmd=\"$test_cmd\""
  ;;
red)
  [ -n "$ref" ] || fail "need base sha"
  [[ $subject == test\(* ]] || fail "last commit subject does not start with test(: $subject"
  extra=$(git diff --name-only "$ref" HEAD | grep -v -x -e "$test_file" -e "$stub" || true)
  [ -z "$extra" ] || fail "RED commit touched files other than $test_file and $stub: $extra"
  [ -z "$(git status --porcelain)" ] || fail "uncommitted changes after RED"
  run_tests
  [ $code -ne 0 ] || fail "tests pass after RED: they do not test new behaviour"
  grep -qE '\[build failed\]|\[setup failed\]|undefined: ' <<<"$out" && fail "build error, not a failing assertion: $(grep -m1 -E 'undefined: |build failed' <<<"$out")"
  n=$(grep -c -- '^--- FAIL' <<<"$out")
  [ "$n" -ge 1 ] || fail "no '--- FAIL' line in the test output"
  echo "GATE red: PASS failing_tests=$n test_lines=$(wc -l <"$test_file")"
  ;;
green)
  [ -n "$ref" ] || fail "need red sha"
  [[ $subject == feat\(* ]] || fail "last commit subject does not start with feat(: $subject"
  touched=$(git diff --name-only "$ref" HEAD -- '*_test.go')
  [ -z "$touched" ] || fail "implementer changed tests: $touched"
  [ -z "$(git status --porcelain)" ] || fail "uncommitted changes after GREEN"
  run_tests
  [ $code -eq 0 ] || fail "tests still red: $(grep -m1 -- '--- FAIL' <<<"$out")"
  (cd services/core && go vet "./${pkg#services/core/}/") >/dev/null 2>&1 || fail "go vet reports problems"
  echo "GATE green: PASS impl_lines=$(wc -l <"$stub")"
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
  (cd services/core && go vet "./${pkg#services/core/}/") >/dev/null 2>&1 || fail "go vet reports problems"
  echo "GATE refactor: PASS impl_lines=$(wc -l <"$stub")"
  ;;
*) fail "unknown phase" ;;
esac
