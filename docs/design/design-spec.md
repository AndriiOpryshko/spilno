---
feature: board-protocol
status: Verified
owner: Andrii Opryshko (product owner, frontend)
updated_at: 2026-10-28
---

# Design spec: Protocol draft — lifecycle on the first screen

## 1. Raw request

> Член правління на /cases/$caseId/protocol бачить стан протоколу в його життєвому циклі (чернетка →
> готовий до голосування → готовий до підпису → підписаний, знятий з розгляду) і готує чернетку до
> попереднього голосування та друку, не зламавши provenance слотів, live validation і readiness, які
> вже є на екрані.

## 2. Why now

- **Problem:** the screen shows one undifferentiated "draft". The board cannot see where the protocol is in its life (draft, informal poll, print, signed), cannot move it forward or back, and the print gate still blocks on tier-3 charter checks that PRD Stage 1 moved to Stage 2.
- **User:** член правління (any current board member edits; CONTEXT.md, PRD.ao-prd.md §10).
- **Trigger:** Stage 1 was rescoped to draft → informal poll → print → sign (PRD.ao-prd.md §1, 2026-10-25); this is the first screen of the module "design system + first screen".
- **Success outcome:** in one glance the member sees the current state and the one next action; moving forward or back is one click; printing is possible exactly when every AI-proposed value is confirmed.

## 3. Evidence sources

| Status | Source | What it proves |
|---|---|---|
| Observed | `apps/web/src/routes/cases.$caseId.protocol.tsx` | Route `/cases/$caseId/protocol` renders `ProtocolScreen` keyed by `caseId` |
| Observed | `apps/web/src/features/governance/protocol-screen.tsx` | TopBar crumbs + "Збережено" indicator; skeleton while pending; `Unavailable` alert on error; body = header, slot rows, side panel (ReadinessCard + Properties); mobile stacks readiness above and properties below |
| Observed | `apps/web/src/features/governance/slot-row.tsx` | Per-slot glyph, label, source line, `OriginTag`; textarea for editable slots, read-only text, `AmountValue` for amounts; Save/Cancel, `⌘ Enter` / `Esc`; "Збереження…"; inline error; tier warnings via `LimitNote` |
| Observed | `apps/web/src/features/governance/readiness.ts`, `readiness-card.tsx` | Print enabled only when no tier-3 warning and no unverified slot; segmented progress; `window.print()` |
| Observed | `apps/web/src/features/governance/validation.ts`, `inline-hint.tsx` | Live hints on every keystroke: suggestion with "Підставити", example chips, "Поле потрібне для протоколу"; never block saving |
| Observed | `apps/web/src/features/governance/api.ts`, `types.ts`, `mock-data.ts` | In-memory API (350 ms latency): `getProtocol`, `updateSlot` (sets `origin: human`); no lifecycle state in `ProtocolDraft`; `c-101` has a draft, `c-102` has none |
| Observed | `apps/web/src/i18n/uk.ts` (`protocol.*`) | All current Ukrainian copy |
| Observed | `docs/design/design-system.md` | Tokens (light/dark), primitives (Button, Badge, Card, Input, Textarea, Skeleton, Kbd), governance components, a11y rules, and Gaps (no lifecycle badge/stepper, no `--success-subtle`) |
| Observed | `docs/design/protocol-brief.md` | Keep-list, states and layouts of the approved redesign (Figma `D06IijJnLn8WEgU85fnD84`) |
| Confirmed | `docs/features/board-protocol/PRD.ao-prd.md` §10, AC-13, AC-17, AC-25, AC-25b, AC-26, AC-46, AC-06 | Stage 1 states and transitions; poll is informal and labelled so; print blocked only by unconfirmed AI values |
| Confirmed | `docs/features/board-protocol/CONTEXT.md` | Glossary: чернетка, готовий до голосування, раунд опитування, друкована копія, підписаний, знятий з розгляду |
| Confirmed | User answers, 2026-10-28 | Scope = чернетка + готовий до голосування + готовий до підпису with transitions; підписаний / знятий з розгляду as read-only specimens; print gate aligned to PRD Stage 1; poll shows state and result only; in-memory model may gain `state`, `round`, poll result |
| Proposed | Sections 9 and 14 | New `ProtocolStateBadge`, `LifecycleStepper`, `--success-subtle` token |

## 4. Current user flow

