# Dynamic workflows · homework 7.4 · parallel wave T17/T33/T43 + OpenAPI contract audit · 2026-11-05

- Demo `modules/7-execution-scale/7.4-dynamic-workflows`: `make verify` RED at start (`15 failed in 0.03s`,
  `search.py` / `dedupe.py` / `export.py` are `NotImplementedError` stubs); `make check-workflow` →
  `OK: both .mjs parse as valid JS`.
- Easy level: three tracker tasks fanned out by a coordinator in one turn, worktree `../spilno-wf-wave`,
  branch `feat/board-protocol-parallel-wave` from `f786235` (origin/main). Not committed yet.
- Hard level: a read-only audit of all 41 operations of `services/core/api/openapi.yaml`, worktree
  `../spilno-wf-audit`, branch `chore/openapi-contract-audit`. Output `audit.md` next to this file; the script
  is saved as `.claude/workflows/openapi-contract-audit.mjs` (`b021212`).
- Isolation: a fresh `cl` session per level, each in its own worktree from a fresh `origin/main`, because
  another session works in the main `spilno` folder.
- Metrics come from the session transcripts (`~/.claude/projects/<dir>/*.jsonl`, assistant `usage`, the
  workflow `journal.jsonl`), not from the panel. "Tokens" = input + cache write + cache read + output, with
  output listed on its own because it is what the cache cannot discount.

## Demo: where the primitives are

| Script | Line | Primitive |
|---|---|---|
| `ship-snippets.mjs` | 28 | `parallel(` — Implement phase, a barrier over the three stories |
| | 29 | `STORIES.map((s) => () =>` — one thunk per story (the iteration; there is no `while`) |
| | 30 | `agent(` — implementer, schema `{id, file, green}` |
| | 45 | `pipeline(built, …)` — Verify phase, no barrier: each story is reviewed as soon as it is built |
| | 48–50 | `parallel(` + `[1, 2].map` + `agent(` — two independent reviewers per story |
| | 57–58 | convergence: `votes === 2`, kept only if both reviewers agree |
| `audit-independence.mjs` | 20–21 | `parallel(` + `STORIES.map` over stories |
| | 23–25 | nested `parallel(` of N=2 auditors, each an `agent(` |
| | 32, 39 | convergence `votes.every(v => v.overlap === false)`; any overlap → `throw` "re-split before parallel run" |

Convergence is held by code, not by the coordinator's judgement: a story or a "no overlap" verdict survives only
when independent agents agree.

## Easy level

### Independence proven up front

| Task | Write set | Read only | Check |
|---|---|---|---|
| T17 domain | new files only: `services/core/internal/governance/domain/{text,factcheck,proposal}{,_test}.go` | existing `domain/*.go` | `go test ./internal/governance/domain/...` + golangci-lint `./internal/governance/...` |
| T33 infra | `infra/**`, `docs/runbooks/backup-restore.md` | `services/core/migrations/` | shellcheck + local drill (dump → age → restore → row counts) |
| T43 ui | `apps/web/src/routes/registry.*.tsx`, `apps/web/src/features/governance/registry/**`, `routeTree.gen.ts` (router plugin only), `i18n/uk.ts` | `src/api/gen/**`, existing governance components | eslint, tsc, vitest unit + stories |

