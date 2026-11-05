#!/usr/bin/env bash
# Binary completion check for T18 (poll rounds and the own-vote rule).
# Exit 0 and print PASS only when every condition holds; otherwise print the
# first FAIL and exit 1. Run from anywhere inside the repo.
set -uo pipefail
cd "$(git rev-parse --show-toplevel)"

BASE="${RALPH_BASE:-origin/main}"
DOMAIN=services/core/internal/governance/domain
TEST_SHA=ff985de8eb9a6fab60477ce0eff5cc97ddb8a2cd9bcd9d8f32fc271e383455b2
ALLOWED='^(services/core/internal/governance/domain/(poll|poll_test|errors)\.go|docs/features/board-protocol/tasks/tracker\.md)$'

fail() { echo "FAIL: $*"; exit 1; }

[ -f "$DOMAIN/poll.go" ] || fail "$DOMAIN/poll.go does not exist"
echo "$TEST_SHA  $DOMAIN/poll_test.go" | sha256sum -c --quiet - >/dev/null 2>&1 \
  || fail "$DOMAIN/poll_test.go differs from the pinned acceptance test"

go -C services/core test ./internal/governance/... >/tmp/ralph-t18-test.log 2>&1 \
  || { tail -30 /tmp/ralph-t18-test.log; fail "go test ./internal/governance/... is red"; }
go -C services/core tool -modfile=tools/go.mod golangci-lint run ./internal/governance/... >/tmp/ralph-t18-lint.log 2>&1 \
  || { tail -30 /tmp/ralph-t18-lint.log; fail "golangci-lint ./internal/governance/... reports issues"; }

grep -qE '^\| \[T18\]\(.*\| done \|$' docs/features/board-protocol/tasks/tracker.md \
  || fail "T18 is not marked done in tracker.md"

git log --format=%s "$BASE..HEAD" | grep -qE '^feat\(governance\): ' \
  || fail "no feat(governance): commit on the branch since $BASE"
committed=$(git diff --name-only "$BASE..HEAD")
for f in "$DOMAIN/poll.go" "$DOMAIN/poll_test.go" docs/features/board-protocol/tasks/tracker.md; do
  grep -qx "$f" <<<"$committed" || fail "$f is not committed"
done
stray=$(grep -vE "$ALLOWED" <<<"$committed" || true)
[ -z "$stray" ] || fail "commit touches files outside T18: $stray"
git diff --quiet HEAD -- services/core docs/features/board-protocol/tasks \
  || fail "uncommitted changes left in services/core or tasks/"

echo PASS
