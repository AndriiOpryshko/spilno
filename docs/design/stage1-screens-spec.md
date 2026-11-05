---
feature: board-protocol
status: Draft
owner: Andrii Opryshko (product owner, frontend)
updated_at: 2026-11-05
---

# Design spec: board-protocol — the remaining Stage 1 screens

Second design spec of the module. The first ([`design-spec.md`](design-spec.md)) covered the protocol
lifecycle on the draft screen and listed the rest as non-goals (§12 there). This one covers the rest of
Stage 1, design-only: frames and components in Figma now, implementation later through tasks T34–T43.

## 1. Raw request

> поки в мене є підписка на фігму, хотів би домалювати всі екрани що можна.
>
> (2026-11-05, on the Sidebar question) доречі ще треба типу бекофічу щоб можна було заводити самі осбб

## 2. Why now

- **Problem:** Stage 1 has 13 user stories, but Figma holds only the cases list and the protocol draft
  with its lifecycle (pages `03`, `05`). Sign-in, settings, roster, meetings, the case with documents,
  the conversation with the помічник протоколу, own poll vote, printing, withdrawal, the signed-copy form
  and the реєстр протоколів have no design, and T34–T43 would invent them in code.
- **User:** член правління (CONTEXT.md, PRD.md §10); for the back office, the operator who sets up an ОСББ
  (today a CLI step, T08 and T12).
- **Trigger:** the Figma subscription is active now; tasks T34–T43 are `todo` and need a design to follow.
- **Success outcome:** every Stage 1 UI task has frames at 1440 and 390 and components with state variants
  to build from, each traced to a PRD AC and an API operation.

## 3. Evidence sources

| Status | Source | What it proves |
|---|---|---|
| Confirmed | `docs/features/board-protocol/PRD.md` (Approved 2026-10-25) US-01…US-13, AC list, §3 non-goals | Stage 1 scope and the AC numbers used below. `PRD.ao-prd.md` is a parallel draft with other numbering; SAD and tasks cite `PRD.md` |
| Confirmed | `docs/features/board-protocol/sad.md` §4 (UI architecture, polling ~2 s), §5 (`routes/`: sign-in · settings (association, roster) · cases · protocols/$id · registry), §6 Flows 1–9, 12, 13 | Surfaces, user-facing branches, error codes per flow |
| Confirmed | `services/core/api/openapi.yaml` | Operations and fields: `AssociationSettingsInput` (meeting_form and signers are free text ≤ 500, register_start_number ≥ 1), `BoardMemberInput` (full_name, sign_in_email, term_start, term_end nullable), `NewMeeting` (title, kind REGULAR/EXTRAORDINARY, held_at, ends_at, location, quorum_required, chair), `CaseDocumentUpload` (file, previous_version_id, document_date), `AssistantTurnInput` (message, case_id, case_document_ids, files), `Job` (pending/running/completed/failed, attempt/max_attempts), `PollChoice` (for/against/abstain), `MeetingFactsInput`, `WithdrawalInput` (reason), registry filters (counterparty, subject, amount exact or from–to, date_from–date_to, cursor) |
| Confirmed | Contract error codes | `identity.session_required`, `identity.not_current_member`, `governance.settings_missing`, `settings_invalid`, `register_start_locked`, `roster_entry_invalid`, `file_invalid`, `storage_unavailable`, `draft_revision_conflict`, `slot_unconfirmed`, `draft_incomplete`, `not_ready_for_signing`, `poll_not_open`, `vote_on_behalf`, `meeting_facts_missing`, `correction_note_required`, `reason_required`, `transition_not_allowed`, `protocol_locked`, `search_invalid` |
| Confirmed | `docs/features/board-protocol/tasks/T34…T43` | UI task boundaries and `files_hint` routes: `routes/sign-in.tsx`, `settings.association.tsx`, `settings.roster.tsx`, `cases.$caseId.index.tsx`, `protocols.$protocolId.tsx`, `registry.index.tsx`, `registry.$protocolId.tsx`; components `assistant-panel`, `decision-text-editor`, `withdraw-dialog`, `signed-copy-form`, `registry/` |
| Confirmed | ADR-0011, ADR-0017, ADR-0018 | Text = fragments + slot chips; number reserved at first print and never reused; meeting facts entered with the scan, lock refused without them |
| Confirmed | T08, T12, ADR-0015, PRD §3, idea-brief §5 | Today the association row is seeded, the first член правління is an ops command (`api bootstrap-member`), Google consent is an operator CLI (`cmd/google-connect`); onboarding other associations is a PRD non-goal |
| Observed | `apps/web/src/routes/` | Live routes: `/`, `/cases`, `/cases/$caseId/protocol` |
| Observed | `apps/web/src/components/app-shell.tsx`, `i18n/uk.ts` `nav` | Sidebar: Справи (link) + Протоколи, Збори, Документи as disabled «Незабаром» |
| Observed | `apps/web/src/components/ui/` | Primitives: badge, button, card, input, kbd, skeleton, textarea; no dialog, select, checkbox, tabs, table, toast |
| Observed | `docs/design/design-system.md` (tokens, a11y, Gaps) | Tokens light/dark incl. `success-subtle`; the Gaps list this spec closes |
| Observed | Figma `D06IijJnLn8WEgU85fnD84`: pages `02 Components`, `03 New screen`, `05 Lifecycle` | Components: Button, StatusGlyph, OriginTag, Kbd, InlineHint, LimitNote, SlotRow, ReadinessCard, CaseRow, CaseCard, Sidebar, ProtocolStateBadge, LifecycleStepper, PollRoundPanel, InlineAlert, TransitionActions; variables Color (Light/Dark), Spacing, Radius; font Onest |
| Confirmed | User answers, 2026-11-05 | The first draft is requested from the case **and** from the meeting; Sidebar = Справи · Збори · Реєстр протоколів · Налаштування; 1440 and 390 for every screen; new parts are components with state variants; add a back office to set up an ОСББ |
| Proposed | §6 BR-20…BR-22, §14 | Back office, the Sidebar change, the poll label correction |

