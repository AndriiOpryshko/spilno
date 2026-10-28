---
status: Living
updated_at: "2026-10-28"
---

# Domain Context — board-protocol

<!--
CONTEXT.md is the domain glossary — not a PRD and not a scratch pad. NO implementation
detail here (no datastore/broker/framework names, no API contracts) — only domain words
and the boundaries between them. Implementation choices live in the SAD and ADRs;
behaviour lives in PRD.md.

Terms get fixed inline, the moment they surface in an interview / PRD / review — never
batched «I'll consolidate later». Empty H2 → prune before commit; keep only the sections
that carry real content. Sections: Glossary · Invariants · Sentinel errors · Org-filter invariant ·
Out of scope (capstone layout, 2026-10-28).
-->

## Glossary

<!-- One line per term: name · one-sentence canonical definition · one-sentence boundary
     (what it is NOT / the concept it gets confused with). Alphabetical once there are a few. -->
- акт виконаних робіт — the document signed with the contractor (usually the management company) that confirms which works were done and for what amount, closing an expense line approved by a decision. NOT кошторис робіт (the planned cost before the work, not the confirmed cost after it); NOT банківська виписка (proof of payment, not of work done).
- блокуюча перевірка — (Stage 2) a deterministic check of a protocol draft against a правило статуту or a missing source (no meeting quorum, an amount above the board's competence limit) that stops a draft from being marked готовий до підпису unless lifted with an обхід блокування. In Stage 1 the only gate is that printing waits until every value proposed by the помічник протоколу is confirmed by a person. NOT попередження (advisory, never stops signing).
- головуючий — the board member who chairs a particular board meeting and signs its protocol: the chair, or a member substituting for the chair as the charter allows. NOT голова правління (the elected position; the chair may be absent and someone else presides).
- готовий до підпису — the state of a protocol draft whose text is frozen and rendered into a print copy for signing; the protocol number is reserved on the first render and never reused. Any later edit returns it to чернетка and voids the printed copy, the number stays with the protocol. NOT a locked protocol (locked only once the signed scan and the meeting facts are uploaded).
- готовий до голосування — the state of a protocol draft whose wording is frozen so that members can cast попереднє голосування on exactly that version; any edit or «повернути в чернетку» returns it to чернетка, and the next poll is a new round. NOT готовий до підпису (frozen for printing).
- джерело коштів — the fund or income line named in the approved budget (e.g. maintenance contributions, repair fund, reserve fund, rental income) that pays for a decision. NOT стаття витрат (where the money comes from vs what it is spent on).
- знятий з розгляду — the final state of a protocol draft that was never printed or signed (the proposal did not gather support or was abandoned), kept with its case, wording and preliminary votes as an internal record, not a legal document. NOT a signed protocol with a «not adopted» item; NOT a deleted draft.
- кворум — the minimum number of board members present for a board meeting to take valid decisions, as set by the association's charter. NOT більшість (presence needed to decide vs votes needed to pass); NOT кворум загальних зборів (that one counts co-owners' votes, not board members).
- кворум пункту — the кворум for one agenda item: the same charter кворум rule applied only to the members present who may vote on that item (a member with неучасть у голосуванні does not count when the правило статуту says so). NOT кворум of the meeting (all members present); NOT більшість (votes needed to pass).
- кошторис робіт — the management company's priced estimate for one job (works, materials, amount) that the board votes on and the protocol cites. NOT кошторис доходів і витрат (the association's annual budget approved by the general meeting); NOT рахунок (a request for payment, issued after or instead of an estimate).
- неучасть у голосуванні — a board member does not vote on one specific agenda item, with a recorded reason: left or joined during the meeting, personal interest, or another legal restriction. NOT відсутність на засіданні (out of the whole meeting); NOT утримався (abstained — voted, counted as present).
- обхід блокування — (Stage 2) a член правління's recorded choice to mark a draft готовий до підпису despite a блокуюча перевірка, with a mandatory reason that stays with the protocol in the реєстр протоколів. NOT signing through a попередження (no reason required, nothing is overridden).
- попередження — an advisory notice on a protocol draft, linked to its source (a charter clause, a related чинне рішення, a possible personal interest, a required wording), that the головуючий may sign through. NOT блокуюча перевірка (deterministic, stops signing).
- попереднє голосування — an informal poll per agenda item that each член правління casts in the system on one frozen version of the draft wording, only to see what passes before the meeting; every return to чернетка starts a new round, earlier rounds stay in history and are not counted. It is labelled in the UI as not the meeting's vote, never pre-fills the meeting result and never stops printing. NOT the binding vote (the signed paper protocol and the meeting facts entered with the scan record that); NOT votes typed in on another member's behalf.
- помічник протоколу — the conversation in which a член правління attaches case documents and asks for a protocol draft; it writes wording and advisory warnings and proposes amounts read from documents, but every number it proposes counts only after a person confirms it. NOT the source of any number in the protocol; NOT the board chat where members discuss.
- правило статуту — one charter parameter kept as data with the date it takes effect (quorum size as a count or a share of the roster in force, the majority an item needs and what it is counted from, the board's competence limit, whether a non-voting member counts toward an item's quorum), applied to meetings held on or after that date. NOT статут (the whole legal document); NOT чинне рішення (a decision, not a standing rule of procedure).
- протокол правління — the signed record of one board meeting: attendance, quorum, agenda items and each decision with its vote count. NOT протокол загальних зборів (different body, different competence and quorum).
- ревізійна комісія — the association's audit body elected by the general meeting that checks the board's decisions and spending against protocols and case documents; it takes no board decisions. NOT правління; NOT a співвласник asking a question (an audit is a formal check with its own mandate).
- реєстр протоколів — the association's searchable record of signed (locked) protocols with their signed copies, decisions, cited case documents and recorded overrides; it belongs to the association, not to any board member; a number taken by a знятий з розгляду protocol shows as a gap marked «знято з розгляду», the withdrawn protocol itself stays an internal record. NOT справа (the case is the input to a protocol; the registry holds the result).
- слот рішення — one typed field of a board decision (amount, counterparty, джерело коштів, document reference with its number and date, term or date, person, votes, неучасть у голосуванні) holding a value, where it came from, and the case document it is confirmed against; the meeting date and головуючий are filled from the meeting record; members present and the vote counts are meeting facts entered from the signed paper together with the scan. NOT вільний текст рішення (connecting prose; any number it mentions must equal a slot value).
- співвласник — an owner of an apartment or non-residential premises in the building, whether or not a member of the association, who pays contributions and votes at the general meeting. NOT мешканець (a resident may neither vote nor pay); NOT член об'єднання (membership is optional for a co-owner).
- справа — the bundle of source documents (contracts, invoices, estimates, owner requests) that a board decision is based on and cites. NOT протокол (the case is what a protocol is built from, not the protocol itself).
- управляюча компанія — the company contracted by the general meeting to maintain the buildings and manage the repair and development fund; it issues estimates and acts and holds the money. NOT правління (the board decides; the management company executes and is paid).
- частка будинку — a building's fixed share of a cost shared between the association's buildings, proportional to its area and approved by a separate inter-building decision. NOT площа будинку (the area is the input; the share is the approved result that may only change by a new decision).
- чинне рішення — a decision of the board or the general meeting that was validly adopted and has not been revoked or superseded, so new decisions must stay consistent with it; it stops being чинне only when a later locked protocol explicitly revokes or replaces it — a decision that is only referred to stays чинне. NOT підписаний протокол (a signed protocol may hold decisions that were later revoked or replaced).
- член правління — a person elected by the general meeting to the board for a term, who may take part and vote at board meetings while the term lasts. NOT головуючий (a per-meeting role a member takes); NOT член об'єднання (a co-owner who joined the association).
- шаблон формулювання — a fixed wording for an agenda-item type that the law or the charter requires to be phrased a certain way; only its слоти рішення may be filled, the text itself is not edited. NOT вільний текст рішення.
- чернетка — the state of a protocol draft that the помічник протоколу prepared and any член правління may edit; the only state where the text changes. NOT a print copy (frozen and rendered for signing).
- раунд опитування — one run of попереднє голосування on one frozen version of the wording; a return to чернетка closes it, and only the latest round is shown as current. NOT the meeting's vote.
- друкована копія — the PDF rendered from the frozen text when a draft becomes готовий до підпису, printed and signed on paper; voided by any later edit. NOT the signed scan (the uploaded photo or scan of the signed paper).
- підписаний — the final state of a protocol after the signed scan and the meeting facts (attendance, vote counts per item, note of handwritten corrections) are uploaded; the protocol is then immutable and changes only through a new referencing protocol. NOT готовий до підпису.

