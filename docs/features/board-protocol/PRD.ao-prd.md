---
status: Approved
skill: ao-prd
owner: "Andrii Opryshko"
reviewers: ["Andrii Opryshko"]
updated_at: "2026-10-24"
feature_size: "L"
stage: "03"
depth: "medium"
ticket: ""
---

# PRD — board-protocol

> **Inputs:** [idea brief](./idea-brief.ao-discovery.md) · [CONTEXT](./CONTEXT.md)
> **Reference context:** `services/core/internal/governance` — board meeting entity (kind regular / extraordinary, held date, optional end and location, quorum-required flag), validation that reports every broken field at once, `ErrMeetingNotFound`; no lifecycle statuses, no authorization checks, no association scope yet.
> **Other channels:** `docs/adr/0001`–`0004` (governance and finance are separate contexts; contracts live in finance; one API contract; migrations per context) · prototype `apps/web/src/features/governance/` (slot origin system / ai / human, slot status verified / needs review / not found, advisory vs blocking warnings, readiness gate before printing) · `PRD.md` from `sdlc:write-prd` (open questions only).

## 1. Context

A board member of a young housing association turns the management company's кошторис робіт into a signed протокол правління 10-30 times a year, 1-3 hours each: the estimate is posted to the board chat, members vote there, and only if the vote passes is the protocol typed from an old text, printed and signed in person by the головуючий and the board. Because it takes so long, protocols are written "when there is time", decisions stay unrecorded for weeks, and акти and bank statements lie apart from the protocol they close. The gap shows at the annual general meeting, where the ревізійна комісія checks spending against protocols and co-owners ask where the money went (brief §2, §3).

Nothing has gone badly wrong yet, but the annual general meeting is a fixed date: by then every spending decision of the year needs a signed protocol and its papers in one place (brief §4). The top risk is a protocol that names a full-presence meeting while the vote really happened in the chat; if the charter does not allow that form, every such decision can be challenged at once (brief §10).

The protocol is assembled from typed слоти рішення: requisites, number and date, members present, кворум, votes, the estimate amount and each частка будинку are rendered from reference data and the справа, each with a visible source. A член правління works in a conversation with the помічник протоколу: attaches the estimate, adds what the decision is for, and gets a draft; the помічник writes the wording and advisory попередження with a source link and proposes amounts read from documents, but no amount reaches print until a person confirms or corrects it; shares, per-building amounts, кворум, the competence-limit check and vote tallies are computed from data. The wording can also be edited by hand outside the conversation, while the numbers stay in the slots. Members may cast попереднє голосування on a frozen version of the wording; it is recommended, never required, and the signed paper outranks it. After printing and signing, the signed copy is uploaded and the protocol joins the реєстр протоколів together with its справа (brief §13, Approach C).

Traceability:
- Read: `services/core/internal/governance/{domain,app,ports}/`, `module.go`; `docs/adr/0001`–`0004`; `apps/web/src/features/governance/` (`types.ts`, `readiness.ts`).
- Deviation from the brief: the brief's §5 kept "the board chat poll is replaced: no, its result is transferred". This PRD adds in-system попереднє голосування by each member on a frozen wording version, optional and non-blocking — user decision 2026-10-24: "трекаємо реальні попередні голоси на ті формулювання, що були підготовлені, але це не блокує друк".
- Deviation from the brief: failed votes are not printed; a draft that is not signed is closed as знятий з розгляду and kept as an internal record — user decision 2026-10-24 (answers §15 "must failed votes be recorded as protocols?").
- Deviation from the brief: §3 named one preparer and the головуючий as signer; rights are not split — any член правління with a term in force may move a protocol through every state, record an обхід блокування and upload the signed copy (user decision 2026-10-24). The головуючий remains a slot of the protocol, not a separate permission.
- Deviation from the brief: a missing правило статуту on meeting forms, кворум or the competence limit only warns instead of blocking (AC-21b) — user decision 2026-10-24: «Лише попередження».
- Deviation from the brief: a головуючий substituting for the chair without a service note in the справа only warns (AC-27); the brief required the note — user decision 2026-10-24: «Лише попередження».
- Deviation from the brief: the помічник протоколу (conversation drafting) is in Stage 1, not later (user decision 2026-10-24).
- Decision override: repository paths and module names in §1 ¶4 (critic F6) — rationale: the ao-prd template requires §1 ¶4 to list the reference paths and ADRs read, for traceability.
- Deviation from the brief: uploading the signed scan means the protocol is adopted as frozen; there is no separate per-item result entry, and the scan's number and date are checked only with a warning — user decision 2026-10-24: «якщо вже дока зі сканом завантажена — ну значить прийняли … чекнути, щоб дати та номера протоколів сходилися».
- Clarification (clarify-prd, 2026-10-24): the first release only records who voted and what the meeting decided; missing charter rules or missing meeting votes warn instead of block, every blocking check except unconfirmed numbers and text-slot mismatch can be overridden with a reason (§10), and the general-meeting reservation check moves to Stage 2 — user: «мені треба просто на цьому етапі зафіксувати хто проголосував».
- Brief §17 said only `system` existed on the server; `governance` (meetings) and `finance` (contracts) now exist and are reused.

