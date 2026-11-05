# Background and schedules · homework 7.5 · PR watch loop + T33 restore drill in the background · 2026-11-05

- Demo `modules/7-execution-scale/7.5-background`: `make matrix` prints the three-level table (exit 0).
  `make verify` RED at start, but for the environment, not the code: `make: pytest: No such file or directory`,
  `Error 127`. Through the demo's own env (`uv run make verify`) it is GREEN: `3 passed in 0.01s`. The Makefile
  says GREEN by design: the demo's artifacts are the recipes and the matrix, not a red build.
- Easy level: `/loop` in a live `cl` session polling CI and review state of PR #31 and #32 through `gh`.
- Hard level: the T33 restore drill (`infra/backup/drill.sh`), which exists only uncommitted in `../spilno-wf-wave`,
  run once as a background session (`/goal` + `/bg`) with least-privilege settings
  ([`drill-session.settings.json`](drill-session.settings.json)).
- Isolation: this report lives in worktree `../spilno-bg`, branch `docs/background-schedule-report` from
  `origin/main` `f786235`; another session works in the main `spilno` folder.
- Metrics come from the transcripts (`~/.claude/projects/<dir>/*.jsonl`), `claude agents --json --all`, git and
  `docker`, not from the summaries.

## Demo: what each level looks like finished

| Level | Finished form (from `recipes/`, `routines/README.md`) | Sees local files | Must be on | Min interval |
|---|---|---|---|---|
| `/loop` | `/loop 5m run make verify; green -> report and stop, else fix and wait` inside an open session; dies with the session | yes | the open session | 1 min |
| desktop scheduled | `/schedule create local "daily 09:00" "..."`; a fresh session each run, survives a reboot | yes | the machine | OS schedule |
| cloud `/schedule` | `/schedule create "daily 03:00" "..."`, `list`, `update`, `run <id>`; pushes only to `claude/...` branches | no, a fresh git clone | nothing of mine | 1 h |
| `/bg` (background session) | `/goal <binary condition>` then `/background`; audited in `claude agents` | yes | the machine | n/a |

Where the safeguards sit:

- **7-day lifetime** and **50 tasks per session** apply to scheduled tasks (`/loop` crons and routines): a task
  fades after a week; a session cannot hold more than 50.
- **Jitter**: a small random offset on the start time so routines do not all fire on the same minute.
- **`CLAUDE_CODE_DISABLE_CRON`**: environment kill switch for every cron task.
- **Supervisor**: unpinned idle background sessions are reaped after about an hour; `Ctrl+T` in `claude agents`
  pins.
- **`/goal`** gives a background session its own binary stop, `or stop after N turns` is the soft cap.
- The hard spend ceiling is not in the tool: a spend alert in the Anthropic Console.

The trap from `routines/README.md`: a routine summarising the git-ignored `drafts/` returns an empty digest,
because the cloud sees only a fresh clone.

## Easy level

### Choice

**Level: `/loop` in a live session.** One sentence: I want PR state while I am at the keyboard, at an interval
tighter than the cloud minimum of one hour, and I need neither a reboot-proof run nor a run with the machine off.

**The trap:** the task needs no local files: `gh pr view` / `gh pr checks` read GitHub, so a cloud routine could
technically run it. It is not chosen for the interval (1 h minimum) and because its result would land in a
`claude/...` branch or a routine log instead of the session I am working in.

### Configuration

Launched by hand in `cl` in `../spilno-bg` (read-only work, so the report worktree is safe to share):

```
/loop 10m Перевір стан сам, не спираючись на попередні цикли чи свою пам'ять: виконай date -Is, потім для PR 31 і 32 команду gh pr view <n> --json number,state,isDraft,reviewDecision,mergeable,headRefOid,updatedAt,statusCheckRollup,reviews,comments. Виведи рівно один рядок на PR: <час> | #n | state | draft | sha(7) | core/web/e2e | review | коментарів. Лише читання: нічого не коміть, не пуш, не коментуй у PR. Якщо будь-який check FAILURE, напиши це першим рядком. Якщо обидва PR MERGED або CLOSED, напиши STOP і зупини цикл.
```

