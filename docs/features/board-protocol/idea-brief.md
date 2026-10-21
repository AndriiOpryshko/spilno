---
status: Confirmed
owner: "Andrii Opryshko"
reviewers: []
updated_at: "2026-10-21"
feature_size: TBD               # set by sdlc:classify-size, not here
stage: "01"
depth: medium
ticket: ""
value_score:
  rice: 2.1
  state: confirmed
  confirmed_at: "2026-10-21"
feasibility_state: confirmed
---

<!-- Stage 01 → see SDLC/plugin/skills/interview/SKILL.md -->

# Idea Brief — Board protocol

## 1. Raw idea
Translated from Ukrainian; the author selected both draft framings and added the third point.

1. The board chair spends hours on a board meeting protocol: copies sums from contracts and invoices into a Word template and makes mistakes. The system should assemble a protocol draft from the case documents, check quorum and board competence, and the human only edits and signs.
2. Board decisions are scattered across Viber, Word and Excel, and half a year later nobody can show an auditor or a co-owner which documents and what quorum a decision was based on. Want a single registry of signed protocols where every number cites a case document.
3. Sometimes a member cannot vote on an item because of a legal restriction, an item needs special wording, or earlier rules or protocols contradict the new decision. The chair does not always know or remembers, and does not want to get into trouble.

## 2. Problem
One association produced 30+ board protocols in under two years, each assembled by hand from contracts, invoices and chat approvals. Post-signature corrections and re-votes have already happened; numbers have no link to their source, and nothing reminds the author of the charter, voting restrictions or prior decisions.

## 3. Users
- **Primary — головуючий**: usually the chair (a volunteer, not a lawyer); sometimes another board member or secretary when the chair is substituted. Writes ~15-20 protocols a year.
- **Secondary — reviewers of the record**: audit commission, bank, co-owners asking "why did we pay this?".
- Board composition changes meeting to meeting: absences and substitutions are normal.

## 4. Why now
The author's own association is under two years old and already has 30+ protocols and at least one correction. The app skeleton and a protocol-draft UI prototype already exist in the spilno repo. Rollout: own association first, charter rules kept as data so neighbouring associations (more protocols per quarter) can follow.

## 5. Out of scope
- General-meeting protocols and co-owner e-voting (covered by existing services, §6).
- Onboarding other associations and co-owner read access (first release serves one association).
- Blocking signing on automated legal judgments.
- Legal advice: the product reminds and cites; it does not decide legality.

## 6. Competitive analysis
| # | Product · URL | Features | Value per feature (1-5) | Gap |
|---|---|---|---|---|
| 1 | Kyiv Digital, OSBB cabinet · https://suspilne.media/640356-novi-funkcii-u-kiiv-cifrovij-zavivsa-servis-dla-osbb/ | Co-owner registry, e-votes with bank ID, protocol upload | 2 | General meetings only; no drafting from case documents, no per-number citation, no charter check; Kyiv only |
| 2 | PrivatBank OSBB voting · https://privatbank.ua/en/news/2025/4/23/privatbank-zapustiv-dlya-osbb-servis-onlayn-golosuvannya-zagalnih-zboriv-dlya-uchasti-u-programi-tepli-krediti | Online general meetings, protocol for a loan programme | 2 | General meetings only; no board protocols, no citations, no restriction or contradiction checks |
| 3 | Мій Дім Онлайн · https://jobs.dou.ua/companies/miydimonline/ | Resident cabinet, receipts, requests to the board | 1-2 | Billing and communication; no evidence of protocol features (thin source) |
| 4 | BoardPro (board portals: OnBoard, Boardable) · https://www.boardpro.com/features/minutes | AI minutes from transcript, e-sign lock, searchable decision register | 3 | Drafts from what was said, not from documents; no Ukrainian charter model, no competence or quorum check, per-user pricing heavy for a volunteer board |
| 5 | AI notetakers (Otter, Fireflies) · https://fireflies.ai/blog/meeting-notes-software/ | Transcripts, summaries, action items | 2 | No signed record, no quorum, no source citation |
| 6 | Status quo: Word template + Viber/Telegram + Excel | Free-form text, chat approvals, spreadsheets | 2 | Nothing enforced; finding a decision means scrolling chats |

Footnotes (all 2026-10-21): 1 — «електронне голосування ОСББ кабінет»; 2-3 — «програма для ОСББ протоколи документи онлайн сервіс»; 4 — «board portal meeting minutes e-signature»; 5 — «AI meeting minutes generator board». Row 6 is the stated status quo; the dropped OSBB service is unidentified (§15).

**Key gap:** no product drafts a board protocol from case documents with a source citation per number and a charter quorum/competence check; none flags non-voting members or contradictions with prior decisions.

## 7. Strategic approaches

