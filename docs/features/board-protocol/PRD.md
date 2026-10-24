---
status: Approved
owner: "Andrii Opryshko"
reviewers: ["Andrii Opryshko"]
updated_at: "2026-10-24"
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

Committed approach (idea-brief §13): **Approach C, staged** — every decision is built from typed слоти рішення confirmed by a human against a case document; charter rules, неучасть у голосуванні, required wording and related чинні рішення surface linked to their source. This PRD covers **Stage 1** (roster and правила статуту bound to the meeting date, attendance and quorum, per-item votes with неучасть у голосуванні, every number cited to a case document, a реєстр протоколів of signed protocols) **and Stage 2** (slot pre-fill from case documents, locked шаблони формулювання, charter попередження, related чинні рішення shown alongside). Stage 2 also brings in conversational drafting: the головуючий describes what the protocol should decide and attaches the case documents, the system proposes a protocol with recommended numbers, and people review and correct it — the part of Stage 3 that writes connecting text is included, bounded to confirmed slots. The binding vote happens outside the system: the system prepares the document, collects the попереднє голосування each член правління casts on wording frozen as готовий до голосування, marks it готовий до підпису, and locks the protocol once the signed scan comes back (the first release accepts signed scans only, §3).

Traceability and deliberate decisions:
- Reference module: the existing board-meeting record is the anchor the protocol hangs on; its "report every broken field at once" validation style carries over to drafts, roster and rules.
- Prototype: slot origin (system / document / human) and status (confirmed / needs review / not found); advisory notices never block, deterministic checks block printing — this PRD keeps that split.
- Decision override: first release = Stage 1 + Stage 2, while idea-brief §13 locked Stage 1 only — rationale: Stage 1 alone does not relieve the "stay out of trouble" pain or the time spent per protocol; delivery stays staged (Stage 1 stories are marked and useful on their own).
- Decision override: conversational drafting with recommended numbers is in the first release and the size moved from M to L — this partly revives approach B (idea-brief §7, parked in §14 until 2+ associations and lawyer-reviewed templates) without waiting for its trigger. Rationale: it is how the author wants to work with the system; the reasons B was parked are excluded here — every recommended number waits for human confirmation, generated text holds no fact outside the slots (AC-28), and blocking stays arithmetic with an override.
- Decision override: deterministic блокуючі перевірки stop a draft from being marked готовий до підпису, while idea-brief §5/§14 rejected blocking — rationale: the rejection targets blocking on automated legal judgments; checks here are arithmetic against правила статуту, and an обхід блокування with a recorded reason keeps the emergency path open. Item quorum after a неучасть у голосуванні and a failed item majority stay попередження (AC-06, AC-05b) until the §8 question on non-voting members is answered; a missing meeting quorum is a блокуюча перевірка that, like the others, can be overridden with a reason (AC-05c).
- Deliberate simplification (clarify-prd, 2026-10-24): the draft lifecycle and the signed scan get a naive first cut — no lifecycle rule blocks, an uploaded signed scan means the protocol was adopted, and what the system reads from the scan only raises попередження (AC-03c, AC-14d); stricter rules can follow once the first protocols run through the system.

## 2. Goals

- The головуючий produces a board protocol in which every amount and counterparty is confirmed against a case document, and quorum and majority are computed from the roster and правила статуту in force on the meeting date (Stage 1, idea-brief §13).
- Any current член правління finds a past decision with its signed copy, votes, quorum and cited documents in the реєстр протоколів, instead of searching chats (Stage 1).
- The головуючий describes what is needed and reviews a proposed protocol instead of copying numbers by hand, and is reminded of what they may not know or remember — required wording, possible personal interest, related чинні рішення — without the system making legal judgments (Stage 2).

## 3. Non-goals