## 2. Goals

- A член правління gets a printable protocol from a кошторис робіт in one sitting, without retyping numbers from documents.
- Every number, date and person in a signed protocol shows the document or the person it came from, and nothing outside the board's competence or the charter's meeting form reaches signing unnoticed.
- Board members find every decision of the year — signed or withdrawn — with its справа, votes and recorded overrides in one place.

## 3. Non-goals

- General-meeting protocols and co-owner voting — different body, competence and quorum (brief §5).
- Collecting contributions and suing debtors — the management company's job (brief §5).
- Electronic signing of the protocol; the binding signature stays on paper — brief §5; the signed copy is uploaded as a scan.
- A binding electronic board vote — попереднє голосування is advisory; the signed paper decides (user decision 2026-10-24).
- Casting a preliminary vote on another member's behalf — the trail must show who voted (user decision 2026-10-24).
- Preparing the annual budget (кошторис доходів і витрат) — brief §5.
- Importing the 30+ past paper protocols into the реєстр протоколів — not needed for this year's annual meeting; revisit after Stage 2.
- Access for співвласники and the управляюча компанія — they receive printed protocols as today; their own access is a separate feature.
- The ревізійна комісія's own access to the реєстр протоколів — left for a later feature («це на потім»); was in US-09 before edit.
- Exporting the year's archive for the ревізійна комісія — left for a later feature together with the commission's access; was in US-09 before edit.

## 4. User stories

### US-01: Keep board composition and charter rules  _(Stage 1)_
**As a** член правління
**I want** to record the board members with their terms, the частки будинку in force and each правило статуту with the date it takes effect
**So that** every protocol takes composition, shares, кворум, the competence limit and the allowed meeting forms as of its meeting date, not as of today

### US-02: Draft a protocol through the assistant  _(Stage 1)_
**As a** член правління
**I want** to open a справа, add documents to it by hand or in the conversation (a newer document is simply added, not marked as replacing an older one), and ask the помічник протоколу for a draft, adding what the decision is for
**So that** I start from a filled draft instead of an old text

### US-03: Confirm every number against its source  _(Stage 1)_
**As a** член правління
**I want** each слот рішення to show its value, where it came from and whether it is confirmed, and to confirm or correct proposed amounts
**So that** no number I did not check reaches the printed protocol

### US-04: Edit the wording by hand  _(Stage 1)_
**As a** член правління
**I want** to edit the protocol text directly, in the conversation or by hand
**So that** I can fix phrasing without losing the link between numbers and their sources

### US-05: Put the wording to a preliminary vote  _(Stage 1)_
**As a** член правління
**I want** to freeze the wording and let each member cast a попереднє голосування on it
**So that** we see support for exactly this text before printing

### US-06: Mark the protocol ready to sign  _(Stage 1)_
**As a** член правління
**I want** to see every блокуюча перевірка and попередження and mark the protocol готовий до підпису once the blocking ones are resolved or overridden with a reason
**So that** what we print is legally sound or the risk is recorded

### US-07: Lock the protocol with its signed copy  _(Stage 1)_
**As a** член правління
**I want** to upload the signed paper copy, which records the protocol as adopted as frozen
**So that** the protocol is locked in the реєстр протоколів with its справа

### US-08: Withdraw a draft that will not be signed  _(Stage 1)_
**As a** член правління
**I want** to close a draft as знятий з розгляду with a reason
**So that** abandoned proposals do not hang open and the history of what was proposed stays

### US-09: Find a past decision with its basis  _(Stage 1)_
**As a** член правління
**I want** to find signed and withdrawn protocols by date, contractor or amount and open each with its signed copy, sources, preliminary votes and overrides
**So that** I can answer "who approved this and on what basis" without collecting papers

### US-10: Use the required wording template  _(Stage 2)_
**As a** член правління
**I want** agenda items whose wording the law or charter fixes to use a шаблон формулювання where only the slots are filled
**So that** required phrasing is never rewritten by the assistant or by hand

### US-11: See related decisions and limit splitting  _(Stage 2)_
**As a** член правління
**I want** advisory warnings when a draft touches a чинне рішення or splits one job across protocols under the limit
**So that** I do not contradict a decision in force or bypass the competence limit by accident

### US-12: Close an expense line with act and statement  _(Stage 2)_
**As a** член правління
**I want** to attach the акт виконаних робіт and the bank statement to the decision they close
**So that** the справа shows each approved expense as open or closed before the annual meeting

## 5. Acceptance criteria

### AC-01 (US-01) — happy path
**Given** a член правління and a board election protocol naming four members with a term until 2028-05-31
**When** the член правління records the four members with that term and attaches the election protocol
**Then** the system lists them as the board in force for any meeting held up to 2028-05-31 and shows the election protocol as the source

