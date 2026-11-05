# Ralph run · T18 (poll rounds and the own-vote rule) · 2026-11-05

- Harness: `scripts/ralph.sh` (cold `claude -p` per iteration, MAX_ITER, MAX_USD/ITER_USD,
  ITER_TIMEOUT, INT/TERM trap, DONE re-verified by `check.sh`, tool allowlist in acceptEdits).
- Prompt: `PROMPT.md`; binary DoD: `check.sh` (pinned sha256 of `poll_test.go`, governance tests
  and lint, tracker `done`, one `feat(governance):` commit with only the T18 files, clean tree).
- Result: commit `15e7815 feat(governance): model poll rounds and the own-vote rule`,
  `check.sh` PASS; full `golangci-lint ./...` 0 issues and `go test ./...` green afterwards.

## Numbers

| | iterations | cost (claude `total_cost_usd`) | wall clock |
|---|---|---|---|
| Run 1 | 1 → BLOCKED | $0.71 | 1 m 50 s |
| Run 2 | 2 → DONE verified | $0.95 | 1 m 21 s |
| **Total** | **3 of 5** | **$1.66** | ~3 min of loop |

ccusage did not see the loop: `--no-session-persistence` writes no transcript, so the
before/after delta ($4.57 → $5.84) is only the supervising interactive session.

## Manual interventions

1. Run 1, iteration 1: the petforge `hooks/guard.sh` PreToolUse hook took the dotted string
   literal `"governance.vote_on_behalf"` in a `domain/` file for a non-stdlib import and blocked
   the edit of `errors.go`. The model could not get round it without breaking the rules, left
   `errors.go` referencing an undefined constant, and correctly wrote `BLOCKED`. Fixed by hand in
   petforge: the guard now reads only import lines. Loop restarted with MAX_ITER=4.
2. Run 2, iteration 1: the model invoked `bash …/check.sh`, the allowlist had only the bare path,
   the headless call was denied, so it committed but did not touch DONE. Iteration 2 only ran the
   check and touched DONE. Fixed in the harness (`Bash(bash $CHECK)` allowed).

## Smoke-test finding in the harness itself

Ctrl-C did not stop the course `ralph.sh` pattern once the child runs under `timeout`: the child
exits 130 instead of dying from SIGINT, so bash never runs the INT trap and starts the next
iteration. The harness treats exit 130 as the interrupt.

## Next time

- Dry-run every allowed command once (`claude -p "run check.sh"`) before the real loop.
- Let the prompt name the exact command form (`docs/…/check.sh`, not `bash …`).
- Run project hooks against the planned diff before the loop; a hook is part of the environment.
- The task was small with the acceptance test pinned; next time give Ralph a story where it
  writes the tests as well (T16), with a mutation or coverage gate instead of a pinned file.
