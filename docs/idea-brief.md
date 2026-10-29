---
status: Confirmed
owner: "Andrii Opryshko"
reviewers: []
updated_at: "2026-10-29"
feature_size: TBD               # product-level brief; features are sized by sdlc:classify-size
stage: "01"
depth: medium
ticket: ""
scope: product                  # product-level brief; feature briefs live in docs/features/<slug>/
value_score:
  rice: 3.2
  state: confirmed
  confirmed_at: "2026-10-29"
feasibility_state: confirmed
---

<!-- Stage 01 → see SDLC/plugin/skills/interview/SKILL.md (product-level run) -->

# Idea Brief — spilno (product)

## 1. Raw idea
The author's own words (Ukrainian, verbatim, from the board-protocol discovery interview, confirmed as the product's starting point on 2026-10-29):

> в моєму випадку в нас є управляюча компанія. зазвичай ми щось просимо її зробити та вона нам кидає кошторис. і на цьому кошторису ми робимо голосування та підкладаємо його. в нас буває коли ми голосуємо за трати будинку, а буває ділемо щось між будиками пропорційно площі. рідко це щось інше. коли в нас є кошторис і опис на будинок чи скидаємось з цією інфою я іду в чат тг правління, кидаю і ми типу голосуємо. якщо все ок - друкую та підписуємо. якщо не набирається голоси - я навіть не друкую, а сенс.

Plus three pains (translated): hours per protocol spent copying sums by hand, with mistakes; decisions scattered across chat, Word and spreadsheets, so half a year later nobody can show an auditor or a co-owner what a decision was based on; and the chair not remembering voting restrictions, required wording or earlier decisions, and not wanting to get into trouble.

Product framing from this interview: in a year spilno is the association's whole management tool; the MVP is board protocols and the decision registry.

## 2. Problem
One association produced 30+ board protocols in under two years, each taking 1-3 hours from the management company's кошторис робіт to a signed paper. The vote happens in the board chat, the protocol in a document editor with help from general-purpose AI chats, and the акт виконаних робіт arrives later and lies apart from both. For almost no spending decision can the board show, in a minute, the chain estimate → signed protocol → act or payment.

## 3. Users
- **MVP — правління** (~5 members). The головуючий (usually the chair, a volunteer, not a lawyer) writes 15-20 protocols a year; members vote and sign; substitutions are normal.
- **MVP, read side — ревізійна комісія** (~3): checks spending against protocols once a year.
- **Vision (year one+) — співвласники** of the same association: "what did our money pay for, and on whose decision".
- Only the author's own association; no other associations are served.

## 4. Why now
The association is under two years old and already has 30+ protocols and at least one post-signature correction; the next annual general meeting will ask the audit commission for the trail. The app skeleton, the case and protocol model and the first screen already exist (board-protocol, 9 of 44 tasks done in 8 days). The repo also serves as the proving ground of a learning course, so the work has a schedule.

## 5. Out of scope
- Other associations, onboarding, pricing, support (strategy is one association).
- Charges, receipts and payments to co-owners; general meetings and co-owner e-voting (existing services, §6; vision, not MVP).
- A co-owner cabinet in the MVP: co-owners get answers through the board.
- Backfilling the 30+ past protocols with a full trail (parked, §14).
- Legal advice: the product reminds and cites, it does not decide legality.

