---
status: Living
updated_at: "2026-10-21"
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
- головуючий — the board member who chairs a particular board meeting and signs its protocol: the chair, or a member substituting for the chair as the charter allows. NOT голова правління (the elected position; the chair may be absent and someone else presides).
- джерело коштів — the fund or income line named in the approved budget (e.g. maintenance contributions, repair fund, reserve fund, rental income) that pays for a decision. NOT стаття витрат (where the money comes from vs what it is spent on).
- кворум — the minimum number of board members present for a board meeting to take valid decisions, as set by the association's charter. NOT більшість (presence needed to decide vs votes needed to pass); NOT кворум загальних зборів (that one counts co-owners' votes, not board members).
- кошторис робіт — the management company's priced estimate for one job (works, materials, amount) that the board votes on and the protocol cites. NOT кошторис доходів і витрат (the association's annual budget approved by the general meeting); NOT рахунок (a request for payment, issued after or instead of an estimate).
- неучасть у голосуванні — a board member does not vote on one specific agenda item, with a recorded reason: left or joined during the meeting, personal interest, or another legal restriction. NOT відсутність на засіданні (out of the whole meeting); NOT утримався (abstained — voted, counted as present).
- протокол правління — the signed record of one board meeting: attendance, quorum, agenda items and each decision with its vote count. NOT протокол загальних зборів (different body, different competence and quorum).
- співвласник — an owner of an apartment or non-residential premises in the building, whether or not a member of the association, who pays contributions and votes at the general meeting. NOT мешканець (a resident may neither vote nor pay); NOT член об'єднання (membership is optional for a co-owner).
- справа — the bundle of source documents (contracts, invoices, estimates, owner requests) that a board decision is based on and cites. NOT протокол (the case is what a protocol is built from, not the protocol itself).
- управляюча компанія — the company contracted by the general meeting to maintain the buildings and manage the repair and development fund; it issues estimates and acts and holds the money. NOT правління (the board decides; the management company executes and is paid).
- частка будинку — a building's fixed share of a cost shared between the association's buildings, proportional to its area and approved by a separate inter-building decision. NOT площа будинку (the area is the input; the share is the approved result that may only change by a new decision).
- чинне рішення — a decision of the board or the general meeting that was validly adopted and has not been revoked or superseded, so new decisions must stay consistent with it. NOT підписаний протокол (a signed protocol may hold decisions that were later revoked or replaced).
