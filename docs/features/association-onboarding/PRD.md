---
status: Approved
skill: ao-prd
owner: "Andrii Opryshko"
reviewers: ["Andrii Opryshko (голова правління, document owner)", "Tech Lead"]
approved_by: "Andrii Opryshko"
approved_at: "2026-11-05"
updated_at: "2026-11-05"
feature_size: "M"
stage: "03"
depth: "easy"
ticket: ""
---

# PRD — association-onboarding

> **Inputs:** [idea brief](./idea-brief.md) (ao-discovery, Confirmed 2026-11-05) · [CONTEXT](./CONTEXT.md) · shared terms in [board-protocol CONTEXT](../board-protocol/CONTEXT.md)
> **Reference context:** `services/core/internal/identity` — module, stub handler (sign-in, session, sign-out), the association, account, session and Google-connection records (`migrations/00004`, `00009`–`00011`; an account carries exactly one association today); `services/core/internal/governance` — roster entry with a term and no position (`migrations/00012`, `domain/board_member.go`: `Term.CoversDate` counts the end date as covered, `BoardMemberEntry.Validate`), the `governance.*` error registry (`domain/errors.go`).
> **Other channels:** board-protocol `PRD.md` §3, AC-19; board-protocol ADR-0003, ADR-0005, ADR-0014, ADR-0015; board-protocol tasks T08 (`bootstrap-member`), T12 (`cmd/google-connect`); `docs/design/stage1-screens-spec.md` BR-21, §14.

## 1. Context

Спільно serves one association today, and that association was set up by hand by the developer: a record for the association, a command that adds the first член правління, another that connects the association's document storage. A second association, or a new голова правління whom the outgoing board never added to the roster, needs the developer again. Nothing records who checked that the person really heads the association: no document shows the register's керівник, nobody's signature proves the applicant is that person, and the ревізійна комісія or a співвласник cannot trace who runs the association in Спільно back to the state register (brief §2, §3).

Why now: the approved board-protocol tasks T08 and T12 would make this setup a developer command; this feature replaces both before they are built, so the first association connected after Stage 1 needs no developer at all. Growing to many associations is not the goal (brief §4).

Committed approach (brief §13, Approach C): anyone signs in with their own sign-in account and, as голова правління, files a заявка на підключення with the ЄДРПОУ, name, address, their position and term, and a витяг з ЄДР signed with their КЕП. Спільно checks the signature itself; the оператор approves or rejects with a reason. A newly elected голова files a запит на передачу into the same queue; its approval adds them to the roster and ends the previous голова's term. The оператор can disable an association (its members keep read and download access to the signed protocols only) and enable it again. The board connects the association's document storage itself, with the association's own account. AI boundary: no AI anywhere in this feature. The applicant types the association data; the signer comes from the КЕП and the керівник from the signed extract; the only checks of the extract's content are the advisory розбіжність імен and the extract's age; the оператор decides.