## 6. Competitive analysis
| # | Product · URL | Features | Value per feature (1-5) | Gap |
|---|---|---|---|---|
| 1 | Kyiv Digital, OSBB service · https://kyivcity.gov.ua/news/novi_funktsi_miskogo_zastosunku_komanda_kiv_tsifroviy_rozpochinaye_beta-testuvannya_servisu_dlya_goliv_osbb/ | Co-owner e-votes with e-signature, protocol upload to e-cabinet | Votes 3 · upload 2 · trail 1 | General meetings only; no board protocols, no link from a number to an estimate or act; Kyiv-specific, beta status unconfirmed ^1 |
| 2 | PrivatBank online voting · https://privatbank.ua/en/news/2025/4/23/privatbank-zapustiv-dlya-osbb-servis-onlayn-golosuvannya-zagalnih-zboriv-dlya-uchasti-u-programi-tepli-krediti | Remote general meetings, minutes for a loan programme, free | Voting 3 · minutes 2 · cost 5 | Loan-programme shaped; no board decisions, citations or registry ^1 |
| 3 | Мій Дім Онлайн · https://jobs.dou.ua/companies/miydimonline/ | Billing, receipts, resident cabinet, metering | Billing 4 · protocols 1 | Fits the later vision (charges, cabinet), not the MVP; thin source ^2 |
| 4 | osbb-online · https://www.education.ua/courses/47766/ | Accounting software sold with a course | Accounting 3 · decisions 1 | Only a course page; accounting, not board decisions ^2 |
| 5 | BoardPro · https://www.boardpro.com/en-gb/lp/board-management-software-uk | Agenda, minutes, actions, decision register | Register 4 · cost 1 | ~$150+/month per board; English; no Ukrainian charter model, no quorum or competence check, no per-number link to an estimate ^3 |
| 6 | Buildium (also PayHOA, Easy HOA) · https://www.buildium.com/association-management-software-lp/ | Board tools, committee votes on requests, resident centre, files | Portal 3 · trail 2 | US rules and language; no protocol from estimates, no quorum check ^4 |
| 7 | Status quo: document editor + AI chats + board chat · n/a | Free-form drafting, chat poll | 2 | Nothing enforces numbers, quorum or restrictions; no trail to source documents |

^1 2026-10-29, "ОСББ онлайн платформа кабінет мешканця голосування протоколи". ^2 2026-10-29, "програма для ОСББ управління будинком облік платежів протоколи документи сервіс ціна". ^3 2026-10-29, "BoardPro minutes board meeting software features pricing". ^4 2026-10-29, "Buildium HOA board management software features architectural requests documents pricing". Row 7 is from the interview, not a search.

**Gap:** no verified product keeps the board-level chain "estimate → drafted protocol → signed protocol → act" with Ukrainian quorum and competence rules; Ukrainian services stop at general meetings, billing or loan paperwork, foreign board portals lack local rules and cost too much for a volunteer board.

## 7. Strategic approaches

### Approach A — Retrofit Decision Trail Registry
- **Thesis**: a registry links each decision's estimate, signed scan and act; past protocols are backfilled, drafting stays where it is.
- **For whom**: правління, ревізійна комісія.
- **Outcome metric**: decisions (past included) with full trail in under a minute, ~0% → ≥80%.
- **Key trade-off**: no time saved, no safeguards, built drafting work parked.
- **Effort signal**: S (≈ 1-2.5 months of evenings).
- **Recommended?** ◯

### Approach B — Co-owner-Facing Accountability Loop
- **Thesis**: each estimate opens a loop closed in public: vote and signing in the product, act checked against estimate, area split, co-owner cabinet.
- **For whom**: співвласники, ревізійна комісія.
- **Outcome metric**: last 12 months' decisions a co-owner traces alone in under a minute, ~0% → 100%.
- **Key trade-off**: a year before board value; personal data widely exposed.
- **Effort signal**: L (≈ 6-18+ months of evenings).
- **Recommended?** ◯

### Approach C — Case-Driven Board Back-office
- **Thesis**: the справа is the backbone: an estimate opens a case, the preliminary vote moves from the chat into it, the protocol is drafted from its documents, and it stays open until the act or payment is attached.
- **For whom**: головуючий, substitute chairs, ревізійна комісія.
- **Outcome metric**: new spending decisions with full trail in under a minute, ~0% → ≥90%; leading signal: cases with no act after 60 days.
- **Key trade-off**: no past history; co-owner cabinet slips past year one.
- **Effort signal**: M (≈ 3-6 months of evenings), reuses board-protocol.
- **Recommended?** ●