### AC-02 (US-01) — traceability
**Given** a правило статуту "кворум: 3 members" in force from 2025-01-01 and a new one "кворум: 4 members" recorded with effect from 2026-11-01
**When** a член правління prepares a protocol for a meeting held on 2026-10-20
**Then** the protocol checks кворум against 3 members, names the rule in force on the meeting date, and the later rule is not applied

### AC-03 (US-02) — happy path
**Given** a член правління with a справа holding a кошторис робіт of the management company for roof repair at 184 500 грн
**When** the член правління asks the помічник протоколу for a protocol draft and adds "ремонт покрівлі, оплата з фонду ремонту"
**Then** the system creates a draft for the chosen meeting with the agenda item, wording, members and слоти рішення filled, the amount slot shows 184 500 грн as proposed and unconfirmed with the estimate as its source, and the draft is in state чернетка

### AC-04 (US-02) — error
**Given** a член правління with an empty справа
**When** the член правління asks the помічник протоколу for a draft
**Then** the system refuses and says that at least one source document must be attached first

### AC-05 (US-02) — error
**Given** a член правління opening a draft for a meeting that is not stored for the association
**When** the член правління chooses that meeting
**Then** the system says the meeting is not found and creates no draft

### AC-06 (US-03) — traceability
**Given** a draft whose amount slot holds 184 500 грн proposed from the кошторис робіт
**When** the член правління opens the draft
**Then** the amount shows the estimate number, date and the fragment it was read from, is marked as proposed by the помічник, and the draft cannot be marked готовий до голосування or готовий до підпису until a person confirms or corrects it; an обхід блокування does not lift this check

### AC-06b (US-03) — traceability
**Given** an estimate whose total is 184 500 грн including VAT and the помічник proposed 153 750 грн
**When** the член правління corrects the amount to 184 500 грн and confirms it
**Then** the slot records the corrected value, who confirmed it and when, keeps the proposed value in the history, and counts the amount as confirmed

### AC-07 (US-03) — traceability
**Given** an amount of 184 500 грн to be shared between two buildings and an inter-building decision in force on the meeting date with shares 58 % and 42 %
**When** the член правління opens the draft
**Then** the per-building amounts 107 010 грн and 77 490 грн are computed from the confirmed amount and the shares, show the inter-building decision as their source, and cannot be typed over

### AC-08 (US-03) — traceability
**Given** a draft whose members, головуючий and meeting date are filled
**When** the член правління opens the draft
**Then** members show the election protocol as of the meeting date, the draft has no protocol number yet, and a головуючий who is not the board chair shows the service note in the справа as the source

### AC-09 (US-03) — domain invariant
**Given** a draft with a decision slot (amount, counterparty or джерело коштів) that lacks either a document of the справа or a person's confirmation
**When** the член правління tries to mark it готовий до підпису
**Then** the system blocks, names the slot and says the value has no confirmed source; an обхід блокування does not lift this check, only a document plus a person's confirmation does (slot sources by type: §10)

### AC-10 (US-03) — cross-context
**Given** the association's financial records hold a maintenance contract in force and a repair-fund contract that ended on 2026-06-30
**When** the член правління picks the джерело коштів for a decision of a meeting held on 2026-10-20
**Then** only the contract in force on the meeting date is offered, and the ended one is refused with its end date named

### AC-11 (US-04) — happy path
**Given** a draft in state чернетка
**When** the член правління rewords a sentence by hand without touching any number
**Then** the system keeps the new wording, records who changed it, and every slot keeps its value and source

### AC-11b (US-04) — domain invariant
**Given** two members of the board have opened version 3 of the same draft
**When** the first saves a change and the second then saves a change made to version 3
**Then** the system refuses the second change, says the draft was changed by the first member, and shows version 4 without losing either change

### AC-12 (US-04) — domain invariant
**Given** a draft whose amount slot is 184 500 грн
**When** the член правління edits the text by hand so that it reads 185 400 грн
**Then** the system marks the mismatch between text and slot, and the draft cannot become готовий до голосування or готовий до підпису until text and slot agree

### AC-13 (US-05) — happy path
**Given** a draft in state чернетка with all amounts confirmed
**When** a член правління marks it готовий до голосування
**Then** the wording is frozen as a numbered version, every member of the board in force sees it to vote, and the draft is in state готовий до голосування

### AC-14 (US-05) — domain invariant
**Given** a draft with an amount still proposed and unconfirmed
**When** a член правління marks it готовий до голосування
**Then** the system refuses and names the unconfirmed slot

### AC-15 (US-05) — happy path
**Given** a draft in state готовий до голосування, version 2
**When** each of three members casts "за" for agenda item 1 and the fourth records неучасть у голосуванні with the reason "особиста заінтересованість"
**Then** the system shows 3 "за" of 3 voting and the preliminary majority as reached, each vote with its member and time, all tied to version 2; this result is advisory and only pre-fills the meeting vote count (AC-15b)

### AC-15b (US-05) — traceability
**Given** a draft for a meeting where 4 members were present
**When** the член правління records the meeting's vote count for agenda item 1 as 3 "за", 1 "утримався"
**Then** the slot shows the count with the member who entered it and when, is pre-filled from the попереднє голосування when there was one, and marks every place where the entered count differs from the preliminary votes