The prompt re-reads GitHub each cycle (`date -Is` + `gh pr view`) instead of trusting the last answer, and has a
binary stop (both PRs merged or closed). The skill turned it into `CronCreate` job `b8463640`: "Every 10 minutes …
Session-only (not written to disk, dies when Claude exits). Auto-expires after 7 days."

### Cycles observed

Session `998831f3-d329-473c-a838-fb9fec56564a`, from `scheduled_task_fire` events and the `gh` tool results:

| Fire (local) | `gh` ran | #31 | #32 |
|---|---|---|---|
| 19:17:06 (immediate run) | 19:17:06 | OPEN, draft, `e32c5fe`, SUCCESS ×3, no review, 0 comments | OPEN, ready, `c36ba81`, SUCCESS ×3 |
| 19:23:36 | 19:23:43 | same | **MERGED** (merged 19:22:32 by the user) |
| 19:33:36 | 19:33:40 | same | MERGED |
| — | — | *no fire 19:43 … 20:53* | |
| 21:03:13 | **21:05:11** | same | MERGED |

- The first scheduled fire came at 19:23, not 19:27: the cron runs on a 10-minute grid with an offset (jitter
  plus grid alignment), then 19:33, 21:03, all on the same grid.
- The gap: after an `away_summary` at 19:36 no transcript on this machine has a single entry until 21:03, then
  the tool call lags the fire by two minutes. The machine (the Mac running OrbStack) was most likely asleep:
  the `/loop` lives only while its session can run, and missed fires were **not replayed** (one fire at 21:03,
  not nine).
- The loop caught the #32 merge on the next cycle; the stop condition stays false while #31 is an open draft.
- Cost of 4 cycles + the setup turn: 7 assistant turns, output 1 607 tokens, cache read 415 055.

### Safeguards that would fire here

| Safeguard | Here |
|---|---|
| 7-day lifetime | **would fire**: #31 is a draft with no reviewer, so the binary stop may not come; the cron says "Auto-expires after 7 days" |
| 50 tasks per session | would not: one job; it matters only if a loop prompt starts scheduling more loops |
| Jitter | fired: first fire at 19:23, off the launch minute |
| `CLAUDE_CODE_DISABLE_CRON` | the kill switch if many loops are left behind; not needed for one |
| Session-only | fired in effect: the sleep window lost nine cycles; closing `cl` ends the job |

## Hard level

### Process

The T33 restore drill: `infra/backup/drill.sh` starts a throwaway source Postgres (`spilno-drill-src`,
`127.0.0.1:55432`), migrates it with goose, seeds a few rows, runs `backup.sh` (pg_dump → age → local
"off-site"), `restore.sh` into a fresh `spilno-drill-dst`, and compares exact per-table row counts. It exists only
uncommitted in `../spilno-wf-wave` (T33 of the 7.4 wave), and it needs the local Docker daemon.

**Level: a background session (`/goal` + `/bg`), run once.** The cloud is out twice: the routine's fresh clone
has no `infra/backup/`, and it has no Docker next to the dev stack. A schedule is out on purpose: the nightly
rhythm already lives inside the backup container (`schedule.sh`, `BACKUP_AT`), so a Claude schedule on top would
only duplicate it.

Correction of the brief: the drill does **not** touch the dev database. It never connects to
`spilno-postgres-1`; its own containers and network are prefixed `spilno-drill-` and removed by `trap cleanup EXIT`.
The real risk was the uncommitted code plus a background agent with Docker on the host.

### Configuration

