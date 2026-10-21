---
status: Confirmed
skill: ao-discovery
owner: "Andrii Opryshko (board member, idea author)"
reviewers: []
updated_at: "2026-10-21"
feature_size:                    # set by classify-size, not here
stage: "01"
depth: medium
ticket: ""
value_score:
  rice: 60.7
  state: confirmed
  confirmed_at: "2026-10-21"
feasibility_state: confirmed
---

<!-- Collision rule: idea-brief.md in this folder was written by another skill; this file sits next to it. -->

# Idea Brief — Board protocol (протокол правління)

## 1. Raw idea
> в моєму випадку в нас є управляюча компанія. зазвичай ми щось просимо її зробити та вона нам кидає кошторис. і на цьому кошторису ми робимо голосування та підкладаємо його. в нас буває коли ми голосуємо за трати будинку, а буває ділемо щось між будиками пропорційно площі. рідко це щось інше. коли в нас є кошторис і опис на будинок чи скидаємось з цією інфою я іду в чат тг правління, кидаю і ми типу голосуємо. якщо все ок - друкую та підписуємо. якщо не набирається голоси - я навіть не друкую, а сенс. іноді може буде так треба, але наразі не було потреби зафіксувати негативне голосування. іноді ми приймаємо якісь інші рішення, типу може якісь дозвіл комусь на основі чогось, типу скління балконів наприклад, хоча на практиці не робили. може ще щось бути, якісь рішеня. може виділити зі спільних площ щось для осбб, ну я тут не супер сильний та і багато ми не робили

## 2. Problem
A board protocol takes 1-3 hours from the management company's кошторис робіт to a signed paper, 10-30 times a year. So protocols are written "when there is time": decisions taken in the board chat stay unrecorded for weeks, and closing papers (акт виконаних робіт, bank statement) lie apart from the protocol. The gap shows at the annual general meeting, where the auditor checks spending and co-owners ask where the money went.

## 3. Users
- **Prepares**: one board member (today the idea author): collects the estimate, runs the chat poll, types and prints.
- **Signs**: the головуючий and the board (4 people); the seal is with the chair or the elected deputy, who substitutes by a service note.
- **Reads later**: the auditor and co-owners at the annual meeting; the management company as the basis for works and payment.
- **Frequency**: 10-30 protocols a year per association; about 13 associations in the target group.

## 4. Why now
Nothing has gone badly wrong yet, but everything rests on "we will do it when there is time". The annual general meeting is the forcing date: every spending decision of the year must have a signed protocol and its papers by then.