### AC-15c (US-05) — happy path
**Given** a draft in state готовий до голосування or чернетка with no attendance or no meeting vote count recorded for agenda item 1
**When** a член правління marks it готовий до підпису
**Then** the draft becomes готовий до підпису with a попередження that кворум and majority for item 1 could not be checked, and the printed copy leaves the attendance and vote count blank to fill in by hand at the meeting

### AC-16 (US-05) — authorization
**Given** a член правління voting on a draft in state готовий до голосування
**When** the член правління tries to cast a vote for another member
**Then** the system refuses and says each member votes only for themselves

### AC-17 (US-05) — happy path
**Given** a draft in state готовий до голосування, version 2, with two preliminary votes
**When** a член правління edits the wording, the agenda or a decision slot
**Then** the draft returns to state чернетка, the two votes stay recorded against version 2, and the next vote starts on version 3 with no votes

### AC-18 (US-05) — domain invariant
**Given** a draft in state чернетка
**When** a член правління tries to cast a preliminary vote
**Then** the system refuses and says voting is open only on a frozen version

### AC-19 (US-06) — domain invariant
**Given** a draft whose attendance shows 2 of 4 members and a правило статуту кворум of 3 in force on the meeting date
**When** the член правління tries to mark it готовий до підпису
**Then** the system blocks, names кворум as not reached with the rule and its effective date, and no item is shown as adopted

### AC-19b (US-06) — domain invariant
**Given** 3 of 4 members present, a правило статуту кворум of 3, a rule that a member with неучасть у голосуванні does not count toward that item's кворум, and one member recording неучасть у голосуванні on item 1 for personal interest
**When** the член правління tries to mark the draft готовий до підпису
**Then** the system blocks item 1 as without кворум, names both rules, and item 2, where all three vote, is not blocked

### AC-19c (US-06) — domain invariant
**Given** a правило статуту requiring 3 votes "за" to adopt a decision, and agenda item 1 with a recorded meeting vote count of 2 "за", 2 "проти"
**When** the член правління tries to mark the draft готовий до підпису with item 1 worded as adopted
**Then** the system blocks, names the majority rule, and lifts the block only when item 1 is worded as not adopted, removed from the agenda, or an обхід блокування is recorded with a reason

### AC-20 (US-06) — legal competence
**Given** a правило статуту setting the board's competence limit at 100 000 грн in force on the meeting date
**When** the член правління tries to mark ready to sign a draft approving 184 500 грн
**Then** the system blocks and names the limit rule, saying the decision needs the general meeting or an обхід блокування with a reason

### AC-21 (US-06) — legal competence
**Given** a правило статуту allowing only meetings held in person, and a draft naming the meeting form "заочне (опитування в чаті)"
**When** the член правління tries to mark it готовий до підпису
**Then** the system blocks and names the rule on allowed meeting forms

### AC-21b (US-06) — legal competence
**Given** no правило статуту on allowed meeting forms, кворум, majority, the competence limit or whether a non-voting member counts toward an item's кворум is recorded for the meeting date
**When** the член правління marks a draft готовий до підпису
**Then** the draft becomes готовий до підпису with a попередження naming each missing rule and saying that check could not be made, which does not stop printing

### AC-22 (US-10) — legal competence  _(Stage 2: needs typed agenda items)_
**Given** a draft whose agenda item, of a type the charter reserves for the general meeting, allots part of the common premises to the association, a decision the charter reserves for the general meeting
**When** the член правління tries to mark it готовий до підпису
**Then** the system blocks and names the charter rule that reserves the decision for the general meeting; an обхід блокування with a reason lifts it

### AC-23 (US-06) — happy path
**Given** a draft in state готовий до голосування with a blocking competence-limit check
**When** a член правління records an обхід блокування with the reason "аварійний ремонт, загальні збори скликаються 2026-11-15" and marks it готовий до підпису
**Then** the draft is in state готовий до підпису, the text is frozen for printing, and the override with its author, time and reason stays with the protocol

### AC-24 (US-06) — error
**Given** a draft with a blocking check
**When** a член правління records an обхід блокування without a reason
**Then** the system refuses and says a reason is required

### AC-25 (US-06) — happy path
**Given** a draft in state чернетка with every blocking check resolved and no preliminary votes
**When** a член правління marks it готовий до підпису
**Then** the draft is in state готовий до підпису, takes the next free number of the association's protocol register (one sequence, no duplicates) printed on the copy, and shows the попередження "попереднього голосування в системі не було", which does not stop printing

### AC-25b (US-06) — happy path
**Given** a draft in state готовий до голосування where the preliminary vote has 2 "за" of 4
**When** a член правління marks it готовий до підпису
**Then** the draft is in state готовий до підпису and shows the попередження "попередня більшість не набрана", which does not stop printing

### AC-26 (US-06) — happy path
**Given** a draft in state готовий до підпису
**When** a член правління edits the wording, the agenda or a decision slot (amount, counterparty, джерело коштів)
**Then** the draft returns to state чернетка, its protocol number is released, and every check runs again; recording meeting facts (attendance, неучасть у голосуванні, meeting vote count, meeting form, головуючий), confirming a value without changing it, or recording an обхід блокування does not return it to чернетка — the checks just rerun