## 4. Current user flow

1. **Start:** a visitor opens `http://localhost:5180/`; there is no sign-in, the health card shows.
2. **Action:** opens `/cases`, picks a case, lands on `/cases/$caseId/protocol` (in-memory data).
3. **System response:** the draft screen with slots, readiness and lifecycle actions (design-spec.md §5).
4. **Success:** the draft reaches готовий до підпису and `window.print()` prints the screen.
5. **Recovery:** missing draft → `role="alert"` with «До справ». No settings, roster, meetings, documents,
   conversation, own vote, signed copy or registry exist in the UI; the association and the first member
   exist only through migrations and ops commands.

## 5. Target user flow

1. **Start:** a visitor opens any page → `/sign-in` («Увійти через Google»). A person who is not a current
   член правління gets the no-access screen that reveals nothing (AC-19).
2. **First run:** asking for a draft before the settings are filled answers `governance.settings_missing`
   naming every missing setting with a link → `/settings/association` (AC-42). The roster lives at
   `/settings/roster` (AC-02).
3. **Prepare:** the member creates a meeting at `/meetings` (date, kind, chair) and a справа at `/cases`,
   uploads documents to the case, marking a revised one as a new version (AC-09).
4. **Ask the помічник:** from the case («Попросити помічника» → pick the meeting) or from the meeting
   («Попросити помічника» → pick the справи and documents) the member writes what to decide. The turn starts
   a job; the panel polls about every 2 s; on success the protocol opens with unconfirmed slots and the
   conversation in a side panel (AC-26, AC-03, AC-03c). A second справа for the same meeting becomes another
   agenda item of the same draft.
5. **Review:** slots show their fragment, alternative values (AC-22), a new value beside a confirmed one
   (AC-28b) and «потребує перевірки» after a new document version (AC-09); text is edited with slot chips
   (AC-29, AC-28); a stale save answers `draft_revision_conflict` with a reload notice.
6. **Poll:** each member casts and changes only their own vote per item; the tally lists members and times
   under the AC-41 label; no verdict (AC-30, AC-31, AC-05e, AC-32).
7. **Print:** «Позначити готовим до підпису» lists every blocker at once (AC-04, AC-21); otherwise a render
   job runs and offers the PDF with its number (AC-14c, AC-33). Return to draft warns that printed copies
   must not be signed (AC-14b).
8. **Lock:** in готовий до підпису the member uploads the scan and enters the meeting facts; every missing
   fact is named at once; corrections need a note; the form is never pre-filled from the poll (AC-14, AC-34,
   AC-15, AC-40, AC-41).
9. **Withdraw:** a dialog with a required reason (AC-35…AC-37); a withdrawn draft offers «Нова чернетка з
   цієї» (AC-39).
10. **Find:** `/registry` searches by counterparty, subject, amount or range and period; withdrawn numbers
    show as gaps; a decision opens with scan, facts, cited documents, correction note and informal poll
    history; changing it is only possible through a new decision (AC-18, AC-17, AC-38).
11. **Back office (Proposed):** an operator creates an ОСББ, connects its Google Drive and adds its first
    член правління (§14).

## 6. Requirements