Pairwise intersections: ∅ / ∅ / ∅. Forbidden for all three (hidden shared writes): `go.mod`/`go.sum`,
`package.json`, `pnpm-lock.yaml`, `mise.toml`, `openapi.yaml`, `tracker.md`. T16 is merged (`f786235`, PR #30) though
`tracker.md` still says `todo`, so T17's dependency holds. `age` was installed before the run; `pg_dump` ran from a
container.

### Evidence of real parallelism

Session `cc23f2d2-ccd5-406e-aa5d-e8eab5eca078` (cwd `spilno-wf-wave`): one assistant message
`msg_011Cgk6AELmkbnzmSgfhKuqA` carries three `Agent` tool_use blocks (jsonl lines 54–56: T17, T33, T43). All three
were launched at 16:23:28Z and finished at T33 16:29:54, T17 16:30:27, T43 16:30:49: about 8 min wall clock for
work estimated at M + M + L. The coordinator made **0** Edit/Write calls; after the fan-out it only ran
`git status` and one `mise run lint` + `mise run test`.

### Result

| Task | Files | Executor check | Exit |
|---|---|---|---|
| T17 | 6 new in `services/core/internal/governance/domain/` | go test + golangci-lint `0 issues.` | 0 |
| T33 | 10 in `infra/backup/` + 1 in `docs/runbooks/` | shellcheck clean; drill `PASS - 38 tables, 48 rows match` | 0 |
| T43 | 14 in `apps/web/src/` | lint, tsc, unit 56/56, stories 55/55 | 0 |

31 files, every one in exactly one write set, nothing outside. `mise run lint` 0, `mise run test` 0. An independent
re-run (`go test -count=1` on the domain package, shellcheck on `infra/backup/*.sh`) was green too.

Tokens: coordinator 0.62 M (out 6.1 k); executors T43 4.99 M (out 45.5 k, 47 turns), T17 1.50 M (out 47.4 k),
T33 1.14 M (out 25.3 k). Total ≈ 8.2 M, out ≈ 124 k; almost all of it is cache reads.

Left for follow-ups, all caused by the narrow write sets (the price of independence): T17 keeps a `votes`
constant in `factcheck.go` instead of `slot.go` and `MergeProposal` does not store fragments (needs
`agenda_item.go`); T33's drill ran against a throwaway database, not the dev one the DoD names; T43 has no menu
link (`app-shell.tsx`), the year in «№ N/YYYY» is added client side and «(3 з 4)» is shown as «N присутні».

## Hard level

### Why this audit

A plain "handler vs contract" audit says little today: 34 of 41 handlers are `ErrNotImplemented` stubs. So each
operation is audited on three axes: the contract against `.claude/rules/openapi.md` "spilno specifics" (R1 tag,
R2 security, R3 401/403 codes, R4 422 + problem+json, R6 pagination / `protocol_revision`), every `code` against
the Go registry (R5), and the handler status (R7). Drift caught now is cheap, before the stubs become code.

### The generated script (seen before approval)

Script `openapi-contract-audit-wf_27380f86-0c6.js`, now `.claude/workflows/openapi-contract-audit.mjs`:

| Line (saved file) | Construct |
|---|---|
| 13–16 | `CAP` + `spawn = (p, o) => { used++; return agent(p, o) }`, the wrapper around every `agent()` |
| 63 | `agent()` (List): one agent collects the operationIds, schema `LIST_SCHEMA` |
| 75 | `pipeline(ops, audit, verify)`, no barrier: an operation goes to Verify as soon as its audit is back |
| 78 | plain `while (attempt < 3)`: re-run a broken or empty auditor, at most 2 retries, then `failed` |
| 80 | `agent()` (Audit), `model: 'sonnet'`, schema `AUDIT_SCHEMA` |
| 97–99 | `parallel(audit.findings.map(…))` + `agent()` (Verify): one independent verifier per finding |
| 109–110 | convergence: only `verdict.confirmed` is kept; the rest are counted in `filtered` |
| 120 | `agent()` (Report): writes `audit.md` only |

The saved file differs from the run in two lines only: a `whenToUse`, and `CAP = (args && args.cap) || 140`
instead of `90` (see the cap note below). The body ends with a top-level `return` that the runtime wraps, so the
check is `node --check` on the body wrapped in an async function (done, parses).

### Run

Session `bef2cf70-c7f9-4254-8010-a094b39eb289`, run `wf_27380f86-0c6`: launched 16:37:38Z, completed 16:54:56Z
(~17 min). 90 agents: List 1, Audit 41 (0 retries, 0 failed), Verify 47, Report 1. Handlers: 7 implemented,
34 stub, 0 missing.

| Measure | Value |
|---|---|
| Auditor findings | 92 |
| Verified | 47 → 46 confirmed, **1 filtered** |
| Unverified (agent cap) | 45 findings in 17 operations, listed in `audit.md` as `fail?(unverified)` |
| Confirmed by rule | R5 27, R7 12, R3 7 |

Main confirmed findings (details in `audit.md`):
1. `identity.session_required` and `identity.not_current_member` exist only in the spec (`openapi.yaml:1265`,
   `:1272`) and the generated TS: there is no Go constant, no session middleware on the governance group
   (`governance/module.go:36-43`), and the validator uses `NoopAuthenticationFunc` (`httpapi/problem.go:58`).
   No Go code can return the advertised 401/403 yet.
2. Six older operations (finance contracts, governance meetings) do not declare 401/403 although the global
   `SessionCookie` protects them; `signOut` declares 401 but not 403.
3. Stubs answer 501 (`httpapi/problem.go:131-134`), which no operation declares.

**The filtered claim.** The `createBoardMember` auditor wrote (R7): "stub … answered as 501 … none of
201/401/403/422 is produced" (`stubs.go:17-19`). The independent verifier returned `confirmed: false`: the
governance group runs `httpapi.RequestValidator` (`module.go:36`), which answers schema violations of
`BoardMemberInput` with a code-less 422 before the stub is reached (`problem.go:63-81`). Only the domain 422
`governance.roster_entry_invalid` is never produced. A plausible half-truth was dropped instead of landing in the
report. (`audit.md` says the filtered text is not kept; it is, in `journal.jsonl`.)

**Cap note.** The prompt said "up to ~90 agents"; the script turned that into a hard `CAP = 90` and logged the
overflow instead of dropping it silently. 41 auditors took nearly half the budget, so 45 of 92 findings were never
verified. That is a prompt-sizing mistake, not a runtime one. The saved command defaults to 140 (1 + 41 + 92 + 1);
a resume with `resumeFromRunId` would take the 90 finished agents from cache.

### Cost: workflow vs one agent

| | Tokens (all) | Output | Wall clock |
|---|---|---|---|
| Workflow sub-agents | 21.17 M (cache read 18.15 M, cache write 2.87 M) | 155 k | ~17 min |
| Main session | 0.41 M | 6 k | |
| **Workflow total** | **≈ 21.6 M** | **≈ 161 k** | |
| One agent, estimate | ≈ 3–5 M | ≈ 40–60 k | ~20–30 min |

The single-agent estimate is grounded in the input size: the contract (2 427 lines), the rules, the registries,
`httpapi` and all handlers are about 155 KB ≈ 45 k tokens. One agent reads them once, then spends ~40–60 turns
grepping per operation with its context growing to ~120 k: ≈ 3–5 M tokens processed, mostly cache reads, plus
~50 k output for a 41-row table and the findings. By phase the workflow spent Audit 12.1 M (sonnet), Verify
8.3 M, Report 0.56 M, List 0.14 M. So the workflow cost about **4–7× the tokens and ~3× the output** of one agent.

### Did breadth pay off here?

Partly. Coverage was guaranteed: 41/41 operations each got a clean context, and verification filtered one
plausible-but-wrong claim. But the findings are not 41 independent problems. 50 of the 92 are R5 repeats of one
root cause (identity codes missing in Go), and the rest are a handful of cross-cutting facts: no session
middleware, 501 stubs, six legacy operations without 401/403. One agent with `grep` would find these in a fraction
of the tokens. The per-finding verifiers spent most of their budget re-confirming the same identity-code fact,
and that is what starved the cap. A better shape for this contract is per rule rather than per operation (7
sweepers), dedupe by root cause behind a barrier, then verify the distinct claims. Per-operation breadth pays off
once the 34 stubs are real handlers and each operation has its own logic to check.

## Takeaways

- Parallelism is proven by the transcript, not by the coordinator's words: three `Agent` blocks in one
  `message.id`.
- The hard part of the easy level was proving independence up front, especially the hidden shared files
  (`go.mod`, lockfile, `mise.toml`, `routeTree.gen.ts`, `tracker.md`). The coordinator never tried to write code.
  Narrow write sets have a cost: every executor left a follow-up that needed a file outside its set.
- In a workflow, "how many agents" is a real budget decision. Size the cap from items × fan-out, and dedupe
  before verifying.
- Convergence earns its keep on half-true claims, which a single pass would have reported as fact.

## Open items

- Easy-level changes in `../spilno-wf-wave` are not committed (three atomic `feat`/`docs`/`feat` commits on request).
- The saved command, `audit.md` and this report are in draft PR #32 (`chore/openapi-contract-audit`).
- From the homework, for the user to attach: a `/workflows` screenshot with agent and token counters. Not seen in
  the transcript: an unrelated prompt sent while the run was in the background. The session transcript shows no
  user message between 16:38 and 16:54.