1. **Start:** member opens `/cases/c-101/protocol` from the cases list.
2. **Action:** reads slots, edits an editable slot, accepts a hint ("Підставити" / chip), saves with Save or `⌘ Enter`, cancels with Cancel or `Esc`.
3. **System response:** skeleton for ~350 ms; on save the slot becomes `human` + `verified`, the top bar shows "Збережено щойно", ReadinessCard recounts.
4. **Success:** when no slot is unverified and no tier-3 warning remains, "Надрукувати протокол" enables and calls `window.print()`.
5. **Recovery:** no draft (`/cases/c-102/protocol`) → `role="alert"` "Чернетки протоколу ще немає" with "До справ"; save failure → inline error under the slot.

## 5. Target user flow

1. **Start:** same route; the header shows the protocol state badge and a lifecycle stepper (чернетка → готовий до голосування → готовий до підпису → підписаний).
2. **Action (чернетка):** edit and confirm slots as today; then either "Подати на попереднє голосування" (AC-13) or "Позначити готовим до підпису" (AC-25).
3. **Action (готовий до голосування):** slots are read-only; a poll panel shows "Раунд N", per-item tallies for/against/abstain and the label «Неформальне опитування — не голосування засідання» (AC-46); actions: "Повернути в чернетку" (AC-17) or "Позначити готовим до підпису" (AC-25b).
4. **Action (готовий до підпису):** slots read-only; protocol number shown; "Надрукувати протокол" prints; "Повернути в чернетку" voids the print copy and keeps the number (AC-26).
5. **System response:** every transition shows a pending button state, then the new badge, stepper step and "Збережено щойно"; editing a field in чернетка behaves exactly as today.
6. **Success:** member reaches готовий до підпису and prints.
7. **Recovery:** a transition failure shows an inline alert in the actions area and leaves the state unchanged; "Позначити готовим до підпису" stays disabled with a reason while an AI-proposed value is unconfirmed (AC-06).

## 6. Requirements

| ID | Status | Requirement | Source |
|---|---|---|---|
| BR-01 | Observed | Route `/cases/$caseId/protocol`, TopBar crumbs (Справи › case › Протокол № N) and the "Збережено / Збережено щойно" indicator remain | `protocol-screen.tsx` |
| BR-02 | Observed | Each slot keeps its status glyph, label, source line (document title / "лише читання" / "Заповнено вручну" / "Не знайдено в документах справи") and origin tag (system / AI / human / editing) | `slot-row.tsx`, `origin-tag.tsx` |
| BR-03 | Observed | In чернетка, editable slots save through `updateSlot()` with Save / `⌘ Enter`, cancel with Cancel / `Esc`; saving shows "Збереження…" and disables both buttons; saving sets `origin: human` | `slot-row.tsx`, `api.ts` |
| BR-04 | Observed | Live validation runs on every keystroke and never blocks saving: suggestion + "Підставити", example chips, "Поле потрібне для протоколу" | `validation.ts`, `inline-hint.tsx` |
| BR-05 | Observed | Amounts render large with fraction and amount in words; non-editable AI and system slots cannot be edited | `slot-row.tsx` `AmountValue` |
| BR-06 | Observed | Loading shows the skeleton with `aria-busy`; missing draft shows the `role="alert"` block with "До справ" | `protocol-screen.tsx` |
| BR-07 | Confirmed | The header shows a state badge for the current lifecycle state with the CONTEXT.md label, and a stepper чернетка → готовий до голосування → готовий до підпису → підписаний with the current step marked | PRD §10, user answer |
| BR-08 | Confirmed | In чернетка the actions are "Подати на попереднє голосування" and "Позначити готовим до підпису" | AC-13, AC-25 |
| BR-09 | Confirmed | In готовий до голосування slots are read-only; the poll panel shows "Раунд N", per-item tallies and the label «Неформальне опитування — не голосування засідання»; actions "Повернути в чернетку" and "Позначити готовим до підпису"; no vote controls in this iteration | AC-13, AC-17, AC-25b, AC-46, user answer |
| BR-10 | Confirmed | In готовий до підпису slots are read-only, the protocol number is shown, "Надрукувати протокол" is enabled, and "Повернути в чернетку" says the printed copy becomes void while the number stays | AC-25, AC-26 |
| BR-11 | Confirmed | The print gate ("Позначити готовим до підпису" and printing) is blocked only while a slot proposed by the помічник (`origin: ai`) is not confirmed; tier-3 warnings stay visible as advisory notes and no longer block | AC-06, PRD §3 Stage 2 list, user answer |
| BR-12 | Confirmed | ReadinessCard keeps its segmented progress and list, but counts only what blocks per BR-11 and names the blocking slots | `readiness-card.tsx`, AC-06 |
| BR-13 | Confirmed | підписаний and знятий з розгляду exist as read-only specimens (badge + stepper + read-only slots, no actions); no scan upload and no withdraw dialog in this iteration | user answer |
| BR-14 | Confirmed | Every transition shows a pending state on its button, then updates badge, stepper and "Збережено щойно"; a failure shows an inline alert in the actions area and keeps the previous state | PRD §5.1 `transition_not_allowed`, design-system Gaps (Alert) |
| BR-15 | Observed | Desktop 1440: main column + 320 px side panel (sticky). Mobile 390: single column, readiness above slots, properties below; lifecycle actions reachable without horizontal scroll | `protocol-screen.tsx`, `protocol-brief.md` |
| BR-16 | Observed | Only existing tokens and primitives (Button, Badge, Card, Skeleton, Kbd, StatusGlyph, LimitNote); no new dependency; `radix-ui` already installed may be used only through an existing wrapper | `design-system.md`, `package.json` |
| BR-17 | Observed | Accessibility: every control has a name, focus is visible, state changes are announced (`aria-live="polite"`), the stepper exposes the current step (`aria-current="step"`), contrast ≥ AA in light and dark; story a11y tests stay at `error` | `design-system.md` a11y, commit d181e56 |
| BR-18 | Confirmed | Data changes stay in the in-memory model (`types.ts`, `mock-data.ts`, `api.ts`): `state`, `round`, poll tallies and transition functions; `services/core/api/openapi.yaml` and the backend are untouched | user answer |

