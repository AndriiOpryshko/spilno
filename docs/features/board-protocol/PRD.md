---
status: Approved
owner: "Andrii Opryshko"
reviewers: ["Andrii Opryshko"]
updated_at: "2026-10-25"
feature_size: "L"
stage: "03"
ticket: ""
---

# PRD — board-protocol

> **Inputs (required):** [idea-brief](./idea-brief.md) · [CONTEXT](./CONTEXT.md)
> **Reference module:** `services/core/internal/governance` — board meeting entity (kind regular / extraordinary, held date, optional end and location, a quorum-required flag), validation that reports every broken field at once, "meeting not found"; no authorization and no lifecycle statuses yet.
> **External context channels used:** Project docs: `docs/adr/0001`–`0004` (bounded contexts: governance and finance are separate contexts; contracts live in finance) · Prototype UI: `apps/web/src/features/governance/` (commit 8ad3124: cases list, protocol draft, slot origin and status, advisory vs blocking warnings, readiness gate before printing).

## 1. Context

A board chair of a young association has produced 30+ board protocols in under two years, each assembled by hand from contracts, invoices and chat approvals. Post-signature corrections and re-votes have already happened; numbers in protocols have no link to their source, and nothing reminds the author of the charter, voting restrictions or prior decisions (idea-brief §2). The primary user is the головуючий — usually the chair, a volunteer and not a lawyer, sometimes a substitute — writing ~15-20 protocols a year; board composition changes from meeting to meeting (§3).

The association is under two years old and already has a correction on record; the app skeleton, a board-meeting record and a protocol-draft UI prototype already exist. Rollout is the author's own association first, with charter rules kept as data so neighbouring associations can follow (§4).

Committed approach (idea-brief §13): **Approach C, staged**. After the product owner's rescope of 2026-10-25, **Stage 1** is the assistant's draft and the keeping of the document. A член правління describes in a conversation what the protocol should decide and attaches the case documents for the meeting's agenda; the system proposes a protocol with recommended numbers, each linked to the case document it came from and unconfirmed until a person confirms or corrects it (US-12, US-09). Any current член правління reviews and edits the text in spilno. The board may run an informal попереднє голосування on the wording frozen as готовий до голосування, only to see what passes — it is not the meeting's vote and not a substitute for it (AC-41) — and may return the draft to чернетка and restart the poll as a new round. Marking готовий до підпису renders the print copy and reserves the protocol number; the board signs on paper, and a член правління uploads the signed scan with the meeting facts written on the paper, which locks the protocol; a draft that will not be signed is withdrawn as знятий з розгляду with a reason. Stage 1 lifecycle and its ACs: (none) → чернетка AC-26 · чернетка → готовий до голосування AC-30 · готовий до голосування → чернетка AC-05d · чернетка → готовий до підпису AC-14c · готовий до голосування → готовий до підпису AC-33 · готовий до підпису → чернетка AC-14b · готовий до підпису → підписаний AC-14 · чернетка / готовий до голосування / готовий до підпису → знятий з розгляду AC-35 / AC-36 / AC-37; refused: any change after lock AC-17, a poll vote outside готовий до голосування AC-32. **Stage 2** brings the deterministic checks with their overrides (quorum, majority, competence limit, source of every number as a blocking rule, finance cross-checks), charter rules with effective dates, locked шаблони формулювання, попередження and related чинні рішення, reading the signed scan, and the export for the ревізійна комісія (§3).

Traceability and deliberate decisions:
- Reference module: the existing board-meeting record is the anchor the protocol hangs on; its "report every broken field at once" validation style carries over to drafts, roster and settings.
- Prototype: slot origin (system / document / human) and status (confirmed / needs review / not found); advisory notices never block, deterministic checks block printing — this PRD keeps that split for Stage 2.
- Decision override (superseded 2026-10-25): first release = Stage 1 + Stage 2, while idea-brief §13 locked Stage 1 only. The rescope below redraws Stage 1 instead.
- Decision override: conversational drafting with recommended numbers is in the first release and the size moved from M to L — this partly revives approach B (idea-brief §7, parked in §14 until 2+ associations and lawyer-reviewed templates) without waiting for its trigger. Rationale: it is how the author wants to work with the system; the reasons B was parked are excluded here — every recommended number waits for human confirmation (AC-21) and generated text holds no fact outside the slots (AC-28).
- Decision override (Stage 2 since 2026-10-25): deterministic блокуючі перевірки stop a draft from being marked готовий до підпису, while idea-brief §5/§14 rejected blocking — rationale: the rejection targets blocking on automated legal judgments; checks here are arithmetic against правила статуту, and an обхід блокування with a recorded reason keeps the emergency path open (AC-05b, AC-05c, AC-06, AC-12).
- Deliberate simplification (clarify-prd, 2026-10-24): the draft lifecycle and the signed scan get a naive first cut — an uploaded signed scan means the protocol was adopted (AC-14); reading the scan is Stage 2 (AC-14d); stricter rules can follow once the first protocols run through the system.
- Stage 1 rescoped by product owner, 2026-10-25: «суть — AI помічник, який складе чернетку, ну і саме ведення документу. Не обов'язково перевірки всякі робити на цьому етапі. Головне чернетка: я перевірив, може підправив; внутрішнє голосування чисто щоб глянути що заходить; можна вернутися назад у чернетку та перезапустити голосування. Потім я друкую, підписуємо і закидую скан.» Consequences: conversational drafting and pre-fill (US-12, US-09) move into Stage 1; blocking checks, overrides, dated rules, scan reading and the export move to Stage 2 (§3); any current член правління edits a draft (AC-03b dropped); withdrawal is in scope (US-13); a handwritten correction on the signed paper is recorded as a note and does not stop the lock (AC-15); the protocol number is reserved at the first print copy and never reused (AC-14c, AC-14b, AC-37).