### Approach A — Case-Linked Protocol Registry
- **Thesis**: a structured protocol form where attendance yields quorum and every amount must be attached to a case document; signed protocols land in a searchable registry; no drafting, guardrails manual.
- **For whom**: the chair and later reviewers (auditor, co-owner).
- **Outcome metric**: time to find a past decision with its case and quorum: often impossible after ~6 months → ≤1 minute.
- **Key trade-off**: staying out of trouble still depends on the chair's memory; time per protocol barely drops.
- **Effort signal**: S
- **Recommended?** ◯ (becomes stage 1 of C)

### Approach B — Autonomous Charter-Aware Drafting
- **Thesis**: upload the case, get a ready-to-sign protocol; the system extracts numbers, infers conflicted members, inserts fixed wording, checks against all prior decisions and blocks signing on any problem; multi-association from day one.
- **For whom**: time-poor or new chairs; later, other associations.
- **Outcome metric**: post-signature corrections: TBD → 0 per year.
- **Key trade-off**: the product makes legal judgments it can get wrong, and a volunteer chair will over-trust it; hard blocks delay shipping.
- **Effort signal**: L
- **Recommended?** ◯

### Approach C — Confirmed Slots, AI Wording
- **Thesis**: each decision is built from typed slots (amount, counterparty, source of funds, votes, non-voting members) pre-filled from a case document and confirmed by a human; AI writes only connecting text; charter rules, non-voting, mandatory wording and related prior decisions surface as non-blocking warnings linked to their source.
- **For whom**: the головуючий of the author's association, including substitutes.
- **Outcome metric**: post-signature corrections TBD → 0 per year; leading: cited and confirmed numbers ~0% → 100%.
- **Key trade-off**: warnings can be signed through; contradictions are surfaced, not detected; one association only.
- **Effort signal**: M
- **Recommended?** ● (staged)

## 8. Multi-perspective feedback

### Engineer
- Volume is tiny; correctness, provenance and immutable signed records matter.
- A: smallest surface, no inference risk; its model is a subset of C, so A → C is incremental.
- B: open-ended reasoning over meaning plus a missing related-party dataset; blocking false positives halt signing.
- C: human confirmation bounds risk; AI text must not introduce numbers or facts.

### Executive
- Value is avoiding errors with outside consequences and fast retrieval, not hours saved.
- A ships fastest but resembles existing stores and the dropped service; B differentiates but is L effort plus liability for a solo builder.
- C fills the market gap at M effort, measurable from the first protocol.

### UX-researcher
- Authors rotate, so the core problem is missing institutional memory.
- A: friction over Word with no drafting payoff. B: hard-block dead ends, over-trust, auto-removal from a vote feels like an accusation.
- C: human in control, warnings linked to rules teach the chair; risk of warning fatigue.

### Synthesis matrix
|         | Engineer | Executive | UX |
|---------|:--------:|:---------:|:--:|
| App. A  | + smallest surface, no inference risk | 0 fast, undifferentiated, little relief | − more friction, no drafting payoff |
| App. B  | − open-ended legal inference, blocking false-positives | − moat, but L-effort solo liability | − over-trust, hard-block dead-ends |
| App. C  | 0 confirmation bounds extraction risk | + market-gap value at M effort | + cited, confirmable, teaches the rules |

## 9. Trade-offs and edge cases

### Trade-offs per approach
| Approach | Pros | Cons |
|---|---|---|
| A | Cheapest; safe; delivers retrieval | Leaves "avoid trouble" and time per protocol unsolved |
| B | Biggest wow; least typing | Legal judgments, over-trust, hard blocks, L effort |
| C | Human control with cited numbers; teaches rules | Warnings ignorable; more clicks than B; one association |

### Edge cases
- Board roster or charter changed between meetings: checks must use the rules and roster in force on the meeting date.
- Member does not vote on one item (left the meeting, personal interest): item quorum and majority recalculated; legal effect on quorum unresolved (§15).
- Right document, wrong number: with/without VAT, superseded estimate, advance vs full amount.
- Item requires mandatory wording; AI text must not change who is authorised to do what.
- New decision conflicts with a чинне рішення of the board or general meeting.
- Approval happened in a messenger and is recorded after the fact as a meeting.
- Paper copy signed with a handwritten correction diverges from the registry.
- Chair is replaced or re-elected; the outgoing chair keeps access.
- Emergency repair needs a decision tonight.