## 7. Routes and frames

| Surface | Route / frame | Required content and layout |
|---|---|---|
| Current | `/cases/c-101/protocol` | Draft with mixed slot statuses and a tier-3 limit (see `docs/design/screenshots/protocol-after-*.png`) |
| Desktop | `1440 x 900` | Sidebar · TopBar · header (eyebrow, title, state badge, stepper, intro) · actions row · slot rows (label 220 px + value) · sticky side panel: ReadinessCard, poll panel (готовий до голосування), Properties |
| Mobile | `390 x 844` | TopBar · header with badge and a compact stepper (current step label + "крок N з 4") · ReadinessCard · actions (full-width buttons) · slots stacked · poll panel · Properties |
| Specimens | `c-101` чернетка, `c-103` готовий до голосування, `c-104` готовий до підпису, plus підписаний and знятий з розгляду | One frame per state at both widths |

## 8. State inventory

| State | Trigger | Visible result | Recovery / next action |
|---|---|---|---|
| Empty | No draft (`c-102`) | `role="alert"` "Чернетки протоколу ще немає" + "До справ" | Back to cases |
| Loading | First load / transition pending | Skeleton (`aria-busy`); during a transition the clicked button shows a pending label and both action buttons are disabled | Wait |
| Ready — чернетка | `state: draft` | Badge "чернетка", stepper step 1, editable slots, actions per BR-08 | Edit, poll, or mark ready for signing |
| Ready — готовий до голосування | `state: ready_for_vote` | Badge, step 2, read-only slots, poll panel with round and tallies, informal label | Return to draft / mark ready for signing |
| Ready — готовий до підпису | `state: ready_for_signing` | Badge, step 3, number, print enabled, return explains the void copy | Print / return to draft |
| Specimen — підписаний | `state: signed` | Badge (success tone), step 4 done, read-only, no actions | — |
| Specimen — знятий з розгляду | `state: withdrawn` | Badge (neutral), stepper muted, reason shown, no actions | — |
| Success | Slot saved / transition done | "Збережено щойно", badge and stepper updated, live region announces the new state | Continue |
| Error | Save or transition rejected | Inline error under the slot / inline alert in the actions area; previous state kept | Retry |
| Disabled / focus / hover | Unconfirmed AI slot; keyboard focus | "Позначити готовим до підпису" disabled with the reason "Підтвердіть: <slot>"; visible focus ring on every control; hover per Button variants | Confirm the slot |

## 9. Reuse and handoff contract