### AC-27 (US-06) — traceability
**Given** a draft whose головуючий is a member other than the board chair and the справа holds no service note for that meeting
**When** the член правління marks it готовий до підпису
**Then** the draft becomes готовий до підпису with a попередження that the головуючий has no service note as a source, which does not stop printing

### AC-28 (US-07) — happy path
**Given** a protocol in state готовий до підпису, printed and signed on paper
**When** a член правління uploads the signed scan
**Then** the protocol is in state підписаний, its decisions count as adopted exactly as worded in the frozen text, it appears in the реєстр протоколів with the scan, its справа, votes and overrides, and keeps the number it took when marked готовий до підпису

### AC-28b (US-07) — traceability
**Given** a protocol No 14 dated 2026-10-20 in state готовий до підпису
**When** a член правління uploads a signed scan on which the number reads 15 or the date reads 2026-10-21
**Then** the protocol becomes підписаний and shows a попередження naming the number or date the помічник read from the scan next to the protocol's own, which does not stop the upload; when the помічник cannot read them, the попередження says the scan was not checked

### AC-29 (US-07) — error
**Given** a protocol in state готовий до підпису
**When** a член правління uploads a file above the size limit or of a type that is not allowed (§6)
**Then** the system refuses, names the allowed types and the size limit, and the state does not change

### AC-30 (US-07) — domain invariant
**Given** a protocol in state підписаний
**When** a член правління tries to change any slot, the wording or the scan
**Then** the system refuses, says a signed protocol cannot be changed, and offers to start a new protocol that references it as a correction

### AC-30b (US-07) — domain invariant
**Given** a protocol in state підписаний
**When** a член правління tries to close it as знятий з розгляду
**Then** the system refuses, says a signed protocol cannot be withdrawn, and offers to start a new protocol that revokes or corrects it

### AC-31 (US-07) — domain invariant
**Given** a draft in state чернетка or готовий до голосування
**When** a член правління tries to upload a signed copy
**Then** the system refuses and says the protocol must first be marked готовий до підпису

### AC-32 (US-07) — authorization
**Given** a person whose term as член правління ended on 2026-05-31
**When** that person tries to open, edit or upload a signed copy for any protocol of the association
**Then** the system refuses, and the protocol is unchanged

### AC-33 (US-07) — authorization
**Given** a член правління of association A
**When** that member opens a protocol or справа of association B
**Then** the system answers that the protocol is not found, without revealing that it exists

### AC-34 (US-08) — happy path
**Given** a draft in state готовий до голосування whose preliminary vote did not gather support
**When** a член правління closes it as знятий з розгляду with the reason "не набрано підтримки, переглянути кошторис"
**Then** the draft is in state знятий з розгляду, keeps its справа, versions and preliminary votes, is never printed, and takes no protocol number

### AC-34b (US-08) — happy path
**Given** a draft in state чернетка that the board decided not to pursue
**When** a член правління closes it as знятий з розгляду with the reason "управляюча компанія відкликала кошторис"
**Then** the draft is in state знятий з розгляду with its справа and versions kept, and takes no protocol number

### AC-34c (US-08) — happy path
**Given** a protocol in state готовий до підпису that was printed but not signed
**When** a член правління closes it as знятий з розгляду with the reason "не підписано на засіданні"
**Then** the protocol is in state знятий з розгляду, its frozen text is kept as never signed, and its reserved number is released

### AC-35 (US-08) — domain invariant
**Given** a draft in state знятий з розгляду
**When** a член правління tries to edit it or mark it готовий до голосування
**Then** the system refuses and offers to start a new draft copied from it

### AC-36 (US-09) — happy path
**Given** a член правління and a реєстр протоколів holding 14 signed and 3 withdrawn protocols for 2026
**When** the член правління searches by the contractor "управляюча компанія" and the year 2026
**Then** the system lists the matching protocols with number, date, amount and state, and each opens with its signed copy (none for withdrawn), справа documents, slot sources, preliminary votes and recorded overrides

### AC-38 (US-10) — domain invariant
**Given** an agenda-item type with a шаблон формулювання
**When** the помічник протоколу or a член правління changes the fixed text of that item
**Then** the system refuses the change and allows only the template's slots to be filled

### AC-39 (US-11) — cross-context
**Given** a чинне рішення of the board approving the same contractor's estimate for the same building within the last 90 days
**When** the член правління prepares a new draft for that contractor and building
**Then** the system shows a попередження linking the earlier decision and its amount, which does not stop printing

### AC-40 (US-11) — legal competence
**Given** two drafts for the same estimate at 60 000 грн and 70 000 грн and a competence limit of 100 000 грн
**When** the член правління marks the second one готовий до підпису
**Then** the system shows a попередження that the same estimate totals 130 000 грн across protocols, above the limit