- General-meeting protocols and co-owner e-voting — covered by existing services (idea-brief §5).
- Running the vote in the system (live meeting, remote participation, written polls, signing inside the system) — the vote and signing happen outside; the system collects the попереднє голосування and accepts the signed result as a scan.
- Access for co-owners or the ревізійна комісія, and onboarding other associations — the first release serves one association and its current board; the audit commission receives exports (idea-brief §5).
- Legal advice or interpretive blocking — only deterministic checks block; anything needing interpretation is a попередження (idea-brief §5).
- Files signed with a qualified electronic signature, and checking who signed — the first release accepts signed scans only; the uploader confirms the scan matches the frozen text and carries the signatures the charter requires (AC-14). Signature upload and validation come later.
- Marking an abandoned draft знятий з розгляду — in the first release an abandoned draft simply stays a draft.
- The autonomous part of approach B — the system never confirms a number, removes a member from a vote or blocks on its own interpretation; approach B stays parked (idea-brief §14).

## 4. User stories

### US-01: Maintain roster and charter rules  _(Stage 1)_
**As a** член правління
**I want** to record the board roster with each member's term, and the правила статуту and шаблони формулювання with the date each takes effect
**So that** every check uses the roster and rules in force on the meeting date, not today's

### US-02: Open a protocol draft from a case  _(Stage 1)_
**As a** головуючий
**I want** to start a protocol draft for a board meeting from one or more справи, with its agenda items
**So that** the protocol is built from the documents that justify it

### US-03: Record attendance and preliminary votes  _(Stage 1)_
**As a** член правління
**I want** the головуючий to record who attended and each неучасть у голосуванні with its reason, to freeze the draft wording as готовий до голосування, and every member to cast their own попереднє голосування on each agenda item
**So that** the system computes the meeting's quorum and each item's quorum and majority from each member's own vote

### US-04: Cite every number to a case document  _(Stage 1)_
**As a** головуючий
**I want** every amount, counterparty and джерело коштів in a decision to be a слот рішення confirmed against a document of the справа
**So that** any number in the protocol can be traced to its source later

### US-05: Resolve blocking checks before locking  _(Stage 1)_
**As a** член правління
**I want** to see every блокуюча перевірка before locking, and record an обхід блокування with a reason when an emergency requires it
**So that** I don't lock an invalid protocol by accident and am never stuck

### US-06: Lock the protocol with its signed copy  _(Stage 1)_
**As a** член правління
**I want** to mark the final protocol готовий до підпису, print or download it, and upload the signed scan to lock it
**So that** the реєстр протоколів holds the legally meaningful copy and the text can no longer change

### US-07: Find a past decision with its basis  _(Stage 1)_
**As a** член правління
**I want** to search the реєстр протоколів and open a decision with its signed copy, votes, quorum, applied rules and cited documents
**So that** I can answer an auditor or a co-owner within a minute

### US-08: Export a protocol for the audit commission  _(Stage 1)_
**As a** член правління
**I want** to export a locked protocol with its signed copy and cited case documents as one package
**So that** the ревізійна комісія can check it without access to the system

### US-09: Pre-fill decision slots from documents  _(Stage 2)_
**As a** головуючий
**I want** слоти рішення pre-filled from the case documents and marked unconfirmed until I confirm each one
**So that** I don't copy sums and names by hand, and nothing unconfirmed slips into a locked protocol

### US-10: Use the required wording template  _(Stage 2)_
**As a** головуючий
**I want** an agenda item of a type that has a шаблон формулювання to use that fixed wording, with only its slots editable
**So that** wording the law or charter requires is never changed by accident

### US-11: See warnings and related decisions  _(Stage 2)_
**As a** головуючий
**I want** to see попередження (possible personal interest, relevant charter clauses) and related чинні рішення alongside the draft, each linked to its source
**So that** I'm reminded of what I may not know or remember, and decide myself

### US-12: Draft a protocol through a conversation  _(Stage 2)_
**As a** головуючий
**I want** to describe in a conversation what the protocol should decide, attach case documents or a note that a document comes later, get a proposed protocol with recommended numbers, and refine it through the conversation or by hand
**So that** I review and correct a protocol instead of assembling it