| Design responsibility | Current code owner / token | Decision |
|---|---|---|
| Page shell, crumbs, saved indicator | `components/app-shell` `TopBar`, `protocol-screen.tsx` | reuse |
| Slot row, hints, origin tag, amount | `slot-row.tsx`, `inline-hint.tsx`, `origin-tag.tsx` | reuse; add a read-only mode for non-draft states |
| Readiness card | `readiness-card.tsx`, `readiness.ts` | extend (gate per BR-11) |
| Tier warnings | `limit-note.tsx` | reuse (advisory only) |
| State badge | `components/ui/badge.tsx` + `StatusGlyph` | new component candidate `ProtocolStateBadge` (Badge variants + glyph) |
| Lifecycle stepper | — | new component candidate `LifecycleStepper` |
| Poll panel | `components/ui/card.tsx` | new component candidate `PollRoundPanel` composed from Card |
| Transition actions + inline error | `components/ui/button.tsx`; alert markup in `protocol-screen.tsx` | reuse Button; extract the existing alert markup into one component candidate `InlineAlert` |
| Signed tone | `--success`, missing `--success-subtle` | new token candidate (Proposed) |
| Copy | `i18n/uk.ts` `protocol.*` | extend with lifecycle labels from CONTEXT.md |
| Data | `types.ts`, `mock-data.ts`, `api.ts` | extend in memory only (BR-18) |

## 10. Responsive and accessibility contract

- **Responsive:** desktop keeps the 1fr + 320 px grid with the sticky side panel; mobile is one column with readiness and actions above the slots; the stepper collapses to "крок N з 4 · <label>" under 768 px; no horizontal scroll at 390 px; action buttons are full width on mobile.
- **Keyboard:** tab order = crumbs → stepper (not focusable, read-only) → action buttons → slots → side panel; `⌘ Enter` / `Esc` keep working in чернетка only; focus returns to the first action after a transition.
- **Semantics:** one `h1`; the badge text is the state name; stepper is an `ol` with `aria-current="step"`; state changes are announced through a polite live region; errors use `role="alert"`; disabled actions expose their reason as visible text, not only a tooltip.
- **Contrast and motion:** all tones ≥ AA (4.5:1 for text) in light and dark; no animation beyond the existing transitions; `prefers-reduced-motion` respected.

## 11. Content and realistic data

- States (CONTEXT.md): «чернетка», «готовий до голосування», «готовий до підпису», «підписаний», «знятий з розгляду».
- Actions: «Подати на попереднє голосування», «Повернути в чернетку», «Позначити готовим до підпису», «Надрукувати протокол».
- Poll label: «Неформальне опитування — не голосування засідання»; «Раунд 2 · 3 з 4 членів правління проголосували».
- Return from готовий до підпису: «Надрукована копія стане недійсною. Номер № 14 залишиться за протоколом.»
- Disabled reason: «Підтвердіть: Сума за договором».
- Data: case `c-101` «Гідроізоляція покрівлі, під’їзд 2», 184 500,00 грн; board of 4 members; protocol № 14.

## 12. Non-goals and forbidden changes

- No vote controls, no scan upload, no meeting facts form, no withdraw dialog (next iteration).
- No change to `services/core/api/openapi.yaml`, Go code, migrations or generated clients.
- No new npm dependency; no new design tokens except the Proposed `--success-subtle`.
- No change to cases list (`cases-screen.tsx`) behaviour.
- Live validation, provenance (source line, origin tag) and amount rendering must not change.

## 13. Design references

- Current UI evidence: `docs/design/screenshots/protocol-after-desktop.png`, `protocol-after-mobile.png`, `protocol-after-edits-*.png`; live route `/cases/c-101/protocol`.
- Approved frame URLs / node IDs: base direction `https://www.figma.com/design/D06IijJnLn8WEgU85fnD84?node-id=23-137` (desktop), `?node-id=23-420` (mobile), `?node-id=23-650` (states).
- Lifecycle page `05 Lifecycle` (`?node-id=31-3`), drawn through the Figma MCP on 2026-10-28:

  | Node | Link | Covers |
  |---|---|---|
  | Components section | `?node-id=31-208` | — |
  | `ProtocolStateBadge` (5 variants) | `?node-id=31-19` | BR-07 |
  | `LifecycleStepper` (Step 1–4, Withdrawn, Compact 1–3) | `?node-id=31-144` | BR-07, BR-17 |
  | `PollRoundPanel` | `?node-id=31-145` | BR-09 |
  | `TransitionActions` (Draft, Ready for vote, Ready for signing, Pending, Disabled reason) | `?node-id=31-207` | BR-08…BR-11, BR-14 |
  | `InlineAlert` | `?node-id=31-161` | BR-14 |
  | чернетка · desktop / mobile | `?node-id=32-21` / `?node-id=32-1850` | BR-08, BR-11, BR-15 |
  | готовий до голосування · desktop / mobile / dark | `?node-id=32-295` / `?node-id=32-2069` / `?node-id=32-2577` | BR-09, BR-15, BR-17 |
  | готовий до підпису · desktop / mobile | `?node-id=32-633` / `?node-id=32-2319` | BR-10, BR-15 |
  | підписаний · desktop (specimen) | `?node-id=32-938` | BR-13 |
  | знятий з розгляду · desktop (specimen) | `?node-id=32-1217` | BR-13 |
  | Transition states (pending, disabled with reason, error) | `?node-id=32-2543` | BR-11, BR-14 |
  | Token `color/success-subtle` (Light `#E7F6F3`, Dark `#0F2B28`) | Color collection | BR-16 |
