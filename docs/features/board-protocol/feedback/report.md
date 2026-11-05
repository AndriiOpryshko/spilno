# Feedback loops · homework 7.6 · live browser channel on a Figma drift + clean-context reviewer on two governance tasks · 2026-11-05

- Demo `modules/7-execution-scale/7.6-feedback-loops`: `make verify` and `make gate` first failed for the environment
  (`sh: 1: vitest: not found`, `make: *** [Makefile:36: verify] Error 127`, no `node_modules`). After `npm ci` from
  the demo's lock both are RED by design (story-28, test-first): `Tests  2 failed | 5 passed (7)`, exit 2; in
  `make gate` `tsc --noEmit` passed, since `&&` reached vitest.
- Easy level: the note under `TransitionActions` drifted from Figma (05 Lifecycle, node 31-207). Closed in 3 turns
  of edit → `playwright-cli eval getComputedStyle` → fix, then locked into the story tests.
- Hard level: two tasks on the governance critical path (readiness gate on «готовий до підпису», slot confirmation
  under the protocol revision), each with (a) the author's self-check and (b) a clean-context reviewer bound by
  [`reviewer-contract.md`](reviewer-contract.md).
- Isolation: worktree `../spilno-fb`, branch `docs/board-protocol-feedback-loops` from `origin/main` `31d54b0`;
  implementers in `../spilno-fb-a` and `../spilno-fb-b` (same base). Another session's dev server holds :5180, so
  this run used its own Storybook on :6016.
- Metrics come from the transcripts (`~/.claude/projects/-Users-andriio-course-learning-spilno/8af49cd0-….jsonl` and
  `…/subagents/agent-*.jsonl`) and git, not from the agents' summaries. Raw artifacts are in
  `/Users/andriio/course/learning/hw-7.6-screens/`.

## Demo: what each finished channel looks like

| Channel | Finished form | Signal | Blind to |
|---|---|---|---|
| `verify-gate` (`/verify-gate story-28`) | `npx tsc --noEmit && npm test` as the only definition of done; the same gate in `scripts/pre-commit` and the opt-in Stop hook `.claude/hooks/verify.sh`, which turns red into exit 2 so the turn cannot end | exit code + the pasted `N passed` line | anything the tests do not describe (`queue.test.ts` is green over an off-by-one) |
| `verify-ui` (`/verify-ui story-26`) | `playwright-cli open :5173` → `snapshot` → per-AC action → `eval getComputedStyle(...)` → `screenshot`; a per-AC PASS/FAIL with actual vs expected (`paddingLeft 8px`, expected `12px`); never fixes code | the live page state, read as numbers | logic the AC does not name; it needs a running dev server |
| `code-reviewer-subagent` | a `Task` subagent that gets only the story AC and `git diff`, and answers exactly one of ACCEPT / WARN / PARTIAL / REJECT with AC-bound reasons (story-27: `saveQueue` never called → REJECT, then ACCEPT) | a discrete verdict | everything outside the diff |
| `baselines/` + `screenshot-diff` | `baselines/edit-card-form.png`, a real 512×300 Chromium shot of the correct form; a new shot in `tmp/`, a share of differing pixels against a 0.5% threshold and a diff map; the baseline is never regenerated to match | % of pixels + where | anything below the threshold (see the easy level) |

## Easy level

### Task and criterion

Figma `05 Lifecycle / TransitionActions` (`?node-id=31-207`, read through the Figma MCP `get_design_context`) draws
the note under the buttons (variants *Ready for signing* 31-186 and *Disabled reason* 31-202) as a 14px glyph,
**6px** gap, text **12px / 17px**. The component used `gap-2` (8px), `text-[13px]` for the disabled reason and
`text-xs` (12/16) for the void-copy note.

**Criterion:** proven by the page only. In the live Storybook iframe, for both stories, `getComputedStyle(note)`
gives `columnGap 6px`, `fontSize 12px`, `lineHeight 17px`; the text starts 20px from the glyph's left edge; and the
group's `getBoundingClientRect().height` is **61px**, the height of both Figma variants (36 + 8 + 17).

### Iterations

| Turn | Time (UTC) | Change | Machine evidence (`playwright-cli eval`) | Caught |
|---|---|---|---|---|
| 0 | 19:49:55 | none (`origin/main`) | disabled-reason `gap 8px · 13px · lh 19.5px · text +22px · group 63.5px`; ready-for-signing `gap 8px · 12px · lh 16px · +22px · group 60px` | FAIL: gap, size, line height |
| 1 | 19:50:19 | `gap-2` → `gap-1.5`, `text-[13px]` → `text-xs` | both `gap 6px · 12px · lh 16px · +20px` | FAIL: `lineHeight 16px`. Tailwind `text-xs` ships its own 16px line height |
| 2 | 19:50:26 | `text-xs` → `text-xs/[17px]` | both `gap 6px · 12px · lh 17px · +20px`, group `61px` | PASS |