## 5. Out of scope
- General-meeting protocols and co-owner voting.
- Collecting contributions and suing debtors (the management company's job).
- Replacing the board chat poll; its result is transferred.
- Electronic signature; paper signing stays.
- Preparing the annual budget (кошторис доходів і витрат).

## 6. Status quo and alternatives

### Status quo
Request to the management company -> кошторис робіт -> poll in the board chat -> with 3 or more of 4 votes "for", the protocol is typed from an old text, printed and signed in person; a failed vote is not recorded. Building shares are taken from the inter-building decision and calculated by hand. Acts and bank statements are kept wherever they land.
**Cost**: 10-90 hours a year per association; protocols lag behind decisions; no record of failed proposals; no place that shows an expense line as closed.

### Alternatives
| # | Product · URL | Features | Value per feature (1-5) | Gap |
|---|---|---|---|---|
| 1 | PrivatBank online voting · https://privatbank.ua/osbb | co-owner voting by area | voting 4 | general meetings only, no board protocol |
| 2 | Kyiv Digital "Електронна демократія" · https://kyivcity.gov.ua | co-owner participation for association heads | participation 3 | Kyiv only, co-owner side |
| 3 | Мій Дім Онлайн · https://jobs.dou.ua/companies/miydimonline/ | accounting, receipts, resident cabinet | accounting 4 | no protocol, no link from decision to estimate and act |

Queries: "ОСББ сервіс протокол засідання правління онлайн голосування", "програма для ОСББ кабінет правління документи кошторис управитель", 2026-10-21. None builds a board protocol from an estimate with a case file.

## 7. Strategic approaches

### Approach A — Fixed-block protocol form
- **Thesis**: one form with fixed blocks; the preparer types every number; papers are plain attachments.
- **For whom**: a preparer who only wants a printable protocol fast.
- **Outcome metric**: time per protocol 1-3 h -> 20 min.
- **Key trade-off**: nothing is checked (amounts, board limit, conflict with a чинне рішення); no closure view.
- **AI boundary**: no AI; only requisites and board composition pre-filled.
- **Effort signal**: S
- **Recommended?** ◯

### Approach B — AI-drafted protocol copilot
- **Thesis**: AI drafts the whole protocol from the справа, amounts, shares and джерело коштів included; warnings may block.
- **For whom**: boards with frequent free-form decisions.
- **Outcome metric**: time per protocol 1-3 h -> 10 min.
- **Key trade-off**: a misread amount goes straight to print and starts a dispute.
- **AI boundary**: AI writes everything; people edit and sign.
- **Effort signal**: L
- **Recommended?** ◯

### Approach C — Data-rendered slots, AI wording
- **Thesis**: typed slots render requisites, date, members, кворум, vote, estimate amount and частка будинку from reference data and the справа; AI writes wording and non-blocking warnings with a source link and proposes amounts read from papers; the справа follows each expense line from estimate to act to bank statement until closed.
- **For whom**: the preparer now; the auditor and co-owners at the annual meeting.
- **Outcome metric**: time per protocol 1-3 h -> 20 min; protocols with a fully closed справа: not measured -> all by the annual meeting.
- **Key trade-off**: less "wow"; rare decision types need a new slot set first.
- **AI boundary**: AI writes wording and warnings and proposes amounts; no amount reaches print until the preparer confirms or corrects it; shares, per-building amounts, limit checks and vote results are computed from data.
- **Effort signal**: M
- **Recommended?** ●

## 8. Multi-perspective feedback

### Engineer
- Reading amounts from scans is the riskiest part; a wrong figure is the likeliest dispute trigger.
- Reference data makes checks deterministic but must be kept as of the protocol date.
- The chat-to-paper transfer of votes has no trail.

### Board chair
- Every number I sign must trace back to the estimate or the agreed shares.
- I will not sign over the 100 000 UAH limit or against a decision in force; stop it before printing.
- Papers kept together is the real time saver; amounts only after I have checked them.

### Resident / auditor
- I check estimate, act and bank statement per expense line; only C shows lines as closed or open.
- A number pointing to its source I can check; a retyped or guessed one I can only trust.
- Failed votes are not recorded, so "why was X not done?" has no answer.

### Synthesis matrix
|        | Engineer | Board chair | Resident / auditor |
|--------|:--------:|:-----------:|:------------------:|
| App. A | + smallest surface, but no checks | 0 fast, my typing errors hidden | 0 readable, no trail to sources |
| App. B | - free reading, huge error surface | - won't sign machine-guessed amounts | - machine-read amounts, cannot verify |
| App. C | + deterministic numbers, AI only advisory | + traceable numbers, papers together, signable | + every number links to source |

## 9. Trade-offs and edge cases

### Trade-offs per approach
| Approach | Pros | Cons |
|---|---|---|
| A | cheapest, fastest first print | no checks, errors hidden |
| B | fastest per protocol, free-form decisions | unverifiable amounts, highest effort |
| C | traceable numbers, limit and quorum checked, papers together | reference data entered once; new slots per decision type |

### Edge cases
- Missing source document: amount slot empty, printing blocked.
- Two documents disagree on a number (act vs approved estimate): warning with both sources; re-vote threshold open (§15).
- Quorum not reached or fewer than 3 votes "for": no "adopted" decision prints.
- Decision outside board competence (over 100 000 UAH, common space): blocking warning.
- Correction after signing: signed version frozen; a correction is a new protocol referencing it.
- One job split into several protocols under the limit: warning on the same estimate or contractor.
- Member with a personal interest: неучасть у голосуванні with reason, not counted.
- Chair absent: the deputy signs only with the service note in the справа.
- Shares changed by a new inter-building decision: the protocol uses those in force on its date.
- Amount misread from a scan (VAT): unconfirmed amounts are marked and block printing.

## 10. Risks
- **Top: a protocol records a meeting that was really a chat poll.** "Present 4 of 4, quorum reached" while the vote happened in the chat; if the charter does not allow absentee board voting, all such decisions can be challenged at once. Signal: auditor remark "no meeting on that date", a co-owner claim. Hits A, B, C; close via §18 before the spec.
- Limit splitting. Signal: co-owners asking "who approved 300 thousand?".
- Wrong figure signed. Signal: the management company bills a different sum.
- Stale reference data after re-election or new shares. Signal: dispute between buildings.
- Act exceeds the approved amount. Signal: auditor remark "spent above the protocol".
- Closing papers withheld by the management company. Signal: case files that never close.
- One person runs it all. Signal: incomplete archive after a board change.

## 11. RICE — Claude proposed
- **Reach (R)**: 260 protocols per year — 13 associations x about 20 (§3)
- **Impact (I)**: 2 — legal risk: board limit, quorum, annual-meeting report (§2, §8 board chair)
- **Confidence (C)**: 0.7 — every §16 value has a known source; the charter rule on absentee board voting is open
- **Effort (E)**: 6 person-weeks — approach C (§7)
- **RICE = 260 x 2 x 0.7 / 6 = 60.7**
- **State**: confirmed (2026-10-21)

## 12. Feasibility — Claude proposed
- [☑] **Tech**: a screen prototype of the case list and the protocol with slots, source tags and two warning tiers exists (§17); the server side, case-file storage and printable output are new.
- [☑] **Skills**: the team built the existing service skeleton; the idea author tried reading amounts from scans with AI by hand with good results.
- [☑] **Time**: people and time confirmed by the idea author.
- **State**: confirmed

## 13. Recommendation
**Selected: Approach C — Data-rendered slots, AI wording.** The status quo costs 10-90 hours a year per association and leaves decisions unrecorded until the annual meeting; no alternative builds a board protocol from an estimate with a case file (§6). C is the only approach the board chair accepts — "traceable numbers, papers together, signable" — and the only one with "+" in every column (§8). RICE 60.7 at 6 person-weeks (§11) on an existing screen prototype (§12). Of 17 values in §16, 14 come from data, computation or a person; AI writes 2 (wording, warnings) and proposes 1 (amount), which counts only after confirmation.

**Locked-in pointer**: every amount, share, date, name and vote is a typed slot with a visible source and a "confirmed" state; AI output never reaches print unconfirmed; the справа holds estimate, act and bank statement per expense line.

## 14. Parked & rejected approaches
| # | Approach | Status | Reason | Revisit trigger |
|---|---|:---:|---|---|
| A | Fixed-block protocol form | parked | no checks; chair and auditor score 0 | C slips past 6 person-weeks and a stop-gap is needed before the annual meeting |
| B | AI-drafted protocol copilot | rejected | "-" from all three lenses | AI-proposed amounts need almost no corrections over a season of confirmed protocols |

## 15. Open questions
- [ ] Does the charter allow absentee / electronic board voting, and how must the protocol name the meeting form? — owner: board chair, due: before write-prd
- [ ] Is the 100 000 UAH limit per protocol, per job or per estimate; which clause? — owner: board chair, due: before write-prd
- [ ] On a 4-member board, is "3 votes for" the quorum, the majority, or both? — owner: board chair, due: before write-prd
- [ ] At what overrun of the act over the approved amount is a new decision needed? — owner: board chair + accountant, due: before write-prd
- [ ] Must failed votes be recorded as protocols? — owner: board, due: before write-prd
- [ ] Will the management company hand over acts and statements per expense line? — owner: idea author, due: before closure criteria in the spec
- [ ] Confirm Reach: 13 associations at about 20 protocols each. — owner: idea author, due: before roadmap scoring
- [ ] AI may propose amounts read from papers (user decision 2026-10-21, wider than this skill's default boundary); the spec defines how a confirmed amount is marked. — owner: tech lead, due: write-prd
- New glossary terms in CONTEXT.md: акт виконаних робіт, кошторис робіт, управляюча компанія, частка будинку.

## 16. Source of truth
| Value in the document | Source | Written by | Who may change it |
|---|---|---|---|
| Association name and requisites | registry extract, charter | data | reference-data editor, with the new document |
| Protocol number and date | association's protocol register | data | nobody after signing |
| Board members and roles | board election protocol | data | reference-data editor, on a new election |
| Головуючий; deputy substitution | election protocol; service note in the справа | data | preparer selects; note required for the deputy |
| Attendance and неучасть with reason | board chat poll | human | preparer, before signing |
| Quorum reached | attendance vs charter quorum | computed | nobody |
| Vote count per item | board chat poll | human | preparer, before signing |
| Agenda item title | the справа | human | preparer |
| Estimate number, date, contractor | кошторис робіт in the справа | data | re-import of the document |
| Estimate amount | кошторис робіт in the справа | data (AI proposes from the scan, preparer confirms) | preparer, before signing |
| Building shares | inter-building decision in force on the date | data | only a new inter-building decision |
| Amount per building | estimate amount x share | computed | nobody |
| Amount in words | confirmed amount | computed | nobody |
| Джерело коштів | general-meeting contracts with the management company (maintenance; repair and development fund) | human, from a list in data | preparer |
| Board limit check | charter clause (§15) | computed | reference-data editor, on a charter change |
| Decision wording | slots above | AI | preparer edits, chair approves |
| Warnings (limit, competence, чинне рішення, act mismatch) | rules + linked source | AI for advisory text; computed for blocking checks | nobody; fix the source |

## 17. Repo reference
- **Bounded context**: new `governance`; the prototype already expects it (`apps/web/src/features/governance/api.ts` stands in for `@/api/gen/endpoints/governance/governance`). Only `system` exists on the server (`services/core/internal/system/`).
- **Adjacent modules**: `apps/web/src/features/governance/` — `types.ts` (`CaseStatus`, `DocumentKind` offer/invoice/contract/act, `SlotOrigin` system/ai/human, warning tiers at `types.ts:58`), `protocol-screen.tsx`, `slot-row.tsx`, `readiness.ts`, `validation.ts`; print via `window.print()` at `readiness-card.tsx:54`; routes `cases.index.tsx`, `cases.$caseId.protocol.tsx`; `design/protocol-brief.md`. Conventions: ADR 0002 (context layout, depguard), 0003 (one OpenAPI spec, tag per context), 0004 (goose, sqlc per context). No server-side upload, rendering or AI yet — new nodes for architecture ADRs.
- **Sentinel error seeds**: no domain sentinel convention yet (`fmt.Errorf` wrapping; problem+json in `services/core/internal/platform/httpapi/problem.go`). Seeds: `governance.quorum_missing`, `governance.outside_board_competence`, `governance.amount_mismatch`, `governance.amount_unconfirmed`, `governance.source_not_found`, `governance.reference_stale`, `governance.version_immutable`.

## 18. Legal and competence check
- **Who decides**: the board within the approved contracts and its limit; the general meeting above the limit and for common space (confirm against the charter).
- **Charter limits**: 100 000 UAH per protocol — source: charter and association documents, clause TBD; board quorum — TBD; absentee board voting — TBD (all §15).
- **Meeting decisions that apply**: maintenance contract with the management company; contract for the repair and development fund; the inter-building decision on shares; the board protocol electing the deputy.
- **What happens if this is wrong**: a decision over the limit or in a form the charter does not allow can be voided and re-voted at the general meeting; a wrong amount or share leads to a dispute with the management company or between buildings.

## Related
- CONTEXT.md (canonical glossary); idea-brief.md (other skill, kept for comparison); design/protocol-brief.md

## DoD self-check
- [x] 18 sections present
- [x] No tech terms outside §17
- [x] Every §16 row has a source and an owner; no AI-only row holds a number
- [x] Length ≤ 6 pages (~2600 words)
- [x] status: Confirmed, RICE and Feasibility confirmed
- [x] §13 cites §6, §8, §11, §12, §16