- Repo-local design file: N/A — Figma is the design tool for this project.

## 14. Open decisions

| Decision | Why it matters | Owner | Due | Status |
|---|---|---|---|---|
| Add `--success-subtle` (light/dark) for the підписаний badge | No subtle success tone exists; without it the badge reuses `--success` text on `--card` | Andrii Opryshko | before design | Confirmed 2026-10-28 — added to Figma `Color` (Light `#E7F6F3`, Dark `#0F2B28`); mirror in `apps/web/src/index.css` |
| `ProtocolStateBadge` as Badge variants vs a separate component | Badge has no status variants and no story | Andrii Opryshko | before design | Confirmed 2026-10-28 — separate `ProtocolStateBadge` (pill + dot) composed from `Badge` in code, with a story per state |
| Mobile stepper as "крок N з 4" text vs a compact 4-dot bar | Space at 390 px | Andrii Opryshko | in Figma | Confirmed 2026-10-28 — both: "Крок N з 4 · <state>" text over a 4-segment bar (`LifecycleStepper` Compact) |
| `ReadinessCard` in чернетка still lists a tier-3 check as blocking (red) | BR-11 makes tier-3 advisory; the existing Figma component predates the rescope | Andrii Opryshko | implementation | Proposed — code renders tier-3 items as attention (amber), not blocked; update the Figma component after implementation |

## 15. Spec-ready gate

- [x] Current product behavior is backed by repository evidence.
- [x] Product decisions are confirmed or visibly marked Proposed.
- [x] Requirements have stable IDs.
- [x] Desktop/mobile, states, reuse, accessibility, and non-goals are covered.
- [x] No route, dependency, component, token, or API has been invented as fact.
- [x] Every requirement maps to a DOD row in section 16 below.

---

# Definition of Done

Completion means every applicable row is checked with evidence. An agent report, generated image,
or visual impression is not evidence by itself.

## 16. Done checks