### AC-41 (US-12) — happy path
**Given** a signed protocol approving 184 500 грн and an акт виконаних робіт for 184 500 грн and a bank statement for 184 500 грн
**When** a член правління attaches both to the decision
**Then** the decision shows the expense line as closed with links to the act and the statement

### AC-42 (US-12) — traceability
**Given** a signed protocol approving 184 500 грн
**When** a член правління attaches an акт виконаних робіт for 196 000 грн
**Then** the expense line stays open and shows a попередження naming both amounts and their documents

## 5.1 Error catalogue

| Code | Kind | Message to the actor | AC |
|---|---|---|---|
| `governance.source_document_required` | validation | Спершу додайте до справи хоча б один документ. | AC-04 |
| `governance.meeting_not_found` | not_found | Засідання не знайдено. | AC-05 |
| `governance.protocol_not_found` | not_found | Протокол не знайдено. | AC-33 |
| `governance.amount_unconfirmed` | invariant | Сума запропонована помічником і ще не підтверджена людиною. | AC-06, AC-14 |
| `governance.source_missing` | invariant | Значення не має підтвердженого джерела у справі. | AC-09 |
| `governance.funding_source_not_in_force` | invariant | Цей договір не чинний на дату засідання. | AC-10 |
| `governance.draft_changed` | conflict | Чернетку щойно змінив інший член правління; перегляньте нову версію. | AC-11b |
| `governance.text_slot_mismatch` | invariant | Число в тексті не збігається зі слотом рішення. | AC-12 |
| `governance.vote_on_behalf` | authorization | Кожен член правління голосує лише за себе. | AC-16 |
| `governance.voting_not_open` | invariant | Голосування відкрите лише для замороженої версії формулювання. | AC-18 |
| `governance.quorum_missing` | invariant | Кворуму немає за правилом статуту, чинним на дату засідання. | AC-19, AC-19b |
| `governance.majority_missing` | invariant | Пункт не набрав більшості за правилом статуту і не може бути записаний як прийнятий. | AC-19c |
| `governance.outside_board_competence` | invariant | Сума перевищує ліміт компетенції правління. | AC-20 |
| `governance.meeting_form_not_allowed` | invariant | Статут не дозволяє таку форму засідання. | AC-21 |
| `governance.reserved_for_general_meeting` | invariant | Це рішення статут відносить до компетенції загальних зборів. | AC-22 |
| `governance.reason_required` | validation | Вкажіть причину. | AC-24 |
| `governance.signed_copy_invalid` | validation | Дозволені PDF, JPEG або PNG до 20 МБ. | AC-29 |
| `governance.protocol_locked` | invariant | Підписаний протокол змінити або зняти не можна; створіть протокол-виправлення. | AC-30, AC-30b |
| `governance.transition_not_allowed` | invariant | Цей перехід недоступний у поточному стані протоколу. | AC-31, AC-35 |
| `governance.not_board_member` | authorization | Дія доступна лише члену правління з чинними повноваженнями. | AC-32 |
| `governance.template_text_fixed` | invariant | Текст шаблону формулювання не редагується; заповніть лише слоти. | AC-38 |

## 6. Non-functional requirements

| Aspect | Target | Measurement |
|---|---|---|
| Latency p95, read or list of one association's records | ≤ 200 ms (default) | server request-duration histogram per route |
| Latency p95, write without file | ≤ 500 ms (default) | same histogram |
| Latency p95, upload of a signed scan | ≤ 3 s for 20 MB (default) | same histogram, upload routes |
| Printable protocol generation | ≤ 5 s p95 (default) | request duration |
| Draft from the помічник протоколу | ≤ 30 s p95 from request to filled draft | assistant request duration |
| Throughput | ≥ 50 req/s per instance on the read path (default) | load smoke test in CI |
| Availability | 99.5 % monthly (default) | uptime check on the health endpoint |
| Durability | RPO ≤ 24 h, RTO ≤ 4 h (default) | daily backup + quarterly restore drill |
| File size | ≤ 20 MB per file, PDF / JPEG / PNG (default) | upload validation |
| Retention | signed and withdrawn protocols kept ≥ 5 years, never hard-deleted (default) | storage policy |
| Integrity | a signed protocol changes only through a new referencing protocol; 0 mismatches between printed numbers and confirmed slots (default + feature) | integrity check job |

## 6.1 Security / privacy

- **Data classification:** Confidential — internal legal and financial records of the association, with personal data.
- **Personal data touched:** names and terms of board members (low); signatures on scans (medium); names of co-owners and apartment numbers that may appear in case documents and owner requests (medium); reasons for неучасть у голосуванні such as personal interest (medium).
- **Authorization impact:** new rights: any член правління with a term in force acts on the association's protocols; a person whose term ended and anyone from another association has no access. Depends on sign-in, which the app does not have yet (§8).
- **Abuse cases:**
  - Cross-association access: a member opens another association's protocol by its identifier — answered as not found (AC-33).
  - Editing a signed record: a member tries to change a signed protocol to match a later bill — refused; correction only through a new referencing protocol (AC-30).
  - Leaking personal data to the text-generation service: case documents with co-owners' names are sent to the помічник — sent only under the agreement decided in §8; nothing is used for training.
  - Uploading a forged or wrong scan: the uploaded copy differs from the frozen text — the number and date read from the scan are compared with the protocol and a mismatch is shown as a попередження (AC-28b); the scan, uploader and time are recorded and visible to board members in the реєстр протоколів.
  - Votes on behalf of others: a member votes for an absent colleague — refused (AC-16).