## 10. Risks
- **Top (devil's advocate): stale ruleset or roster.** Checks run against today's charter and board instead of those in force on the meeting date, and the registry presents the wrong result as verified. Likely in an association's first two years. Signal: the audit commission recalculates quorum; re-votes rise after a charter change.
- Signed-through warnings later show the chair "was warned".
- A valid citation makes a wrong number look trustworthy.
- Adoption: at 15-20 protocols a year the chair relearns the tool each time; a similar service was already abandoned.
- Personal data in case documents exposed to wider access.

## 11. RICE — Claude proposed
- **Reach (R)**: 8 protocols per quarter — the author's association is the lower bound (~5), neighbouring associations produce more (§3).
- **Impact (I)**: 1.5 — between medium and high per the author: signed protocols carry legal weight and corrections have happened, but no external incident yet and time savings are modest (§2, §8 Executive). Off the standard scale by the author's choice.
- **Confidence (C)**: 0.7 — pain confirmed by experience and the market gap, but nine open questions remain, including two missing baselines and an unverified legal rule (§15).
- **Effort (E)**: 4 person-weeks — mid M from §7.
- **RICE = 8 × 1.5 × 0.7 / 4 = 2.1** (same inputs: A ≈ 4.3 at E 1.5 and I 1; B ≈ 1.3 at E 9, I 3, C 0.5).
- **State**: confirmed

## 12. Feasibility — Claude proposed

- [☐] **Tech**: no adjacent module in petforge (a plugin repo). The target app spilno has an app skeleton with only platform modules, one migration and a UI prototype of the cases list and protocol draft (commit 8ad3124); no members, votes or documents domain yet.
- [☑] **Skills**: the author works inside the domain (charter, quorum, competence) and owns the app's stack; the legal side may be thin but is not blocking now (§15).
- [☐] **Time**: built in spare time, so 4 person-weeks stretch to months; no shipped feature to calibrate against. Delivery must be staged.
- **State**: confirmed

## 13. Recommendation
**Selected: Approach C, staged** — C scores RICE 2.1 (§11) against A's 4.3, but A's lead comes only from cheapness: it leaves the "avoid trouble" pain unsolved and scores − on UX ("more friction, no drafting payoff", §8), while C scores + for both Executive and UX. C targets the documented market gap — no product drafts a board protocol from case documents with a citation per number and a charter check (§6). Feasibility is 1/3 ☑ (§12): Tech and Time are open, so C is delivered in stages where each stage is useful alone, starting from A's core, which is a strict subset of C's model (§8 Engineer).

- **Stage 1**: roster and charter rules bound to the meeting date (answers the top risk, §10), attendance and quorum, per-item votes with неучасть у голосуванні, every number cited to a case document, registry of signed protocols.
- **Stage 2**: slot pre-fill from case documents, locked mandatory-wording templates, charter warnings, related чинні рішення shown alongside.
- **Stage 3**: AI connecting text that cannot introduce numbers or facts.

**Locked-in pointer**: the PRD scopes Stage 1 as the first release, with Stages 2-3 as named later increments; no blocking legal checks; one association.

## 14. Parked & rejected approaches
| # | Approach | Status | Reason | Revisit trigger |
|---|---|:---:|---|---|
| A | Case-Linked Protocol Registry | parked | Absorbed as Stage 1 of C, not a product on its own | If C Stage 2 stalls for more than a quarter, ship A alone |
| B | Autonomous Charter-Aware Drafting | parked | L effort, legal liability, over-trust | C Stage 3 in use by 2+ associations and templates reviewed by a lawyer |
| - | Blocking signing on automated checks | rejected | Dead ends during emergencies push the chair back to Word | - |

## 15. Open questions
- [ ] Baseline time per protocol — time the next 2-3 protocols — owner: Andrii Opryshko, due: 2026-11-30
- [ ] How many of the 30+ protocols needed post-signature correction or re-vote — owner: Andrii Opryshko, due: 2026-11-04
- [ ] Does the OSBB law or the charter restrict a board member's vote on personal interest, and does a non-voting member count toward that item's quorum — owner: Andrii Opryshko, due: before PRD
- [ ] Does the charter allow remote or written board voting — owner: Andrii Opryshko, due: before PRD
- [ ] Which specialised OSBB service was tried and why it was dropped — owner: Andrii Opryshko, due: 2026-11-04
- [ ] Which item types need mandatory wording, and who reviews the templates — owner: Andrii Opryshko, due: before Stage 2
- [ ] Registry access on chair change: who holds it and how it is handed over — owner: Andrii Opryshko, due: before PRD
- [ ] Who may see personal data in case documents — owner: Andrii Opryshko, due: before PRD
- [ ] Confirm the feature is built in spilno and map its architecture — owner: Andrii Opryshko, due: before architecture-design

## Related
- [CONTEXT.md](./CONTEXT.md) — glossary (головуючий, джерело коштів, кворум, неучасть у голосуванні, протокол правління, співвласник, справа, чинне рішення)
- spilno repo: protocol draft UI prototype (commit 8ad3124)

## DoD self-check
- [x] 15 sections present
- [x] No anti-pattern terms (Postgres/Redis/etc.)
- [x] Length ≤ 5 pages (~2200 words)
- [x] Frontmatter status: Confirmed
- [x] RICE confirmed (state: confirmed)
- [x] Feasibility confirmed (state: confirmed)
- [x] Recommendation present with rationale citing 4 upstream sections (§6, §8, §11, §12)
