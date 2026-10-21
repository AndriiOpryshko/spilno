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
- головуючий — the board member who chairs a particular board meeting and signs its protocol: the chair, or a member substituting for the chair as the charter allows. NOT голова правління (the elected position; the chair may be absent and someone else presides).
- джерело коштів — the fund or income line named in the approved budget (e.g. maintenance contributions, repair fund, reserve fund, rental income) that pays for a decision. NOT стаття витрат (where the money comes from vs what it is spent on).
- кворум — the minimum number of board members present for a board meeting to take valid decisions, as set by the association's charter. NOT більшість (presence needed to decide vs votes needed to pass); NOT кворум загальних зборів (that one counts co-owners' votes, not board members).
- неучасть у голосуванні — a board member does not vote on one specific agenda item, with a recorded reason: left or joined during the meeting, personal interest, or another legal restriction. NOT відсутність на засіданні (out of the whole meeting); NOT утримався (abstained — voted, counted as present).
- протокол правління — the signed record of one board meeting: attendance, quorum, agenda items and each decision with its vote count. NOT протокол загальних зборів (different body, different competence and quorum).
- співвласник — an owner of an apartment or non-residential premises in the building, whether or not a member of the association, who pays contributions and votes at the general meeting. NOT мешканець (a resident may neither vote nor pay); NOT член об'єднання (membership is optional for a co-owner).
- справа — the bundle of source documents (contracts, invoices, estimates, owner requests) that a board decision is based on and cites. NOT протокол (the case is what a protocol is built from, not the protocol itself).
- чинне рішення — a decision of the board or the general meeting that was validly adopted and has not been revoked or superseded, so new decisions must stay consistent with it. NOT підписаний протокол (a signed protocol may hold decisions that were later revoked or replaced).
