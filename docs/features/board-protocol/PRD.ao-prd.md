---
status: Approved
skill: ao-prd
owner: "Andrii Opryshko"
reviewers: ["Andrii Opryshko"]
updated_at: "2026-10-25"
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

Nothing has gone badly wrong yet, but the annual general meeting is a fixed date: by then every spending decision of the year needs a signed protocol and its papers in one place (brief §4). The top risk is a protocol that names a full-presence meeting while the vote really happened in the chat; if the charter does not allow that form, every such decision can be challenged at once (brief §10). Stage 1 does not check this; it stores the meeting form the charter allows and prints it on every copy (AC-45), and it labels the board's in-system poll as informal so it can never pass for an absentee vote (AC-46).

Stage 1 is the помічник протоколу's draft and the keeping of the document until it is signed. A член правління picks the meeting, attaches the case documents for its agenda items and asks the помічник for a draft; the помічник writes the wording and proposes amounts and other values read from the documents, each with a visible link to the document it came from, and no proposed value reaches the print copy until a person confirms or corrects it. Any current член правління reviews and edits the text in spilno (чернетка). To see what passes, the board may run an informal попереднє голосування on the current text (готовий до голосування); it is not the meeting's vote and does not replace it, and any edit or «повернути в чернетку» ends the round. A член правління then renders the print copy (готовий до підпису), which reserves the protocol number; the board signs on paper, and a член правління uploads the signed scan with the meeting facts written on the paper — members present, the vote count per item and a note of any handwritten corrections — which locks the protocol (підписаний). A draft that will not be signed is closed as знятий з розгляду with a reason. Blocking checks and overrides, reading the scan with AI and the export for the ревізійна комісія are Stage 2 (§3).

Traceability:
- Read: `services/core/internal/governance/{domain,app,ports}/`, `module.go`; `docs/adr/0001`–`0004`; `apps/web/src/features/governance/` (`types.ts`, `readiness.ts`).
- Deviation from the brief: the brief's §5 kept "the board chat poll is replaced: no, its result is transferred". This PRD adds an in-system попереднє голосування by each member on the current wording, optional and non-blocking — user decision 2026-10-24: "трекаємо реальні попередні голоси на ті формулювання, що були підготовлені, але це не блокує друк".
- Deviation from the brief: failed votes are not printed; a draft that is not signed is closed as знятий з розгляду and kept as an internal record — user decision 2026-10-24 (answers §15 "must failed votes be recorded as protocols?").
- Deviation from the brief: §3 named one preparer and the головуючий as signer; rights are not split — any член правління with a term in force may move a protocol through every state and upload the signed copy (user decision 2026-10-24). The головуючий remains a slot of the protocol, not a separate permission.
- Deviation from the brief: a missing правило статуту on meeting forms, кворум or the competence limit only warns instead of blocking (AC-21b) — user decision 2026-10-24: «Лише попередження». Stage 2 since 2026-10-25.
- Deviation from the brief: a головуючий substituting for the chair without a service note in the справа only warns (AC-27); the brief required the note — user decision 2026-10-24: «Лише попередження». Stage 2 since 2026-10-25.
- Deviation from the brief: the помічник протоколу (conversation drafting) is in Stage 1, not later (user decision 2026-10-24).
- Decision override: repository paths and module names in §1 ¶4 (critic F6) — rationale: the ao-prd template requires §1 ¶4 to list the reference paths and ADRs read, for traceability.
- Deviation from the brief: uploading the signed scan means the protocol is adopted as signed — user decision 2026-10-24: «якщо вже дока зі сканом завантажена — ну значить прийняли». Since 2026-10-25 the meeting facts are entered from the paper at upload (AC-28) and the number and date on the scan are not read in Stage 1 (AC-28b, Stage 2).
- Clarification (clarify-prd, 2026-10-24): the first release only records who voted and what the meeting decided — user: «мені треба просто на цьому етапі зафіксувати хто проголосував». The overridable-checks part of that clarification and the general-meeting reservation check are Stage 2.
- Stage 1 rescoped by product owner, 2026-10-25: «суть — AI помічник, який складе чернетку, ну і саме ведення документу. Не обов'язково перевірки всякі робити на цьому етапі. Головне чернетка: я перевірив, може підправив; внутрішнє голосування чисто щоб глянути що заходить; можна вернутися назад у чернетку та перезапустити голосування. Потім я друкую, підписуємо і закидую скан.» Stage 1 is the lifecycle in §10; blocking checks, overrides, charter-rule effective dates, reading the scan with AI and the export move to Stage 2 (§3); the poll never pre-fills the meeting vote count (AC-15b dropped); a protocol keeps its number from the first print copy and the number is never reused (AC-26, AC-34c).
- Brief §17 said only `system` existed on the server; `governance` (meetings) and `finance` (contracts) now exist and are reused.

## 2. Goals

- A член правління gets a printable protocol from a кошторис робіт in one sitting: the помічник протоколу drafts it, a person confirms every proposed number and edits the text, without retyping numbers from documents.
- Every number the помічник proposed shows the document it came from and who confirmed it, and the board's informal poll stays visibly separate from the meeting's vote on paper.
- Board members find every decision of the year — signed or withdrawn — with its signed scan, meeting facts, справа and poll history in one place.

## 3. Non-goals