Traceability and deviations:
- Read: the reference contexts in the Inputs block; board-protocol ADR-0003 (sign-in refuses a person whose address matches no roster entry — this feature reverses that rule for the application flow, ADR needed), ADR-0005 (storage on the association's account), ADR-0015 (association isolation by explicit filter, the cross-association guard deferred until a second association — this feature is that moment, ADR needed).
- Amends board-protocol PRD §3 (the non-goal «onboarding other associations») and AC-19 (a signed-in person off the roster now sees their own заявка, still nothing of any association). Replaces tasks T08 and T12.
- Brief §15 «who checks the КЕП signature» closed 2026-11-05: Спільно checks it at filing; an invalid signature blocks filing.
- Brief §15 «what the applicant sees while the оператор is away more than a week» closed: the filing date and the оператор's contact address (AC-05); the risk «single operator unavailable» (brief §10) is measured by the decision-time KPI (§7).
- Deviation: a витяг з ЄДР for a different ЄДРПОУ than the one typed blocks filing (brief §16 names only the advisory name check) — reason: an extract of another association is never a proof of office, so no operator judgement is needed.
- Deviation: an extract older than 30 days raises an advisory warning to the оператор (AC-11b) — reason: brief §9 «ЄДР lags after an election» works the other way too; a stale extract may name a голова already replaced. It never blocks and needs no named document.
- Deviation: approving over an open розбіжність імен requires the оператор to name the document relied on (brief §9 «nothing is decided without a document»).
- Deviation: a second заявка на підключення for a code that already has one pending is accepted and shown to the оператор as competing (brief §9 refused it) — reason: otherwise an impostor who files first locks the real голова out, the top risk of brief §10.
- Brief §9 «the document storage is connected only with the association's own account» is enforced as far as Спільно can see: an account whose address is a член правління's sign-in address is refused (AC-28b); a personal account with another address cannot be told apart (§6.1).
- Product owner 2026-11-05: a person may be голова or член правління of several associations at once («можеш бути головою хоч всіх ОСББ Києва, якщо за тебе проголосують»); membership comes from each association's roster, not from the account.

## 2. Goals

- A голова правління connects their association, or takes it over after an election, without the developer, and the оператор decides on a signed proof of office.
- Every connected association and every change of голова carries a checkable chain: the КЕП-signed витяг з ЄДР, the signer, the оператор's decision and its reason.
- Data of one association is never visible to a person who is not on its roster, guarded by a check that blocks a release.

## 3. Non-goals

- The оператор disabling one person's account — the board manages its own roster (brief §5, interview 2026-11-05).
- Periodic re-checks of ЄДР and alerts on a change of керівник — parked Approach B, effort L and a paid, lagging register source (brief §14).
- Filling the association's data from ЄДР by code — no free per-code source of the current керівник (brief §6, §14).
- Payment, tariffs, a trial period — not a goal at this scale (brief §5).
- Several operators and operator roles — one product owner runs Спільно (brief §5).
- Co-owners' accounts and the ревізійна комісія's own access — a separate feature (brief §5, board-protocol §3).
- Notifications about a decision by mail or messenger — Спільно has no delivery channel (board-protocol ADR-0003); the applicant sees the decision after signing in.
- Filing without a КЕП, or with only an election protocol — proof of office is the КЕП-signed extract; the form explains how to get a КЕП (brief §9).

## 4. User stories

### US-01: File an application to connect the association
**As a** голова правління
**I want** to sign in and file a заявка на підключення with the association's data and the КЕП-signed витяг з ЄДР
**So that** my association starts using Спільно without the developer

### US-02: Follow, fix and refile the application
**As a** голова правління
**I want** to see whether my заявка на підключення is очікує, схвалена or відхилена, read the reason of a rejection, fix it and file again
**So that** I know what happens and can correct a mistake without starting over

### US-03: Decide an application
**As an** оператор
**I want** to review a заявка на підключення with the signer, the керівник named in the extract and any розбіжність імен, then approve it or reject it with a reason
**So that** only a person proven by a signed document becomes the first член правління of a new association

### US-04: Request a transfer after an election
**As a** голова правління
**I want** to file a запит на передачу for an association already connected, with a витяг з ЄДР naming me and signed with my КЕП
**So that** I take over the association in Спільно even when the previous board does not add me

### US-05: Decide a transfer request
**As an** оператор
**I want** to approve or reject a запит на передачу with a reason, seeing competing requests for the same association side by side
**So that** the new голова joins the roster and the previous голова's access ends, on a document and with a recorded reason

### US-06: Disable and enable an association
**As an** оператор
**I want** to disable a підключене ОСББ with a reason and enable it again
**So that** a disputed or abandoned association stops changing records while its signed protocols stay readable

### US-07: Work in a disabled association
**As a** член правління
**I want** to see «доступ призупинено» with the reason and still read and download the signed protocols of the реєстр протоколів
**So that** the association keeps its record even while it is вимкнене ОСББ

### US-08: Connect the association's document storage
**As a** член правління
**I want** to connect the association's own storage account and see which account holds the files
**So that** documents stay with the association, not with a member who may leave

### US-09: See how the голова was confirmed
**As a** член правління
**I want** to see on the association's card which витяг з ЄДР, signed by whom, the оператор relied on, and when the decision was made
**So that** I can show the ревізійна комісія who runs the association in Спільно and why

## 5. Acceptance criteria

### AC-01 (US-01) — happy path
**Given** a person signed in with their sign-in account, whether or not they are a член правління of another association, with a витяг з ЄДР for ЄДРПОУ 12345678 signed with their valid КЕП
**When** they file a заявка на підключення with that ЄДРПОУ, the association's name and address, their position «голова правління» and their term start
**Then** the system records the заявка as очікує with the filing date, keeps the signed extract with it, shows the signer and the керівник read from the extract, and shows «Заявку подано, очікує розгляду оператором»

### AC-02 (US-01) — error
**Given** a голова правління filling a заявка на підключення
**When** they file it with the ЄДРПОУ not of 8 digits, the name or address empty, the term start missing, or no extract attached
**Then** the system refuses and names every missing or wrong field at once, and nothing is recorded

### AC-03 (US-01) — domain invariant
**Given** a голова правління attaching a витяг з ЄДР whose signature is broken, made with a certificate that is not qualified, or revoked or expired on the signing date
**When** they file the заявка на підключення
**Then** the system refuses, says «Підпис КЕП на витягу недійсний» with the reason, and nothing is recorded

### AC-03b (US-01) — domain invariant (edge case: extract of another association)
**Given** a голова правління who typed ЄДРПОУ 12345678 and attached a validly signed витяг з ЄДР for ЄДРПОУ 87654321
**When** they file the заявка на підключення
**Then** the system refuses with «Витяг стосується іншого ОСББ — перевірте код ЄДРПОУ або витяг», and nothing is recorded

### AC-04 (US-01) — domain invariant (code already connected)
**Given** an association with ЄДРПОУ 12345678 that is підключене ОСББ
**When** a person files a заявка на підключення for ЄДРПОУ 12345678
**Then** the system refuses with «Це ОСББ вже підключене — зверніться до голови вашого ОСББ або подайте запит на передачу», does not show who heads it, and offers the запит на передачу form

### AC-04b (US-01) — domain invariant (edge case: duplicate by the same person)
**Given** a person with a заявка на підключення for ЄДРПОУ 12345678 that очікує
**When** they try to file a second заявка for the same ЄДРПОУ
**Then** the system refuses with «У вас уже є заявка на це ОСББ» and shows the open one

### AC-04c (US-03) — domain invariant (edge case: competing applications)
**Given** a заявка на підключення for ЄДРПОУ 12345678 that очікує, filed by one person
**When** another person files a заявка for the same ЄДРПОУ with their own signed extract
**Then** the system records it as очікує without telling the second applicant that another заявка exists; the оператор sees both side by side marked «конкуруюча заявка»; once one is approved, approving the other is refused as for a connected code (AC-04), and the оператор rejects it with a reason

### AC-04d (US-01) — cross-context (edge case: member of several associations)
**Given** a current член правління of ОСББ «Сонячний» whose заявка на підключення for ОСББ «Липки» was approved
**When** they sign in
**Then** they choose which association to work in, and the records, roster and settings of each association are shown only within that association and only to a person on its own roster

### AC-05 (US-02) — happy path
**Given** a голова правління whose заявка на підключення очікує since 2026-11-10
**When** they sign in
**Then** they see their заявка with the status «очікує розгляду з 10.11.2026» and the оператор's contact address, and nothing of the association it names

### AC-06 (US-02) — happy path (refile after rejection)
**Given** a голова правління whose заявка на підключення is відхилена with the reason «у витягу інший керівник, документа про обрання не надано»
**When** they attach a new витяг, correct the fields and file it again
**Then** the system records the same заявка as очікує again with the new filing date, keeps the rejected version and its reason in the заявка's history, and shows the new status

### AC-07 (US-02) — authorization
**Given** a person signed in who is not on the roster of ОСББ «Липки» (an applicant whose заявка очікує, a member of another association, or nobody)
**When** they try to open the реєстр протоколів, a справа or the roster of ОСББ «Липки»
**Then** the system denies access the same way whether the record exists or not

### AC-08 (US-03) — happy path
**Given** an оператор reviewing a заявка на підключення that очікує, with a valid signature and no розбіжність імен
**When** the оператор approves it
**Then** the system creates the підключене ОСББ with the ЄДРПОУ, name and address from the заявка, records the decision with its date and the оператор, and the applicant becomes its first член правління with the position голова правління and the term from the заявка

### AC-09 (US-03) — cross-context
**Given** a заявка на підключення the оператор has just approved
**When** the applicant signs in
**Then** they appear in the association's roster as a current член правління from the term start in the заявка, can open the реєстр протоколів and add the other члени правління, and see that the document storage is not yet connected

### AC-10 (US-03) — traceability
**Given** an оператор opening a заявка на підключення
**When** the review page shows the association's data and the applicant
**Then** every value names its source: ЄДРПОУ, name, address, position and term «введено заявником»; the керівник «з витягу з ЄДР від <дата витягу>»; the signer and the signing date «з КЕП»; and a value the system could not read from the extract is marked «не прочитано — перевірте вручну»

### AC-11 (US-03) — traceability (advisory розбіжність імен)
**Given** a заявка на підключення or a запит на передачу where the applicant's typed name, the signer of the КЕП and the керівник in the extract are not all the same person
**When** the оператор opens it
**Then** the system shows a розбіжність імен warning naming which values differ, and does not block approval or rejection by itself

### AC-11b (US-03) — traceability (advisory: extract age)
**Given** a заявка на підключення or a запит на передачу whose витяг з ЄДР is older than 30 days on the filing date
**When** the оператор opens it
**Then** the system shows «Витяг сформовано <дата> — старший за 30 днів», and approval needs nothing more than without the warning

### AC-12 (US-03) — legal competence
**Given** a заявка на підключення or a запит на передачу with an open розбіжність імен (for example, the extract still names the previous голова after an election)
**When** the оператор approves it without naming the document relied on
**Then** the system refuses and says «Схвалення з розбіжністю потребує документа: вкажіть, на чому ґрунтується рішення» — the оператор checks documents and does not decide who heads the association, which the general meeting or the board does per the charter and ЄДР records

### AC-13 (US-03) — error
**Given** an оператор reviewing a заявка на підключення or a запит на передачу that очікує
**When** the оператор rejects it with an empty reason
**Then** the system refuses with «Вкажіть причину — її побачить заявник», and the заявка stays очікує

### AC-14 (US-03) — happy path (reject)
**Given** an оператор reviewing a заявка на підключення that очікує
**When** the оператор rejects it with the reason «у витягу інший керівник, документа про обрання не надано»
**Then** the system records the заявка as відхилена with the reason, the date and the оператор, no association is created, and the applicant sees the reason after signing in

### AC-15 (US-03) — domain invariant (refused transitions)
**Given** a заявка на підключення or a запит на передачу that is схвалена, or відхилена and not refiled
**When** the оператор tries to approve or reject it again
**Then** the system refuses with «Рішення за цією заявкою вже ухвалено» and the recorded decision stays unchanged

### AC-16 (US-03) — authorization
**Given** a person signed in who is not the оператор (a член правління, a голова with a pending заявка, anyone)
**When** they try to open the list of заявки, approve, reject, disable or enable an association
**Then** the system denies access and shows nothing of the operator's queue

### AC-16b (US-03) — authorization (operator sees no association records)
**Given** the оператор signed in
**When** they try to open a протокол правління, a справа or the roster of any association
**Then** the system denies access: the оператор decides access to Спільно and sees заявки, запити and their extracts only, never an association's records

### AC-17 (US-04) — happy path
**Given** a підключене ОСББ with ЄДРПОУ 12345678 and a newly elected голова правління who is not on its roster, with a витяг з ЄДР naming them and signed with their valid КЕП
**When** they file a запит на передачу for ЄДРПОУ 12345678 with their term start
**Then** the system records the запит as очікує, keeps the signed extract, and shows «Запит на передачу подано, очікує розгляду оператором» without revealing who is on the current roster

### AC-18 (US-04) — domain invariant (not connected)
**Given** no підключене ОСББ with ЄДРПОУ 12345678
**When** a голова правління files a запит на передачу for it
**Then** the system refuses with «Це ОСББ ще не підключене — подайте заявку на підключення» and offers the заявка form

### AC-18b (US-04) — domain invariant (edge case: signature on a transfer)
**Given** a голова правління filing a запит на передачу with an extract whose КЕП signature is invalid, or that is for another ЄДРПОУ
**When** they file it
**Then** the system refuses the same way as for a заявка на підключення (AC-03, AC-03b), and nothing is recorded

### AC-19 (US-05) — happy path + cross-context
**Given** a запит на передачу that очікує for a підключене ОСББ whose current голова was recorded by an earlier approval, with the new голова's term start 2026-11-20 typed in the запит
**When** the оператор approves it on 2026-11-25
**Then** the new голова joins the roster as a current член правління with the position голова правління from 2026-11-20; the previous голова's term ends on 2026-11-19, the day before the new term start (never later than the day before approval); the previous голова's access ends at the moment of approval; both changes are recorded with the extract and the decision

### AC-19b (US-05) — happy path (edge case: two factions)
**Given** two запити на передачу that очікує for the same association
**When** the оператор opens either one
**Then** the system shows both side by side marked «конкуруючий запит»; approving one does not decide the other, which stays очікує until the оператор decides it with a reason

### AC-19c (US-05) — cross-context (edge case: association set up by hand)
**Given** the association set up by hand before onboarding, with no ЄДРПОУ and no голова recorded by an approval, and its голова who is a current член правління
**When** they file a запит на передачу from the association's card with its ЄДРПОУ and a витяг з ЄДР naming them, and the оператор approves it
**Then** the запит is attached to that association, its ЄДРПОУ is recorded from the запит, their roster entry gets the position голова правління, no other term ends, and the card shows the extract instead of «підключено вручну, без витягу»

### AC-20 (US-05) — traceability
**Given** an approved запит на передачу
**When** a член правління opens the roster
**Then** the entries of the new and the previous голова show «призначено / повноваження завершено за запитом на передачу від <дата>, схвалено оператором <дата>» with a link to the extract

### AC-21 (US-06) — happy path (disable)
**Given** a підключене ОСББ
**When** the оператор disables it with the reason «спір щодо складу правління»
**Then** the association becomes вимкнене ОСББ, the reason, the date and the оператор are recorded, and on their next action its members see «Доступ призупинено: спір щодо складу правління»

### AC-22 (US-06) — error
**Given** a підключене ОСББ
**When** the оператор disables it with an empty reason
**Then** the system refuses with «Вкажіть причину — її побачать члени правління», and the association stays підключене

### AC-23 (US-06) — happy path (enable)
**Given** a вимкнене ОСББ
**When** the оператор enables it
**Then** the association becomes підключене again, every record is as it was before disabling, and the enabling is recorded with the date and the оператор

### AC-24 (US-06) — domain invariant (refused transitions)
**Given** a вимкнене ОСББ, or a підключене ОСББ
**When** the оператор tries to disable the already вимкнене one, or to enable the already підключене one
**Then** the system refuses with «ОСББ уже в цьому стані» and records nothing

### AC-24b (US-06) — cross-context (edge case: transfer for a disabled association)
**Given** a вимкнене ОСББ
**When** a newly elected голова files a запит на передачу and the оператор approves it
**Then** the roster changes as in AC-19 and the association stays вимкнене until the оператор enables it

### AC-25 (US-07) — authorization
**Given** a member of a вимкнене ОСББ
**When** they try to edit a draft, upload a document, change the roster or the settings, or connect the document storage
**Then** the system refuses with «Доступ призупинено» and the reason, and nothing changes

### AC-26 (US-07) — happy path
**Given** a member of a вимкнене ОСББ
**When** they open the реєстр протоколів and download a signed protocol
**Then** the system shows the registry and serves the signed copy as before disabling

### AC-27 (US-08) — happy path
**Given** a підключене ОСББ without a connected document storage
**When** a член правління connects the association's own storage account
**Then** the system records the connection with the account's address and the member who connected it, and shows «Сховище документів: <адреса акаунта>» to every член правління

### AC-28 (US-08) — cross-context
**Given** a підключене ОСББ without a connected document storage
**When** a член правління tries to attach a document to a справа
**Then** the system refuses with «Спершу підключіть сховище документів ОСББ» and says that any член правління can connect it

### AC-28b (US-08) — domain invariant (edge case: personal account)
**Given** a член правління connecting a storage account whose address is the sign-in address of a член правління of this association
**When** they confirm the connection
**Then** the system refuses with «Це особистий акаунт члена правління — підключіть акаунт ОСББ», and nothing is connected; on any later запит на передачу the оператор sees which account holds the files

### AC-29 (US-09) — traceability
**Given** a підключене ОСББ
**When** a член правління opens the association's card
**Then** the card shows the current голова, the витяг з ЄДР they were confirmed by with its date and signer, the оператор's decision date, and every later запит на передачу and disabling with its reason; an association connected before onboarding existed shows «підключено вручну, без витягу» until AC-19c

### AC-30 (US-02) — legal competence (term of the applicant)
**Given** a голова правління whose term in the заявка or the запит ended before the filing date
**When** they file the заявка на підключення or the запит на передачу
**Then** the system refuses with «Строк повноважень голови вже завершився — подайте витяг після переобрання», because a person whose term has ended does not head the association

## 5.1 Error catalogue

| Code | Kind | Message to the actor | AC |
|---|---|---|---|
| `identity.application_invalid` | validation | Перевірте поля заявки: <поля> | AC-02 |
| `identity.signature_invalid` | invariant | Підпис КЕП на витягу недійсний: <причина> | AC-03, AC-18b |
| `identity.extract_code_mismatch` | invariant | Витяг стосується іншого ОСББ — перевірте код ЄДРПОУ або витяг | AC-03b, AC-18b |
| `identity.association_already_connected` | conflict | Це ОСББ вже підключене — зверніться до голови вашого ОСББ або подайте запит на передачу | AC-04, AC-04c |
| `identity.application_already_open` | conflict | У вас уже є заявка на це ОСББ | AC-04b |
| `identity.not_current_member` | authorization | Немає доступу (existing, board-protocol ADR-0003) | AC-07, AC-16b |
| `identity.mismatch_needs_document` | invariant | Схвалення з розбіжністю потребує документа: вкажіть, на чому ґрунтується рішення | AC-12 |
| `identity.reason_required` | validation | Вкажіть причину — її побачить заявник (AC-13) / її побачать члени правління (AC-22) | AC-13, AC-22 |
| `identity.application_not_pending` | conflict | Рішення за цією заявкою вже ухвалено | AC-15 |
| `identity.not_operator` | authorization | Немає доступу | AC-16 |
| `identity.association_not_connected` | invariant | Це ОСББ ще не підключене — подайте заявку на підключення | AC-18 |
| `identity.association_state_unchanged` | conflict | ОСББ уже в цьому стані | AC-24 |
| `identity.association_disabled` | authorization | Доступ призупинено: <причина> | AC-21, AC-25 |
| `identity.term_ended` | invariant | Строк повноважень голови вже завершився — подайте витяг після переобрання | AC-30 |
| `identity.storage_account_personal` | invariant | Це особистий акаунт члена правління — підключіть акаунт ОСББ | AC-28b |
| `governance.drive_not_connected` | invariant | Спершу підключіть сховище документів ОСББ (planned — not yet in `governance` ErrorRegistry) | AC-28 |

Objects of another association, and another person's заявка, answer as not existing (same body as `identity.not_current_member`), so their existence is not revealed. All `identity.*` codes except `identity.not_current_member` are planned in the existing `identity` context.

## 6. Non-functional requirements

| Aspect | Target | Measurement |
|---|---|---|
| Latency p95, read or list (application page, operator queue, association card) | ≤ 200 ms (default) | request-duration histogram per route |
| Latency p95, write without file (decide, disable, enable) | ≤ 500 ms (default) | same histogram |
| Latency p95, filing with the extract and the signature check | ≤ 5 s for a 5 MB file | same histogram, filing routes |
| Throughput | ≥ 50 req/s per instance on the read path (default) | load smoke test in CI |
| Availability | 99.5 % monthly (default) | uptime check on the health endpoint |
| Durability | RPO ≤ 24 h, RTO ≤ 4 h (default) | daily backup + quarterly restore drill |
| File size | extract ≤ 5 MB, signed PDF or a PDF with its detached signature | upload validation |
| Retention | заявки, запити на передачу, extracts and decisions kept ≥ 5 years, never hard-deleted | storage policy |
| Isolation | 0 records of association B readable by a person not on B's roster, for every kind of record, in a release-blocking check run against two seeded associations and one person on both rosters | CI check, blocks release on any hit |
| Effect of disabling and of a transfer | 0 requests of an affected member served with the old rights after the decision is recorded | integration test |
| Integrity | a decided заявка or запит changes only through a new filing or a new decision | integrity check, 0 mismatches |

## 6.1 Security / privacy

- **Data classification:** Confidential — the feature decides who reads an association's protocols and keeps signed proofs of office.
- **Personal data touched:** the applicant's full name, sign-in address, position and term; the signer's name and the certificate data inside the КЕП (it contains the РНОКПП — shown to nobody, kept only inside the signed file); the керівник's name from the extract; the storage account address. Medium sensitivity; the РНОКПП is high.
- **Authorization impact:** a new role оператор with access to the application queue only; a signed-in person sees an association's records only while on its roster, and may be on several rosters; a member of a вимкнене ОСББ is limited to reading the registry; a transfer ends the previous голова's access at the moment of approval.
- **Abuse cases:**
  - Impostor with the public name of the голова and a matching account name: no signed extract, no filing (AC-03); a stranger's КЕП on a real extract raises розбіжність імен, and approving over it requires a document (AC-11, AC-12); filing first does not lock the real голова out (AC-04c).
  - Forged or edited extract: the signature check fails on any change of the file (AC-03).
  - Probing ЄДРПОУ codes to learn who heads which association: no refusal names the голова or another applicant, and a pending заявка is not revealed (AC-04, AC-04c, AC-17); only «already connected» is revealed (brief §9).
  - Cross-association read, including by a person on two rosters: denied with the same answer as a missing record, guarded by the release-blocking isolation check (AC-04d, AC-07, §6 Isolation).
  - Ex-голова keeps the document storage under a personal account: refused when the address is a member's sign-in address (AC-28b); a personal account with another address is not detectable — the оператор sees the account on every transfer.
  - Takeover of the оператор's account: the оператор sees no association's records (AC-16b); every decision records who made it and when.
- **Security review:** Required — a new sign-in path for people off the roster, a new role, signature checking, membership in several associations and the first real second association.

## 7. Metrics / KPIs

- **Developer actions to connect an association** — baseline: 3 (a migration and two commands) per association, target: 0 for every association connected after release, measured over the first 6 months.
- **Time from filing to the оператор's decision** — baseline: none (no flow), target: ≥ 90 % of заявки and запити decided within 5 working days, over the first 6 months.
- **Cross-association exposures** — baseline: 0 with one association, target: 0 with 2+ associations, continuously (release-blocking check + reports from members).
- **Connected associations backed by a valid signed extract** — baseline: 0 %, target: 100 % of associations connected after release; the hand-made one gets its extract through AC-19c within 3 months of release.

## 8. Open questions

- [ ] How many associations actually wait to join (RICE Reach)? Default now: 6 filings a year. — owner: Andrii Opryshko, due: 2026-12-01
- [ ] How Спільно checks the КЕП signature (own library for the Ukrainian algorithms or a public verification service) and which signature formats it accepts? Default now: Спільно checks itself at filing; formats per SAD. — owner: Andrii Opryshko, due: SAD (ao-architecture)
- [ ] Where the operator's back office lives (inside Спільно behind the оператор role or a separate surface)? Default now: inside Спільно, a separate section shown only to the оператор. — owner: Andrii Opryshko, due: SAD
- [ ] How long rejected заявки and their extracts are kept (personal data of a person who never got access)? Default now: ≥ 5 years like all decisions. — owner: Andrii Opryshko, due: before implementation
- [ ] How the system reads the керівник from the extract (the extract's layout from the register)? Default now: read from the extract's text; unreadable → «не прочитано — перевірте вручну» (AC-10). — owner: Andrii Opryshko, due: SAD

## 9. Migration plan

- **New records:** заявки на підключення and запити на передачу, each with its signed extract file, the signature check result (signer, signing date, certificate validity), the керівник and the extract date read from the extract, every filing and the оператор's decision with its reason and the document relied on; disable and enable records with reason; the association gets its ЄДРПОУ, address, state (підключене / вимкнене) and a link to the decision that confirmed its current голова.
- **Changed records:** a signed-in account no longer carries one association: it may have none (an applicant before approval) or several (a person on several rosters); which associations a person may open comes from the rosters. A roster entry gains a position (голова правління or член правління) and, for a голова, a link to the decision that put them there or ended their term.
- **The signed extract is kept by Спільно itself**, not in the association's document storage: at filing time no storage is connected, and the proof must survive a change of the storage account.
- **Existing records:** the one association set up by hand becomes підключене ОСББ with no заявка, no ЄДРПОУ, marked «підключено вручну, без витягу»; its address is entered by the board in the settings, its ЄДРПОУ and голова come from its голова's запит на передачу (AC-19c). Existing roster entries get the position член правління; nobody is marked голова until AC-19c. Its accounts keep their association as a roster-derived membership; sessions, protocols and the storage connection stay untouched.
- **Board-protocol plan:** tasks T08 and T12 are replaced by this feature; board-protocol PRD §3 and AC-19 are amended (§1 ¶4).
- **Reversibility:** rolling the release back keeps every заявка, extract and decision; new associations then go back to manual setup. Signed extracts and decisions are never deleted. Restoring «one association per account» is possible only while no account has none or several.
- **Rollout order:**
  1. Isolation guard: the release-blocking cross-association check against two seeded associations and a person on both rosters; positions on roster entries; the hand-made association's backfill. Members see no change.
  2. Sign-in for people off the roster and on several rosters, заявки на підключення and the оператор's review. New associations connect without the developer.
  3. Запити на передачу, the roster change on approval, and AC-19c for the hand-made association.
  4. Disable / enable, and the board connecting its own document storage (replaces T12). Between steps 2 and 4 the developer still connects the storage for a new association.

## 10. RBAC and state machine

### Roles and actions

| Role | File заявка / запит | See own заявка | Decide заявка / запит | Disable / enable ОСББ | Read association records | Change association records | Connect document storage |
|---|---|---|---|---|---|---|---|
| голова правління (applicant, not on that association's roster) | ✓ | own only | ✗ | ✗ | ✗ | ✗ | ✗ |
| член правління of a підключене ОСББ (incl. its голова) | ✓ (for any other association) | own only | ✗ | ✗ | ✓ own association | ✓ own association | ✓ |
| член правління of a вимкнене ОСББ | ✓ (запит for it, or for any other association) | own only | ✗ | ✗ | read only: signed protocols of the registry | ✗ | ✗ |
| оператор | ✗ | ✗ | ✓ | ✓ | ✗ | ✗ | ✗ |
| ревізійна комісія, співвласник | ✗ (no access in this feature) | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ |

Scope is one association at a time: a person on several rosters has these rights in each association separately; a член правління's rights come from a term in force on the date of the action; a person whose term ended keeps no rights there.

### States

Заявка на підключення and запит на передачу (same machine):
- **Очікує** — filed with a valid signed extract, waiting for the оператор.
- **Схвалена** — approved; the association is created (заявка) or the roster changed (запит). Final.
- **Відхилена** — rejected with a reason; the applicant may fix and refile it.

ОСББ:
- **Підключене** — created by an approved заявка (or by hand before onboarding); members work as usual.
- **Вимкнене** — disabled by the оператор with a reason; members read and download signed protocols only.

### Allowed transitions

| From → To | Who | Condition | AC |
|---|---|---|---|
| — → Очікує (заявка) | голова правління | all fields valid, signature valid, extract for the typed ЄДРПОУ, code not connected, no open заявка of the same person for that code, term not ended; another person's pending заявка does not block | AC-01, AC-04c |
| — → Очікує (запит) | голова правління | the association is підключене or вимкнене, signature valid, extract for the ЄДРПОУ, term not ended | AC-17, AC-19c, AC-24b |
| Очікує → Схвалена (заявка) | оператор | code not connected by then; a document named when розбіжність імен is open | AC-08, AC-12 |
| Очікує → Схвалена (запит) | оператор | a document named when розбіжність імен is open | AC-19, AC-19c, AC-12 |
| Очікує → Відхилена | оператор | a reason given | AC-14 |
| Відхилена → Очікує | голова правління (the same applicant) | refiled with the same checks as filing | AC-06 |
| — → Підключене | system, on approval of a заявка | — | AC-08 |
| Підключене → Вимкнене | оператор | a reason given | AC-21 |
| Вимкнене → Підключене | оператор | — | AC-23 |

### Refused transitions

| From → To | Why refused | AC |
|---|---|---|
| — → Очікує with missing or wrong fields | the заявка must be complete to be checked | AC-02 |
| — → Очікує with an invalid signature or an extract of another code | no proof of office | AC-03, AC-03b, AC-18b |
| — → Очікує (заявка) for a connected code | one association, one record | AC-04 |
| — → Очікує (заявка) by the same person for a code they already have pending | duplicate | AC-04b |
| — → Очікує with an applicant's term already ended | a person whose term ended does not head the association | AC-30 |
| — → Очікує (запит) for a code not connected | nothing to transfer | AC-18 |
| Очікує → Схвалена (заявка) when another заявка for the code was approved | the code is connected now | AC-04c |
| Очікує → Схвалена over розбіжність імен without a document | the оператор checks documents, does not decide elections | AC-12 |
| Очікує → Відхилена without a reason | the applicant must see why | AC-13 |
| Схвалена → any, Відхилена → Схвалена / Відхилена | a decision is final; a change is a new filing or decision | AC-15 |
| Any → any by a person who is not the оператор | only the оператор decides access | AC-16 |
| Підключене → Вимкнене without a reason | members must see why | AC-22 |
| Вимкнене → Вимкнене, Підключене → Підключене | no change to record | AC-24 |