## Invariants

<!-- Rules above any single AC: «X always must / can never». Each names the decision that owns it. -->
- A підписаний protocol can never change: its text, meeting facts and signed scan are frozen; a correction is always a new protocol that references it (ADR-0007, AC-17).
- A protocol can never be deleted, in any state; знятий з розгляду is its terminal record, not a removal (ADR-0007, AC-35).
- A protocol number, once printed on a друкована копія, is never reused or cleared — not by a return to чернетка, not by withdrawal; the реєстр протоколів may have gaps (ADR-0017, AC-14b, AC-37).
- A protocol can never become підписаний without meeting facts: members present and a vote count for every agenda item, plus a note when the paper carries handwritten corrections (ADR-0018, AC-34).
- A друкована копія can never be issued while a value proposed by the помічник протоколу is unconfirmed by a person (AC-21, BR-11 of docs/design/design-spec.md).
- The попереднє голосування is always informal: it never counts as the meeting's vote and never pre-fills the meeting facts (AC-41).
- Any edit returns the draft to чернетка: an open раунд опитування closes (its votes stay in history and are not counted) and an issued друкована копія becomes void (AC-05d, AC-14b).
- Only a current член правління — a person whose term covers today — can read or change anything of the association (ADR-0003, AC-19).

## Sentinel errors

