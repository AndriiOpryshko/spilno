# Execution · module 7 homework · one task, tests first, Ralph vs /goal · 2026-11-06

- Task: **T13a**, a slice of T13 that needs no Google account: the `Files` port, the in-memory fake and the
  shared contract suite the Drive adapter (T13b, after T12) will also run
  ([story](../tasks/T13a-files-port-fake.md)). Half a day, layer infra, binary "done".
- Method 1 **Ralph** (7.2): `scripts/ralph.sh`, a cold `claude -p` per iteration, state only on disk
  ([PROMPT.md](T13a/PROMPT.md)). Method 2 **`/goal`** (7.3): one headless `claude -p "/goal …"`, a fresh
  evaluator model after every turn ([goal.sh](T13a/goal.sh)). Same tool allowlist, `acceptEdits`, $15 cap, model
  Opus 5.5. Why these two: the task is autonomous with a yes/no exit, which is the case both are made for; the
  question is the hand-rolled loop against the built-in one.
- One machine proof for both: [check.sh](T13a/check.sh) (pinned sha256 of the two test files, governance tests and
  lint, the port and fake on stdlib only, tracker `done`, one `feat(governance):` commit with only T13a files, a
  clean tree).
- Isolation: base `feat/governance-files-port` from `origin/main` `9e34674`; each method in its own worktree and
  branch from the red commit `62b6e6d` (`feat/governance-files-port-ralph`, `feat/governance-files-port-goal`).
  The two runs ran at the same time on the same VM.

## Tests first

`62b6e6d test(governance): add failing tests for the Files port and fake (T13a)`, written from the story before
any code: `filestest/contract.go` (12 contract cases: digest/size/type, small reads, round trip, empty body, stat,
two uploads → two ids, read-only, untrash, unknown id → `ErrFileNotFound` ×4, a body failing mid-read, a cancelled
context) and `filesfake/fake_test.go` (contract on the fake, outage → `governance.storage_unavailable` from all 5
methods with nothing changed, wrong reported sha256 → `ErrChecksumMismatch` with nothing stored, the code in
`ErrorRegistry` as 503).

- First run: red on the build (`undefined: ports.Files`, `FAIL … filesfake [build failed]`).
- To prove the assertions bite and not only the imports, a throwaway naive stub (types plus a fake that just keeps
  bytes) was dropped in and removed again: **14 of 15 behaviours failed**; green were only "open streams back the
  bytes" and the registry check, which any plain storage passes.

## Results

| | Ralph | /goal |
|---|---|---|
| Implementation commit | `4d901f5` at 1 m 54 s, iteration 1 | `25fdc00` at 2 m 01 s; first green `go test` at 1 m 34 s |
| Run 1 (broken `check.sh`) | 3 iterations, 5 m 56 s, $1.36, then `BLOCKED` with a precise cause | 35 turns, 6 m 00 s, $1.16; first «not met» at 2 m 38 s, then 13 of them in 3 m 22 s, ended by the stop-hook block cap (9 in a row), not by its own «stop after 15 turns» clause; process exit 0, `subtype: success` |
| Run 2 (fixed `check.sh`) | 1 iteration, 38 s, $0.29, `DONE verified by check.sh` | 5 turns, 40 s, $0.34, goal met; independent `check.sh` PASS |
| Total | **$1.65** | **$1.50** |
| Permission denials | ≥ 4 (`check.sh; echo exit=$?`, a multi-line `echo >> progress.md`; run 1 iteration 1 lost, see below) | 5 over both runs (compound commands: `cd … && git …`, `git add … && git commit`, `…check.sh; echo exit=$?`) |
| Hand edits to code | 0 | 0 |

The time to green is the commit time: in both runs the code was right on the first pass and the tests were green
before the commit. Both `fake.go` files came out almost the same (mutex, `lookup` under the lock, sha256 through
`io.MultiWriter`, `fake-file-N` ids); the `/goal` one wraps `ErrFileNotFound` with the id, so it went to the PR
(`744e937`, cherry-picked onto the base). On it `mise run lint` (0 issues, eslint, tsc) and `mise run test` (Go +
Vitest, 43 + 43) are green; `e2e` was not run, no UI is touched.

## Manual interventions

1. **Both runs, one cause, my bug:** check rule 4 ran `go list` on the whole `ports` package, and `ports.go` has
   imported `github.com/google/uuid` since before T13a. Neither model touched `check.sh` or `ports.go` (both
   forbidden) and both named the exact fix. Fixed by hand: `google/uuid` is allowed, as in the domain depguard
   rule. Then each method ran again from where it stopped.
2. Nothing else: no code, test or commit message was edited by hand.

## Comparison

Ralph gave more control: its exit on a broken criterion is a rule I wrote ("same FAIL three iterations in a row →
`BLOCKED`"), it stopped after three iterations with a file naming the cause and the fix, and every iteration's cost
is in `cost.log`. `/goal` was a little faster to the first green and needed no prompt file, progress file or loop
script, only a condition, but on a criterion that could not be met it kept nudging a model that had nothing more to
do: 3 m 22 s of a 6-minute run spent on repeating «not met», closed by the harness cap rather than by my turn
clause, and the process still exited 0 with `success`, so only the separate `check.sh` showed the run had
failed. Both runs lost turns to permission denials on compound commands (`…; echo exit=$?`), so the allowlist, not the
method, is what to fix there. Manual interventions were the same, one each, from the same bug in my check. For this kind of task (one story, a binary
check, run unattended) I would choose Ralph: the stop rule and the cost log are mine, and a broken criterion ends
the run with an answer instead of an evaluator loop. `/goal` is the better pick when I stay in the session and
the check is already proven, for example on the second run here.

## Next time

- `scripts/ralph.sh` names its logs `iter-N.json` per run, so run 2 overwrote run 1's `iter-1.json` (its
  denials are lost); the files need a run prefix.
- Allow the check through a wrapper that prints its own exit code, or tell both methods to run it bare: the
  `; echo exit=$?` habit was denied in both.

- Run the check against the baseline and against a reference change before the loop: rule 4 would have failed on
  the baseline for the wrong reason, which a dry run against a fake "done" state would have shown.
- In a `/goal` condition, add an explicit "if the check fails for a reason outside the allowed files, say
  IMPOSSIBLE and stop" so the evaluator can close the goal as impossible instead of looping to the cap.
- Read the exit of a headless `/goal` from the independent check, never from `rc` or `subtype`.