## 8. Multi-perspective feedback

### Engineer
- A: smallest build, but backfilling mixed-format scans is manual toil and splits the registry into two integrity levels.
- B: widest surface (co-owner accounts, per-unit access, per-unit areas that do not exist yet) and a check that breaks when the management company changes its format.
- C: reuses the case, protocol, poll and snapshot model already built; adds reminders that can be missed or duplicated.

### Executive
- A: fastest audit win and the only one covering past protocols, but strands built work.
- B: best fit to the vision, but a year-long bet for one evening developer.
- C: pays off sunk work, moves both metrics, gives the board an ongoing control.

### UX-researcher
- A: no habit change, but a filing chore after signing that quietly goes incomplete.
- B: biggest habit change; board payoff comes last.
- C: mirrors today's flow; moving a one-tap chat poll behind a login is the risky step.

### Synthesis matrix
|        | Engineer | Executive | UX |
|--------|:--------:|:---------:|:--:|
| App. A | + smallest build; backfill toil, model fork | 0 fast audit win, strands built work | 0 low friction, but a chore |
| App. B | − widest surface, external format dependency | − year-long bet, core pain delayed | − heavy onboarding, delayed board payoff |
| App. C | + reuses built core; adds scheduler | + reuses built work, hits both metrics | + mirrors current flow; vote migration risky |

## 9. Trade-offs and edge cases

### Trade-offs per approach
| Approach | Pros | Cons |
|---|---|---|
| A | Smallest; covers past decisions; no habit change | No time saved; no safeguards; parks built work; two integrity levels |
| B | Best fit to the vision; co-owner self-service | ~1 year to board value; personal-data exposure; external format dependency |
| C | Reuses built work; both metrics; trail enforced at entry | No past history; vote must leave the chat; reminders add noise risk |

### Edge cases
- Board roster changed (resignation, new term) but not updated in the product.
- A cost split between buildings needs its own decision on the area share.
- A proposal without enough votes is withdrawn, never printed.
- A fix after signing goes only through a new referencing protocol.
- The act arrives late, never, or for a different amount.
- A substitute presides; a successor chair takes over the product.