## 5. Acceptance criteria

### AC-01 (US-01) — happy path
**Given** a член правління and a правило статуту "quorum size" with value 3 in effect since the association was founded
**When** the член правління records a new value 4 taking effect on 2026-12-01
**Then** the system keeps both values with their periods and shows that meetings before 2026-12-01 use 3 and meetings on or after it use 4

### AC-02 (US-01) — error
**Given** a член правління adding a person to the board roster
**When** they save the entry without a term start, or with a term end before the term start
**Then** the system rejects the entry and names every missing or inconsistent field at once

### AC-02b (US-01) — domain invariant
**Given** a член правління changing a правило статуту or a board roster entry
**When** they save the change without a reference to the charter clause or the decision it comes from
**Then** the system rejects it; every saved change is kept with its reference, author and time, and is shown on every protocol that applied it

### AC-02c (US-01) — happy path
**Given** a правило статуту "quorum size" recorded as a share of the board, "at least 2/3", in force on the meeting date
**When** the board roster in force on that date has 6 members
**Then** the system computes the meeting quorum as 4 members present, and recomputes it from the roster of each meeting rather than storing a fixed number

### AC-03 (US-02) — happy path
**Given** a головуючий, a recorded board meeting and a справа with a contract and an estimate
**When** the головуючий opens a protocol draft for that meeting from the справа
**Then** the draft shows the meeting date, the board roster in force on that date, the case documents and an agenda item for the справа

### AC-03b (US-02) — authorization
**Given** a board meeting whose record names its головуючий, and a protocol draft for it
**When** another член правління tries to change a slot or the text of that draft
**Then** the system refuses and names the головуючий; any current член правління may change who presides over the meeting, and that change is recorded with its author and time

### AC-03c (US-02) — happy path
**Given** a board meeting that already has a protocol draft
**When** the головуючий opens a draft for the same meeting from another справа, or asks for one in the conversation
**Then** the system adds the справа as a new agenda item to the existing draft instead of opening a second one, and no lifecycle rule blocks adding items while it is a draft

### AC-04 (US-02) — error
**Given** a головуючий opening a protocol draft
**When** the draft has no agenda item, or an agenda item without a subject
**Then** the system does not let the draft be marked готовий до підпису and names every missing part at once

### AC-05 (US-03) — happy path
**Given** a board of 5 members, a quorum rule of 3 and a majority rule "more than half of the members present who may vote on the item" in force on the meeting date, 4 members marked present, and the item's wording frozen as готовий до голосування
**When** 3 members cast "for" and 1 casts "against"
**Then** the system shows the meeting quorum as reached and the item as passed, with the counts and the rule versions it applied

### AC-05b (US-03) — domain invariant
**Given** 4 members present who may vote on an item and the majority rule from AC-05
**When** 2 members cast "for" and 2 abstain
**Then** the system shows the item as lacking a majority (an abstaining member counts as present, not as "for"), raises a попередження, and still allows marking the draft готовий до підпису, with the item printed as not adopted or removed by the головуючий

### AC-05c (US-03) — error
**Given** a quorum rule of 3 in force on the meeting date and 2 members marked present
**When** the головуючий checks the draft before marking it готовий до підпису
**Then** the system shows a блокуюча перевірка "no meeting quorum" naming the rule and its charter clause, passable only with an обхід блокування (AC-12)

### AC-05d (US-03) — domain invariant
**Given** a draft frozen as готовий до голосування on which 2 members have already voted
**When** the головуючий changes a slot or the text
**Then** the draft returns to draft, the cast votes stay with the old wording version and no longer count, and members vote again once the new wording is frozen

### AC-05e (US-03) — authorization
**Given** a член правління casting a попереднє голосування
**When** they try to cast or change another member's vote
**Then** the system refuses: each member casts and changes only their own vote

