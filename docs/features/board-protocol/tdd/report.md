# TDD discipline · homework 7.7 · tests first by hand, then three isolated agents vs one session · 2026-11-06

- Demo `modules/7-execution-scale/7.7-tdd-discipline`: `make test` first failed for the environment
  (`make: pytest: No such file or directory`, `Error 127`); through the demo's own env (`uv run make test`) it is RED
  as designed: `11 failed in 0.58s`, every failure `NotImplementedError: Implement per tasks/story-24-sm2.md`.
  `make demo` prints the `/tdd story-24-sm2` steps (exit 0). The three agents and the `tdd` skill were read as the
  template.
- Stories: T17 and T19 split into three TDD-sized stories with an interface, rules and 5-6 AC each
  (`tasks/T19a-meeting-facts.md`, `T17b-amount-recommendation.md`, `T19b-snapshot-fingerprint.md`).
- Isolation: base `chore/claude-tdd-agents` from `origin/main` `ba0462c`; every run in its own worktree and branch
  from `9954fc9`. Metrics from the transcripts (`~/.claude/projects/-Users-andriio-course-learning-spilno-tdd-*/`,
  subagents included) and git; times below are wall clock from the prompt to the last commit.

## Easy level · T19a by hand, then "do not change the tests"

Branch `feat/governance-meeting-facts`:

- `f46325e test(governance): add failing tests per AC for T19a meeting facts`: 6 tests, one per AC, written
  from the story before any code; the stub only declares the interface and returns "not implemented".
  `go test`: 6 × `--- FAIL` on assertions, no build error; `gate.sh red` → `GATE red: PASS failing_tests=6`.
- `ea40f4c feat(governance): implement T19a meeting facts to make tests pass`, by a Claude subagent whose
  instruction said «НЕ ЗМІНЮЙ ТЕСТИ». **1 round to green** (2 `go test` runs, both green), 0 writes to
  `*_test.go` (transcript `agent-acdccca4944e3cc65`), `gate.sh green` → PASS. 55 s.
- No refactor commit: one validator and one transition, nothing to extract.

The hard part was not the agent but writing the tests before the code: AC-1 ("complete facts → nil") passes on
any stub that returns nil, and a Go `panic` stub aborts the whole package run. The stub has to return an error, a
rule that went into the test-writer agent (`9954fc9`).

## Hard level

### Agents and coordinator

- [`.claude/agents/tdd-test-writer.md`](../../../../.claude/agents/tdd-test-writer.md): RED; one test per AC,
  compile-only stub, RED must be assertion failures, not `[build failed]`.
- [`.claude/agents/tdd-implementer.md`](../../../../.claude/agents/tdd-implementer.md): GREEN; `*_test.go` is a
  read-only contract, a contradiction is reported (`GREEN blocked: ...`), never fixed in the test.
- [`.claude/agents/tdd-refactorer.md`](../../../../.claude/agents/tdd-refactorer.md): REFACTOR; tests after
  every step, no behaviour change, "no change" allowed instead of the demo's "at least 2 helpers".
- [`.claude/skills/tdd/SKILL.md`](../../../../.claude/skills/tdd/SKILL.md) +
  [`gate.sh`](../../../../.claude/skills/tdd/gate.sh): the coordinator calls the three agents with the Agent tool
  and runs a deterministic gate after each (commit subject, test exit code and `--- FAIL` vs build error,
  `git diff <red> HEAD -- '*_test.go'` empty, `go vet`, `gofmt`). `--review-tests` stops after RED;
  `--continue` resumes.

Go adaptations of the Python template: the test command comes from the story's front matter; the tests live next
to the code (`*_test.go`), so the immutable contract is a glob, not a `tests/` folder; RED needs a stub to compile.

### Runs