| ID | Status | Requirement | Source |
|---|---|---|---|
| BR-01 | Confirmed | `/sign-in`: brand, one «Увійти через Google» action (startSignIn), states default, session expired (`identity.session_required`), callback refused for an unknown, expired or replayed state (completeSignIn 401), after sign-out (signOut 204) | T34, ADR-0003, Flow 7, contract `/identity/*` |
| BR-02 | Confirmed | No-access screen for `identity.not_current_member`: the same body whatever was requested, no record names; shown before any data loads. Two cases: at the callback no session is created, so the action is «Увійти іншим Google-акаунтом»; with a session whose term has ended, the action is «Вийти» | AC-19, AC-16, T34, contract `completeSignIn` 403, `getSession` 403 |
| BR-03 | Confirmed | `/settings/association`: text fields «Форма засідання за статутом», «Хто підписує протокол» (≤ 500 chars), number «Номер, з якого продовжується реєстр» (≥ 1); states empty (first run, banner naming the missing settings), saved (author and time), `settings_invalid` (every field at once), `register_start_locked` (number read-only with the reason, other fields editable) | AC-42, Flow 12, T35 |
| BR-04 | Confirmed | `/settings/roster`: members with full name, sign-in email, term start–end, «чинний» or «повноваження завершено»; add / edit form; `roster_entry_invalid` names every broken field at once (term start missing, end before start); empty state | AC-02, Flow 12, T35 |
| BR-05 | Confirmed | `/meetings` list (date, title, kind, protocol state) and a meeting page: details, agenda items from справи, the protocol draft link and state, «Попросити помічника»; new-meeting form (title, kind, date and time, end, location, quorum required, головуючий from the roster) | User answer 2026-11-05, `NewMeeting`, AC-03 |
| BR-06 | Confirmed | `/cases/$caseId`: title, documents with document date, uploader, version chain (newer version marked, older «замінено новою версією»), upload with «нова версія документа…» and document date; client hint «PDF, JPEG або PNG до 20 МБ»; `file_invalid` inline, `storage_unavailable` with retry; «Попросити помічника» with a meeting picker; the draft link when the case is already an agenda item | AC-09, AC-03c, T36, Flow 8 |
| BR-07 | Confirmed | Conversation panel of the помічник протоколу on the protocol screen (side panel at 1440, a section or sheet at 390): turns by member and помічник with time, attached documents, composer with document picker; job states pending / running («Помічник готує чернетку…»), failed with «Спробувати ще раз» (retryJob, attempt N з M), refusal `settings_missing` with the link | AC-26, AC-42, T39, Flows 1, 8 |
| BR-08 | Confirmed | Slot additions on the protocol screen: alternative values with their documents beside the recommended one (AC-22); a new value from the conversation beside the confirmed one, «потребує перевірки» (AC-28b); after a newer document version: confirmed value vs value in the current document (AC-09); головуючий next to the meeting date, «не вказано» when null | T37 |
| BR-09 | Confirmed | Decision text editor: plain text with inline slot chips; a typed number is highlighted with «Зробити слотом» and does not block saving | AC-28, AC-29, ADR-0011, T38 |
| BR-10 | Confirmed / Proposed | Revision conflict: a 409 `draft_revision_conflict` reloads the draft and shows a notice naming that another member changed it (Confirmed); the member's unsaved text stays on screen with «Скопіювати в поле» (Proposed — the flow says only «reload») | Flow 1, T37 |
| BR-11 | Confirmed | Own poll vote: per agenda item a за / проти / утримався choice for the signed-in member only, changeable while the round is open; the tally per item lists each member, choice and time; the label is exactly «Попереднє внутрішнє опитування правління — не є голосуванням засідання і не замінює його»; no verdict on quorum or majority; outside готовий до голосування the choice is absent | AC-30, AC-31, AC-41, AC-05e, AC-32, T40 |
| BR-12 | Confirmed | «Позначити готовим до підпису» with blockers lists every blocker at once (no agenda item, item without subject, unconfirmed proposed slot); otherwise a pending render job, then «Завантажити PDF · № N» | AC-04, AC-21, AC-14c, AC-33, T41 |
| BR-13 | Confirmed | Return to draft from готовий до підпису is a confirm dialog: «Надруковані раніше примірники не можна підписувати. Номер № N залишиться за протоколом.»; the voided copy stays in history marked «недійсний» | AC-14b, ADR-0017, T41 |
| BR-14 | Confirmed | Withdraw dialog from чернетка, готовий до голосування or готовий до підпису: required reason (`reason_required`), consequence text per state (no number taken / number № N stays and is never reused); a withdrawn draft shows the reason and «Нова чернетка з цієї» | AC-35…AC-37, AC-39, Flow 9, T41 |
| BR-15 | Confirmed | Signed-copy form, only in готовий до підпису: scan drop zone (PDF, JPEG, PNG ≤ 20 MB; uploading, uploaded, `file_invalid`, `storage_unavailable`), present members (roster in force), за / проти / утримався counts per agenda item, «На папері є рукописні виправлення» with a required note; `meeting_facts_missing` names every missing fact at once; never pre-filled from the poll; outside готовий до підпису the form is absent and `not_ready_for_signing` explains why | AC-14, AC-34, AC-15, AC-40, AC-41, ADR-0018, T42 |
| BR-16 | Confirmed | Locked protocol: no edit, poll, withdraw or upload action; refusals render `protocol_locked` as «Підписаний протокол змінюється лише новим рішенням правління, що на нього посилається» with that offer | AC-17, AC-38, T41 |
| BR-17 | Confirmed | `/registry`: filters counterparty, subject, amount (exact or from–to), period; results with number, date, subject, counterparty, amount; withdrawn numbers as gap rows «№ N · знято з розгляду»; «Показати ще» (cursor); empty «Нічого не знайдено» with filters kept; `search_invalid` marks every broken filter | AC-18, Flow 13, T43 |
| BR-18 | Confirmed | `/registry/$protocolId`: number, date, meeting; frozen text with the correction note beside it; signed scan download; meeting facts with who entered them and when; cited documents with versions; poll history under the AC-41 label; the locked explanation of BR-16 | AC-18, AC-15, AC-41, AC-17, T43 |
| BR-19 | Observed | Shell: Sidebar + TopBar crumbs as in `app-shell.tsx`; tokens and Onest from the design system; one `h1` per screen; Button variants as in `02 Components` | `app-shell.tsx`, design-system.md |
| BR-20 | Proposed | Sidebar nav = Справи · Збори · Реєстр протоколів · Налаштування; «Документи» removed (no Stage 1 screen); the current member and «Вийти» at the bottom | User answer 2026-11-05 |
| BR-21 | Proposed | Back office (operator, outside the board UI): ОСББ list; «Нове ОСББ» (name, address); Google Drive connection state with «Підключити Google Drive» (consent, creates the `spilno/` folder); first член правління (full name, sign-in email, term start); refusal when the roster already has a member | User answer 2026-11-05; replaces T08 `bootstrap-member` and T12 `cmd/google-connect`; conflicts with PRD §3 (§14) |
| BR-22 | Proposed | The existing Figma `PollRoundPanel` label «Неформальне опитування — не голосування засідання» is replaced with the exact AC-41 text | AC-41 vs page `05 Lifecycle` |
| BR-23 | Confirmed | Every screen has 1440 and 390 frames; at 390 the Sidebar collapses to a top bar with a menu button, tables become cards, dialogs become full-width bottom sheets, side panels move below the main content | User answer 2026-11-05, design-system.md breakpoints |
| BR-24 | Confirmed | New parts are Figma components with state variants on page `06 Stage 1 · components`: Field, Checkbox, VoteChoice, Dialog, Toast, Tabs, FileDrop, DocumentRow, ChatMessage, RegistryRow, MemberRow, MeetingRow | User answer 2026-11-05 |
| BR-25 | Observed | Accessibility as in design-system.md: AA contrast light and dark, visible focus, `aria-invalid` fields with the message under the field, errors with `role="alert"`, dialogs with a title and focus trap, status never color-only | design-system.md |

