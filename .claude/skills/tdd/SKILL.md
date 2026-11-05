---
name: tdd
description: Run the full red-green-refactor cycle on one spilno story through three isolated agents (tdd-test-writer, tdd-implementer, tdd-refactorer) with deterministic shell gates between phases. Use when the user types `/tdd <story-id>` (e.g. `/tdd T17b`), `/tdd <story-id> --review-tests` (stop after RED for a human review), or `/tdd <story-id> --continue` (resume after that review), or asks to "run the TDD pipeline on <story>".
argument-hint: <story-id> [--review-tests | --continue]
allowed-tools: Bash, Read, Agent
---

# tdd — coordinator of the red-green-refactor cycle

You conduct; you do not play. You never write a test or a line of production code yourself. Each phase is one
`Agent` call to a project agent from `.claude/agents/`, which runs in its own context with its own system prompt:
the test writer never sees the implementer's reasoning, the implementer never sees how the tests were chosen.
Between phases you run `.claude/skills/tdd/gate.sh`, a deterministic script; its verdict is final.

## Arguments

- `<story-id>`: required, e.g. `T17b`; the story is `docs/features/*/tasks/<story-id>-*.md`.
- `--review-tests`: stop after the RED gate so a human reviews the tests before any code is written.
- `--continue`: resume a run stopped by `--review-tests` (skips preflight and RED).

## Rules

- Run every gate exactly as written and paste its `GATE ...` line into your report.
- On `GATE ...: FAIL`, stop at once with that line and the phase's SHA. Do not retry, do not fix, do not run the
  next agent. A pipeline that "almost works" is a failed pipeline.
- Record the time (`date +%H:%M:%S`) before and after each agent call.
- Pass agents only the story id and file path. No summary of earlier phases: isolation is the point.

## Preflight (skip with --continue)

```bash
.claude/skills/tdd/gate.sh preflight <story-id>
```

Note `story=` and `base=` from the PASS line as `STORY` and `BASE`.

## Phase 1 — RED (skip with --continue)

```
Agent(subagent_type="tdd-test-writer", description="RED: failing tests for <story-id>",
      prompt="Story <story-id>: <STORY>. Follow your RED workflow: one failing test per AC, compile-only stub,
              prove the failure is an assertion and not a build error, commit, print `RED phase commit: <SHA>`.")
```

Gate 1:

```bash
.claude/skills/tdd/gate.sh red <story-id> <BASE>
```

With `--review-tests`, stop here and print:

```
RED done for <story-id> (<RED_SHA>). Review the tests before any code exists:
  git show <RED_SHA> -- '*_test.go'
Continue: /tdd <story-id> --continue      Abort: git revert <RED_SHA>
```

## Phase 2 — GREEN

With `--continue`: first check `git log -1 --pretty=%s` starts with `test(` and mentions `<story-id>`, take
`RED_SHA = HEAD`, take `BASE = RED_SHA~1`, and rerun Gate 1 (`gate.sh red <story-id> <BASE>`) before going on.

```
Agent(subagent_type="tdd-implementer", description="GREEN: implement <story-id>",
      prompt="Story <story-id>: <STORY>. The tests committed in <RED_SHA> are a read-only contract: do not
              modify any *_test.go file. Follow your GREEN workflow, commit, print `GREEN phase commit: <SHA>`.")
```

Gate 2:

```bash
.claude/skills/tdd/gate.sh green <story-id> <RED_SHA>
```

## Phase 3 — REFACTOR

```
Agent(subagent_type="tdd-refactorer", description="REFACTOR: tidy <story-id>",
      prompt="Story <story-id>: <STORY>. Tests are green and strictly read-only. Follow your REFACTOR workflow;
              print `REFACTOR phase commit: <SHA>` or `REFACTOR phase: no change - <why>`.")
```

Gate 3:

```bash
.claude/skills/tdd/gate.sh refactor <story-id> <GREEN_SHA>
```

## Final report

```
TDD pipeline <PASS|STOPPED at Gate n> for <story-id>
  RED      <sha> <subject>   (<start>-<end>)
  GREEN    <sha> <subject>   (<start>-<end>)
  REFACTOR <sha|no change>   (<start>-<end>)
Gates: <the four GATE lines>
Lines: tests <n> / implementation <m>
Check: git log --oneline <BASE>..HEAD ; git diff <RED_SHA> HEAD -- '*_test.go'   (must be empty)
```

## Anti-patterns

- Doing a phase inline in this context. That brings back the context pollution the three agents remove.
- Interpreting a gate ("it failed, but only on vet"). The script decides.
- Squashing the commits. RED, GREEN and REFACTOR stay separate so the history shows the tests came first.