| | `/tdd T17b` | one session T17b | `/tdd T19b --review-tests` | one session T19b |
|---|---|---|---|---|
| Commits | `51611cc` test · `009795d` feat · `ddf9874` refactor | `9a7c187` · `54fabfd` · `cc8ecdd` | `3981add` · `323750b` · `4685a3b` | `989b17e` · `cda2dbc` · `b700658` |
| Wall time | 2 m 28 s (RED 52 s, GREEN 27 s, REFACTOR 38 s) | 1 m 39 s | 3 m 14 s machine + 1 m 09 s review | 1 m 27 s |
| Rounds to green | 1 | 1 | 1 | 1 |
| Tests / implementation lines | 163 / 78 (2.1) | 143 / 81 (1.8) | 175 / 157 (1.1) | 157 / 116 (1.4) |
| Attempts to change a test | 0 | 0 | 0 | 0 |
| Gate stops | 0 (4 × PASS) | n/a; gates run afterwards: 3 × PASS | 0 (5 × PASS) | n/a; afterwards 3 × PASS |
| Output tokens | 13.3 k (coordinator 3.2 k + agents 10.1 k) | 9.4 k | 17.7 k (6.6 k + 11.0 k) | 9.5 k |
| Cache read | 1.04 M | 0.65 M | 1.45 M | 0.74 M |

Every RED in all four runs failed on assertions, not on the build. Every test suite passes on the other run's
implementation (cross-run check), so the four implementations behave the same on everything any test asks.

### Test quality: the same mutants against both test suites

Each mutant breaks one story rule in the `/tdd` implementation; both suites run against it.

| Mutant | `/tdd` tests | one-session tests |
|---|---|---|
| T17b upload date decides instead of document date | killed | killed |
| T17b tie by id reversed / VAT not added / VAT for non-payer / advance eligible / input mutated | killed | killed |
| T17b invoices no longer win over a newer estimate | **survived** | **survived** |
| T19b `<`, `>`, `&` HTML-escaped | **survived** | killed |
| T19b time not in UTC / meeting date as RFC 3339 / `nil` not `null` | killed | killed |

- The invoice-priority gap is in the story: no AC puts an invoice next to a newer estimate. Both suites test the
  AC, so both miss it.
- The HTML gap is in the `/tdd` tests only. The test-writer's prompt says "do not test what the AC does not say";
  the rule lives in the story's Rules section, not in an AC. The one-session run tested it anyway.
- `--review-tests` found exactly this gap: at the stop the coordinator listed three untested rules (HTML
  escaping, other unsupported types, `nil` / midnight dates) and proposed a new AC-6. The run continued without
  adding it, so the gap stayed; the implementer still wrote `SetEscapeHTML(false)` from the story's rules.

### Mini-essay

Isolation did what it promises mechanically: the test writer never saw an implementation, the implementer could
only read a committed contract, and `git diff -- '*_test.go'` stayed empty in every run, but it stayed empty in the
single sessions too: on these stories the model never tried to change a test, so the guard had nothing to stop.
The tests were not better with isolation; they were narrower. The isolated writer followed its "AC only" rule and
left a stated rule untested, while the single session, which read the whole story as one task, tested it. A
mutation run showed the difference where coverage numbers would not. `--review-tests` worked as a human review
would: it read the tests against the story, named the same gap a reviewer would, and was useless the moment the
run continued without acting on it. The price is real: 1.5-2.2× the wall time and 1.4-1.9× the output tokens for
the same four commits, and no measurable gain. On small, well-specified domain functions where the first
implementation is green, three agents are overhead. Their worth starts where the implementer is tempted to bend a
test: unclear AC, slow or flaky tests, many red rounds. These stories were too simple to show it; next time a story
with an integration test and a real conflict between the AC and the existing code.

## Branches

- easy: `feat/governance-meeting-facts`
- agents and coordinator: `chore/claude-tdd-agents`
- `/tdd`: `feat/governance-amount-recommendation-tdd` (PR #36), `feat/governance-snapshot-fingerprint-tdd`
  (PR #37); both merged, so T17b, T19b, the agents and the coordinator are on `main`
- one session: `feat/governance-amount-recommendation-solo`, `feat/governance-snapshot-fingerprint-solo`; kept as
  evidence, not merged (they duplicate T17b and T19b)