The fix is commit `0692472`. The criterion is then written into the stories (`expectFigmaNote` in
`transition-actions.stories.tsx`). A mutation run puts the old component back and the stories go red:
`Tests  2 failed | 5 passed (7)` (`story-assert-red-on-old-code.txt`). With the fix: `test:stories` `43 passed`,
`tsc -b` and eslint clean.

**What the channel caught that the usual tests did not:** all 43 story tests were green over the drift, because
they assert roles, text and disabled state, not geometry. The pixel diff with the demo's default 0.5% threshold
would have missed it too: before vs after differs by 0.187% and 0.363% of the frame, confined to the note row
(bbox y 69-84). Only reading the computed style against the Figma numbers turned the drift into a FAIL, and turn 1
showed that a plausible fix still left the line height wrong.

Artifacts: `ta-*-before.png`, `ta-*-after.png`, `ta-*-diff.png`, `figma-transition-actions-31-207.png`,
`easy-iterations.txt`.

## Hard level

### Tasks (critical path: protocol state and data integrity)

- [**T-A**](tasks/T-A.md): `Protocol.MarkReadyForSigning` never consulted `Readiness()`, so a draft with an
  unconfirmed slot or no agenda could be printed (ADR-0009, AC-21, AC-04). AC A1-A5: the same `CodedFieldsError` as
  `EnsureReadyForSigning`, nothing changed on refusal, a fixed refusal order (guard → already printed → readiness →
  number), unchanged happy path, no weakened tests.
- [**T-B**](tasks/T-B.md): `Slot.Confirm` / `Slot.Correct` were reachable past the revision guard (ADR-0009:
  every slot change bumps the revision; a stale write is `409 governance.draft_revision_conflict`). AC B1-B5:
  `Protocol.ConfirmSlot` / `CorrectSlot`, guard first, unknown item or key refused without a panic, success as a
  content change (revision +1, back to draft with `ClosePollRound` / `VoidPrintCopy`), tests for each.

### Reviewer contract

[`reviewer-contract.md`](reviewer-contract.md), based on the demo's `code-reviewer-subagent`. Input: the task's AC
and `git diff -U25`, nothing else: no conversation, plan, test output or author's summary, and no repository access.
Output: exactly one of ACCEPT / WARN / PARTIAL / REJECT, findings bound to an AC id or `collateral` and a diff line,
and a 2-4 sentence rationale. Mode (a) is the same procedure run by the author with its history, plus `Ready: yes|no`.

### Run

| | Mode | Agent | Turns | Tools | Output tok | Cache read | Wall |
|---|---|---|---|---|---|---|---|
| T-A | (a) author implements + self-checks | general-purpose, worktree `spilno-fb-a` | 12 | Bash 11 | 11 162 | 716 592 | 123 s |
| T-A | (b) clean-context reviewer | general-purpose, Read only | 2 | Read 3 | 4 077 | 62 137 | 47 s |
| T-B | (a) author implements + self-checks | general-purpose, worktree `spilno-fb-b` | 10 | Bash 7, Write 2 | 9 254 | 551 204 | 99 s |
| T-B | (b) clean-context reviewer | general-purpose, Read only | 2 | Read 3 | 2 937 | 62 137 | 34 s |

The reviewers' transcripts show exactly three `Read` calls each: the contract, the AC and the diff. That is the
clean context. The same diff text went to both modes.

### Findings

Every finding is checked against the code. *Real* = true and actionable; *diff-blind* = an unknown that comes from
seeing only the diff, settled by the code or a run; *false* = wrong.

| Task | Mode | Verdict | Findings | Real | Diff-blind | False |
|---|---|---|---|---|---|---|
| T-A | (a) self | WARN, `Ready: yes` | 3 minor: fixture shifts revision literals 1/2 → 2/3; one order test tells errors apart by type; slot statuses not re-asserted after a refusal | 3 | 0 | 0 |
| T-A | (b) clean | WARN | 3 minor: suite-green unprovable from the diff; no multi-blocker case through `MarkReadyForSigning`; `guard` body not in the diff | 1 | 2 | 0 |
| T-B | (a) self | WARN, `Ready: yes` | 3 minor: **the old bypass stays open** (`Slot.Confirm` exported, `AgendaItem.Slot` hands out `*Slot`); `ErrSlotNotFound` has no HTTP mapping; no withdrawn + stale row | 3 | 0 | 0 |
| T-B | (b) clean | WARN | 4 minor: `Confirm`/`Correct` errors dropped (**false**: both return nothing); `reopen`/`guard` not in the diff; helpers not in the diff; refusal snapshot skips `confirmedAt` and `alternative` (**fixed**) | 1 | 2 | 1 |