## 7. Routes and frames

| Surface | Route / frame | Required content and layout |
|---|---|---|
| Current | `/cases`, `/cases/$caseId/protocol` | Figma pages `03`, `05` (unchanged here) |
| Sign-in | `/sign-in` · 1440, 390 | Centered card, brand, Google action; specimen: session expired |
| No access | any route · 1440, 390 | Same card layout, message, «Вийти» |
| Settings | `/settings/association` · 1440, 390 | Sidebar · tabs Об'єднання / Склад правління · form card; specimens: first run, error, locked number |
| Roster | `/settings/roster` · 1440, 390 | Table of members (cards at 390), add/edit dialog; specimens: empty, invalid entry |
| Meetings | `/meetings`, `/meetings/$meetingId` · 1440, 390 | List; meeting page with agenda and protocol; new-meeting dialog |
| Case | `/cases/$caseId` · 1440, 390 | Documents with versions, upload, «Попросити помічника» dialog with meeting picker; specimens: upload error, storage down |
| Protocol + assistant | `/protocols/$protocolId` · 1440, 390 | Existing draft screen + conversation panel; specimens: job pending, job failed, settings missing, conflict, alternative values, text editor |
| Own vote | `/protocols/$protocolId` (готовий до голосування) · 1440, 390 | Poll panel with own choice per item and the tally |
| Print / return / withdraw | dialogs on `/protocols/$protocolId` · 1440, 390 | Blockers, rendering, PDF ready; return confirm; withdraw with reason; withdrawn → copy |
| Signed copy | `/protocols/$protocolId` (готовий до підпису) · 1440, 390 | Form; specimens: missing facts, corrections, file error |
| Registry | `/registry`, `/registry/$protocolId` · 1440, 390 | Search with results and gaps; empty; decision view |
| Back office (Proposed) | route not decided (§14) · 1440, 390 | ОСББ list, new ОСББ, Google connection, first member |