### AC-06 (US-03) — domain invariant
**Given** 3 of 5 members present, a quorum rule of 3, and the rule "a non-voting member does not count toward an item's quorum" in force
**When** the головуючий records a неучасть у голосуванні for one member on one item
**Then** the system shows the meeting quorum as reached and that item as lacking quorum, names the rule that excluded the member, and raises a попередження for the item rather than a блокуюча перевірка

### AC-07 (US-03) — error
**Given** a головуючий recording votes on an item
**When** they mark a member as not voting without choosing a reason
**Then** the system does not accept it and asks for one of: left or joined during the meeting, personal interest, another legal restriction

### AC-08 (US-03) — domain invariant
**Given** a person whose board term ended before the meeting date
**When** the головуючий tries to mark that person present
**Then** the system refuses and explains that only members whose term covers the meeting date can attend and vote

### AC-09 (US-04) — domain invariant
**Given** an amount slot confirmed against a document of the справа
**When** that document is changed, or replaced in the справа by a newer one (a revised estimate, an amendment to the contract)
**Then** the system marks the slot as needing review, shows the confirmed value next to the value in the current document, and no longer counts the slot as confirmed; a draft marked готовий до підпису returns to draft, as in AC-14b

### AC-10 (US-04) — domain invariant
**Given** a decision whose amount slot has a value but no confirmed source document
**When** the головуючий checks the draft before marking it готовий до підпису
**Then** the system shows a блокуюча перевірка "number without a confirmed source" pointing at that slot

### AC-11 (US-05) — domain invariant
**Given** a правило статуту "board competence limit 150,000 UAH" in force on the meeting date
**When** a decision's confirmed amount is 184,500 UAH
**Then** the system shows a блокуюча перевірка naming the rule and its charter clause, and does not allow marking the draft готовий до підпису without an обхід блокування

### AC-12 (US-05) — happy path
**Given** a draft with one блокуюча перевірка
**When** a член правління records an обхід блокування with the reason "emergency roof repair; general meeting to ratify"
**Then** the system allows marking the draft готовий до підпису and keeps the override with its reason, author and time next to the protocol in the реєстр протоколів

### AC-12b (US-05) — domain invariant
**Given** an обхід блокування recorded for "board competence limit" on an amount of 184,500 UAH
**When** the amount is later changed to 300,000 UAH, or another блокуюча перевірка appears
**Then** the override no longer applies: it covers only the one check and the values it was recorded for, a new override with its own reason is needed, and the old one stays in the draft's history

### AC-13 (US-05) — error
**Given** a draft with a блокуюча перевірка
**When** a член правління attempts an обхід блокування with an empty reason
**Then** the system refuses and explains that an override needs a written reason

### AC-14 (US-06) — happy path
**Given** a draft marked готовий до підпису
**When** a член правління uploads the signed scan and confirms that it matches the frozen text, carries the signatures the charter requires and has no handwritten corrections
**Then** the system locks the protocol, adds it to the реєстр протоколів with the signed scan and the uploader's confirmation, author and time, and the text can no longer be edited; who actually signed is not checked by the system

### AC-14b (US-06) — domain invariant
**Given** a draft marked готовий до підпису
**When** a член правління changes a slot or the text
**Then** the system returns it to draft, and tells them that any copy printed or downloaded earlier must be re-issued before signing

### AC-14c (US-06) — happy path
**Given** a draft with every блокуюча перевірка resolved or overridden
**When** the головуючий marks it готовий до підпису
**Then** the system proposes the next protocol number in the реєстр протоколів, lets the головуючий change it, and prints the number and the meeting date on the copy for signing

### AC-14d (US-06) — domain invariant
**Given** a signed scan being uploaded for a draft marked готовий до підпису
**When** the protocol number or meeting date the system reads from the scan differs from the draft's, or cannot be read
**Then** the system shows a попередження with both values side by side and still lets the член правління lock the protocol: an uploaded signed scan counts as the adopted protocol