- **No verdict the self-check missed:** all four verdicts are WARN, and neither mode raised a blocker.
- **The clean reviewer's unique catch:** the T-B snapshot gap (`confirmedAt`, `alternative`). It was fixed before
  the commit.
- **The self-check's unique catch:** the most substantive finding, the bypass that T-B leaves open. A diff-only
  reviewer cannot see it: the bypass lives in `slot.go` and `agenda_item.go`, which the diff does not touch.
  Half of the clean reviewer's findings (4 of 7) were diff-blind or false, for the same reason.
- **What both modes missed:** both authors said `Ready: yes`, and each diff was green on its own branch. Applied
  together on this branch, `go test ./internal/governance/...` went red with **18 failing subtests** in T-B
  (`setup to signed: governance.slot_unconfirmed: /agenda_items/0/slots/amount`). T-B's fixture drove a draft with
  a needs-review slot to готовий до підпису, which is exactly what T-A now forbids. Neither reviewer could see it:
  each saw one diff. The deterministic gate caught it on integration. The fixture now confirms the slot before the
  transition and offers the later value after it. Result: 212 domain test cases (162 before), commits `43862d0`
  and `67528ea`.

### Calibration

| Change | Channel | Evidence here |
|---|---|---|
| Critical path (protocol state, revision, readiness) | visible channel (deterministic gate: `go test` on the integrated branch, `mise run test`) **+** clean-context reviewer | gate caught the 18-subtest cross-task break; reviewer added one real finding over the self-check |
| Cosmetic change next to it (the note spacing) | `tsc` **+** screenshot comparison (computed style / pixel diff against the previous render) | `tsc -b` clean; diff confined to the note row, 0.19% / 0.36% |

The reasoning: a wrong state transition or a lost revision check corrupts a legal record that cannot be fixed
after signing, so it pays for two independent signals (a machine one that sees the whole branch and a reviewer
that does not share the author's assumptions), while a misplaced 2px only needs proof that it compiles and that the
render changed where intended and nowhere else.

### The self-check's blind spot and the honest price

The self-check knows the codebase and the intent, so it saw the most substantive problem: the bypass T-B leaves
open. It still answered `Ready: yes`. It judged against the AC it was given, and the AC did not ask for the bypass
to be closed. The blind spot is not ignorance but scope plus a stake in finishing: the author checks the task as
framed and reads its own fixtures as correct. The clean reviewer removes the stake but also the knowledge. It added
one real finding (the snapshot gap), produced one false positive, and spent half its findings on "not in the diff".
Neither mode could see the one defect that mattered, the cross-task break, because each looked at one diff. Only
the deterministic gate on the integrated branch saw the whole system.

The price: a review costs ~3-4k output tokens, ~100k input (62k cache read + 36-41k cache create) and 34-47 s per
diff, against ~10k output, 550-720k cache read and 100-125 s for implementation plus self-check. That is cheap
next to a print that ignores unconfirmed amounts. But in this run the reviewer's marginal yield over the
self-check was one minor finding. The course's warning applies: the diffs were small, domain-only and already
test-heavy, so the expensive failure was in integration, not inside a diff. Next time the reviewer should get a
path where a diff alone can do damage (the SQL/adapter side of the revision check or the lock of a signed
protocol) and read access to the files the diff calls into, so "not in the diff" stops eating its findings.

## Verification

- `mise run lint`: exit 0. `mise run test`: exit 0 (Go unit + integration with testcontainers, Vitest `43 passed`,
  stories `43 passed`; log `mise-test.log`).
- `mise run e2e` was not run: it needs the core API on :8080, and the :5180 server belongs to another session. The
  changes are a domain method with no caller outside `domain/` yet and a CSS class on a component.

## Follow-ups

- Close the slot bypass: unexport `Slot.Confirm`/`Correct` or stop `AgendaItem.Slot` from handing out a mutable
  `*Slot`, once the app layer calls `ConfirmSlot`/`CorrectSlot` (T-B self-check F1). Until then the T-B test fixture
  itself offers a value on a signed protocol's slot through that path.
- Map `ErrSlotNotFound` to a problem response when the HTTP operation lands.
- The Figma `ReadinessCard` tier-3 note in `docs/design/design-spec.md` §14 is still Proposed.
