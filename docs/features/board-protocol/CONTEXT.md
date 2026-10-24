---
status: Living
updated_at: "2026-10-24"
---

# Domain Context — board-protocol

<!--
CONTEXT.md is the domain glossary — not a PRD and not a scratch pad. NO implementation
detail here (no datastore/broker/framework names, no API contracts) — only domain words
and the boundaries between them. Implementation choices live in the SAD and ADRs;
behaviour lives in PRD.md.

Terms get fixed inline, the moment they surface in an interview / PRD / review — never
batched «I'll consolidate later». Empty H2 → prune before commit; keep only the sections
that carry real content. ## Glossary is mandatory; the other two are optional.
-->

## Glossary

<!-- One line per term: name · one-sentence canonical definition · one-sentence boundary
     (what it is NOT / the concept it gets confused with). Alphabetical once there are a few. -->
- акт виконаних робіт — the document signed with the contractor (usually the management company) that confirms which works were done and for what amount, closing an expense line approved by a decision. NOT кошторис робіт (the planned cost before the work, not the confirmed cost after it); NOT банківська виписка (proof of payment, not of work done).
- блокуюча перевірка — a deterministic check of a protocol draft against a правило статуту or a missing source (no meeting quorum, an amount above the board's competence limit, a number with no confirmed source in the case) that stops a draft from being marked готовий до підпису; a член правління may lift it with an обхід блокування, except a number with no confirmed source, which only a source or a person's confirmation lifts. NOT попередження (advisory, needs interpretation, never stops signing).
- головуючий — the board member who chairs a particular board meeting and signs its protocol: the chair, or a member substituting for the chair as the charter allows. NOT голова правління (the elected position; the chair may be absent and someone else presides).
- готовий до підпису — the state of a protocol draft whose text is frozen for printing or download and signing, reached only when every блокуюча перевірка is resolved or overridden; any later edit returns it to draft. NOT a locked protocol (locked only once the signed copy — a scan or a file with a qualified electronic signature — is uploaded to the реєстр протоколів).
- готовий до голосування — the state of a protocol draft whose wording is frozen so that members can cast попереднє голосування on exactly that version; any edit returns it to draft and the votes stay with the old version. NOT готовий до підпису (frozen for printing, after the blocking checks).
- джерело коштів — the fund or income line named in the approved budget (e.g. maintenance contributions, repair fund, reserve fund, rental income) that pays for a decision. NOT стаття витрат (where the money comes from vs what it is spent on).
- знятий з розгляду — the final state of a protocol draft that was never printed or signed (the proposal did not gather support or was abandoned), kept with its case, wording and preliminary votes as an internal record, not a legal document. NOT a signed protocol with a «not adopted» item; NOT a deleted draft.
- кворум — the minimum number of board members present for a board meeting to take valid decisions, as set by the association's charter. NOT більшість (presence needed to decide vs votes needed to pass); NOT кворум загальних зборів (that one counts co-owners' votes, not board members).
- кошторис робіт — the management company's priced estimate for one job (works, materials, amount) that the board votes on and the protocol cites. NOT кошторис доходів і витрат (the association's annual budget approved by the general meeting); NOT рахунок (a request for payment, issued after or instead of an estimate).
- неучасть у голосуванні — a board member does not vote on one specific agenda item, with a recorded reason: left or joined during the meeting, personal interest, or another legal restriction. NOT відсутність на засіданні (out of the whole meeting); NOT утримався (abstained — voted, counted as present).
- обхід блокування — a член правління's recorded choice to mark a draft готовий до підпису despite a блокуюча перевірка, with a mandatory reason that stays with the protocol in the реєстр протоколів. NOT signing through a попередження (no reason required, nothing is overridden).
- попередження — an advisory notice on a protocol draft, linked to its source (a charter clause, a related чинне рішення, a possible personal interest, a required wording), that the головуючий may sign through. NOT блокуюча перевірка (deterministic, stops signing).
- попереднє голосування — the votes per agenda item that each член правління casts in the system on one frozen version of the draft wording, before the meeting result is fixed on paper; quorum and majority are checked against it, it is tied to that wording version, and a missing or failed preliminary majority warns but never stops printing. NOT the binding vote (the signed paper protocol records that; the paper outranks the system); NOT votes typed in by someone on another member's behalf.
- помічник протоколу — the conversation in which a член правління attaches case documents and asks for a protocol draft; it writes wording and advisory warnings and proposes amounts read from documents, but every number it proposes counts only after a person confirms it. NOT the source of any number in the protocol; NOT the board chat where members discuss.
- правило статуту — one charter parameter kept as data with the date it takes effect (quorum size, the board's competence limit, whether a non-voting member counts toward an item's quorum), applied to meetings held on or after that date. NOT статут (the whole legal document); NOT чинне рішення (a decision, not a standing rule of procedure).
- протокол правління — the signed record of one board meeting: attendance, quorum, agenda items and each decision with its vote count. NOT протокол загальних зборів (different body, different competence and quorum).
- ревізійна комісія — the association's audit body elected by the general meeting that checks the board's decisions and spending against protocols and case documents; it takes no board decisions. NOT правління; NOT a співвласник asking a question (an audit is a formal check with its own mandate).
- реєстр протоколів — the association's searchable record of locked protocols with their signed copies, decisions, cited case documents and recorded overrides; it belongs to the association, not to any board member. NOT справа (the case is the input to a protocol; the registry holds the result).
- слот рішення — one typed field of a board decision (amount, counterparty, джерело коштів, votes, неучасть у голосуванні) holding a value, where it came from, and the case document it is confirmed against. NOT вільний текст рішення (connecting prose that may hold no numbers or facts).
- співвласник — an owner of an apartment or non-residential premises in the building, whether or not a member of the association, who pays contributions and votes at the general meeting. NOT мешканець (a resident may neither vote nor pay); NOT член об'єднання (membership is optional for a co-owner).
- справа — the bundle of source documents (contracts, invoices, estimates, owner requests) that a board decision is based on and cites. NOT протокол (the case is what a protocol is built from, not the protocol itself).
- управляюча компанія — the company contracted by the general meeting to maintain the buildings and manage the repair and development fund; it issues estimates and acts and holds the money. NOT правління (the board decides; the management company executes and is paid).
- частка будинку — a building's fixed share of a cost shared between the association's buildings, proportional to its area and approved by a separate inter-building decision. NOT площа будинку (the area is the input; the share is the approved result that may only change by a new decision).
- чинне рішення — a decision of the board or the general meeting that was validly adopted and has not been revoked or superseded, so new decisions must stay consistent with it. NOT підписаний протокол (a signed protocol may hold decisions that were later revoked or replaced).
- член правління — a person elected by the general meeting to the board for a term, who may take part and vote at board meetings while the term lasts. NOT головуючий (a per-meeting role a member takes); NOT член об'єднання (a co-owner who joined the association).
- шаблон формулювання — a fixed wording for an agenda-item type that the law or the charter requires to be phrased a certain way; only its слоти рішення may be filled, the text itself is not edited. NOT вільний текст рішення.