### AC-15 (US-06) — domain invariant
**Given** a член правління uploading the signed copy
**When** they report that the paper copy carries handwritten corrections
**Then** the system does not lock the protocol and keeps the draft open, so the text is corrected and the protocol re-printed and re-signed

### AC-16 (US-06) — authorization
**Given** a person whose board term has ended, and a draft they worked on while in office
**When** they try to upload its signed copy or record an обхід блокування
**Then** the system refuses: only a current член правління may lock a protocol or override a блокуюча перевірка, and each such action is recorded with its author

### AC-17 (US-06) — domain invariant
**Given** a locked protocol in the реєстр протоколів
**When** any член правління tries to change its text, votes or signed copy
**Then** the system refuses and explains that a locked protocol changes only through a new board decision that references it

### AC-17b (US-06) — domain invariant
**Given** a locked protocol
**When** a правило статуту or a roster term is later recorded with an effective date before its meeting, or a case document it cites is changed or replaced
**Then** the protocol keeps the rule versions, roster, computed quorum and majority, and the document versions its slots were confirmed against, and the реєстр протоколів and the export show those; the back-dated entry affects only drafts

### AC-17c (US-06) — domain invariant
**Given** a new decision that references a чинне рішення in the реєстр протоколів
**When** the головуючий marks the reference as "revokes", "replaces" or "only refers to", and the new protocol is locked
**Then** a revoked or replaced decision is shown in the реєстр протоколів as no longer чинне, with a link to the decision that ended it; a decision that is only referred to stays чинне

### AC-18 (US-07) — happy path
**Given** 30 locked protocols in the реєстр протоколів
**When** a член правління searches by counterparty, amount (exact or a from–to range), subject or date (a period)
**Then** the system lists the matching decisions and opens each with its signed copy, votes, quorum, applied rule versions, cited documents and any overrides

### AC-19 (US-07) — authorization
**Given** a person whose board term has ended, or who was never a член правління
**When** they try to open the реєстр протоколів or a справа
**Then** the system denies access and does not reveal whether any protocol or case exists

### AC-20 (US-08) — happy path
**Given** a locked protocol
**When** a член правління exports it
**Then** the system produces one downloadable file with the signed scan, a decision summary with votes and quorum, and every cited case document in the version the protocol cites

### AC-21 (US-09) — happy path
**Given** a справа with a contract stating the total amount and the counterparty
**When** the головуючий opens a draft for it
**Then** the amount and counterparty slots are pre-filled, each marked "from document, unconfirmed" with the quoted fragment, and become confirmed only when the головуючий confirms them

### AC-22 (US-09) — happy path
**Given** a справа whose documents state different amounts for the same works (with and without VAT, an estimate superseded by a newer one, an advance and the full amount)
**When** the system pre-fills the amount slot
**Then** it recommends the final amount billed — from the document with the latest document date (not upload date) that states the full cost of the works, VAT included, or the amount payable for a contractor who is not a VAT payer; with no invoice yet, from the latest estimate or contract — marks it unconfirmed, and lists the other values with their documents so the головуючий sees them before confirming

### AC-23 (US-10) — domain invariant
**Given** an agenda-item type with a шаблон формулювання
**When** the головуючий tries to edit the fixed text of that item
**Then** the system refuses, names the template and its source, and lets them fill only the template's slots

### AC-23b (US-10) — domain invariant
**Given** a locked protocol whose agenda item used a шаблон формулювання
**When** a член правління records a new version of that template taking effect on a later date
**Then** the locked protocol keeps its original wording, and only drafts for meetings on or after that date use the new version

### AC-24 (US-11) — happy path
**Given** an agenda item approving a contract with a counterparty
**When** the головуючий records the votes
**Then** the system shows a попередження reminding them to record неучасть у голосуванні for any member with personal interest, linked to the charter or law clause, and does not stop the draft from being marked готовий до підпису