## 8. State inventory

| State | Trigger | Visible result | Recovery / next action |
|---|---|---|---|
| Empty | No settings / roster / meetings / documents / registry results | Empty-state text with the one next action | Fill / add / change filters |
| Loading | First load, page change | Skeleton (`aria-busy`) | Wait |
| Pending job | Assistant turn, print render | Message «Помічник готує чернетку…» / «Готуємо друкований примірник…», actions disabled, polling ~2 s | Wait |
| Success | Saved, voted, rendered, locked | Toast or inline «Збережено», new state | Continue |
| Validation error | `settings_invalid`, `roster_entry_invalid`, `meeting_facts_missing`, `correction_note_required`, `reason_required`, `search_invalid`, `file_invalid` | Every broken field marked at once, message under each, summary alert on top | Fix and resubmit |
| Conflict | `draft_revision_conflict`, `register_start_locked` | Notice, data reloaded / field read-only with reason | Re-apply / keep |
| Refused transition | `transition_not_allowed`, `poll_not_open`, `not_ready_for_signing`, `protocol_locked`, `vote_on_behalf` | InlineAlert with the reason and the allowed next action | Follow the offer |
| Service down | `storage_unavailable`, failed job | Alert «нічого не збережено» + «Спробувати ще раз» | Retry |
| No access | `identity.session_required`, `identity.not_current_member` | Sign-in / no-access screen | Sign in / sign out |
| Disabled / focus / hover | Blockers; keyboard | Disabled with visible reason; focus ring; hover per Button variants | — |

## 9. Reuse and handoff contract

| Design responsibility | Current code owner / token | Decision |
|---|---|---|
| Shell, crumbs | `components/app-shell.tsx` `Sidebar`, `TopBar`; Figma `Sidebar` | reuse; nav items per BR-20 (Proposed) |
| Buttons, badges, cards, inputs | `components/ui/button`, `badge`, `card`, `input`, `textarea`, `skeleton`, `kbd` | reuse |
| Protocol screen parts | `features/governance/*` and Figma `SlotRow`, `ReadinessCard`, `ProtocolStateBadge`, `LifecycleStepper`, `PollRoundPanel`, `TransitionActions`, `InlineAlert`, `OriginTag`, `StatusGlyph`, `LimitNote` | reuse; `PollRoundPanel` gains own-vote and per-member rows (BR-11, BR-22) |
| Field (label, control, help, error) | `input`, `textarea` | new component candidate (Figma `Field`) |
| Checkbox, VoteChoice | — (`radix-ui` installed, no wrapper) | new component candidates |
| Dialog / sheet | — | new component candidate `Dialog` |
| Toast | — | new component candidate `Toast` |
| Tabs | — | new component candidate `Tabs` |
| File drop zone | — | new component candidate `FileDrop` |
| Document row with versions | — | new component candidate `DocumentRow` (`features/governance/case-documents.tsx` per T36) |
| Conversation message | — | new component candidate `ChatMessage` (`assistant-panel.tsx` per T39) |
| Registry, member, meeting rows | `CaseRow` pattern | new component candidates `RegistryRow`, `MemberRow`, `MeetingRow` |
| Signed copy form, withdraw dialog, text editor | — | compositions per T42, T41, T38 |
| Copy | `i18n/uk.ts` | new keys per screen; error texts from PRD §5.1 / contract |
| Tokens | Figma Color / Spacing / Radius | reuse; no new token |

## 10. Responsive and accessibility contract

- **Responsive:** 1440 = Sidebar 240 + content (1fr + 320 side panel where a panel exists); 390 = top bar
  with menu, one column, 16 px gutters, full-width buttons, tables as cards, dialogs as bottom sheets, the
  conversation panel below the slots; no horizontal scroll.
- **Keyboard:** tab order follows reading order; dialogs trap focus and return it to the opener; Esc closes a
  dialog without saving; the vote choice is a radio group per item (arrow keys).
- **Semantics:** one `h1`; forms with `<label>`; errors `aria-invalid` + text, summary `role="alert"`; job
  progress in a polite live region; tables with headers; the AC-41 label is text, not an image.
- **Contrast and motion:** AA in light and dark; no new animation beyond the skeleton pulse.

## 11. Content and realistic data

- ОСББ «Січових Стрільців, 12»; members О. Коваленко (голова), І. Мельник, Т. Шевчук, Н. Бондар
  (fictional, as on page `05`); emails `user-<n>@example.test`.
