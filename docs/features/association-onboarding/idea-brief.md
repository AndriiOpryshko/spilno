---
status: Confirmed
skill: ao-discovery
owner: "Andrii Opryshko"
reviewers: []
updated_at: "2026-11-05"
stage: "01"
depth: medium
ticket: ""
value_score:
  rice: 2
  state: confirmed
  confirmed_at: "2026-11-05"
feasibility_state: confirmed
---

<!-- Why: capture the idea in the language of the board before it is retold as a feature list. -->
<!-- Self-check: no tech terms outside §17; no number in an AI-written row of §16. -->

# Idea Brief — Підключення ОСББ до Спільно

## 1. Raw idea
«умовно флоу такий. будь хто може через гмейл зробити логі. на початку його просять там щось по осбб
ввести і він типу подає заявку щоб користуватися. в бекофісі його апрувят і він тоді може
користуватися. також може бути дізейбл через бекофіс. підключають до гугл драйву самі осбб»

## 2. Problem
Today Спільно serves one association, and it was set up by hand by the developer: a database record for
the association, a developer command for the first член правління, another for the Google Drive
connection. Every new association, and every new голова whom nobody on the board added, would need the
developer again. Nothing records who checked that the person really heads the association.

## 3. Users
- **Голова правління** (applicant): files the application, proves the office with an ЄДР extract signed
  by their КЕП, then becomes the first член правління and adds the rest of the board. Once per
  association, again after a change of голова.
- **Оператор** (the product owner, one person): reviews applications and transfer requests, approves or
  rejects with a reason, disables or re-enables an association.
- **Члени правління**: add themselves through the roster, see «доступ призупинено» if the association is
  disabled.
- **Ревізійна комісія, співвласники** (later readers): need to trust who runs the association in Спільно
  and trace it to the ЄДР extract the operator saw.

## 4. Why now
To stop creating associations by hand. The approved board-protocol tasks T08 and T12 would make setup a
developer command; this replaces both before they are built. Growing to many associations is not the
goal yet.

## 5. Out of scope
- Disabling a single person's account by the operator: the board manages its own roster (interview
  2026-11-05).
- Periodic re-checks of ЄДР and alerts on a change of head (Approach B, §14).
- Automatic filling of the association's data from ЄДР: no free per-code source of the current head was
  found (§6); the applicant types it and the operator checks it.
- Payment, tariffs, a trial period.
- Several operators and their roles.
- Co-owners' accounts and the ревізійна комісія's own access.

## 6. Status quo and alternatives

### Status quo
The developer runs a migration and two commands per association. Cost: developer time for every
association and every change of голова; no record of how the голова was verified; a new голова whom the
old board does not add has no way in.

