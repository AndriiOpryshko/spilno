# Reviewer contract · clean-context diff review (homework 7.6)

Base: the demo skill `code-reviewer-subagent` (`modules/7-execution-scale/7.6-feedback-loops/.claude/skills/`),
tightened for a Go domain change on the governance critical path.

## Input (nothing else)

1. The acceptance criteria of the task, verbatim (`tasks/<id>.md`).
2. `git diff -U25 <base>...<head>` of the change, as text.

The reviewer gets no conversation, no plan, no test output, no author's summary and no history of how the code
was written. It does not open the repository: if a fact is not in the diff or the AC, it is unknown, and an
unknown that an AC depends on is a finding, not a pass.

## Procedure

1. List the AC as numbered checks.
2. For each AC, find the lines in the diff that satisfy it, and the test that proves it. "Satisfied" needs both
   the production line and a test assertion that would fail without it.
3. Look for what the AC forbids: state, revision, number, transitions or slots changed on a refused call;
   a check ordered so that a cheaper refusal hides the one the AC names; an error code other than the AC's.
4. Look for collateral damage visible in the diff: tests deleted or weakened, assertions loosened, fixtures
   changed so an existing test stops covering its case.

## Output (exactly this shape)

```
Verdict: ACCEPT | WARN | PARTIAL | REJECT
Findings:
  F1 [blocker|major|minor] <AC id or "collateral"> - <file:line in the diff> - <what is wrong, one sentence>
  ...
Rationale: <2-4 sentences tied to AC ids>
```

- **ACCEPT** - every AC is met and proven by a test; no findings above minor.
- **WARN** - every AC is met, with minor findings (naming, a missing extra test, small risk).
- **PARTIAL** - at least one AC is met and at least one is not met or not proven by a test.
- **REJECT** - a key AC is violated by the code, or the change breaks an invariant it touches.

Exactly one verdict. No "mostly fine". Every finding cites an AC id or `collateral` and a diff location.

## Self-check variant (mode a)

The author agent applies the same procedure and the same output to its own diff, with its full history in
context, and adds one line: `Ready: yes | no`.