- General-meeting protocols and co-owner voting — different body, competence and quorum (brief §5).
- Collecting contributions and suing debtors — the management company's job (brief §5).
- Electronic signing of the protocol; the binding signature stays on paper — brief §5; the signed copy is uploaded as a scan.
- A binding electronic board vote, or an absentee board meeting held in the system — попереднє голосування is an informal internal poll, not the legal vote of the meeting and not a substitute for it; the signed paper decides (user decision 2026-10-24, product owner 2026-10-25).
- Casting a preliminary vote on another member's behalf — the trail must show who voted (user decision 2026-10-24).
- Preparing the annual budget (кошторис доходів і витрат) — brief §5.
- Importing the 30+ past paper protocols into the реєстр протоколів — not needed for this year's annual meeting; the register continues from a number set in the association settings (AC-45); revisit after Stage 2.
- Access for співвласники and the управляюча компанія — they receive printed protocols as today; their own access is a separate feature.
- The ревізійна комісія's own access to the реєстр протоколів — left for a later feature («це на потім»); was in US-09 before edit.
- More than one association — Stage 1 serves one association; records carry the association, but there is no multi-association scope (onboarding a second association is a later feature).
- **Deferred to Stage 2** (product owner, 2026-10-25: Stage 1 is the assistant's draft and keeping the document; the legal checks stay with the board on paper, as today). The ACs keep their ids and text in §5 «Deferred to Stage 2»:
  - Blocking checks and overrides — кворум, кворум пункту and majority (AC-19, AC-19b, AC-19c, AC-15c), the competence limit 100 000 грн (AC-20), the meeting form (AC-21, AC-21b), the general-meeting reservation (AC-22), an обхід блокування with a reason (AC-23, AC-24), the головуючий's service note (AC-27) — why: they need the charter answers in §8 and would stop the first protocols on missing data.
  - Source of truth as a blocking rule — a decision value without a document and a confirmation (AC-09), text that differs from a slot (AC-12) — why: in Stage 1 every number shows its source document as a visible link and only the помічник's proposals need a person's confirmation (AC-06).
  - Cross-check with the finance contracts — the джерело коштів must be a contract in force (AC-10) — why: needs a fund line on contracts (§8).
  - Charter rules with effective dates (AC-02) — Stage 1 stores only the current rules in the association settings — why: the rules have not changed since the association was founded.
  - Reading the signed scan with AI (AC-28b) — why: the uploader enters the facts from the paper; reading reliability is open (§8).
  - Dropped from Stage 1, not deferred: the poll gate on unconfirmed amounts (AC-14) — why: the poll is informal, the only number gate is before printing (AC-06); pre-filling the meeting vote count from the poll (AC-15b) — why: the poll must never become evidence of an absentee vote; the count comes from the signed paper (AC-28).
  - Exporting the year's archive for the ревізійна комісія — left for a later feature together with the commission's access; was in US-09 before edit.
  - Required wording templates, related decisions and limit splitting, expense closure (US-10–US-12, AC-38–AC-42).

## 4. User stories

### US-01: Keep board composition and association settings  _(Stage 1)_
**As a** член правління
**I want** to record the board members with their terms, the частки будинку in force and the association settings — the meeting form the charter allows, who signs the protocol, the current правила статуту and the number the protocol register continues from
**So that** every protocol shows the board as of its meeting date and every print copy carries the meeting form and signers (charter rules with effective dates: Stage 2)

### US-02: Draft a protocol through the assistant  _(Stage 1)_
**As a** член правління
**I want** to open a справа for an agenda item of the meeting, add documents to it by hand or in the conversation (a revised document may be marked as a new version of an earlier one), and ask the помічник протоколу for a draft, adding what the decision is for
**So that** I start from a filled draft instead of an old text

### US-03: Confirm what the assistant proposed  _(Stage 1)_
**As a** член правління
**I want** each value the помічник proposed to show where it came from as a link and whether it is confirmed, and to confirm or correct it
**So that** no number I did not check reaches the printed protocol

### US-04: Edit the wording by hand  _(Stage 1)_
**As a** член правління
**I want** to edit the protocol text directly in spilno, in the conversation or by hand
**So that** I can fix phrasing without losing the link between numbers and their sources

### US-05: Run an informal preliminary poll  _(Stage 1)_
**As a** член правління
**I want** to put the current wording to an informal попереднє голосування among the board, return it to чернетка and restart the poll as a new round
**So that** we see what passes before printing, without this counting as the meeting's vote

### US-06: Print the protocol for signing  _(Stage 1)_
**As a** член правління
**I want** to render the print copy from the current text once every proposed value is confirmed, with the protocol number reserved
**So that** we sign on paper exactly the text kept in spilno

### US-07: Lock the protocol with its signed copy and the meeting facts  _(Stage 1)_
**As a** член правління
**I want** to upload the signed paper copy and enter the meeting facts written on it, which records the protocol as adopted as signed
**So that** the protocol is locked in the реєстр протоколів with its справа

### US-08: Withdraw a draft that will not be signed  _(Stage 1)_
**As a** член правління
**I want** to close a draft as знятий з розгляду with a reason
**So that** abandoned proposals do not hang open and the history of what was proposed stays

### US-09: Find a past decision with its basis  _(Stage 1)_
**As a** член правління
**I want** to find signed and withdrawn protocols by date, contractor or amount and open each with its signed copy, meeting facts, sources and poll history
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

### AC-45 (US-01) — legal competence
**Given** an association whose settings do not yet hold the meeting form the charter allows for the board (e.g. "засідання за особистої присутності"), who signs the protocol (e.g. "головуючий і секретар") or the number the protocol register continues from
**When** a член правління asks the помічник протоколу for the first protocol draft
**Then** the system refuses, names every missing setting at once and links to the settings; once they are saved, every print copy shows the meeting form and the signers from the settings, and in Stage 1 no state change is blocked by comparing them with the protocol

### AC-03 (US-02) — happy path
**Given** a член правління, a stored meeting with the agenda item "ремонт покрівлі" and a справа for it holding a кошторис робіт of the management company at 184 500 грн
**When** the член правління asks the помічник протоколу for a protocol draft and adds "оплата з фонду ремонту"
**Then** the system creates a draft in state чернетка for that meeting with the agenda item, wording and members filled; the amount shows 184 500 грн as proposed by the помічник and unconfirmed, with a link to the estimate

### AC-04 (US-02) — error
**Given** a член правління with an empty справа
**When** the член правління asks the помічник протоколу for a draft
**Then** the system refuses and says that at least one source document must be attached first

### AC-05 (US-02) — error
**Given** a член правління opening a draft for a meeting that is not stored for the association
**When** the член правління chooses that meeting
**Then** the system says the meeting is not found and creates no draft

### AC-06 (US-03) — traceability
**Given** a draft whose amount 184 500 грн was proposed by the помічник from the кошторис робіт
**When** the член правління opens the draft
**Then** the amount shows a link to the estimate with its number, date and the fragment it was read from, is marked as proposed by the помічник, and the draft cannot be marked готовий до підпису until a person confirms or corrects it

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
**Then** members show the election protocol as of the meeting date as their source, the головуючий is the member a член правління picked, and the draft has no protocol number yet

### AC-43 (US-03) — domain invariant
**Given** a draft whose amount 184 500 грн the помічник proposed from a кошторис робіт and a person confirmed
**When** a член правління adds a revised estimate to the справа and marks it as a new version of the earlier one
**Then** the system marks the wording and the values the помічник took from the earlier version as needing review, shows the confirmed value next to the value in the new version, and counts the amount as unconfirmed again (AC-06); like any change to the draft, this returns a draft in готовий до голосування or готовий до підпису to чернетка (AC-17, AC-26)

### AC-11 (US-04) — happy path
**Given** a draft in state чернетка
**When** a член правління other than the one who created it rewords a sentence by hand without touching any number
**Then** the system keeps the new wording, records who changed it, and every slot keeps its value and source

### AC-11b (US-04) — domain invariant
**Given** two members of the board have opened version 3 of the same draft
**When** the first saves a change and the second then saves a change made to version 3
**Then** the system refuses the second change, says the draft was changed by the first member, and shows version 4 without losing either change

### AC-13 (US-05) — happy path
**Given** a draft in state чернетка
**When** a член правління starts the попереднє голосування
**Then** the current text is frozen as round 1 (or as the next round after a return to чернетка), every member of the board in force sees it with the label of AC-46, and the draft is in state готовий до голосування; values still unconfirmed are shown as unconfirmed and do not prevent the poll

### AC-15 (US-05) — happy path
**Given** a draft in state готовий до голосування, round 2
**When** three members cast "за" for agenda item 1 and the fourth casts "утримався"
**Then** the system shows the round-2 tally 3 "за", 1 "утримався" with each vote's member and time, labelled as informal (AC-46), and gives no verdict on кворум, majority or adoption

### AC-46 (US-05) — legal competence
**Given** a draft in state готовий до голосування, or a protocol whose drafts had poll rounds
**When** any member views the poll, its tally, the print copy or the protocol in the реєстр протоколів
**Then** the poll is labelled "Попереднє внутрішнє опитування правління — не є голосуванням засідання і не замінює його"; its tallies never appear on the print copy, never pre-fill the meeting facts (AC-28), and appear in the реєстр протоколів only as the draft's history

### AC-16 (US-05) — authorization
**Given** a член правління voting on a draft in state готовий до голосування
**When** the член правління tries to cast a vote for another member
**Then** the system refuses and says each member votes only for themselves

### AC-17 (US-05) — happy path
**Given** a draft in state готовий до голосування, round 2, with two votes cast
**When** a член правління edits the wording, the agenda or a value, or chooses «повернути в чернетку»
**Then** the draft is in state чернетка, round 2 is closed with its two votes kept for history and no longer counted, and starting the poll again opens round 3 with no votes (AC-13)

### AC-18 (US-05) — domain invariant
**Given** a draft in state чернетка, готовий до підпису, підписаний or знятий з розгляду
**When** a член правління tries to cast a poll vote
**Then** the system refuses and says the poll is open only while the draft is готовий до голосування

### AC-25 (US-06) — happy path
**Given** a draft in state чернетка with every value the помічник proposed confirmed or corrected by a person
**When** a член правління renders the print copy
**Then** the draft is in state готовий до підпису with its text frozen, a PDF print copy is rendered from the current text with the meeting form and signers from the settings (AC-45) and blank fields for members present and the vote count per item, to fill in by hand at the meeting, and a protocol without a number takes the next free number of the association's protocol register (one sequence, no duplicates)

### AC-25b (US-06) — happy path
**Given** a draft in state готовий до голосування, round 2, with every proposed value confirmed
**When** a член правління renders the print copy
**Then** the draft is in state готовий до підпису as in AC-25, round 2 is closed with its votes kept for history, and the print copy carries no poll result

### AC-26 (US-06) — happy path
**Given** protocol No 14 in state готовий до підпису whose print copy was rendered
**When** a член правління edits the wording, the agenda or a value, or chooses «повернути в чернетку»
**Then** the protocol is in state чернетка, the rendered print copy is marked void and kept in the history, the protocol keeps No 14, and its next print copy carries No 14 again; a number, once taken, is never given to another protocol, so the register may have gaps

### AC-28 (US-07) — happy path
**Given** protocol No 14 in state готовий до підпису, printed and signed on paper at the meeting
**When** a член правління uploads the signed scan and enters the meeting facts written on the paper — the members present, the vote count per agenda item ("за", "проти", "утримався") and whether the paper carries handwritten corrections
**Then** the protocol is in state підписаний and locked; it appears in the реєстр протоколів with No 14, the scan, the meeting facts with who entered them and when, its справа and poll history; in Stage 1 the system does not check кворум or majority against these facts

### AC-47 (US-07) — error
**Given** a protocol in state готовий до підпису
**When** a член правління uploads the signed scan without the members present or without a vote count for every agenda item
**Then** the system refuses, names every missing fact at once, and the state does not change

### AC-44 (US-07) — domain invariant
**Given** a член правління uploading the signed scan of protocol No 14
**When** they mark that the paper carries handwritten corrections
**Then** the system requires a note describing each correction, keeps it with the scan, locks the protocol as in AC-28, and shows the note next to the frozen text in the реєстр протоколів; the frozen text is not changed, and the signed paper with its corrections prevails

### AC-29 (US-07) — error
**Given** a protocol in state готовий до підпису
**When** a член правління uploads a file above the size limit or of a type that is not allowed (§6)
**Then** the system refuses, names the allowed types and the size limit, and the state does not change

### AC-30 (US-07) — domain invariant
**Given** a protocol in state підписаний
**When** a член правління tries to change any value, the wording, the meeting facts or the scan
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

### AC-33 (US-07) — error
**Given** a член правління of the association
**When** the член правління opens a protocol or справа by an identifier that is not stored for the association
**Then** the system answers that the protocol is not found and changes nothing

### AC-34 (US-08) — happy path
**Given** a draft in state готовий до голосування whose poll did not gather support
**When** a член правління closes it as знятий з розгляду with the reason "не набрано підтримки, переглянути кошторис"
**Then** the draft is in state знятий з розгляду, keeps its справа, versions and poll rounds, is never printed, and takes no protocol number

### AC-34b (US-08) — happy path
**Given** a draft in state чернетка that the board decided not to pursue
**When** a член правління closes it as знятий з розгляду with the reason "управляюча компанія відкликала кошторис"
**Then** the draft is in state знятий з розгляду with its справа and versions kept, and takes no protocol number

### AC-34c (US-08) — happy path
**Given** protocol No 14 in state готовий до підпису that was printed but not signed
**When** a член правління closes it as знятий з розгляду with the reason "не підписано на засіданні"
**Then** the protocol is in state знятий з розгляду, its frozen text and print copy are kept marked as never signed, and it keeps No 14, which is never given to another protocol

### AC-35 (US-08) — domain invariant
**Given** a draft in state знятий з розгляду
**When** a член правління tries to edit it or start a poll on it
**Then** the system refuses and offers to start a new draft copied from it

### AC-36 (US-09) — happy path
**Given** a член правління and a реєстр протоколів holding 14 signed and 3 withdrawn protocols for 2026
**When** the член правління searches by the contractor "управляюча компанія" and the year 2026
**Then** the system lists the matching protocols with number, date, amount and state, and each opens with its signed copy and meeting facts (none for withdrawn), справа documents, slot sources and poll history labelled as informal

### Deferred to Stage 2

These keep their ids and Stage 1 text as approved on 2026-10-24, so traceability survives; each is re-validated when Stage 2 is specified (§3).

#### AC-02 (US-01) — traceability  _(Stage 2: charter-rule effective dates)_
**Given** a правило статуту "кворум: 3 members" in force from 2025-01-01 and a new one "кворум: 4 members" recorded with effect from 2026-11-01
**When** a член правління prepares a protocol for a meeting held on 2026-10-20
**Then** the protocol checks кворум against 3 members, names the rule in force on the meeting date, and the later rule is not applied

#### AC-09 (US-03) — domain invariant  _(Stage 2: source of truth as a blocking rule)_
**Given** a draft with a decision slot (amount, counterparty or джерело коштів) that lacks either a document of the справа or a person's confirmation
**When** the член правління tries to mark it готовий до підпису
**Then** the system blocks, names the slot and says the value has no confirmed source; an обхід блокування does not lift this check, only a document plus a person's confirmation does

#### AC-10 (US-03) — cross-context  _(Stage 2: cross-check with finance contracts)_
**Given** the association's financial records hold a maintenance contract in force and a repair-fund contract that ended on 2026-06-30
**When** the член правління picks the джерело коштів for a decision of a meeting held on 2026-10-20
**Then** only the contract in force on the meeting date is offered, and the ended one is refused with its end date named

#### AC-12 (US-04) — domain invariant  _(Stage 2: source of truth as a blocking rule)_
**Given** a draft whose amount slot is 184 500 грн
**When** the член правління edits the text by hand so that it reads 185 400 грн
**Then** the system marks the mismatch between text and slot, and the draft cannot become готовий до голосування or готовий до підпису until text and slot agree

#### AC-14 (US-05) — domain invariant  _(dropped from Stage 1: the poll is informal)_
**Given** a draft with an amount still proposed and unconfirmed
**When** a член правління marks it готовий до голосування
**Then** the system refuses and names the unconfirmed slot

#### AC-15b (US-05) — traceability  _(dropped from Stage 1: the poll never pre-fills the meeting facts, AC-46)_
**Given** a draft for a meeting where 4 members were present
**When** the член правління records the meeting's vote count for agenda item 1 as 3 "за", 1 "утримався"
**Then** the slot shows the count with the member who entered it and when, is pre-filled from the попереднє голосування when there was one, and marks every place where the entered count differs from the preliminary votes

#### AC-15c (US-05) — happy path  _(Stage 2: кворум and majority checks)_
**Given** a draft in state готовий до голосування or чернетка with no attendance or no meeting vote count recorded for agenda item 1
**When** a член правління marks it готовий до підпису
**Then** the draft becomes готовий до підпису with a попередження that кворум and majority for item 1 could not be checked, and the printed copy leaves the attendance and vote count blank to fill in by hand at the meeting

#### AC-19 (US-06) — domain invariant  _(Stage 2: blocking checks)_
**Given** a draft whose attendance shows 2 of 4 members and a правило статуту кворум of 3 in force on the meeting date
**When** the член правління tries to mark it готовий до підпису
**Then** the system blocks, names кворум as not reached with the rule and its effective date, and no item is shown as adopted

#### AC-19b (US-06) — domain invariant  _(Stage 2: blocking checks)_
**Given** 3 of 4 members present, a правило статуту кворум of 3, a rule that a member with неучасть у голосуванні does not count toward that item's кворум, and one member recording неучасть у голосуванні on item 1 for personal interest
**When** the член правління tries to mark the draft готовий до підпису
**Then** the system blocks item 1 as without кворум, names both rules, and item 2, where all three vote, is not blocked

#### AC-19c (US-06) — domain invariant  _(Stage 2: blocking checks)_
**Given** a правило статуту requiring 3 votes "за" to adopt a decision, and agenda item 1 with a recorded meeting vote count of 2 "за", 2 "проти"
**When** the член правління tries to mark the draft готовий до підпису with item 1 worded as adopted
**Then** the system blocks, names the majority rule, and lifts the block only when item 1 is worded as not adopted, removed from the agenda, or an обхід блокування is recorded with a reason

#### AC-20 (US-06) — legal competence  _(Stage 2: blocking checks)_
**Given** a правило статуту setting the board's competence limit at 100 000 грн in force on the meeting date
**When** the член правління tries to mark ready to sign a draft approving 184 500 грн
**Then** the system blocks and names the limit rule, saying the decision needs the general meeting or an обхід блокування with a reason

#### AC-21 (US-06) — legal competence  _(Stage 2: blocking checks)_
**Given** a правило статуту allowing only meetings held in person, and a draft naming the meeting form "заочне (опитування в чаті)"
**When** the член правління tries to mark it готовий до підпису
**Then** the system blocks and names the rule on allowed meeting forms

#### AC-21b (US-06) — legal competence  _(Stage 2: blocking checks)_
**Given** no правило статуту on allowed meeting forms, кворум, majority, the competence limit or whether a non-voting member counts toward an item's кворум is recorded for the meeting date
**When** the член правління marks a draft готовий до підпису
**Then** the draft becomes готовий до підпису with a попередження naming each missing rule and saying that check could not be made, which does not stop printing

#### AC-22 (US-10) — legal competence  _(Stage 2: needs typed agenda items)_
**Given** a draft whose agenda item, of a type the charter reserves for the general meeting, allots part of the common premises to the association, a decision the charter reserves for the general meeting
**When** the член правління tries to mark it готовий до підпису
**Then** the system blocks and names the charter rule that reserves the decision for the general meeting; an обхід блокування with a reason lifts it

#### AC-23 (US-06) — happy path  _(Stage 2: overrides)_
**Given** a draft in state готовий до голосування with a blocking competence-limit check
**When** a член правління records an обхід блокування with the reason "аварійний ремонт, загальні збори скликаються 2026-11-15" and marks it готовий до підпису
**Then** the draft is in state готовий до підпису, the text is frozen for printing, and the override with its author, time and reason stays with the protocol

#### AC-24 (US-06) — error  _(Stage 2: overrides)_
**Given** a draft with a blocking check
**When** a член правління records an обхід блокування without a reason
**Then** the system refuses and says a reason is required

#### AC-27 (US-06) — traceability  _(Stage 2: blocking checks and warnings)_
**Given** a draft whose головуючий is a member other than the board chair and the справа holds no service note for that meeting
**When** the член правління marks it готовий до підпису
**Then** the draft becomes готовий до підпису with a попередження that the головуючий has no service note as a source, which does not stop printing

#### AC-28b (US-07) — traceability  _(Stage 2: reading the signed scan with AI)_
**Given** a protocol No 14 dated 2026-10-20 in state готовий до підпису
**When** a член правління uploads a signed scan on which the number reads 15 or the date reads 2026-10-21
**Then** the protocol becomes підписаний and shows a попередження naming the number or date the помічник read from the scan next to the protocol's own, which does not stop the upload; when the помічник cannot read them, the попередження says the scan was not checked

#### AC-38 (US-10) — domain invariant  _(Stage 2)_
**Given** an agenda-item type with a шаблон формулювання
**When** the помічник протоколу or a член правління changes the fixed text of that item
**Then** the system refuses the change and allows only the template's slots to be filled

#### AC-39 (US-11) — cross-context  _(Stage 2)_
**Given** a чинне рішення of the board approving the same contractor's estimate for the same building within the last 90 days
**When** the член правління prepares a new draft for that contractor and building
**Then** the system shows a попередження linking the earlier decision and its amount, which does not stop printing

#### AC-40 (US-11) — legal competence  _(Stage 2)_
**Given** two drafts for the same estimate at 60 000 грн and 70 000 грн and a competence limit of 100 000 грн
**When** the член правління marks the second one готовий до підпису
**Then** the system shows a попередження that the same estimate totals 130 000 грн across protocols, above the limit

#### AC-41 (US-12) — happy path  _(Stage 2)_
**Given** a signed protocol approving 184 500 грн and an акт виконаних робіт for 184 500 грн and a bank statement for 184 500 грн
**When** a член правління attaches both to the decision
**Then** the decision shows the expense line as closed with links to the act and the statement

#### AC-42 (US-12) — traceability  _(Stage 2)_
**Given** a signed protocol approving 184 500 грн
**When** a член правління attaches an акт виконаних робіт for 196 000 грн
**Then** the expense line stays open and shows a попередження naming both amounts and their documents

## 5.1 Error catalogue

| Code | Kind | Message to the actor | AC |
|---|---|---|---|
| `governance.association_settings_missing` | validation | Спершу заповніть налаштування об'єднання: форму засідання за статутом, хто підписує протокол і номер, з якого продовжується реєстр. | AC-45 |
| `governance.source_document_required` | validation | Спершу додайте до справи хоча б один документ. | AC-04 |
| `governance.meeting_not_found` | not_found | Засідання не знайдено. | AC-05 |
| `governance.protocol_not_found` | not_found | Протокол не знайдено. | AC-33 |
| `governance.amount_unconfirmed` | invariant | Значення запропоноване помічником і ще не підтверджене людиною. | AC-06 |
| `governance.draft_changed` | conflict | Чернетку щойно змінив інший член правління; перегляньте нову версію. | AC-11b |
| `governance.vote_on_behalf` | authorization | Кожен член правління голосує лише за себе. | AC-16 |
| `governance.voting_not_open` | invariant | Попереднє опитування відкрите лише в стані «готовий до голосування». | AC-18 |
| `governance.meeting_facts_required` | validation | Внесіть із підписаного примірника присутніх і підсумок голосування за кожним пунктом. | AC-47 |
| `governance.corrections_note_required` | validation | Опишіть рукописні виправлення на підписаному примірнику. | AC-44 |
| `governance.signed_copy_invalid` | validation | Дозволені PDF, JPEG або PNG до 20 МБ. | AC-29 |
| `governance.protocol_locked` | invariant | Підписаний протокол змінити або зняти не можна; створіть протокол-виправлення. | AC-30, AC-30b |
| `governance.transition_not_allowed` | invariant | Цей перехід недоступний у поточному стані протоколу. | AC-31, AC-35 |
| `governance.not_board_member` | authorization | Дія доступна лише члену правління з чинними повноваженнями. | AC-32 |

Deferred to Stage 2 with their ACs (§3): `governance.source_missing` (AC-09), `governance.funding_source_not_in_force` (AC-10), `governance.text_slot_mismatch` (AC-12), `governance.quorum_missing` (AC-19, AC-19b), `governance.majority_missing` (AC-19c), `governance.outside_board_competence` (AC-20), `governance.meeting_form_not_allowed` (AC-21), `governance.reserved_for_general_meeting` (AC-22), `governance.reason_required` (AC-24), `governance.template_text_fixed` (AC-38).

## 6. Non-functional requirements

| Aspect | Target | Measurement |
|---|---|---|
| Latency p95, read or list of the association's records | ≤ 200 ms (default) | server request-duration histogram per route |
| Latency p95, write without file | ≤ 500 ms (default) | same histogram |
| Latency p95, upload of a signed scan | ≤ 3 s for 20 MB (default) | same histogram, upload routes |
| Print copy (PDF) rendering | ≤ 5 s p95 (default) | request duration |
| Draft from the помічник протоколу | ≤ 30 s p95 from request to filled draft | assistant request duration |
| Availability | 99.5 % monthly (default) | uptime check on the health endpoint |
| Durability | RPO ≤ 24 h (nightly backup), RTO ≤ 4 h (default) | nightly backup log + quarterly restore drill |
| File size | ≤ 20 MB per file, PDF / JPEG / PNG (default) | upload validation |
| Retention | signed and withdrawn protocols, print copies and scans kept ≥ 5 years, never hard-deleted (default) | storage policy |
| Integrity | a signed protocol changes only through a new referencing protocol: 0 changes to its text, meeting facts or scan after lock (default + feature) | integrity check job |

## 6.1 Security / privacy

- **Data classification:** Confidential — internal legal and financial records of the association, with personal data.
- **Personal data touched:** names and terms of board members (low); signatures on scans (medium); names of co-owners and apartment numbers that may appear in case documents and owner requests (medium); handwritten-correction notes that may name people (medium).
- **Authorization impact:** new rights: any член правління with a term in force acts on the association's protocols; a person whose term ended has no access. One association in Stage 1. Depends on sign-in, which the app does not have yet (§8).
- **Abuse cases:**
  - Former member access: a person whose term ended opens or changes a protocol — refused (AC-32); an identifier not stored for the association answers not found (AC-33).
  - Editing a signed record: a member tries to change a signed protocol to match a later bill — refused; correction only through a new referencing protocol (AC-30).
  - The poll passed off as the meeting's vote: a poll tally is shown as an absentee board decision — the poll is labelled informal everywhere, never printed and never pre-fills the meeting facts (AC-46).
  - Leaking personal data to the text-generation service: case documents with co-owners' names are sent to the помічник — sent only under the agreement decided in §8; nothing is used for training.
  - Instructions hidden in a document: a case document contains text meant to steer the draft — the помічник only proposes; every value it proposed stays unconfirmed until a person confirms it (AC-06).
  - Uploading a forged or wrong scan: the uploaded copy differs from the frozen text — the scan, the meeting facts, the uploader and the time are recorded and visible to board members in the реєстр протоколів; reading the scan is Stage 2 (AC-28b).
  - Votes on behalf of others: a member votes for an absent colleague — refused (AC-16).
- **Security review:** Required — first feature with authorization, personal data and an external text-generation service.

## 7. Metrics / KPIs

- **Time per protocol, from estimate to print** — baseline: 1-3 h (brief §2, author's estimate), target: ≤ 20 min median within 3 months of Stage 1 release.
- **Lag from board decision to locked protocol** — baseline: weeks (brief §2, not measured), target: ≤ 7 days median within 3 months of Stage 1.
- **Spending decisions with a locked protocol before the annual meeting** — baseline: not measured, target: 100 % of 2027 spending decisions by the 2027 annual meeting.
- **AI-proposed amounts corrected by a person** — baseline: none, target: measured on every protocol; ≤ 5 % corrected over one season (brief §14 revisit trigger for Approach B).
- **Protocols with an in-system poll** — baseline: 0, target: ≥ 50 % of signed protocols within 6 months of Stage 1.

## 8. Open questions

- [ ] Which meeting forms does the charter allow for the board (in person, absentee, mixed), and which clause? Default now: a член правління records the allowed form in the association settings before the first protocol and it is printed on every copy (AC-45); the check is Stage 2 (AC-21). — owner: board chair, due: before Stage 1 release
- [ ] Is the 100 000 грн limit per protocol, per job or per estimate; which clause? Default now: per decision line; the limit check is Stage 2 (AC-20, AC-40). — owner: board chair, due: before Stage 2
- [ ] On a 4-member board, is "3 votes for" the кворум, the majority, or both? Default now: two separate правила статуту, кворум 3 and majority 3, stored in the settings; checks are Stage 2. — owner: board chair, due: before Stage 2
- [ ] At what overrun of the act over the approved amount is a new decision needed? Default now: any overrun keeps the line open with a warning (AC-42). — owner: board chair + accountant, due: before Stage 2
- [ ] Will the management company hand over acts and statements per expense line? Default now: attached by a член правління. — owner: Andrii Opryshko, due: before Stage 2
- [ ] How does a член правління sign in, given the app has no sign-in yet? Default now: a prerequisite decided in architecture. — owner: Tech Lead, due: before /sdlc:architecture-design
- [ ] The app stores no association yet; are existing meetings and contracts assigned to the association? Default now: yes, to the single association, confirmed by a член правління. — owner: Tech Lead, due: before /sdlc:architecture-design
- [ ] May case documents with personal data go to a text-generation service outside the association's hosting? Default now: only under an agreement that excludes training on the data. — owner: Tech Lead, due: before /sdlc:architecture-design
- [x] Where is the wording edited by hand? Decided 2026-10-25 by the product owner: in spilno; the editor is an architecture choice; matching numbers in the text against slots is Stage 2 (AC-12). — owner: Tech Lead
- [ ] Do the financial records hold each contract's джерело коштів? Default now: in Stage 1 the джерело коштів is wording the помічник proposes and a person confirms, linked to a contract where one is in the справа; a fund line on contracts comes with AC-10 in Stage 2 (§9). — owner: Tech Lead, due: before Stage 2
- [ ] Time the next 2-3 protocols to replace the estimated KPI baselines (1-3 h per protocol, weeks of lag). Default now: author's estimate from the brief. — owner: Andrii Opryshko, due: before Stage 1 release
- [ ] How does the помічник read the protocol number and date from a signed scan (AC-28b), and how reliable must it be? Default now: not read in Stage 1; best effort in Stage 2. — owner: Tech Lead, due: before Stage 2
- [ ] Confirm Reach: 13 associations at about 20 protocols each. — owner: Andrii Opryshko, due: before roadmap scoring

## 9. Migration plan

- **New records:** association settings (allowed meeting form, signers, the number the protocol register continues from, the current правила статуту without effective dates); board terms with their election protocols; частки будинку with the inter-building decision; справи with source documents and document versions; protocols with wording versions, слоти рішення and their sources and confirmations, poll rounds with their votes, print copies with the reserved number (void ones kept), signed copies with the meeting facts and the handwritten-corrections note. Each protocol links to exactly one meeting already stored by `governance`, and a meeting has at most one protocol; each agenda item of the protocol has its own справа.
- **Existing meetings (`governance`):** untouched. A meeting may have no protocol; a protocol is created for an existing meeting or together with a new one. The meeting's quorum-required flag stays. No backfill. Members present and vote counts are recorded on the protocol at upload, not on the meeting, so existing meetings get nothing filled in.
- **Existing contracts (`finance`):** untouched in Stage 1; a contract may be a document of a справа. The fund line on contracts and the in-force check come with AC-10 in Stage 2.
- **Association scope:** one association; the association record is created once and existing meetings and contracts are assigned to it, confirmed by a член правління (§8). Records carry the association, but no multi-association scope is built.
- **Past paper protocols:** not imported (§3); the settings hold the number the register continues from (AC-45).
- **Reversibility:** rolling back Stage 1 or Stage 2 removes the screens and actions; stored protocols, print copies, signed copies, meeting facts and poll rounds are kept read-only. Signed and withdrawn protocols are never deleted.
- **Rollout:** Stage 0 (prerequisite): sign-in and the association record. Stage 1: US-01–US-09 as in §5 (without the ACs deferred to Stage 2); before the first protocol a член правління records the board, shares and association settings (AC-45). Stage 2: the deferred ACs of §5 and US-10–US-12; between stages users see no blocking checks, overrides, templates, related-decision warnings or expense closure, and drafts behave as in Stage 1.

## 10. RBAC and state machine

### Roles and actions

| Role | Keep board and association settings | Create / edit draft, ask the помічник | Start the poll / return to чернетка | Cast a poll vote | Render the print copy (готовий до підпису) | Upload signed copy and meeting facts | Withdraw | Read signed / withdrawn | Read drafts |
|---|---|---|---|---|---|---|---|---|---|
| член правління (term in force) | ✓ | ✓ | ✓ | own only | ✓ | ✓ | ✓ | ✓ | ✓ |
| головуючий | ✓ (as член правління) | ✓ | ✓ | own only | ✓ | ✓ | ✓ | ✓ | ✓ |
| ревізійна комісія | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ (later feature) | ✗ |
| член правління (term ended) | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ |
| співвласник | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ |
| управляюча компанія | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ |

Scope is one association. Rights to act come from a board term in force on the date of the action; who may cast a poll vote is taken from the board in force on that date; members shown in the draft come from the board on the meeting date.

### Slot sources

| Slot | Counts as sourced when |
|---|---|
| amount, counterparty, джерело коштів and any other value the помічник proposed | a person's confirmation or correction; the document it was read from is shown as a link (a blocking source rule is Stage 2, AC-09) |
| members, protocol number, частки будинку, meeting form, signers | the reference record (election protocol, protocol register, inter-building decision, association settings); no confirmation needed |
| members present, vote count per item, handwritten-corrections note | a член правління's entry from the signed paper at upload, recorded with who and when |

### States

- **чернетка** — the draft is drafted by the помічник and edited in spilno by any current член правління; no protocol number until the first print copy.
- **готовий до голосування** — the current text is frozen as a numbered poll round; members cast an informal попереднє голосування, labelled as not the meeting's vote (AC-46).
- **готовий до підпису** — every proposed value confirmed; text frozen; the PDF print copy rendered; the protocol number reserved.
- **підписаний** — the signed scan and the meeting facts are uploaded; the protocol is locked and immutable in the реєстр протоколів.
- **знятий з розгляду** — closed with a reason without being signed; kept as an internal record; a number it took is never reused.

### Allowed transitions

| From → To | Who | Condition | AC |
|---|---|---|---|
| (none) → чернетка | член правління | association settings filled (AC-45); справа has at least one source document; meeting exists | AC-03 |
| чернетка → готовий до голосування | член правління | none; opens round 1 or the next round | AC-13 |
| готовий до голосування → чернетка | член правління | an edit of wording, agenda or a value, or «повернути в чернетку»; the round is closed, its votes kept for history and not counted | AC-17 |
| чернетка → готовий до підпису | член правління | every value the помічник proposed confirmed or corrected; print copy rendered; number reserved if none | AC-25 |
| готовий до голосування → готовий до підпису | член правління | as above; the poll round is closed; no poll result is printed | AC-25b |
| готовий до підпису → чернетка | член правління | an edit or «повернути в чернетку»; the print copy is void; the number stays with the protocol | AC-26 |
| готовий до підпису → підписаний | член правління | signed scan uploaded with members present and a vote count for every item (and a note if there are handwritten corrections) | AC-28 |
| чернетка → знятий з розгляду | член правління | reason given | AC-34b |
| готовий до голосування → знятий з розгляду | член правління | reason given | AC-34 |
| готовий до підпису → знятий з розгляду | член правління | reason given; print copy kept as never signed; number never reused | AC-34c |

### Refused transitions and actions

| From → To / action | Why refused | AC |
|---|---|---|
| підписаний → any edit (values, wording, meeting facts, scan) | a signed protocol is immutable; corrections go through a new referencing protocol | AC-30 |
| підписаний → знятий з розгляду | a signed protocol cannot be withdrawn | AC-30b |
| cast a poll vote in any state but готовий до голосування | the poll is open only on a frozen round | AC-18 |
| чернетка / готовий до голосування → підписаний | a signed copy is accepted only for a printed text | AC-31 |
| чернетка / готовий до голосування → готовий до підпису with a value the помічник proposed and nobody confirmed | no AI-proposed number reaches print unconfirmed | AC-06 |
| готовий до підпису → підписаний without the meeting facts | the facts on paper must be recorded with the scan | AC-47 |
| знятий з розгляду → any | final; a new draft is copied instead | AC-35 |

Blocking checks, overrides and their table of which checks an обхід блокування lifts are Stage 2 (§3; AC-09, AC-12, AC-19–AC-24, AC-27).