- **Security review:** Required — first feature with authorization, personal data and an external text-generation service.

## 7. Metrics / KPIs

- **Time per protocol, from estimate to print** — baseline: 1-3 h (brief §2, author's estimate), target: ≤ 20 min median within 3 months of Stage 1 release.
- **Lag from board decision to locked protocol** — baseline: weeks (brief §2, not measured), target: ≤ 7 days median within 3 months of Stage 1.
- **Spending decisions with a locked protocol before the annual meeting** — baseline: not measured, target: 100 % of 2027 spending decisions by the 2027 annual meeting.
- **AI-proposed amounts corrected by a person** — baseline: none, target: measured on every protocol; ≤ 5 % corrected over one season (brief §14 revisit trigger for Approach B).
- **Protocols with in-system preliminary votes** — baseline: 0, target: ≥ 50 % of signed protocols within 6 months of Stage 1.

## 8. Open questions

- [ ] Which meeting forms does the charter allow for the board (in person, absentee, mixed), and which clause? Default now: until recorded as a правило статуту, the check is skipped with a попередження (AC-21b). — owner: board chair, due: before Stage 1 release
- [ ] Is the 100 000 грн limit per protocol, per job or per estimate; which clause? Default now: per decision line; splitting across protocols is a Stage 2 warning (AC-40). — owner: board chair, due: before /sdlc:architecture-design
- [ ] On a 4-member board, is "3 votes for" the кворум, the majority, or both? Default now: two separate правила статуту, кворум 3 and majority 3. — owner: board chair, due: before Stage 1 release
- [ ] At what overrun of the act over the approved amount is a new decision needed? Default now: any overrun keeps the line open with a warning (AC-42). — owner: board chair + accountant, due: before Stage 2
- [ ] Will the management company hand over acts and statements per expense line? Default now: attached by a член правління. — owner: Andrii Opryshko, due: before Stage 2
- [ ] How does a член правління sign in, given the app has no sign-in yet? Default now: a prerequisite decided in architecture. — owner: Tech Lead, due: before /sdlc:architecture-design
- [ ] The app stores no association yet; are existing meetings and contracts assigned to the first association? Default now: yes, confirmed by a член правління. — owner: Tech Lead, due: before /sdlc:architecture-design
- [ ] May case documents with personal data go to a text-generation service outside the association's hosting? Default now: only under an agreement that excludes training on the data. — owner: Tech Lead, due: before /sdlc:architecture-design
- [ ] Where is the wording edited by hand (the author proposes a shared online document editor such as Google Docs), and how are numbers in that text checked against slots? Default now: numbers live only in slots and are re-checked before each state change (AC-12). — owner: Tech Lead, due: before /sdlc:architecture-design
- [ ] Do the financial records hold each contract's джерело коштів? Default now: a fund line is added to contracts and filled for contracts in force (§9). — owner: Tech Lead, due: before /sdlc:architecture-design
- [ ] Time the next 2-3 protocols to replace the estimated KPI baselines (1-3 h per protocol, weeks of lag). Default now: author's estimate from the brief. — owner: Andrii Opryshko, due: before Stage 1 release
- [ ] Is a джерело коштів stored as a contract with the management company (its fund taken from the contract) or as a fund line checked against contracts separately? Default now: a contract in force, as in AC-10 and §9. — owner: Tech Lead, due: before /sdlc:architecture-design
- [ ] How does the помічник read the protocol number and date from a signed scan (AC-28b), and how reliable must it be? Default now: best effort; unreadable means a попередження "не перевірено". — owner: Tech Lead, due: before /sdlc:architecture-design
- [ ] Confirm Reach: 13 associations at about 20 protocols each. — owner: Andrii Opryshko, due: before roadmap scoring

## 9. Migration plan

- **New records:** правила статуту with effective dates; board terms with their election protocols; частки будинку with the inter-building decision; справи with source documents; protocols with wording versions, слоти рішення and their sources, preliminary votes, overrides and signed copies. Each protocol links to exactly one meeting already stored by `governance`, and a meeting has at most one protocol; each agenda item of the protocol has its own справа; a decision's джерело коштів links to a contract stored by `finance`.
- **Existing meetings (`governance`):** untouched. A meeting may have no protocol; a protocol is created for an existing meeting or together with a new one. The meeting's quorum-required flag stays; the quorum value comes from the правило статуту in force on the meeting date. No backfill. The meeting form (in person, absentee, mixed) is a slot of the protocol, not a field of the meeting, so existing meetings get nothing filled in.
- **Existing contracts (`finance`):** read as the list of джерела коштів. If they lack a fund line (§8), it is added and filled by a член правління for contracts in force before the first protocol; ended contracts stay empty.
- **Association scope:** existing meetings and contracts are assigned to the first association, confirmed by a член правління (§8).
- **Past paper protocols:** not imported (§3).
- **Reversibility:** rolling back Stage 1 or Stage 2 removes the screens and actions; stored protocols, signed copies, votes and overrides are kept read-only. Signed and withdrawn protocols are never deleted.
- **Rollout:** Stage 0 (prerequisite): sign-in and association scope. Stage 1: US-01-US-09; before the first protocol a член правління records the board, shares and charter rules. Stage 2: US-10-US-12; between stages users see no templates, related-decision warnings or expense closure, and drafts behave as in Stage 1.

## 10. RBAC and state machine

### Roles and actions

| Role | Keep board and charter rules | Create / edit draft, ask the помічник | Mark готовий до голосування | Cast preliminary vote | Override a blocking check | Mark готовий до підпису | Upload signed copy | Withdraw | Read signed / withdrawn | Read drafts |
|---|---|---|---|---|---|---|---|---|---|---|
| член правління (term in force) | ✓ | ✓ | ✓ | own only | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| головуючий | ✓ (as член правління) | ✓ | ✓ | own only | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| ревізійна комісія | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ (later feature) | ✗ |
| член правління (term ended) | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ |
| співвласник | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ |
| управляюча компанія | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ |

Scope is one association. Rights to act come from a board term in force on the date of the action; who may cast a preliminary vote, who counts as present and the кворум are taken from the board on the meeting date.

### Slot sources

| Slot | Counts as sourced when |
|---|---|
| amount, counterparty, джерело коштів | a document of the справа **and** a person's confirmation (an override never replaces this) |
| members, protocol number, частки будинку | the reference record (election protocol, protocol register, inter-building decision); no confirmation needed |
| attendance, неучасть у голосуванні, meeting vote count, meeting form, головуючий | a член правління's entry, recorded with who and when |

### Which blocking checks an обхід блокування lifts

| Check | Override with a reason |
|---|---|
| кворум, кворум пункту, majority (AC-19, AC-19b, AC-19c) | yes |
| competence limit (AC-20) | yes |
| meeting form not allowed (AC-21) | yes |
| reserved for the general meeting (AC-22, Stage 2) | yes |
| decision number unconfirmed or without a document (AC-06, AC-09, AC-14) | no |
| text differs from a slot (AC-12) | no |

### States

- **чернетка** — the draft is edited in the conversation or by hand; checks run on every change; no protocol number yet.
- **готовий до голосування** — wording frozen as a numbered version; members cast попереднє голосування.
- **готовий до підпису** — text frozen for printing; every number confirmed with a source, every other блокуюча перевірка resolved or overridden.
- **підписаний** — the signed copy is uploaded; the protocol is locked in the реєстр протоколів.
- **знятий з розгляду** — closed without printing; kept as an internal record.

### Allowed transitions

| From → To | Who | Condition | AC |
|---|---|---|---|
| (none) → чернетка | член правління | справа has at least one source document; meeting exists | AC-03 |
| чернетка → готовий до голосування | член правління | no unconfirmed amount, text matches slots | AC-13 |
| готовий до голосування → чернетка | член правління | edit of wording, agenda or a decision slot; votes stay with the old version | AC-17 |
| чернетка → готовий до підпису | член правління | every number confirmed with a source; other blocking checks resolved or overridden; warns without preliminary votes | AC-25 |
| готовий до голосування → готовий до підпису | член правління | every number confirmed with a source; other blocking checks resolved or overridden; warns without preliminary majority | AC-23, AC-25b |
| готовий до підпису → чернетка | член правління | edit of wording, agenda or a decision slot; number released; checks rerun | AC-26 |
| готовий до підпису → підписаний | член правління | signed copy uploaded; decisions adopted as frozen; number and date on the scan checked with a warning | AC-28, AC-28b |
| чернетка / готовий до голосування / готовий до підпису → знятий з розгляду | член правління | reason given | AC-34, AC-34b, AC-34c |

### Refused transitions

| From → To | Why refused | AC |
|---|---|---|
| підписаний → any (edit or знятий з розгляду) | a signed protocol is immutable; corrections go through a new referencing protocol | AC-30, AC-30b |
| чернетка / готовий до голосування → підписаний | a signed copy is accepted only for a text frozen for printing | AC-31 |
| чернетка / готовий до голосування → готовий до підпису with an open blocking check | blocking checks must be resolved or overridden with a reason | AC-19, AC-19b, AC-19c, AC-20, AC-21, AC-22, AC-24 |
| чернетка / готовий до голосування → готовий до підпису with an unconfirmed or unsourced number | numbers must be confirmed with a source; an override does not lift this | AC-06, AC-09 |
| чернетка → готовий до голосування / готовий до підпису with text that differs from slots | the printed text must match the confirmed slots | AC-12 |
| чернетка → готовий до голосування with an unconfirmed amount | members vote only on confirmed numbers | AC-14 |
| знятий з розгляду → any | final; a new draft is copied instead | AC-35 |