| ID | Covers | Layer | Pass condition | Evidence | Done |
|---|---|---|---|---|---|
| DOD-01 | BR-07…BR-13, BR-18 | Specification | Scope, sources, Proposed items and non-goals are explicit and approved. | Sections 1–14 review | [ ] |
| DOD-02 | BR-07, BR-09, BR-15 | Design structure | Frames for 5 states × 1440/390 exist and are named by state; badge, stepper and poll panel are components with variants. | Figma node inspection: page `05 Lifecycle` (`?node-id=31-3`), 5 desktop + 3 mobile + dark + transition-states frames, 5 components (§13); mobile frames for the two read-only specimens are covered by the Compact stepper and the mobile vote/signing frames | [x] |
| DOD-03 | BR-16 | Tokens and reuse | Only existing tokens and primitives are used, except an approved `--success-subtle`; new pieces are the candidates in §9. | Figma variables + `git diff apps/web/src/index.css apps/web/src/components` | [x] see `first-screen/README.md` «verify-ui results» |
| DOD-04 | BR-06, BR-08…BR-14 | States | Empty, loading, each lifecycle state, pending transition, transition error and disabled-with-reason are reachable. | Browser: `/cases/c-102/protocol`, `c-101`, `c-103`, `c-104`; screenshots per state | [x] see `first-screen/README.md` «verify-ui results» |
| DOD-05 | BR-01…BR-05 | Preserved behavior | Editing in чернетка works as before: hint "Підставити" fills the value, `⌘ Enter` saves (origin → human, "Збережено щойно"), `Esc` restores, amounts unchanged. | Live browser scenario on `c-101` + `governance.test.ts` green | [x] see `first-screen/README.md` «verify-ui results» |
| DOD-06 | BR-08, BR-09, BR-10, BR-14 | Behavior | чернетка → готовий до голосування → чернетка → готовий до підпису works by clicks; badge, stepper and live region update; slots become read-only outside чернетка. | Live browser scenario + DOM check of `aria-current="step"` | [x] see `first-screen/README.md` «verify-ui results» |
| DOD-07 | BR-11, BR-12 | Behavior (gate) | With an unconfirmed `origin: ai` slot the "Позначити готовим до підпису" button is disabled with the slot named; after confirming it enables; a tier-3 warning alone does not block. | Unit test for `getReadiness` + browser check | [x] see `first-screen/README.md` «verify-ui results» |
| DOD-08 | BR-15 | Responsive | At 1440 and 390 every action is visible without horizontal scroll (`document.documentElement.scrollWidth <= innerWidth`). | Viewport screenshots + DOM eval | [x] see `first-screen/README.md` «verify-ui results» |
| DOD-09 | BR-17 | Accessibility | Every button has an accessible name, focus ring visible, stepper `ol` with `aria-current`, polite live region announces state; story a11y tests pass at `error`; contrast ≥ AA light and dark. | Keyboard run + `pnpm --filter @spilno/web test:stories` | [x] see `first-screen/README.md` «verify-ui results» |
| DOD-10 | BR-03, BR-11, BR-18 | Deterministic gates | `pnpm --filter @spilno/web lint`, `typecheck`, `test`, `test:stories`, `build` pass. | Command output | [x] see `first-screen/README.md` «verify-ui results» |
| DOD-11 | BR-16, BR-18 | Scope | Diff touches only `apps/web/src/features/governance/*`, `apps/web/src/components/*`, `apps/web/src/i18n/uk.ts`, `apps/web/src/index.css` (token) and stories; no `services/`, no `openapi.yaml`, no new dependency. | `git diff --stat` + `git diff package.json` | [x] see `first-screen/README.md` «verify-ui results» |
| DOD-12 | BR-01…BR-17 | Live evidence | `verify-ui` gives PASS/FAIL with expected vs actual for DOD-04…DOD-09, screenshots saved in `docs/design/first-screen/screenshots/`. | `verify-ui` report + `docs/design/first-screen/README.md` «verify-ui results» | [x] see `first-screen/README.md` «verify-ui results» |

## 17. Requirement coverage

| Requirement | Covered by DoD IDs | Gap |
|---|---|---|
| BR-01 | DOD-05, DOD-12 | none |
| BR-02 | DOD-05, DOD-12 | none |
| BR-03 | DOD-05, DOD-10, DOD-12 | none |
| BR-04 | DOD-05, DOD-12 | none |
| BR-05 | DOD-05, DOD-12 | none |
| BR-06 | DOD-04, DOD-12 | none |
| BR-07 | DOD-01, DOD-02, DOD-12 | none |
| BR-08 | DOD-01, DOD-04, DOD-06, DOD-12 | none |
| BR-09 | DOD-01, DOD-02, DOD-04, DOD-06, DOD-12 | none |
| BR-10 | DOD-01, DOD-04, DOD-06, DOD-12 | none |
| BR-11 | DOD-01, DOD-04, DOD-07, DOD-10, DOD-12 | none |
| BR-12 | DOD-01, DOD-04, DOD-07, DOD-12 | none |
| BR-13 | DOD-01, DOD-04, DOD-12 | none |
| BR-14 | DOD-04, DOD-06, DOD-12 | none |
| BR-15 | DOD-02, DOD-08, DOD-12 | none |
| BR-16 | DOD-03, DOD-11 | none |
| BR-17 | DOD-09, DOD-12 | none |
| BR-18 | DOD-01, DOD-10, DOD-11 | none |

## 18. Final gate

- [x] Every requirement is covered by at least one DoD row.
- [x] Every applicable DoD row is checked and links to evidence.
- [ ] Every `Proposed` decision (section 14) was approved or removed before implementation. — open: the Figma `ReadinessCard` repaint (tier-3 amber); code already follows BR-11.
- [x] Deterministic checks and live browser verification both passed.
- [x] The final repository diff matches the approved scope.