## 10. Risks
- **Top (devil's advocate): false legal authority.** The roster, charter rules or a member's restriction are out of date, the draft still says "quorum OK", the chair signs because the product said so, and the decision is challenged. Signal: an audit finding; the chair stops trusting checks. Mitigation: checks only warn and show the data and date they used.
- Drift back to the chat: votes too rare to build a habit; signal: in-product votes only from the chair.
- Amounts misread from estimates; signal: act or payment differs from the protocol.
- Acts not delivered; signal: cases without an act after 60 days keep growing.
- One person holds everything: burnout or leaving the board leaves the record unmaintained.
- Value too late, and personal-data exposure from a self-run service.

## 11. RICE — Claude proposed
- **Reach (R)**: 8 — people touched per quarter: правління ~5 + ревізійна комісія ~3 (§3).
- **Impact (I)**: 2 — the core pain of §2 (hours per protocol, decisions not provable at audit); Executive rates C "+" for moving both metrics (§8).
- **Confidence (C)**: 0.8 — facts are concrete (30+ protocols, built core), but §15 holds open questions on vote migration, act delivery and the time baseline.
- **Effort (E)**: 4 person-weeks — C's signal is M (§7), reduced by reuse of board-protocol.
- **RICE = 8 × 2 × 0.8 / 4 = 3.2**
- **State**: confirmed (2026-10-29)

## 12. Feasibility — Claude proposed

- [☑] **Tech**: governance holds meetings, cases, protocols, poll rounds and snapshots, finance holds contracts, background jobs already exist; reminders are one more job (`services/core/internal/{governance,finance}`, 27 migrations, cases and protocol screens).
- [☑] **Skills**: the author built the skeleton, three contexts, the full board-protocol spec package and 107 commits in 8 days, and knows the domain from the inside. Legal rules stay advisory in the MVP; open legal questions are in §15.
- [☑] **Time**: C fits 3-6 months of evenings; board-protocol phase 0 (9 of 44 tasks) shipped in one run. First real protocol in the product: December 2026.
- **State**: confirmed (2026-10-29)

## 13. Recommendation
**Selected: Approach C — Case-Driven Board Back-office.** It is the only approach scored "+" by all three perspectives (§8: Engineer "reuses built core; adds scheduler", Executive "reuses built work, hits both metrics", UX "mirrors current flow"). It targets the gap no product in §6 fills: a board-level chain estimate → protocol → act with Ukrainian quorum and competence rules, where local services stop at general meetings and billing and foreign board portals lack local rules. With RICE 3.2 (§11) against A's lower impact and B's year-long effort, and Feasibility 3/3 ☑ (§12) grounded in the governance context already built, C gives the board value within the year while keeping the co-owner vision reachable later. Its weak spots (no history, chat-to-product vote) are in §10 and §15.

**Locked-in pointer**: the справа is the product's unit of work; every spending decision starts from a case with its estimate and ends with an act or payment attached; checks warn, never decide; the MVP serves the board and the audit commission of one association. The first feature on this path is board-protocol (`docs/features/board-protocol/`).

## 14. Parked & rejected approaches
| # | Approach | Status | Reason | Revisit trigger |
|---|---|:---:|---|---|
| A | Retrofit Decision Trail Registry | parked | Saves no time, parks built work; its backfill is the one piece C lacks | The audit commission asks for a pre-launch decision the registry cannot show; then backfill scans only |
| B | Co-owner-Facing Accountability Loop | parked | ~1 year before board value, personal-data exposure, external format dependency | C is in daily use for a full quarter and ≥90% of new decisions have a full trail |
| - | Selling to or onboarding other associations | rejected | Strategy is one association plus the course | - |

## 15. Open questions
- [ ] Will members vote in the product, or does the vote need a notice in the board chat to work at all — owner: Andrii Opryshko, due: before the board-protocol poll UI ships
- [ ] How often and how late does the management company deliver acts; is 60 days the right "open case" threshold — owner: Andrii Opryshko, due: 2026-11-30
- [ ] Baseline time per protocol: time the next 2-3 protocols — owner: Andrii Opryshko, due: 2026-11-30
- [ ] Backfill old signed scans (date and number only) or leave them outside — owner: Andrii Opryshko, due: before the registry ships
- [ ] Vision includes co-owners while strategy is one association: confirm co-owners of this association only — owner: Andrii Opryshko, due: before the co-owner phase
- [ ] Who takes over the product if the author leaves the board — owner: Andrii Opryshko, due: 2027-03-31
- [ ] Pending glossary term for `/sdlc:fix-term`: **слід рішення** (decision trail: estimate → signed protocol → act or payment, as one chain) — owner: Andrii Opryshko, due: before the next PRD
- [ ] Recommendation fixes the справа as the unit of work, a structural choice: confirm in the SAD, not here — owner: Tech Lead, due: gate 3

## Related
- Feature: `docs/features/board-protocol/` (idea-brief, PRD, sad, data-model, contracts, tasks, test-plan).
- Glossary: `docs/features/board-protocol/CONTEXT.md` (справа, акт виконаних робіт, кошторис робіт, реєстр протоколів, ревізійна комісія, співвласник).
- Architecture decisions: `docs/adr/`.

## DoD self-check
- [x] 15 sections present
- [x] No anti-pattern terms (datastore/broker names, latency targets)
- [x] Length ≤ 5 pages (~2200 words)
- [x] Frontmatter status: Confirmed
- [x] RICE confirmed (state: confirmed)
- [x] Feasibility confirmed (state: confirmed)
- [x] Recommendation present with rationale citing 4 upstream sections (§6, §8, §11, §12)