### Alternatives
| # | Product · URL | Features | Value per feature (1-5) | Gap |
|---|---|---|---|---|
| 1 | Київ Цифровий, service for ОСББ · [life.kyiv.ua](https://life.kyiv.ua/news/yak-shvidko-zarestruvati-osbb-cherez-kiv-cifroviy), [suspilne.media](https://suspilne.media/640356-novi-funkcii-u-kiiv-cifrovij-zavivsa-servis-dla-osbb/) | голова sends an ЄДР extract naming the голова, signed with the голова's КЕП; the service registers the association; then co-owners join | proof of head 5, self-service 3 | Kyiv only; approval criteria and change of голова not described |
| 2 | Дах · [dah-online.com](https://dah-online.com/), [tariffs](https://dah-online.com/tariffs) | consultation request or application; building data by e-mail; the team registers the building and opens the board's cabinet | onboarding 2 | fully manual; verification of the голова not published |
| 3 | ЄДР data sources · [youcontrol](https://youcontrol.com.ua/catalog/company_details/45807697/), [opendatabot](https://opendatabot.ua/c/45807697), Мін'юст open data | name, address, керівник by code | head lookup 4 | full data behind registration or payment; free open data is weekly bulk without addresses |

Searched 2026-11-05: «ОСББ онлайн сервіс реєстрація ОСББ заявка голова ОСББ верифікація», «ДАХ сервіс ОСББ
сайт реєстрація кабінет голови», «Opendatabot API ЄДР компанія за кодом ЄДРПОУ керівник ціна».

## 7. Strategic approaches

### Approach A — Manual gate, single guard
- **Thesis**: the smallest loop: sign in, apply, the operator approves; the operator checks the register by
  hand; a new голова contacts the operator outside Спільно.
- **For whom**: the operator onboarding a few known associations.
- **Outcome metric** (baseline -> target): cross-association exposures 0 (one association) -> 0 with 2+.
- **Key trade-off**: no transfer path and no trail of the check.
- **AI boundary**: no AI; everything typed by the applicant, judged by the operator.
- **Effort signal**: S
- **Recommended?** ◯

### Approach B — Register-verified trust network
- **Thesis**: only boards the register confirms get in and stay confirmed; register changes raise alerts
  and open transfers; every decision is logged with the register snapshot.
- **For whom**: associations with contested boards; auditors.
- **Outcome metric**: 0 -> 0, two independent isolation guards.
- **Key trade-off**: largest build and a paid, lagging register source for a scale that is not the goal.
- **AI boundary**: no AI; data from the register; operator decides.
- **Effort signal**: L
- **Recommended?** ◯

### Approach C — Operator-decided onboarding with proof of office
- **Thesis**: every confirmed decision once and well: application with the КЕП-signed ЄДР extract,
  operator decides with a name-mismatch warning, transfer requests in the same queue and removing the
  previous голова, disable and re-enable an association.
- **For whom**: the голова of an ordinary ОСББ and the operator.
- **Outcome metric**: 0 -> 0 with 2+ associations, guarded by a release-blocking check that tries to read
  every kind of record across associations.
- **Key trade-off**: no ongoing register monitoring; a change of голова surfaces only when someone files a
  transfer request.
- **AI boundary**: no AI; the applicant types the association data; the signer and the head come from the
  signed extract; the operator decides.
- **Effort signal**: M
- **Recommended?** ●

## 8. Multi-perspective feedback

### Engineer
- A is smallest, but one missed association filter leaks data exactly when a second association arrives.
- C adds a pre-approval state for people off the roster, a signature check and a release-blocking
  cross-association test; low ongoing cost.
- B depends on a register source with no free per-code head and a lag of about a day; false alerts.
- All three reverse the rule that nobody off the roster gets a session.

### Board chair
- A: retypes data, no in-app path for a new голова.
- C: would refuse it without proof of office (anyone with the public name could apply) and without
  removing an ex-голова on transfer — both now in C.
- Needs the signed originals safe if the member who connected Drive leaves, and read access to the
  registry if the association is disabled — both now in C.

### Resident / auditor
- A leaves no trail of who verified the голова.
- C with the signed extract kept next to the decision gives a checkable chain from Спільно to ЄДР.
- An operator settling a disputed election without a document cannot be checked by anyone.

### Synthesis matrix
|        | Engineer | Board chair | Resident / auditor |
|--------|:--------:|:-----------:|:------------------:|
| App. A | + smallest, single fragile guard | - manual, no transfer path | - operator check leaves no trail |
| App. B | - largest surface, flaky register | + takeover alerts, logged decisions | + snapshot per decision |
| App. C | + one check, test-backed guard | 0 → + with signed extract | 0 → + extract kept with decision |

## 9. Trade-offs and edge cases

### Trade-offs per approach
| Approach | Pros | Cons |
|---|---|---|
| A | fastest, no outside dependency | impostor and ex-голова risks open, no trail |
| B | strongest trust, alerts | effort L, paid register, false alerts after elections |
| C | proof of office, transfer, disable, trail | голова without КЕП cannot apply; no monitoring |

### Edge cases
- **Applicant has no КЕП**: the application cannot be filed; the form explains how to get one.
- **ЄДР lags after an election**: the signed extract still names the old head → advisory warning to the
  operator, who may ask for the election protocol; never blocks by itself.
- **Code already connected or pending**: refused with «зверніться до голови вашого ОСББ»; the head's name is
  not revealed; a голова with a signed extract naming them may file a transfer request instead.
- **Transfer request**: on approval the new голова joins the roster and the previous голова's term ends;
  both are recorded with the extract.
- **Two factions**: two transfer requests on one code → the operator decides only on the signed extract
  and the reason is recorded; nothing is decided without a document.
- **Association disabled**: members can read and download the signed protocols of the registry only;
  nothing changes; re-enabling restores everything.
- **Drive owned by a personal account**: Drive is connected only with the association's own account; on a
  transfer the operator sees whose account holds the files.
- **Operator unavailable**: applications wait; the applicant sees «очікує розгляду» and the date.
- **Rejection**: reason required, visible to the applicant, who may fix and resubmit.

## 10. Risks
- **Top: impostor taking over an association.** The head's name is public, and a matching Gmail name passes
  a code-only check. Mitigated by the КЕП-signed extract. Signal: a real голова writes that «someone else
  is голова of our ОСББ»; two approved claims for one code.
- **Cross-association leak** when the second association joins (isolation was deferred until this moment).
  Signal: a report «I see another building's protocol»; the release-blocking check fails.
- **Operator as judge of disputed elections.** Signal: repeated transfer requests on one code.
- **Single operator unavailable.** Signal: applications older than a week.

## 11. RICE — Claude proposed
- **Reach (R)**: 6 events per year (applications and transfer requests) — §3
- **Impact (I)**: 2 — the cost of a mistake is a stranger reading an association's protocols (§2, §8)
- **Confidence (C)**: 0.5 — the number of associations waiting is not confirmed (§15)
- **Effort (E)**: 3 person-weeks — Approach C (§7)
- **RICE = R × I × C / E = 6 × 2 × 0.5 / 3 = 2**
- **State**: confirmed (2026-11-05)

## 12. Feasibility — Claude proposed
- [☑] **Tech**: sign-in, sessions and the association record exist (§17); the change is to let a person
  without an association sign in and to add applications, transfers and disabling. Signature checking of
  the extract is new.
- [☑] **Skills**: the product owner builds Спільно with agents; the same stack.
- [☑] **Time**: Andrii Opryshko, after board-protocol Stage 1 (T34–T44), Dec 2026 – Jan 2027.
- **State**: confirmed

## 13. Recommendation
**Selected: Approach C** — The status quo needs the developer for every association and leaves no trail
(§6), and the only comparable public flow, Київ Цифровий, proves the голова with a КЕП-signed ЄДР extract
(§6). The board-chair column of §8 rejects a code-only check because anyone with the public name could get
in; C with the signed extract, transfer that removes the previous голова, read-only registry for a
disabled association and Drive on the association's own account answers every «would refuse» point.
RICE is 2 with confidence 0.5 (§11), feasible by the owner after Stage 1 (§12). No value in the
association's record is written by AI; the signer and head come from the signed extract, the rest is
typed by the applicant and checked by the operator (§16).

**Locked-in pointer**: the PRD must amend board-protocol PRD §3 and AC-19, and the SAD needs an ADR on
sessions for people off the roster and on isolating associations (superseding the deferral in ADR-0015);
T08 and T12 are replaced.

## 14. Parked & rejected approaches
| # | Approach | Status | Reason | Revisit trigger |
|---|---|:---:|---|---|
| 1 | A — Manual gate | rejected | no proof of office, no transfer, no trail | — |
| 2 | B — Register monitoring | parked | effort L, paid register source, false alerts | more than ~20 associations or a takeover incident |
| 3 | Fill association data from ЄДР by code | parked | no free per-code source of the head | a free or contracted register source |
| 4 | Operator disables single accounts | rejected | the board manages its own roster (2026-11-05) | abuse by a member the board cannot remove |

## 15. Open questions
- [ ] How many associations actually wait to join (Reach)? — owner: Andrii Opryshko, due: 2026-12-01
- [ ] How the signature on the ЄДР extract is checked (by Спільно or by the operator with a public
  service) — owner: Andrii Opryshko, due: before the PRD
- [ ] What the applicant sees while the operator is away more than a week — owner: Andrii Opryshko, due:
  PRD
- [ ] Where the back office lives (inside Спільно behind an operator role or a separate surface) — owner:
  Andrii Opryshko, due: SAD
- Glossary terms added: `docs/features/association-onboarding/CONTEXT.md`.

## 16. Source of truth
| Value in the document | Source | Written by | Who may change it |
|---|---|---|---|
| ЄДРПОУ code | applicant | person | applicant until approval; then nobody |
| Association name, address | applicant, checked against the extract | person | the board in the settings |
| Head named in ЄДР | the КЕП-signed ЄДР extract | data | nobody; a new extract with a transfer request |
| Signer of the extract | the КЕП on the extract | data | nobody |
| Applicant's name, position, term | applicant | person | the board in the roster after approval |
| Name-mismatch warning | applicant's name vs head in the extract | data (advisory) | — |
| Approval / rejection and its reason | operator | person | operator |
| Disabled state and its reason | operator | person | operator |
| Google Drive account and folder | the board's connection | person | any член правління |

No value is written by AI. The only check is the name-mismatch warning, and it is advisory: it never
blocks approval by itself.

## 17. Repo reference
- **Bounded context**: `identity` (existing: associations, accounts, sessions, Google connection) +
  `governance` roster.
- **Adjacent modules**: `services/core/internal/identity/` (skeleton handler, `adapters/postgres`);
  `services/core/migrations/00004_create_associations.sql` (one seeded row);
  `00009_create_accounts.sql` (`accounts.association_id NOT NULL` — an applicant has no association
  yet, so this changes); `00011_create_google_connections.sql`; tasks T08 (`bootstrap-member`), T12
  (`cmd/google-connect`); Figma `D06IijJnLn8WEgU85fnD84` page 07 (onboarding `61:6918…`, Drive `63:7130…`,
  back office section `51:6511`); `docs/design/stage1-screens-spec.md` BR-21, BR-26…BR-29.
- **Sentinel error seeds**: `identity.application_invalid`, `identity.signature_invalid`,
  `identity.association_already_connected`, `identity.application_not_pending`,
  `identity.association_disabled`, `governance.drive_not_connected`.

## 18. Legal and competence check
- **Who decides**: the operator decides access to Спільно; who heads the association is decided by the
  association (general meeting, board) and recorded in ЄДР — the operator only checks the document.
- **Charter limits**: none for this feature — source: the association's charter does not govern access to
  a third-party service.
- **Meeting decisions that apply**: the election of the голова (general meeting or board, per charter),
  reflected in ЄДР.
- **What happens if this is wrong**: a stranger reads or edits protocols of an association; a legitimate
  голова is locked out; the operator is drawn into a disputed election. Personal data of board members is
  exposed to the wrong person.

## Related
- `docs/features/association-onboarding/CONTEXT.md`
- `docs/features/board-protocol/PRD.md` §3, AC-19; ADR-0003, ADR-0005, ADR-0015; tasks T08, T12
- `docs/design/stage1-screens-spec.md` §14

## DoD self-check
- [x] 18 sections present
- [x] No tech terms outside §17
- [x] Every §16 row has a source and an owner; no AI row holds a number
- [x] Length ≤ 6 pages (~2600 words)
- [x] status: Confirmed, RICE and Feasibility confirmed
- [x] §13 cites §6, §8, §11, §12, §16