<!-- The domain refusals a caller can meet, one canonical code each (<context>.<error_name>). The code is
     the contract: the same string in the domain, the problem response and the UI copy. -->
- `identity.session_required` — nobody is signed in (AC-19).
- `identity.not_current_member` — the signed-in person is not a current член правління; the same answer for an existing and a missing record (ADR-0003, AC-19).
- `governance.settings_missing` — association settings (meeting form, signers, register start) are not filled, so no draft can start (AC-42).
- `governance.settings_invalid` / `governance.roster_entry_invalid` — a settings or roster value breaks a rule (AC-42, AC-02).
- `governance.register_start_locked` — the register start number changes after a number was issued (ADR-0017).
- `governance.draft_revision_conflict` — a write against a stale draft revision (ADR-0009, AC-29).
- `governance.transition_not_allowed` — a lifecycle move the current state does not allow, e.g. withdrawing a signed protocol or copying into a meeting with an open draft (AC-38, AC-39; sad.md §6 Flow 6).
- `governance.protocol_locked` — any change to a підписаний protocol (AC-17, AC-32; sad.md §6 Flow 6).
- `governance.poll_not_open` / `governance.vote_on_behalf` — a vote outside готовий до голосування, or for another member (AC-32, AC-05e).
- `governance.slot_unconfirmed` / `governance.draft_incomplete` — printing with an unconfirmed помічник value or an incomplete draft (AC-21, AC-04; sad.md §6 Flow 4).
- `governance.not_ready_for_signing` — uploading the signed copy while the protocol is not готовий до підпису or against a voided print copy (AC-40; sad.md §6 Flow 5).
- `governance.meeting_facts_missing` / `governance.correction_note_required` — a lock without meeting facts, or without the handwritten-corrections note (ADR-0018, AC-34, AC-15).
- `governance.reason_required` — withdrawal without a reason (AC-35).
- `governance.file_invalid` / `governance.storage_unavailable` — a file outside type/size limits (PRD §6, 20 MB PDF/JPEG/PNG), or file storage down during an upload (sad.md §6 Flows 5 and 8).
- `governance.search_invalid` — a registry search the rules reject (AC-18).

## Org-filter invariant

<!-- The tenancy rule every read and write obeys. -->
- Every record of this feature belongs to exactly one association (ОСББ) and is read and written only within it: each owned record carries its association, and every query filters by the caller's association explicitly — there is no implicit, database-level scoping (ADR-0015).
- Stage 1 has a single association, but the filter is still mandatory on every new query, so a second association never sees the first one's cases, protocols, files or roster.
- Records of another association answer as missing (not found), never as forbidden, so their existence is not revealed (ADR-0003, ADR-0015).
- Records of other bounded contexts (finance contracts, identity accounts) are referenced by id only, never joined across contexts (repo ADR 0002).

## Out of scope

<!-- Placed outside this domain for Stage 1, with the reason — see PRD.md §3 and the Stage 2 list. -->
- Blocking charter checks and their override (quorum, majority, competence limit, meeting form) · moved to Stage 2 by the product owner on 2026-10-25; Stage 1 shows tier-3 checks as advice only.
- Reading the signed scan with the помічник · Stage 2; meeting facts are entered by a person from the paper.
- Export package for the ревізійна комісія · Stage 2.
- Cross-check of amounts with finance contracts in force, and the джерело коштів link to a finance contract · Stage 2; finance is untouched in Stage 1.
- Effective-dated charter rules and building shares · Stage 2; Stage 1 keeps only the current settings.
- More than one association, invitations and row-level security · not needed for one ОСББ; the org filter keeps the door open (ADR-0015).
- Qualified electronic signature · a later release; Stage 1 signs on paper.