## 2. Goals

- A член правління gets a printable protocol by reviewing and correcting a draft the system proposed from a conversation and the case documents, instead of assembling it; no number the system proposed reaches print without a person's confirmation (Stage 1).
- Any current член правління finds a past decision with its signed copy, meeting facts, poll history and cited documents in the реєстр протоколів, instead of searching chats (Stage 1).
- Quorum and majority are computed from the roster and правила статуту in force on the meeting date, and the головуючий is reminded of required wording, possible personal interest and related чинні рішення — without the system making legal judgments (Stage 2).

## 3. Non-goals

- General-meeting protocols and co-owner e-voting — covered by existing services (idea-brief §5).
- Running the vote in the system (live meeting, remote participation, written polls, signing inside the system) — the vote and signing happen outside; the in-system попереднє голосування is an informal internal poll, labelled as not the meeting's vote and not a substitute for it (AC-41), and the signed result is accepted as a scan.
- Access for co-owners or the ревізійна комісія, and onboarding other associations — the first release serves one association (records carry the association, but there is no multi-association scope) and its current board (idea-brief §5).
- Legal advice or interpretive blocking — only deterministic checks block, and those come in Stage 2; anything needing interpretation is a попередження (idea-brief §5).
- Files signed with a qualified electronic signature, and checking who signed — the first release accepts signed scans only; the uploader enters the meeting facts written on the paper (AC-14). Signature upload and validation come later.
- The autonomous part of approach B — the system never confirms a number, removes a member from a vote or blocks on its own interpretation; approach B stays parked (idea-brief §14).
- **Deferred to Stage 2** (product owner, 2026-10-25: Stage 1 is the assistant's draft and keeping the document; the legal checks stay with the board on paper, as today). The ACs keep their ids and text in §5 «Deferred to Stage 2»:
  - Blocking checks and overrides — quorum and majority (AC-05, AC-05b, AC-05c, AC-06), attendance and non-voting rules (AC-07, AC-08), competence limit (AC-11), override with a reason (AC-12, AC-12b, AC-13) — why: they need the charter answers in §8 and would stop the first protocols on missing data.
  - Source of truth as a blocking rule — a number without a confirmed source document (AC-10, AC-27) — why: in Stage 1 every number shows its source document as a visible link and only the system's proposals need a person's confirmation (AC-21).
  - Charter rules and roster with effective dates and rule-change references (AC-01, AC-02b, AC-02c), and snapshots of rules and documents on a locked protocol (AC-17b) — why: Stage 1 stores only the current rules in the association settings (AC-42).
  - Reading the signed scan (AC-14d) — why: the uploader enters the facts from the paper.
  - The export package for the ревізійна комісія (US-08, AC-20) — why: comes together with the commission's own access.
  - Revoking or replacing a чинне рішення through a reference (AC-17c), required wording, warnings and related decisions (US-10, US-11, AC-23, AC-23b, AC-24, AC-25, AC-25b) — why: Stage 2 as before.
  - Dropped, not deferred: only the головуючий edits a draft (AC-03b) — product owner: any current член правління edits (AC-29).

## 4. User stories

### US-01: Maintain roster and association settings  _(Stage 1)_
**As a** член правління
**I want** to record the board roster with each member's term, and the association settings — the meeting form the charter allows, who signs the protocol, the current правила статуту and the number the protocol register continues from
**So that** drafts show the roster in force on the meeting date and every print copy carries the meeting form and signers (rules and шаблони формулювання with effective dates: Stage 2)

### US-02: Open a protocol draft from a case  _(Stage 1)_
**As a** член правління
**I want** a protocol draft for a board meeting built from one or more справи, with its agenda items, that any current член правління can edit
**So that** the protocol is built from the documents that justify it

### US-03: Run an informal preliminary poll  _(Stage 1)_
**As a** член правління
**I want** to freeze the current wording as готовий до голосування, have every member cast their own informal попереднє голосування on each agenda item, and return the draft to чернетка and restart the poll as a new round
**So that** we see what passes before printing, without it counting as the meeting's vote

### US-04: Link every number to a case document  _(Stage 1)_
**As a** член правління
**I want** every amount, counterparty and джерело коштів in a decision to show the case document it came from as a link, and to be told when that document gets a newer version
**So that** any number in the protocol can be traced to its source later (a block on a number without a source: Stage 2)

### US-05: Resolve blocking checks before locking  _(Stage 2)_
**As a** член правління
**I want** to see every блокуюча перевірка before locking, and record an обхід блокування with a reason when an emergency requires it
**So that** I don't lock an invalid protocol by accident and am never stuck

### US-06: Print and lock the protocol with its signed copy  _(Stage 1)_
**As a** член правління
**I want** to mark the final protocol готовий до підпису to render the print copy with a reserved number, sign it on paper, and upload the signed scan with the meeting facts written on it to lock it
**So that** the реєстр протоколів holds the legally meaningful copy and the text can no longer change

### US-07: Find a past decision with its basis  _(Stage 1)_
**As a** член правління
**I want** to search the реєстр протоколів and open a decision with its signed copy, meeting facts, poll history and cited documents
**So that** I can answer an auditor or a co-owner within a minute

### US-08: Export a protocol for the audit commission  _(Stage 2)_
**As a** член правління
**I want** to export a locked protocol with its signed copy and cited case documents as one package
**So that** the ревізійна комісія can check it without access to the system

### US-09: Pre-fill decision slots from documents  _(Stage 1)_
**As a** член правління
**I want** слоти рішення pre-filled from the case documents and marked unconfirmed until a person confirms each one
**So that** I don't copy sums and names by hand, and nothing unconfirmed slips into a printed protocol

### US-10: Use the required wording template  _(Stage 2)_
**As a** головуючий
**I want** an agenda item of a type that has a шаблон формулювання to use that fixed wording, with only its slots editable
**So that** wording the law or charter requires is never changed by accident

### US-11: See warnings and related decisions  _(Stage 2)_
**As a** головуючий
**I want** to see попередження (possible personal interest, relevant charter clauses) and related чинні рішення alongside the draft, each linked to its source
**So that** I'm reminded of what I may not know or remember, and decide myself

### US-12: Draft a protocol through a conversation  _(Stage 1)_
**As a** член правління
**I want** to describe in a conversation what the protocol should decide, attach case documents, get a proposed protocol with recommended numbers, and refine it through the conversation or by hand
**So that** I review and correct a protocol instead of assembling it

### US-13: Withdraw a draft that will not be signed  _(Stage 1)_
**As a** член правління
**I want** to close a draft as знятий з розгляду with a reason
**So that** abandoned proposals do not hang open and the history of what was proposed stays

## 5. Acceptance criteria

### AC-42 (US-01) — domain invariant
**Given** an association whose settings do not yet hold the meeting form the charter allows for the board, who signs the protocol, or the number the protocol register continues from
**When** a член правління asks for the first protocol draft
**Then** the system refuses, names every missing setting at once and links to the settings; once they are saved, every print copy shows the meeting form and the signers from the settings, and in Stage 1 they block nothing by being compared with the protocol

### AC-02 (US-01) — error
**Given** a член правління adding a person to the board roster
**When** they save the entry without a term start, or with a term end before the term start
**Then** the system rejects the entry and names every missing or inconsistent field at once

### AC-26 (US-12) — happy path
**Given** a член правління, filled association settings (AC-42), a recorded board meeting with its agenda and a справа holding a contract and an invoice
**When** they ask in the conversation to approve the roof waterproofing contract from the repair fund and attach both documents
**Then** the system creates a protocol draft for that meeting with that agenda item and its слоти рішення filled with recommended values, each marked unconfirmed with its source fragment and a link to the document, and keeps the request, the conversation and every attached document with the draft

### AC-03 (US-02) — traceability
**Given** a draft the system proposed for a recorded board meeting (AC-26)
**When** a член правління opens it
**Then** the draft shows the meeting date, the board roster in force on that date, the case documents and an agenda item for each справа

### AC-03c (US-02) — happy path
**Given** a board meeting that already has a protocol draft
**When** a член правління opens a draft for the same meeting from another справа, or asks for one in the conversation
**Then** the system adds the справа as a new agenda item to the existing draft instead of opening a second one, and no lifecycle rule blocks adding items while it is a draft

### AC-29 (US-02) — happy path
**Given** a draft for a meeting whose головуючий is another member
**When** a current член правління rewords a sentence of the draft in spilno
**Then** the system keeps the new wording with its author and time, and every slot keeps its value and source; any current член правління may edit a draft, not only the головуючий

### AC-04 (US-02) — error
**Given** a член правління working on a protocol draft
**When** the draft has no agenda item, or an agenda item without a subject
**Then** the system does not let the draft be marked готовий до підпису and names every missing part at once

### AC-30 (US-03) — happy path
**Given** a draft in чернетка
**When** a член правління starts the попереднє голосування
**Then** the current wording is frozen as round 1 (or as the next round after a return to draft), the draft is готовий до голосування, and every current член правління sees it with the label of AC-41; slots still unconfirmed are shown as unconfirmed and do not prevent the poll

### AC-31 (US-03) — happy path
**Given** a draft готовий до голосування, round 1, and a board of 5
**When** 3 members cast "for", 1 casts "against" and 1 abstains on an item
**Then** the system shows the round-1 tally for the item with each member and time, labelled as informal (AC-41), and gives no verdict on quorum, majority or adoption

### AC-41 (US-03) — legal competence
**Given** a draft готовий до голосування, or a protocol whose drafts had poll rounds
**When** any member views the poll, its tally, the print copy or the protocol in the реєстр протоколів
**Then** the poll is labelled "Попереднє внутрішнє опитування правління — не є голосуванням засідання і не замінює його"; its tallies never appear on the print copy, never pre-fill the meeting facts entered at lock (AC-14), and appear in the реєстр протоколів only as the draft's history

### AC-05d (US-03) — domain invariant
**Given** a draft frozen as готовий до голосування, round 2, on which 2 members have already voted
**When** a член правління changes a slot or the text, or chooses «повернути в чернетку»
**Then** the draft returns to draft, round 2 is closed with its votes kept for history and no longer counted, and starting the poll again opens round 3 with no votes (AC-30)

### AC-05e (US-03) — authorization
**Given** a член правління casting a попереднє голосування
**When** they try to cast or change another member's vote
**Then** the system refuses: each member casts and changes only their own vote

### AC-32 (US-03) — domain invariant
**Given** a draft in чернетка or готовий до підпису, or a protocol that is locked or знятий з розгляду
**When** a член правління tries to cast a poll vote
**Then** the system refuses and says the poll is open only while the draft is готовий до голосування

### AC-09 (US-04) — domain invariant
**Given** an amount slot the system proposed from a document of the справа and a person confirmed
**When** that document is changed, or a member adds a newer version of it to the справа (a revised estimate, an amendment to the contract)
**Then** the system marks the proposed text and the slots that cited it as needing review, shows the confirmed value next to the value in the current document, and no longer counts the slot as confirmed (AC-21); like any change, this returns a draft готовий до голосування or готовий до підпису to draft (AC-05d, AC-14b)

### AC-21 (US-09) — domain invariant
**Given** a справа with a contract stating the total amount and the counterparty
**When** the system proposes the draft for it
**Then** the amount and counterparty slots are pre-filled, each marked "from document, unconfirmed" with the quoted fragment and a link to the document; they become confirmed only when a член правління confirms or corrects them, and the draft cannot be marked готовий до підпису while any of them is unconfirmed

### AC-22 (US-09) — happy path
**Given** a справа whose documents state different amounts for the same works (with and without VAT, an estimate superseded by a newer one, an advance and the full amount)
**When** the system pre-fills the amount slot
**Then** it recommends the final amount billed — from the document with the latest document date (not upload date) that states the full cost of the works, VAT included, or the amount payable for a contractor who is not a VAT payer; with no invoice yet, from the latest estimate or contract — marks it unconfirmed, and lists the other values with their documents so the член правління sees them before confirming

### AC-28 (US-12) — domain invariant
**Given** a proposed draft with confirmed слоти рішення
**When** the system writes or rewrites the connecting text of a decision, including after a change asked for in the conversation
**Then** the text holds no amount, date, document reference, name or count that is not one of the decision's слоти рішення (amount, counterparty, джерело коштів, document reference with its number and date, term or date, person, votes, неучасть у голосуванні) or a meeting fact from the meeting record (meeting date, головуючий); members present and vote counts stay blank fields on the print copy and are entered from the signed paper with the scan (AC-14); a requested change that would add one becomes a new slot to confirm instead

### AC-28b (US-12) — domain invariant
**Given** a слот рішення a член правління has already confirmed
**When** a change asked for in the conversation would alter its value
**Then** the system does not overwrite it: it shows the new value next to the confirmed one and marks the slot as needing review until a член правління confirms one of them

### AC-14c (US-06) — happy path
**Given** a draft in чернетка with every slot the system proposed confirmed or corrected by a person
**When** a член правління marks it готовий до підпису
**Then** the system freezes the text and renders a PDF print copy with the meeting date, the meeting form and the signers from the settings (AC-42) and blank fields for members present and the vote count per item, to fill in by hand at the meeting; a protocol without a number takes the next number of the реєстр протоколів (one sequence, no duplicates), printed on the copy

### AC-33 (US-06) — happy path
**Given** a draft готовий до голосування, round 2, with every proposed slot confirmed
**When** a член правління marks it готовий до підпису
**Then** the draft is готовий до підпису with its print copy and number as in AC-14c, round 2 is closed with its votes kept for history, and the print copy carries no poll result

### AC-14b (US-06) — domain invariant
**Given** protocol No 14 marked готовий до підпису with a rendered print copy
**When** a член правління changes a slot or the text, or chooses «повернути в чернетку»
**Then** the system returns it to draft, marks the print copy void and keeps it in the history, tells them that copies printed earlier must not be signed, and the protocol keeps No 14 for its next print copy; a number, once taken, is never given to another protocol, so the register may have gaps

### AC-14 (US-06) — happy path
**Given** a protocol marked готовий до підпису, printed and signed on paper at the meeting
**When** a член правління uploads the signed scan and enters the meeting facts written on the paper — members present, the vote count per agenda item, and whether the paper carries handwritten corrections
**Then** the system locks the protocol and adds it to the реєстр протоколів with the signed scan and the meeting facts, who entered them and when, and the text can no longer be edited; who actually signed, quorum and majority are not checked by the system in Stage 1

### AC-34 (US-06) — error
**Given** a protocol marked готовий до підпису
**When** a член правління uploads the signed scan without members present or without a vote count for every agenda item
**Then** the system does not lock it, names every missing fact at once, and the protocol stays готовий до підпису

### AC-15 (US-06) — domain invariant
**Given** a член правління uploading the signed copy
**When** they report that the paper copy carries handwritten corrections
**Then** the system requires a note describing each correction, keeps it with the scan, locks the protocol as in AC-14, and shows the note next to the frozen text in the реєстр протоколів; the frozen text is not changed, and the signed paper with its corrections prevails

### AC-40 (US-06) — domain invariant
**Given** a draft in чернетка or готовий до голосування
**When** a член правління tries to upload a signed copy
**Then** the system refuses and says the protocol must first be marked готовий до підпису and printed

### AC-16 (US-06) — authorization
**Given** a person whose board term has ended, and a draft they worked on while in office
**When** they try to upload its signed copy
**Then** the system refuses: only a current член правління may lock a protocol, and each such action is recorded with its author

### AC-17 (US-06) — domain invariant
**Given** a locked protocol in the реєстр протоколів
**When** any член правління tries to change its text, meeting facts or signed copy
**Then** the system refuses and explains that a locked protocol changes only through a new board decision that references it

### AC-18 (US-07) — happy path
**Given** 30 locked protocols in the реєстр протоколів
**When** a член правління searches by counterparty, amount (exact or a from–to range), subject or date (a period)
**Then** the system lists the matching decisions and opens each with its signed copy, meeting facts, cited documents and poll history labelled as informal

### AC-19 (US-07) — authorization
**Given** a person whose board term has ended, or who was never a член правління
**When** they try to open the реєстр протоколів or a справа
**Then** the system denies access and does not reveal whether any protocol or case exists

### AC-35 (US-13) — happy path
**Given** a draft in чернетка that the board decided not to pursue
**When** a член правління closes it as знятий з розгляду with the reason "управляюча компанія відкликала кошторис"
**Then** the draft is знятий з розгляду with its справи and versions kept, and takes no protocol number

### AC-36 (US-13) — happy path
**Given** a draft готовий до голосування whose poll did not gather support
**When** a член правління closes it as знятий з розгляду with the reason "не набрано підтримки, переглянути кошторис"
**Then** the draft is знятий з розгляду, keeps its справи, versions and poll rounds, is never printed, and takes no protocol number

### AC-37 (US-13) — happy path
**Given** protocol No 14 готовий до підпису that was printed but not signed
**When** a член правління closes it as знятий з розгляду with the reason "не підписано на засіданні"
**Then** the protocol is знятий з розгляду, its frozen text and print copy are kept as never signed, and it keeps No 14, which is never given to another protocol

### AC-38 (US-13) — domain invariant
**Given** a locked protocol
**When** a член правління tries to close it as знятий з розгляду
**Then** the system refuses and offers a new board decision that revokes or corrects it

### AC-39 (US-13) — domain invariant
**Given** a draft знятий з розгляду
**When** a член правління tries to edit it or start a poll on it
**Then** the system refuses and offers to start a new draft copied from it

### Deferred to Stage 2

These keep their ids and text as approved on 2026-10-24, so traceability survives; each is re-validated when Stage 2 is specified (§3).

#### AC-01 (US-01) — happy path  _(Stage 2: effective-dated rules)_
**Given** a член правління and a правило статуту "quorum size" with value 3 in effect since the association was founded
**When** the член правління records a new value 4 taking effect on 2026-12-01
**Then** the system keeps both values with their periods and shows that meetings before 2026-12-01 use 3 and meetings on or after it use 4

#### AC-02b (US-01) — domain invariant  _(Stage 2: effective-dated rules)_
**Given** a член правління changing a правило статуту or a board roster entry
**When** they save the change without a reference to the charter clause or the decision it comes from
**Then** the system rejects it; every saved change is kept with its reference, author and time, and is shown on every protocol that applied it

#### AC-02c (US-01) — happy path  _(Stage 2: quorum computation)_
**Given** a правило статуту "quorum size" recorded as a share of the board, "at least 2/3", in force on the meeting date
**When** the board roster in force on that date has 6 members
**Then** the system computes the meeting quorum as 4 members present, and recomputes it from the roster of each meeting rather than storing a fixed number

#### AC-03b (US-02) — authorization  _(dropped 2026-10-25: any current член правління edits, AC-29)_
**Given** a board meeting whose record names its головуючий, and a protocol draft for it
**When** another член правління tries to change a slot or the text of that draft
**Then** the system refuses and names the головуючий; any current член правління may change who presides over the meeting, and that change is recorded with its author and time

#### AC-05 (US-03) — happy path  _(Stage 2: quorum and majority)_
**Given** a board of 5 members, a quorum rule of 3 and a majority rule "more than half of the members present who may vote on the item" in force on the meeting date, 4 members marked present, and the item's wording frozen as готовий до голосування
**When** 3 members cast "for" and 1 casts "against"
**Then** the system shows the meeting quorum as reached and the item as passed, with the counts and the rule versions it applied

#### AC-05b (US-03) — domain invariant  _(Stage 2: quorum and majority)_
**Given** 4 members present who may vote on an item and the majority rule from AC-05
**When** 2 members cast "for" and 2 abstain
**Then** the system shows the item as lacking a majority (an abstaining member counts as present, not as "for"), raises a попередження, and still allows marking the draft готовий до підпису, with the item printed as not adopted or removed by the головуючий

#### AC-05c (US-03) — error  _(Stage 2: blocking checks)_
**Given** a quorum rule of 3 in force on the meeting date and 2 members marked present
**When** the головуючий checks the draft before marking it готовий до підпису
**Then** the system shows a блокуюча перевірка "no meeting quorum" naming the rule and its charter clause, passable only with an обхід блокування (AC-12)

#### AC-06 (US-03) — domain invariant  _(Stage 2: quorum and majority)_
**Given** 3 of 5 members present, a quorum rule of 3, and the rule "a non-voting member does not count toward an item's quorum" in force
**When** the головуючий records a неучасть у голосуванні for one member on one item
**Then** the system shows the meeting quorum as reached and that item as lacking quorum, names the rule that excluded the member, and raises a попередження for the item rather than a блокуюча перевірка

#### AC-07 (US-03) — error  _(Stage 2: attendance and non-voting rules)_
**Given** a головуючий recording votes on an item
**When** they mark a member as not voting without choosing a reason
**Then** the system does not accept it and asks for one of: left or joined during the meeting, personal interest, another legal restriction

#### AC-08 (US-03) — domain invariant  _(Stage 2: attendance and non-voting rules)_
**Given** a person whose board term ended before the meeting date
**When** the головуючий tries to mark that person present
**Then** the system refuses and explains that only members whose term covers the meeting date can attend and vote

#### AC-10 (US-04) — domain invariant  _(Stage 2: source of truth as a blocking rule)_
**Given** a decision whose amount slot has a value but no confirmed source document
**When** the головуючий checks the draft before marking it готовий до підпису
**Then** the system shows a блокуюча перевірка "number without a confirmed source" pointing at that slot

#### AC-11 (US-05) — domain invariant  _(Stage 2: blocking checks)_
**Given** a правило статуту "board competence limit 150,000 UAH" in force on the meeting date
**When** a decision's confirmed amount is 184,500 UAH
**Then** the system shows a блокуюча перевірка naming the rule and its charter clause, and does not allow marking the draft готовий до підпису without an обхід блокування

#### AC-12 (US-05) — happy path  _(Stage 2: overrides)_
**Given** a draft with one блокуюча перевірка
**When** a член правління records an обхід блокування with the reason "emergency roof repair; general meeting to ratify"
**Then** the system allows marking the draft готовий до підпису and keeps the override with its reason, author and time next to the protocol in the реєстр протоколів

#### AC-12b (US-05) — domain invariant  _(Stage 2: overrides)_
**Given** an обхід блокування recorded for "board competence limit" on an amount of 184,500 UAH
**When** the amount is later changed to 300,000 UAH, or another блокуюча перевірка appears
**Then** the override no longer applies: it covers only the one check and the values it was recorded for, a new override with its own reason is needed, and the old one stays in the draft's history

#### AC-13 (US-05) — error  _(Stage 2: overrides)_
**Given** a draft with a блокуюча перевірка
**When** a член правління attempts an обхід блокування with an empty reason
**Then** the system refuses and explains that an override needs a written reason

#### AC-14d (US-06) — domain invariant  _(Stage 2: reading the signed scan)_
**Given** a signed scan being uploaded for a draft marked готовий до підпису
**When** the protocol number or meeting date the system reads from the scan differs from the draft's, or cannot be read
**Then** the system shows a попередження with both values side by side and still lets the член правління lock the protocol: an uploaded signed scan counts as the adopted protocol

#### AC-17b (US-06) — domain invariant  _(Stage 2: rule and document snapshots)_
**Given** a locked protocol
**When** a правило статуту or a roster term is later recorded with an effective date before its meeting, or a case document it cites is changed or replaced
**Then** the protocol keeps the rule versions, roster, computed quorum and majority, and the document versions its slots were confirmed against, and the реєстр протоколів and the export show those; the back-dated entry affects only drafts

#### AC-17c (US-06) — domain invariant  _(Stage 2: revoking a чинне рішення)_
**Given** a new decision that references a чинне рішення in the реєстр протоколів
**When** the головуючий marks the reference as "revokes", "replaces" or "only refers to", and the new protocol is locked
**Then** a revoked or replaced decision is shown in the реєстр протоколів as no longer чинне, with a link to the decision that ended it; a decision that is only referred to stays чинне

#### AC-20 (US-08) — happy path  _(Stage 2: export)_
**Given** a locked protocol
**When** a член правління exports it
**Then** the system produces one downloadable file with the signed scan, a decision summary with votes and quorum, and every cited case document in the version the protocol cites

#### AC-23 (US-10) — domain invariant  _(Stage 2)_
**Given** an agenda-item type with a шаблон формулювання
**When** the головуючий tries to edit the fixed text of that item
**Then** the system refuses, names the template and its source, and lets them fill only the template's slots

#### AC-23b (US-10) — domain invariant  _(Stage 2)_
**Given** a locked protocol whose agenda item used a шаблон формулювання
**When** a член правління records a new version of that template taking effect on a later date
**Then** the locked protocol keeps its original wording, and only drafts for meetings on or after that date use the new version

#### AC-24 (US-11) — happy path  _(Stage 2)_
**Given** an agenda item approving a contract with a counterparty
**When** the головуючий records the votes
**Then** the system shows a попередження reminding them to record неучасть у голосуванні for any member with personal interest, linked to the charter or law clause, and does not stop the draft from being marked готовий до підпису

#### AC-25 (US-11) — cross-context  _(Stage 2)_
**Given** a decision spending from a джерело коштів for which the association's financial records already hold other contracts approved by decisions that are still чинні (AC-17c)
**When** the головуючий drafts the decision
**Then** the system shows those decisions and contracts alongside as a попередження with links to them, and does not stop the draft from being marked готовий до підпису

#### AC-25b (US-11) — domain invariant  _(Stage 2)_
**Given** a чинне рішення in the реєстр протоколів on the same counterparty, or on the same subject — for now the same agenda-item type (matching rule, §8)
**When** the головуючий drafts a new decision
**Then** the system shows the earlier decision alongside as a попередження with a link to it, and does not stop the draft from being marked готовий до підпису

#### AC-27 (US-12) — domain invariant  _(Stage 2: source of truth as a blocking rule)_
**Given** the головуючий writes that the estimate will come later instead of attaching it
**When** the system proposes the draft
**Then** the amount slot is marked as resting on a note, not a document, and stays a блокуюча перевірка "number without a confirmed source" until a document is attached and the slot confirmed

## 6. Non-functional requirements

| Aspect | Target | Measurement |
|---|---|---|
| Latency p95 — read: open a draft or a registry decision, registry search | ≤ 200 ms | request-duration metric in the service telemetry |
| Latency p95 — write without a file | ≤ 500 ms | request-duration metric in the service telemetry |
| Signed-copy and case-document upload | files up to 20 MB, PDF / JPEG / PNG; p95 ≤ 3 s for 20 MB | upload-duration metric; rejected-size count |
| Print copy (PDF) rendering | p95 ≤ 5 s | request-duration metric |
| Proposed protocol from a conversation | p95 ≤ 30 s from request to filled draft | proposal-duration metric in the service telemetry |
| Integrity of locked protocols | 0 changes to text, meeting facts or signed copy after lock | periodic fingerprint comparison of every locked protocol, alert on mismatch |
| Durability of the registry | RPO ≤ 24 h (nightly backup), RTO ≤ 4 h; restore verified quarterly | nightly backup log + quarterly restore drill record |
| Retention | locked and withdrawn protocols, print copies and scans kept ≥ 5 years, never hard-deleted | storage policy |
| Availability | 99.5 % monthly | uptime check, monthly window |

## 6.1 Security / privacy

- **Data classification:** confidential — signed legal records of the association plus personal data of co-owners, board members and contractors.
- **Personal data touched:** board members' names and terms (internal); names, apartment numbers and payment details of co-owners and sole-proprietor contractors inside case documents (confidential); handwritten signatures on signed copies, handwritten-correction notes and conversation requests that may name people (confidential).
- **AuthZ/AuthN impact:** new capability boundary — only current члени правління (term covers today) can see cases, drafts and the реєстр протоколів; access ends on the term end date; any current член правління edits a draft (AC-29), starts or ends a poll, prints, uploads a signed scan with the meeting facts, withdraws a draft, and keeps the roster and settings, and the author of each is recorded; each member casts only their own попереднє голосування (AC-05e). One association in Stage 1. The app has no sign-in yet: identifying a член правління is a prerequisite (§8).
- **Abuse cases:**
  - Former member access: a person whose term ended opens the registry → access denied, existence of records hidden (AC-19).
  - Rewriting history: someone edits a locked protocol to cover a decision → refused; corrections only through a new referencing decision; integrity check alerts on any mismatch (AC-17, §6).
  - The poll passed off as the meeting's vote: a poll tally is presented as an absentee board decision → the poll is labelled informal everywhere, never printed and never pre-fills the meeting facts (AC-41).
  - Wrong signed copy: a scan of a different protocol is uploaded → the scan, the meeting facts, the uploader and the time are kept for audit (AC-14); reading the scan is Stage 2 (AC-14d).
  - Instructions hidden in a document: a case document contains text meant to steer the proposed protocol → the system only proposes; every number stays unconfirmed until a person confirms it, and generated text holds no fact outside the slots (AC-21, AC-28).
  - Malicious upload: an uploaded file carries active content → only document and image formats are accepted, and files are served for download, never rendered as active content.
- **Security review:** Required — new authorization boundary and new personal data.

## 7. Metrics / KPIs

- **Post-signature corrections and re-votes** — baseline: TBD (count across the 30+ existing protocols, §8), target: 0 per year within 12 months of release.
- **Confirmed proposed numbers** — baseline: ~0% (no protocol links numbers to sources today), target: 100% of the numbers the system proposed in locked protocols confirmed or corrected by a person, each with a link to its case document, from the first locked protocol.
- **Time to find a past decision with its basis** — baseline: often impossible after ~6 months, target: ≤ 1 minute for any locked protocol within 3 months of release.
- **Time per protocol** — baseline: TBD (time the next 2-3 protocols by hand, §8), target: ≤ half the baseline within 3 months of Stage 1.
- **Adoption** — baseline: 0%, target: ≥ 80% of the association's board protocols locked in the system within 3 months of release.
- **Override rate** (Stage 2) — baseline: 0 (new), target: ≤ 10% of locked protocols carry an обхід блокування, reviewed quarterly.

## 8. Open questions

- [ ] Does the OSBB law or the charter restrict a member's vote on personal interest, and does a non-voting member count toward that item's quorum? Default now: not counted, as a правило статуту; checks are Stage 2. — owner: Andrii Opryshko, due: before Stage 2
- [ ] Which meeting forms does the charter allow for the board, and who signs the protocol? Default now: recorded in the association settings before the first protocol and printed on every copy (AC-42). — owner: Andrii Opryshko, due: before Stage 1 release
- [ ] How many of the 30+ protocols needed post-signature correction or a re-vote (KPI baseline)? — owner: Andrii Opryshko, due: 2026-11-04
- [ ] What is the time per protocol today (time the next 2-3)? — owner: Andrii Opryshko, due: 2026-11-30
- [ ] How does a член правління sign in, given the app has no sign-in yet? Default now: decided in architecture as a prerequisite. — owner: Andrii Opryshko, due: before /sdlc-architecture-design
- [ ] Which agenda-item types need a шаблон формулювання, and who reviews the templates? Default now: none until reviewed. — owner: Andrii Opryshko, due: before Stage 2
- [ ] Where do general-meeting чинні рішення come from for related-decision warnings? Default now: only board decisions in the реєстр протоколів. — owner: Andrii Opryshko, due: before Stage 2
- [ ] Do the financial records hold each contract's джерело коштів and a link to the decision that approved it (needed by AC-25)? Default now: in Stage 1 the джерело коштів is a slot the system proposes and a person confirms, linked to its document; the finance fields come with Stage 2. — owner: Andrii Opryshko, due: before Stage 2
- [ ] May case documents and conversation requests with personal data be processed by a text-generation service outside the association's own hosting? Default now: only with an agreement that excludes training on the data. — owner: Andrii Opryshko, due: before /sdlc-architecture-design
- [ ] How is "the same subject" matched for related-decision warnings (AC-25b): agenda-item type, keywords, or meaning? Default now: the same agenda-item type. — owner: Andrii Opryshko, due: before Stage 2
