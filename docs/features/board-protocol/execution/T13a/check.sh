#!/usr/bin/env bash
# Binary completion check for T13a (Files port + in-memory fake), shared by
# both execution methods (Ralph loop and /goal). Exit 0 and print PASS only
# when every condition holds; otherwise print the first FAIL and exit 1.
set -uo pipefail
cd "$(git rev-parse --show-toplevel)"

BASE="${T13A_BASE:-62b6e6d}" # the red-baseline commit with the acceptance tests
GOV=services/core/internal/governance
MOD=github.com/AndriiOpryshko/spilno/services/core
PINNED="354ff65da5a700d1e9992b3bf96a0e603de35f84efa503d77f6078bda8a6d269  $GOV/adapters/filestest/contract.go
4e7a705ff20c204ceda59105b8089fc4753c937ff4ff06ff7321c24fcb43dbd1  $GOV/adapters/filesfake/fake_test.go"
ALLOWED="^($GOV/(ports/files|domain/errors|adapters/filesfake/[a-z_]+)\.go|docs/features/board-protocol/tasks/tracker\.md)$"
LOGDIR="${TMPDIR:-/tmp}/t13a-check.$$"
mkdir -p "$LOGDIR"

fail() { echo "FAIL: $*"; exit 1; }

for f in "$GOV/ports/files.go" "$GOV/adapters/filesfake/fake.go"; do
  [ -f "$f" ] || fail "$f does not exist"
done
sha256sum -c --quiet - <<<"$PINNED" >/dev/null 2>&1 \
  || fail "the acceptance tests (filestest/contract.go, filesfake/fake_test.go) differ from the pinned versions"

go -C services/core test -count=1 ./internal/governance/... >"$LOGDIR/test.log" 2>&1 \
  || { tail -30 "$LOGDIR/test.log"; fail "go test ./internal/governance/... is red"; }
go -C services/core tool -modfile=tools/go.mod golangci-lint run ./internal/governance/... >"$LOGDIR/lint.log" 2>&1 \
  || { tail -30 "$LOGDIR/lint.log"; fail "golangci-lint ./internal/governance/... reports issues"; }

# The fake and the port stay on the standard library, this module and
# google/uuid (ports.go has imported it since before T13a; domain allows it too).
for pkg in ports adapters/filesfake; do
  stray=$(go -C services/core list -f '{{join .Imports "\n"}}' "./internal/governance/$pkg" \
    | grep '\.' | grep -v -e "^$MOD/" -e '^github.com/google/uuid$' || true)
  [ -z "$stray" ] || fail "governance/$pkg imports outside stdlib, the module and uuid: $stray"
done

grep -qE '^\| \[T13a\]\(.*\| done \|$' docs/features/board-protocol/tasks/tracker.md \
  || fail "T13a is not marked done in tracker.md"

git log --format=%s "$BASE..HEAD" | grep -qE '^feat\(governance\): ' \
  || fail "no feat(governance): commit on the branch since $BASE"
committed=$(git diff --name-only "$BASE..HEAD")
for f in "$GOV/ports/files.go" "$GOV/adapters/filesfake/fake.go" docs/features/board-protocol/tasks/tracker.md; do
  grep -qx "$f" <<<"$committed" || fail "$f is not committed"
done
stray=$(grep -vE "$ALLOWED" <<<"$committed" || true)
[ -z "$stray" ] || fail "the commits touch files outside T13a: $stray"
git diff --quiet HEAD -- services/core docs/features/board-protocol/tasks \
  || fail "uncommitted changes left in services/core or tasks/"
[ -z "$(git ls-files --others --exclude-standard services/core)" ] \
  || fail "untracked files left in services/core"

rm -rf "$LOGDIR"
echo PASS