```
cd /Users/andriio/course/learning/spilno-wf-wave
cl -n "T33 drill" --permission-mode dontAsk --settings <this folder>/drill-session.settings.json
  /goal infra/backup/drill.sh запущено рівно один раз з кореня цього worktree і він надрукував "RESULT: PASS" або
        "RESULT: FAIL"; після цього створено файл /home/andriio/.local/state/spilno-drill/drill-<UTC-час>.md з дослівним
        рядком RESULT, таблицею row counts, виводом docker ps -a --filter name=spilno-drill (має бути порожнім) і
        git status --porcelain; у репозиторії нічого не змінено, spilno-postgres-1 не зачеплено; при FAIL не виправляй
        і не перезапускай, лише зафіксуй; or stop after 8 turns
  /bg
```

[`drill-session.settings.json`](drill-session.settings.json): allow only `infra/backup/drill.sh`, `docker ps|images`,
`git status|diff`, `date`, reads in the worktree, writes in `~/.local/state/spilno-drill/`; deny `Edit`, every git
write, `docker rm|stop|kill|exec|compose|volume|system`, `psql`, `rm`, `gh`, `curl`, web; `CLAUDE_CODE_DISABLE_CRON=1`.
`dontAsk` turns everything outside the allow list into a refusal instead of a "Needs input" prompt.

### Safety net under the process

| Layer | What it holds | Fired? |
|---|---|---|
| Worktree isolation | the drill reads `../spilno-wf-wave`, nothing lands in the main `spilno` folder | held |
| Least privilege (`dontAsk` + settings) | no repo edits, no git writes, no Docker against the dev stack | **fired 3×**: a compound `mkdir … && git status > …` and `git -C … status` (pattern is `git status:*`) refused; the report `Write` refused too (my bug, below) |
| The script's own net | `trap cleanup EXIT`, `spilno-drill-` prefix, port bound to 127.0.0.1, `mktemp` work dir | held: no drill containers or network left |
| `/goal` binary condition | "exactly once", "on FAIL do not fix" | held: one drill run, no retry |
| `or stop after 8 turns` (soft) | stop a stuck goal | **did not stop it**: the evaluator kept answering "unmet" past turn 8 |
| Harness stop-hook cap (hard) | a Stop hook may block at most 9 times in a row | **fired**: "A hook blocked the turn from ending 9 consecutive times — overriding and ending turn", then "Goal paused" |
| `CLAUDE_CODE_DISABLE_CRON` | the session cannot schedule anything | not tested (no cron attempted) |
| Supervisor (~1 h idle reap, unpinned) | forgotten background session | not needed; session not pinned |
| `claude agents` audit + revision vs baseline | catches what the summary hides | below |

### Run

Session `5c6c3eb1-c3b6-4e79-9434-eef0fc5e3b8a` (`claude-opus-5-5`), 17:18:00Z – 17:23:36Z:

- 17:18:20 `infra/backup/drill.sh` (once) → 17:18:34 `RESULT: PASS - 38 tables, 48 rows match; backup 0s, restore 1s,
  total 14s`. Image `spilno-backup:drill` built, goose to version 27, encrypted dump 111 817 bytes, header
  `age-encryption.org/v1`.
- 17:18:51 `Write ~/.local/state/spilno-drill/…` → **denied**. Cause, mine: a bare `"Edit"` in `deny` covers every
  file-editing tool including `Write`, and deny beats allow, so the `Write(//home/…/spilno-drill/**)` allow could
  never match. The agent did not work around it through the shell; it wrote the report into its scratchpad.
- 17:19:14 – 17:21:07: **17** Stop-hook blocks ("goal unmet"); the agent repeated "cannot write, will not bypass".
  Turn 8 passed without a stop; the harness cap ended it after 9 consecutive blocks → "Goal paused".
- 17:23:21 I copied the scratchpad report into `~/.local/state/spilno-drill/` by hand (`!` bash); the evaluator then
  read the file and the goal closed.
- Cost: 28 assistant turns, output **13 324** tokens, cache read **2 042 946**, cache write 90 030, input 92: the
  14-second drill was a small fraction; almost all of it is the block loop.

### Revision against reality

Compared with a baseline taken at 19:12:13+02:00, before launch:

| Check | Baseline | After | Verdict |
|---|---|---|---|
| dev DB writes `sum(n_tup_ins/upd/del)` | 64/0/27 | 64/0/27 | untouched |
| `spilno-postgres-1` StartedAt | 2026-10-27T08:33:34Z | same | not restarted |
| `git status --porcelain` in `spilno-wf-wave` (13 lines) | | identical | no repo change |
| sha256 of the 31 modified/untracked files there | | identical | no file change |
| `docker ps -a`, `docker network ls` for `spilno-drill` | none | none | cleaned up |
| `docker images spilno-backup` | none | `spilno-backup:drill` | **residue**: the drill does not remove its image |
| `~/.local/state/spilno-drill/drill-20261105T171839Z.md` | | RESULT line, 38-row table, empty `docker ps` | matches the raw output |

Sessions in `claude agents --json --all` (baseline and after):

| Session | State | Verdict |
|---|---|---|
| `5c6c3eb1` T33 drill | idle after the goal closed | clear with `Ctrl+X`; do not pin |
| `998831f3` loop (`spilno-bg`) | idle between cycles | lives with the loop; stop the cron after the homework |
| `09d732b4` "Контракт openapi.yaml" (`spilno-wf-audit`) | **blocked**, from 7.4 | forgotten: was already blocked before this homework |
| `6a5cf57e` (`spilno`) | waiting / idle | forgotten candidate in the main folder |
| `d8011a17` this session, title "Домашка 7.3 /goal" | working | `/clear` kept the old title |

## Essay

For the drill I chose a one-shot background session (`/goal` + `/bg`) over any schedule, because the code exists
only in an uncommitted worktree and needs the local Docker daemon: the cloud would wake up to a fresh clone with
no `infra/backup/` at all, and the nightly rhythm already belongs to `schedule.sh` inside the backup container.
The most important layer of the net turned out to be least privilege with `dontAsk`: the agent got exactly one
script and no way to touch the repo or the dev database, and the revision confirms it (dev DB writes 64/0/27
before and after, 31 files byte-identical). The same layer also hurt: a bare `Edit` in `deny` silently covered
`Write`, so the agent could not finish its own goal, and the soft `or stop after 8 turns` did not stop it; only the
harness cap of nine consecutive Stop-hook blocks ended a loop that burned 2 M cached tokens around a 14-second
script. The cloud trap was not where the brief put it: the drill never touches the dev database, but its whole
code and its Docker are local, so the cloud is out before the first command. Where it did not pay off: the drill
itself takes 14 s and is deterministic, so a plain `infra/backup/drill.sh` in a terminal gives the same evidence
for a fraction of the cost; the agent earns its keep only if it judges a FAIL. The PR loop did not pay off either:
GitHub notifications already tell me about a merge, and the loop lost nine cycles while the laptop slept.

## Takeaways

- Check permission rules against the tool they really match: `Edit` in `deny` is every file write; test a
  restrictive settings file with one dry run before going AFK.
- `or stop after N turns` is soft; the hard stop is the harness Stop-hook cap. A goal that can become unreachable
  (a denied write) needs a condition the agent can always settle, e.g. "or report why it is impossible".
- `/loop` is honest about its level: no session (or a sleeping machine), no cycles, no catch-up.
- Revise against a baseline taken **before** launch: without the 19:12 snapshot "the dev DB is untouched" would
  be the agent's word.

## Open items

- `spilno-backup:drill` image left by the drill (the script cleans containers and network, not the image).
- Forgotten sessions: `09d732b4` (blocked since 7.4, `spilno-wf-audit`) and `6a5cf57e` (main folder); stop or
  `Ctrl+X` after a look.
- The PR loop `b8463640` still runs; stop it (`CronDelete` or close the session) once the homework is handed in.
- Fix `drill-session.settings.json`: drop the bare `Edit` deny, deny `Edit(//Users/andriio/course/learning/**)`
  instead, so the state-dir write works.