- Settings: «Засідання правління за особистої присутності»; «Голова правління і секретар засідання»;
  register start 12.
- Meeting «Засідання правління 21.10.2026», чергове, головуючий О. Коваленко.
- Case «Гідроізоляція покрівлі, під’їзд 2»: Договір № 14/10 від 14.10.2026, Рахунок № 233 від 16.10.2026,
  Кошторис v2 від 18.10.2026 (нова версія); 184 500,00 грн; ТОВ «Дах-Сервіс»; Фонд ремонту.
- Poll round 2: 3 з 4 проголосували.
- Withdraw reasons from AC-35…AC-37. Correction note from the contract example.
- Registry: № 9–17/2026, № 12 знято з розгляду.

## 12. Non-goals and forbidden changes

- No Stage 2 surfaces: blocking checks and overrides, charter rules with effective dates, templates,
  попередження, related decisions, export for the ревізійна комісія, reading the scan.
- No change to the screens on pages `03` and `05` except BR-22 and the instances they share with BR-20.
- No code, contract, migration or task change in this iteration; the back office is drawn as a proposal
  only.
- No new design tokens; no new font.

## 13. Design references

- Figma file `D06IijJnLn8WEgU85fnD84`; existing pages `02 Components`, `03 New screen`, `05 Lifecycle`.
- Repo-local design file: N/A — Figma is the design tool for this project.
- Drawn through the Figma MCP on 2026-11-05. Links are `https://www.figma.com/design/D06IijJnLn8WEgU85fnD84?node-id=<id>`.

Page `06 Stage 1 · components` (`?node-id=35-2`):

| Component | Node | Variants | Covers |
|---|---|---|---|
| `Field` | `36:95` | Kind Input / Textarea / Date × State Default / Filled / Error / Disabled / Read-only; props Label, Value, Help | BR-03, BR-04, BR-15, BR-24 |
| `Checkbox` | `36:112` | Checked × State Default / Disabled / Error | BR-15 |
| `VoteChoice` | `36:154` | Selected None / For / Against / Abstain / Disabled | BR-11 |
| `Toast` | `36:175` | Success / Error / Info | §8 |
| `Tabs` | `36:197` | Active First / Second / Third | BR-03, BR-04 |
| `Dialog` | `37:70` | Tone Default / Destructive × Layout Desktop / Mobile (sheet); props Title, Body, Show field | BR-13, BR-14 |
| `FileDrop` | `37:108` | Empty / Uploading / Uploaded / Error | BR-06, BR-15 |
| `DocumentRow` | `37:152` | Current / New version / Superseded | BR-06, BR-18 |
| `ChatMessage` | `37:190` | Member / Assistant / Pending / Failed | BR-07 |
| `RegistryRow` | `38:49` | Decision / Gap × Desktop / Mobile | BR-17 |
| `MemberRow` | `38:82` | Current / Ended × Desktop / Mobile | BR-04 |
| `MeetingRow` | `38:117` | None / Draft / Signed protocol × Desktop / Mobile | BR-05 |
| `Sidebar` (page `02`, `20:433`) | changed | nav Справи · Збори · Реєстр протоколів · Налаштування | BR-20 (Proposed) |

Page `07 Stage 1 · screens` (`?node-id=35-3`), 84 frames; desktop / mobile node ids:

| Surface · state | Desktop | Mobile | Covers |
|---|---|---|---|
| Вхід | `39:3` | `39:62` | BR-01 |
| Вхід · сесія завершилась | `39:21` | `39:80` | BR-01 |
| Немає доступу (сесія є, повноваження скінчились) | `39:48` | `39:105` | BR-02 |
| Вхід · посилання застаріло (callback 401) | `59:6820` | `59:6888` | BR-01 |
| Вхід · акаунт не в складі правління (callback 403, сесії немає) | `59:6845` | `59:6913` | BR-02 |
| Вхід · після виходу (signOut) | `59:6863` | `59:6931` | BR-01 |
| Меню акаунта з «Вийти» / меню навігації | `59:6956` | `59:6988` | BR-20 |
| Налаштування об'єднання · перший запуск / помилки / номер заблоковано | `40:19` / `40:92` / `40:165` | `40:239` / `40:283` / `40:327` | BR-03 |
| Склад правління · список / додати з помилками | `41:221` / `41:316` | `41:455` / `41:512` | BR-04 |
| Збори · список / нове засідання | `42:487` / `42:746` | `42:860` / `42:1044` | BR-05 |
| Засідання · попросити помічника / помічник готує | `42:553` / `42:656` | `42:901` / `42:979` | BR-05, BR-07 |
| Справа · документи / файл не прийнято / сховище недоступне / вибір засідання | `43:796` / `43:950` / `43:1105` / `43:1276` | `43:1455` / `43:1582` / `43:1710` / `43:1847` | BR-06 |
| Протокол · помічник · розмова / збій | `43:3356` / `43:3399` | `43:3528` / `43:3570` | BR-07 |
| Протокол · конфлікт версій | `43:3442` | `43:3612` | BR-10 |
| Протокол · альтернативні значення, нове поруч із підтвердженим, текст зі слотами | `43:3485` | `43:3654` | BR-08, BR-09 |
| Протокол · власний голос | `43:3696` | `43:3740` | BR-11 |
| Протокол · блокери друку / готується / PDF готовий | `43:3780` / `43:3823` / `43:3866` | `43:4036` / `53:6273` / `43:4078` | BR-12 |
| Протокол · повернути в чернетку | `43:3909` | `43:4120` | BR-13 |
| Протокол · зняти з розгляду / знятий, нова чернетка з цієї | `43:3952` / `43:3995` | `43:4162` / `53:6358` | BR-14 |
| Протокол · підписаний примірник · форма / помилки / виправлення | `43:4204` / `43:4247` / `43:4290` | `43:4333` / `43:4375` / — | BR-15 |
| Протокол · підписаний, зміни заборонено | `43:4417` | `53:6641` | BR-16 |
| Реєстр · результати / нічого не знайдено / помилки фільтрів | `49:5844` / `49:5953` / `49:6036` | `49:6232` / `49:6317` / `49:6382` | BR-17 |
| Реєстр · рішення № 17/2026 | `49:6119` | `49:6447` | BR-18 |
| Бекофіс (section `51:6511` «Proposed — outside PRD Stage 1») · список / Drive не підключено / Drive підключено / перший член уже є | `51:6223` / `51:6256` / `51:6299` / `51:6340` | `51:6399` / `51:6429` / `51:6465` | BR-21 |

Protocol frames are clones of the approved page `05` frames with the new parts added; in their draft
copies the title reads «Чернетка · …» because a draft has no number until the first print (ADR-0017).

## 14. Open decisions

