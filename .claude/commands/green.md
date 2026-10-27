---
description: Run the repo's "green" check (lint, test, e2e) in order, stop on the first failure and summarize
argument-hint: "[lint|test|e2e to start from that step]"
allowed-tools: Bash(mise run *), Bash(curl *), Bash(date *), Bash(git branch *), Bash(git rev-parse *)
---

"Green" in this repo (CLAUDE.md) is `mise run lint`, `mise run test` and `mise run e2e`, all
passing. Run them from the repo root, one Bash call per step, in this order, starting at
`$ARGUMENTS` if it names a step:

1. `mise run lint` - golangci-lint (depguard context boundaries), eslint, tsc.
2. `mise run test` - Go unit + integration (testcontainers, needs Docker), Vitest, stories.
   Use a 10-minute timeout.
3. `mise run e2e` - first check the core API: `curl -sf localhost:8080/system/health`. If it is
   down, stop and ask the user to start `mise run dev`; do not start it yourself. Playwright
   starts the web dev server on :5180 itself.

Stop at the first failing step. Do not fix anything in this command; it only reports.

Report, in the user's language:

- one line per step run: `lint | test | e2e -> pass | FAIL | skipped`;
- for the failing step: the failing package / test / file:line and the first relevant error
  lines (at most 15), plus your one-sentence guess at the cause;
- generated-code drift counts as a failure: if the error points into `*.gen.go`, `sqlcgen/` or
  `src/api/gen/`, say "run `mise run gen`" rather than suggesting an edit there.

Metrics log: if `git branch --show-current` is `feature-1/*` or `feature-2/*`, append one line to
`LOG.md` with the timestamp from `date +%Y-%m-%dT%H:%M:%S%z`:
`- <ts> | <F1|F2> | check | mise run lint + test + e2e -> <pass | FAIL at <step>: <short reason>>`.