### AC-25 (US-11) — cross-context
**Given** a decision spending from a джерело коштів for which the association's financial records already hold other contracts approved by decisions that are still чинні (AC-17c)
**When** the головуючий drafts the decision
**Then** the system shows those decisions and contracts alongside as a попередження with links to them, and does not stop the draft from being marked готовий до підпису

### AC-25b (US-11) — domain invariant
**Given** a чинне рішення in the реєстр протоколів on the same counterparty, or on the same subject — for now the same agenda-item type (matching rule, §8)
**When** the головуючий drafts a new decision
**Then** the system shows the earlier decision alongside as a попередження with a link to it, and does not stop the draft from being marked готовий до підпису

### AC-26 (US-12) — happy path
**Given** a головуючий and a справа holding a contract and an invoice
**When** they ask in the conversation to approve the roof waterproofing contract from the repair fund and attach both documents
**Then** the system proposes a protocol draft with that agenda item and its слоти рішення filled with recommended values, each marked unconfirmed with its source fragment, and keeps the request, the conversation and every attached document with the draft

### AC-27 (US-12) — domain invariant
**Given** the головуючий writes that the estimate will come later instead of attaching it
**When** the system proposes the draft
**Then** the amount slot is marked as resting on a note, not a document, and stays a блокуюча перевірка "number without a confirmed source" until a document is attached and the slot confirmed

### AC-28 (US-12) — domain invariant
**Given** a proposed draft with confirmed слоти рішення
**When** the system writes or rewrites the connecting text of a decision, including after a change asked for in the conversation
**Then** the text holds no amount, date, document reference, name or count that is not one of the decision's слоти рішення (amount, counterparty, джерело коштів, document reference with its number and date, term or date, person, votes, неучасть у голосуванні) or a meeting fact the system fills from the meeting record and the roster (meeting date, головуючий, members present); a requested change that would add one becomes a new slot to confirm instead

### AC-28b (US-12) — domain invariant
**Given** a слот рішення the головуючий has already confirmed
**When** a change asked for in the conversation would alter its value
**Then** the system does not overwrite it: it shows the new value next to the confirmed one and marks the slot as needing review until the головуючий confirms one of them

## 6. Non-functional requirements

| Aspect | Target | Measurement |
|---|---|---|
| Latency p95 — open a draft or a registry decision | ≤ 1 s with ≤ 500 locked protocols | request-duration metric in the service telemetry |
| Latency p95 — registry search | ≤ 1 s with ≤ 500 locked protocols | request-duration metric in the service telemetry |
| Signed-copy upload | scans up to 50 MB accepted; p95 ≤ 20 s including reading the protocol number and date | upload-duration metric; rejected-size count |
| Proposed protocol from a conversation | p95 ≤ 60 s for up to 10 attached documents | proposal-duration metric in the service telemetry |
| Integrity of locked protocols | 0 changes to text, votes or signed copy after lock | periodic fingerprint comparison of every locked protocol, alert on mismatch |
| Durability of the registry | data loss ≤ 24 h (RPO); restore verified quarterly | daily backup log + quarterly restore drill record |
| Quorum and rule accuracy | 100% agreement with hand calculation on the association's past protocols | regression suite over the 30+ historical protocols, run on every change |
| Availability | 99% monthly | uptime check, monthly window |

## 6.1 Security / privacy