| Decision | Why it matters | Owner | Due | Status |
|---|---|---|---|---|
| Back office for setting up an ОСББ | PRD §3 and idea-brief §5 reject onboarding other associations in the first release; ADR-0015 serves one association and defers RLS; T08 and T12 make setup an operator CLI. A UI needs a PRD amendment, an operator role, operations in the contract and possibly an ADR | Andrii Opryshko | before implementation | Proposed — drawn as exploration only |
| Sidebar nav (BR-20) | Changes the shared Figma `Sidebar` and `app-shell.tsx`; «Документи» disappears | Andrii Opryshko | in Figma | Proposed per user answer |
| Poll label text (BR-22) | Page `05` shows a shorter label than AC-41 requires verbatim | Andrii Opryshko | in Figma | Proposed |
| Meeting screens are not in T34–T43 | Starting from the meeting needs a task for `/meetings` | Andrii Opryshko | before implementation | Proposed — add a UI task |
| Settings as tabs (Об'єднання / Склад правління) vs two Sidebar items | One nav item keeps the Sidebar short | Andrii Opryshko | in Figma | Proposed — tabs |
| Sidebar component already changed in Figma | The shared `Sidebar` (`20:433`) now shows the BR-20 nav, so pages `03` and `05` show it too; undo = rename back and restore «Документи» | Andrii Opryshko | review | Applied in Figma, pending approval |
| Keep the unsaved text on a revision conflict (BR-10) | Needs the client to hold the unsent edit across the reload | Andrii Opryshko | before T37 | Proposed |
| Page `05` draft frames show «№ 17/2026» | ADR-0017 reserves the number at the first print; a draft has none. Page `07` copies are corrected, page `05` is not | Andrii Opryshko | in Figma | Proposed — fix page `05` |
| `ChatMessage`, `Toast`, `DocumentRow` variants share one text property | The component sheet shows the same text in every variant; instances override it | Andrii Opryshko | in Figma | Accepted as is |

## 15. Spec-ready gate

- [x] Current product behavior is backed by repository evidence.
- [x] Product decisions are confirmed or visibly marked Proposed.
- [x] Requirements have stable IDs.
- [x] Desktop/mobile, states, reuse, accessibility, and non-goals are covered.
- [x] No route, dependency, component, token, or API has been invented as fact (routes from SAD §5 and
      task `files_hint`; meetings and back-office routes marked Proposed).
- [x] Every requirement maps to a DOD row in section 16 below.

---

# Definition of Done

Completion means every applicable row is checked with evidence. An agent report, generated image,
or visual impression is not evidence by itself. This iteration is design-only: DOD-01…DOD-07 close it;
DOD-08…DOD-10 close when T34–T43 are implemented.

## 16. Done checks

| ID | Covers | Layer | Pass condition | Evidence | Done |
|---|---|---|---|---|---|
| DOD-01 | BR-20…BR-22, §14 | Specification | Proposed items are approved, changed or dropped by the product owner | §14 statuses updated | [ ] |
| DOD-02 | BR-24, BR-25 | Components | Page `06 Stage 1 · components` holds the 12 components of §9 with state variants (default, error, disabled, focus where interactive) bound to the Color, Spacing and Radius variables | Figma node inspection: component sets, variant names, bound variables | [ ] 12 sets with state variants and bound variables (§13); no separate focus variants yet |
| DOD-03 | BR-01…BR-18, BR-23 | Design structure | Page `07 Stage 1 · screens` holds a 1440 and a 390 frame for every surface of §7, named `<surface> · <state> · desktop|mobile`, built from component instances | Figma node inspection; node ids in §13 | [x] 84 frames; one state without a mobile frame (signed copy with corrections) |
| DOD-04 | BR-03, BR-04, BR-06, BR-07, BR-10…BR-15, BR-17 | States | Each state of §8 that applies to a surface has a specimen frame: empty, pending job, validation error with every field marked, conflict, refused transition, service down, no access | Frame list against §8 | [x] see §13 |
| DOD-05 | BR-11, BR-15, BR-18, BR-22 | Content | The AC-41 label appears verbatim on every poll view and on the decision view; the signed-copy form shows no poll numbers | Figma text search for the AC-41 string | [ ] verbatim on page `07` (4 frames); the `PollRoundPanel` component on page `05` still has the short label (BR-22) |
| DOD-06 | BR-19, BR-25 | Tokens and a11y | No hard-coded colors on new nodes; text on its background ≥ 4.5:1 in light and dark; one dark-mode frame per new component set | Figma variable inspection + contrast check | [ ] |
| DOD-07 | BR-21 | Proposal | Back-office frames sit in a section labelled «Proposed — outside PRD Stage 1» | Figma section name | [x] section `51:6511` |
| DOD-08 | BR-01…BR-18 | Implementation | T34–T43 build against these frames; stories mirror the Figma variants and pass `test:stories` | `pnpm --filter @spilno/web test:stories` | [ ] |
| DOD-09 | BR-23, BR-25 | Live evidence | `verify-ui` at 1440 and 390 per route: no horizontal scroll, focus visible, errors announced | `verify-ui` report + screenshots | [ ] |
| DOD-10 | BR-01…BR-18 | Deterministic gates | `mise run lint`, `mise run test`, `mise run e2e` green | Command output | [ ] |

## 17. Requirement coverage

| Requirement | Covered by DoD IDs | Gap |
|---|---|---|
| BR-01 | DOD-03, DOD-08, DOD-10 | none |
| BR-02 | DOD-03, DOD-08, DOD-10 | none |
| BR-03 | DOD-03, DOD-04, DOD-08 | none |
| BR-04 | DOD-03, DOD-04, DOD-08 | none |
| BR-05 | DOD-03, DOD-08 | needs a UI task (§14) |
| BR-06 | DOD-03, DOD-04, DOD-08 | none |
| BR-07 | DOD-03, DOD-04, DOD-08 | none |
| BR-08 | DOD-03, DOD-08 | none |
| BR-09 | DOD-03, DOD-08 | none |
| BR-10 | DOD-03, DOD-04, DOD-08 | none |
| BR-11 | DOD-03, DOD-04, DOD-05, DOD-08 | none |
| BR-12 | DOD-03, DOD-04, DOD-08 | none |
| BR-13 | DOD-03, DOD-04, DOD-08 | none |
| BR-14 | DOD-03, DOD-04, DOD-08 | none |
| BR-15 | DOD-03, DOD-04, DOD-05, DOD-08 | none |
| BR-16 | DOD-03, DOD-08 | none |
| BR-17 | DOD-03, DOD-04, DOD-08 | none |
| BR-18 | DOD-03, DOD-05, DOD-08 | none |
| BR-19 | DOD-06 | none |
| BR-20 | DOD-01 | none |
| BR-21 | DOD-01, DOD-07 | none |
| BR-22 | DOD-01, DOD-05 | none |
| BR-23 | DOD-03, DOD-09 | none |
| BR-24 | DOD-02 | none |
| BR-25 | DOD-02, DOD-06, DOD-09 | none |

## 18. Final gate

- [x] Every requirement is covered by at least one DoD row.
- [ ] Every applicable DoD row is checked and links to evidence.
- [ ] Every `Proposed` decision (section 14) was approved or removed before implementation.
- [ ] Deterministic checks and live browser verification both passed.
- [ ] The final repository diff matches the approved scope.