- **Data classification:** confidential — signed legal records of the association plus personal data of co-owners, board members and contractors.
- **Personal data touched:** board members' names and terms (internal); names, apartment numbers and payment details of co-owners and sole-proprietor contractors inside case documents (confidential); handwritten signatures on signed copies and conversation requests that may name people (confidential).
- **AuthZ/AuthN impact:** new capability boundary — only current члени правління (term covers today) can see cases, drafts and the реєстр протоколів; access ends on the term end date; only the головуючий of a meeting edits its draft (AC-03b); each member casts only their own попереднє голосування (AC-05e); any current член правління can upload a signed scan, record an обхід блокування, or change a правило статуту or the roster with a reference, and the author of each is recorded. The app has no sign-in yet: identifying a член правління is a prerequisite (§8).
- **Abuse cases:**
  - Former member access: a person whose term ended opens the registry → access denied, existence of records hidden (AC-19).
  - Rewriting history: someone edits a locked protocol to cover a decision → refused; corrections only through a new referencing decision; integrity check alerts on any mismatch (AC-17, §6).
  - Wrong signed copy: a scan of a different protocol is uploaded → the uploader confirms the copy matches, the system warns when the number or date read from the scan differs, and the upload is kept with author and time for audit (AC-14, AC-14d).
  - Relaxing a rule to clear a block: a member raises the competence limit to pass their own draft → every rule change needs a charter or decision reference, is kept with its author and is shown on the protocols that applied it (AC-02b).
  - Routine overrides: blocks are overridden by habit → every override is visible in the registry with its reason; override rate is a tracked KPI (§7).
  - Instructions hidden in a document: a case document contains text meant to steer the proposed protocol → the system only proposes; every number stays unconfirmed until a person confirms it, and generated text holds no fact outside the slots (AC-21, AC-28).
  - Malicious upload: an uploaded file carries active content → only document and image formats are accepted, and files are served for download, never rendered as active content.
- **Security review:** Required — new authorization boundary and new personal data.

## 7. Metrics / KPIs

- **Post-signature corrections and re-votes** — baseline: TBD (count across the 30+ existing protocols, §8), target: 0 per year within 12 months of release.
- **Cited and confirmed numbers** — baseline: ~0% (no protocol links numbers to sources today), target: 100% of amounts and counterparties in locked protocols confirmed against a case document, excluding recorded overrides, from the first locked protocol.
- **Time to find a past decision with its basis** — baseline: often impossible after ~6 months, target: ≤ 1 minute for any locked protocol within 3 months of release.
- **Time per protocol** — baseline: TBD (time the next 2-3 protocols by hand, §8), target: ≤ half the baseline within 3 months of Stage 2.
- **Override rate** — baseline: 0 (new), target: ≤ 10% of locked protocols carry an обхід блокування, reviewed quarterly.
- **Adoption** — baseline: 0%, target: ≥ 80% of the association's board protocols locked in the system within 3 months of release.

## 8. Open questions

- [ ] Does the OSBB law or the charter restrict a member's vote on personal interest, and does a non-voting member count toward that item's quorum? Default now: not counted, as a правило статуту. — owner: Andrii Opryshko, due: 2026-11-30
- [ ] How many of the 30+ protocols needed post-signature correction or a re-vote (KPI baseline)? — owner: Andrii Opryshko, due: 2026-11-04
- [ ] What is the time per protocol today (time the next 2-3)? — owner: Andrii Opryshko, due: 2026-11-30
- [ ] How does a член правління sign in, given the app has no sign-in yet? Default now: decided in architecture as a prerequisite. — owner: Andrii Opryshko, due: before /sdlc-architecture-design
- [ ] Which agenda-item types need a шаблон формулювання, and who reviews the templates? Default now: none until reviewed. — owner: Andrii Opryshko, due: before /sdlc-break-tasks
- [ ] Where do general-meeting чинні рішення come from for related-decision warnings? Default now: only board decisions in the реєстр протоколів. — owner: Andrii Opryshko, due: before /sdlc-break-tasks
- [ ] Do the financial records hold each contract's джерело коштів and a link to the decision that approved it (needed by AC-25)? Default now: both added to the financial records as part of this feature. — owner: Andrii Opryshko, due: before /sdlc-architecture-design
- [ ] May case documents and conversation requests with personal data be processed by a text-generation service outside the association's own hosting? Default now: only with an agreement that excludes training on the data. — owner: Andrii Opryshko, due: before /sdlc-architecture-design
- [ ] How is "the same subject" matched for related-decision warnings (AC-25b): agenda-item type, keywords, or meaning? Default now: the same agenda-item type. — owner: Andrii Opryshko, due: before /sdlc-break-tasks
