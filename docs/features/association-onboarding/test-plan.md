---
status: Draft
owner: "Andrii Opryshko"
reviewers: ["Andrii Opryshko"]
updated_at: "2026-11-10"
feature_size: "M"
---

# Test plan — association-onboarding

> **Inputs:** [PRD](./PRD.md) §5 (42 acceptance criteria), §5.1 error catalogue, §6 NFRs, §10 roles and states ·
> [sad.md](./sad.md) §6 flows 1-12 and the coverage table, §10 quality scenarios · [data-model.md](./data-model.md) ·
> [contracts/openapi.yaml](./contracts/openapi.yaml) · `target_surfaces: [backend-service, web-frontend]`, so the UI tiers apply.

A голова правління connects an association, or takes it over after an election, by filing a заявка на
підключення or a запит на передачу with a signed витяг з ЄДР; the оператор records the check and decides with
a reason, can disable and enable an association, and the board connects its own document storage. With a
second association, nothing of one association is readable by a person off its roster. Every criterion below
has at least one test that `implement-tasks` writes red first.

## Level decisions (product owner, 2026-11-10)

| Group | Levels chosen | Default that was proposed |
|---|---|---|
| Refusals by a rule (AC-02, 03, 03b, 04, 04b, 04c, 11, 11b, 12, 13, 15, 16c, 18, 18b, 22, 24, 28, 28b, 30) | unit + integration + contract + e2e-through-UI, each refusal | unit + integration; the message checked at the component level |
| Access and isolation (AC-04d, 07, 16, 16b, 25, 26) | unit (the access matrix) + integration + contract + e2e-through-UI | integration + contract |
| Happy paths | integration, plus one e2e-through-UI flow per user story (9 flows) | 4 chained flows |
| Visual-regression | on for 3 screens: the filing form, the оператор's review page, the association card | N/A, as in board-protocol |

## Levels

| Level | Scope | Strategy (generic — no tool names) |
|---|---|---|
| Unit | Pure `identity` domain logic: the filing validator, the application and association state machines, the name comparison, the extract-age rule, the term rules, the approval guards, the access matrix of PRD §10; in `governance`, the end date of the previous голова's entry and the re-seat. | In-memory, no I/O; table-driven cases per rule; dates from an injected clock in Europe/Kyiv. |
| Integration | A use case through its module against the real database it owns, with row-level security active and the request-serving database role. | A fresh migrated ephemeral database per test (the repo's existing per-test database helper); the sign-in provider and the storage provider replaced by in-process fakes behind their ports. |
| Contract | The HTTP surface against the feature's OpenAPI delta: response shapes, problem bodies with their catalogue code, every broken field named, the same body for a missing and a forbidden id. | Validate real responses against the agreed document; no hand-rolled stubs. |
| E2E | <!-- N/A: every flow is driven through the UI (product owner's choice); an API-only tier would repeat it --> | — |
| Load | The numeric NFRs of PRD §6: read, write and filing latency, read throughput. | The load tool already in the repo, or a load-testing tool appropriate for the stack. |
| Component | A UI piece in isolation: the filing form, the status line, the source labels and warnings of the review page, the competing marker, the «доступ призупинено» notice, the association card, the storage line. | Render in the repo's component harness with its accessibility checks; assert output + interaction. |
| Visual-regression | Three screens: the filing form, the оператор's review page, the association card. | Snapshot the render in one fixed environment; fail on an unintended diff; the baseline is approved once and updated deliberately. |
| E2E-through-UI | One flow per user story (9) and every refusal and access denial, driven through the rendered web app. | Real web app + core service + ephemeral database; sign-in and storage providers faked. |

## AC coverage

Expected outcomes are in plain words. «Flow US-NN» is the one e2e-through-UI flow of that user story; a row
that names it is covered inside that flow, not by a separate browser test.

### US-01 — file an application

| AC (PRD §5) | Test name (intent-based) | Level | Expected outcome |
|---|---|---|---|
| AC-01 happy path | a signed-in person files a complete заявка with the signed extract | integration | recorded as очікує with the filing date; the stored file equals the uploaded one byte for byte; the same for a person who is a член правління of another association |
| AC-01 | голова files a заявка through the form (Flow US-01) | e2e-through-UI | «Заявку подано, очікує розгляду оператором» |
| AC-01 | the position is shown as «голова правління» and cannot be chosen | component + visual-regression (filing form) | no position control; term end optional |
| AC-02 error | every missing or wrong field of a заявка and of a запит is named at once | unit (table: code not of 8 digits, empty name, empty address, empty full name, another position, no term start, term start after the filing date, no file; the запит's own fields) | all broken fields listed together |
| AC-02 | an incomplete filing records nothing | integration + contract | refused; no application, filing or extract stored; the answer names each broken field in the agreed shape |
| AC-02 | the form marks every broken field after one submit | e2e-through-UI | each field carries its message; nothing filed |
| AC-04 domain invariant | a заявка for a connected code is refused without naming who heads it | unit + integration + contract | refused as already connected; nothing recorded; the answer carries no person |
| AC-04 | the refusal offers the запит на передачу form | e2e-through-UI | «Це ОСББ вже підключене…» with a link that opens the запит form |
| AC-04b domain invariant | a second заявка by the same person for the same code is refused and the open one shown | unit + integration + contract + e2e-through-UI | «У вас уже є заявка на це ОСББ»; the open заявка is shown; nothing recorded |
| AC-04d cross-context | a person on two rosters works in each association separately | unit (access matrix) + integration + contract | the list of own associations has both; records, roster and settings of each are returned only under that association and only to a person on its roster |
| AC-04d | two windows hold two associations after sign-in | e2e-through-UI | the person chooses an association; each window keeps its own; every page names its association |
| AC-30 legal competence | a filing by a голова whose term already ended is refused | unit (table: end before the filing date refused; end on the filing date accepted; no end accepted; заявка and запит) | refused only when the end is earlier than the filing date |
| AC-30 | the ended term is refused at filing and nothing is recorded | integration + contract + e2e-through-UI | «Строк повноважень голови вже завершився — подайте витяг після переобрання» |

### US-02 — follow, fix and refile

| AC (PRD §5) | Test name (intent-based) | Level | Expected outcome |
|---|---|---|---|
| AC-05 happy path | the applicant sees their pending заявка and nothing of the association | integration | own заявка with status, filing date and the оператор's contact address; no association data in any answer |
| AC-05 | the applicant signs in and reads the status (Flow US-02) | e2e-through-UI + component (status line) | «очікує розгляду з 10.11.2026» and the contact address |
| AC-06 happy path | only the same applicant may refile a rejected заявка | unit | Відхилена → Очікує allowed for the applicant, refused for anyone else |
| AC-06 | a rejected заявка is refiled with a new extract | integration | the same заявка is очікує with the new filing date; the rejected version, its extract and reason stay in the history; the earlier check is not carried over |
| AC-06 | the applicant reads the reason, fixes and refiles (Flow US-02) | e2e-through-UI | the new status is shown; the history shows the rejection |
| AC-07 authorization | a person off the roster is denied the registry, a справа and the roster | unit (access matrix) + integration (an applicant with a pending заявка, a member of another association, a person with nothing; an existing and a missing id) | denied in every combination; nothing read |
| AC-07 | the denial is the same for an existing and a missing record | contract | identical body in both cases |
| AC-07 | an applicant opening an association address sees no access | e2e-through-UI | a no-access page; no association data rendered |

### US-03 — decide an application

| AC (PRD §5) | Test name (intent-based) | Level | Expected outcome |
|---|---|---|---|
| AC-03 domain invariant | approval is refused until the check of the extract is recorded | unit + integration + contract + e2e-through-UI | «Спершу перевірте підпис КЕП…»; the заявка stays очікує; no association created |
| AC-03b domain invariant | approval is refused when the extract names another code | unit + integration + contract + e2e-through-UI | «Витяг стосується іншого ОСББ…»; the заявка stays очікує |
| AC-04c domain invariant | a competing заявка is accepted silently and shown side by side | integration + contract | the second applicant gets the same answer as a lone applicant; the оператор's view marks both as competing |
| AC-04c | approving the second of two competing заявки is refused once the first is approved | unit + integration + contract | refused as for a connected code; rejecting it with a reason still works; two approvals racing create exactly one association |
| AC-04c | the оператор sees «конкуруюча заявка» and is refused the second approval | e2e-through-UI + component (competing marker) | both listed together; the refusal message shown |
| AC-08 happy path | approving a checked заявка creates the association and its first член правління | integration | association with the code, name and address of the заявка; decision with date and оператор; the applicant on the roster as голова правління with the term of the заявка; all or nothing |
| AC-08 | the оператор records the check and approves (Flow US-03) | e2e-through-UI | the заявка shows схвалена with the decision date |
| AC-09 cross-context | the approved applicant works in the new association | integration | a current member from the term start; can open the registry and add a член правління; the storage shows as not connected |
| AC-09 | the approved applicant signs in and adds a board member (Flow US-03) | e2e-through-UI | the association opens; the member is added; «сховище не підключене» visible |
| AC-10 traceability | every value on the review page names its source | integration + contract | typed values marked as entered by the applicant; checked values marked with the check date; «не перевірено» before the check |
| AC-10 | source labels on the review page | component + visual-regression (review page) + e2e-through-UI (Flow US-03) | the three source labels and «не перевірено» render as specified |
| AC-11 traceability | three names are compared after levelling their spelling | unit (table: case, extra spaces, apostrophe forms, word order are the same person; another surname, a missing по батькові, Latin for Cyrillic are a розбіжність) | same person or a розбіжність naming which values differ |
| AC-11 | the розбіжність імен warning appears on recording and on reopening and blocks nothing by itself | integration + contract | warning with the differing values in the review answer; rejection works; approval works once a document is named |
| AC-11 | the warning is shown to the оператор | component + e2e-through-UI | the warning names the differing values |
| AC-11b traceability | an extract older than 30 days raises an advisory | unit (30 days none, 31 days warning) + integration + contract | «Витяг сформовано <дата> — старший за 30 днів»; approval needs nothing more |
| AC-11b | the age warning is shown to the оператор | component + e2e-through-UI | the warning with the extract date; approve stays available |
| AC-12 legal competence | approving over a розбіжність імен needs the document relied on | unit + integration + contract | refused without a reference; approved with a reference of up to 500 characters, which is stored with the decision; a longer one is refused as a broken field |
| AC-12 | the оператор is asked for the document | e2e-through-UI | «Схвалення з розбіжністю потребує документа…»; approval succeeds after typing it |
| AC-13 error | rejecting with an empty reason is refused | unit + integration + contract + e2e-through-UI | «Вкажіть причину — її побачить заявник»; stays очікує |
| AC-14 happy path | rejecting with a reason records it and creates nothing | integration | відхилена with the reason, date and оператор; no association; the applicant's own view carries the reason |
| AC-14 | the оператор rejects and the applicant reads the reason (Flow US-02) | e2e-through-UI | the reason is shown to the applicant after sign-in |
| AC-15 domain invariant | a decided заявка or запит cannot be decided again | unit (every refused transition of PRD §10) + integration + contract + e2e-through-UI | «Рішення за цією заявкою вже ухвалено»; the recorded decision is unchanged |
| AC-16 authorization | a person who is not the оператор gets nothing of the queue | unit (access matrix) + integration (a член правління, an applicant, a person with nothing; list, approve, reject, disable, enable) + contract | denied each time; no заявка listed; nothing changed |
| AC-16 | the operator section is absent and its address shows no access | e2e-through-UI | no menu entry; the direct address gives a no-access page |
| AC-16b authorization | the оператор role opens no association's records | unit (access matrix) + integration + contract | off the roster: denied with the same body as any outsider; on the roster of A: a member's rights in A and nothing in B |
| AC-16b | the оператор sees only their own association besides the queue | e2e-through-UI | the queue and the оператор's own association; another association's address gives no access |
| AC-16c domain invariant | the оператор cannot decide their own filing | unit + integration (approve and reject; заявка and запит) + contract + e2e-through-UI | «Це ваша власна заявка…»; stays очікує |

### US-04 — request a transfer

| AC (PRD §5) | Test name (intent-based) | Level | Expected outcome |
|---|---|---|---|
| AC-17 happy path | a запит на передачу is filed for a connected association by a person on or off its roster | integration + contract | recorded as очікує with the extract; the answer reveals nobody of the current roster |
| AC-17 | the new голова files a запит (Flow US-04) | e2e-through-UI | «Запит на передачу подано, очікує розгляду оператором» |
| AC-18 domain invariant | a запит for a code that is not connected is refused | unit + integration + contract + e2e-through-UI | «Це ОСББ ще не підключене…» with a link to the заявка form; nothing recorded |

### US-05 — decide a transfer

| AC (PRD §5) | Test name (intent-based) | Level | Expected outcome |
|---|---|---|---|
| AC-18b domain invariant | a запит is not approved without a recorded check or with an extract of another code | unit + integration + contract + e2e-through-UI | refused the same way as a заявка; stays очікує; the roster unchanged |
| AC-19 happy path | the previous голова's entry ends the day before the new term start | unit (table: start 20.11 → end 19.11; never before the first day of the previous entry; an applicant already on the roster is re-seated, no second entry) | the end date and the re-seat follow the rule |
| AC-19 | approving a запит changes the roster in one step | integration | the new голова is current from the typed term start; the previous голова's whole entry is ended; other entries untouched; both changes link to the extract and the decision; all or nothing |
| AC-19 | the previous голова loses access at the moment of approval | integration | the next request of their open session is denied |
| AC-19 | the оператор approves and the new голова takes over (Flow US-05) | e2e-through-UI | the new голова opens the association; the previous голова gets no access |
| AC-19b happy path | two запити for one association are shown as competing and decided independently | integration | approving one leaves the other очікує; rejecting it needs a reason; approving it later ends the term of the голова confirmed first |
| AC-19b | the «конкуруючий запит» marker | component | both запити listed together with the marker |
| AC-19c cross-context | a запит for the association set up by hand attaches the extract and the code | integration | the code is recorded from the запит; the голова's entry gets the position; no term ends; while that голова is the оператор the запит cannot be decided |
| AC-19c | the card replaces «підключено вручну, без витягу» with the extract | component | the mark before approval; the extract, date and signer after |
| AC-20 traceability | roster entries of the new and the previous голова name the запит and the decision | integration + contract | appointed / ended notes with the запит date, the decision date and a link to the extract |
| AC-20 | the notes are shown on the roster (Flow US-05) | component + e2e-through-UI | «призначено / повноваження завершено за запитом на передачу від …, схвалено оператором …» |

### US-06 — disable and enable

| AC (PRD §5) | Test name (intent-based) | Level | Expected outcome |
|---|---|---|---|
| AC-21 happy path | disabling records the reason and takes effect on the member's next action | integration | вимкнене with reason, date and оператор; the next change of a member's open session is refused with the reason |
| AC-21 | the оператор disables an association (Flow US-06) | e2e-through-UI | members see «Доступ призупинено: спір щодо складу правління» |
| AC-22 error | disabling with an empty reason is refused | unit + integration + contract + e2e-through-UI | «Вкажіть причину — її побачать члени правління»; stays підключене |
| AC-23 happy path | enabling restores the association as it was | integration | підключене; every record equal to its state before disabling; the enabling recorded with date and оператор |
| AC-23 | the оператор enables it again (Flow US-06) | e2e-through-UI | members change records again |
| AC-24 domain invariant | disabling a disabled and enabling an enabled association is refused | unit + integration + contract + e2e-through-UI | «ОСББ уже в цьому стані»; no state change recorded |
| AC-24b cross-context | a transfer is approved for a disabled association | integration | the roster changes as in AC-19; the association stays вимкнене |

### US-07 — work in a disabled association

| AC (PRD §5) | Test name (intent-based) | Level | Expected outcome |
|---|---|---|---|
| AC-25 authorization | a member of a disabled association can change nothing | unit (access matrix) + integration (edit a draft, upload a document, change the roster, change the settings, connect the storage) + contract | each refused with the reason; nothing changes |
| AC-25 | the change controls answer «Доступ призупинено» | e2e-through-UI + component (notice) | the notice with the reason; no change saved |
| AC-26 happy path | the registry and signed copies stay readable in a disabled association | unit (access matrix) + integration + contract | the registry is listed; the signed copy is served unchanged |
| AC-26 | a member downloads a signed protocol while disabled (Flow US-07) | e2e-through-UI | the file downloads |

### US-08 — connect the document storage

| AC (PRD §5) | Test name (intent-based) | Level | Expected outcome |
|---|---|---|---|
| AC-27 happy path | a член правління connects the association's own storage account | integration (fake storage provider) | the connection is recorded with the account address and the member; every член правління reads it |
| AC-27 | the member connects the storage (Flow US-08) | e2e-through-UI + component (storage line) | «Сховище документів: <адреса акаунта>» |
| AC-28 cross-context | attaching a document before the storage is connected is refused | integration + contract + e2e-through-UI | «Спершу підключіть сховище документів ОСББ»; says any член правління can connect it; nothing attached |
| AC-28b domain invariant | a member's own sign-in account is refused as the storage account | unit + integration + contract + e2e-through-UI | «Це особистий акаунт члена правління — підключіть акаунт ОСББ»; nothing connected |
| AC-28b | the оператор sees which account holds the files on a later запит | integration | the review answer of a запит carries the storage account address |

### US-09 — how the голова was confirmed

| AC (PRD §5) | Test name (intent-based) | Level | Expected outcome |
|---|---|---|---|
| AC-29 traceability | the association card shows the confirmation chain | integration + contract | the current голова, the extract with its date and signer, the decision date, every later запит and disabling with its reason; the hand-made association is marked as connected by hand |
| AC-29 | the card renders the chain | component + visual-regression (association card) | all parts present in both variants (with an extract, connected by hand) |
| AC-29 | a member opens the card and the extract (Flow US-09) | e2e-through-UI | the card is shown; the extract downloads |

### Quality checks behind the criteria (PRD §6)

| Requirement | Test name (intent-based) | Level | Expected outcome |
|---|---|---|---|
| Isolation (release-blocking) | every association-scoped operation of the contract is called with another association's id | integration (two seeded associations, a person on A, a person on both, the оператор on A's roster only) | zero records of B returned to a person off B's roster; a new scoped operation is picked up without editing the test |
| Isolation | a query on a scoped table without an association set reads nothing | integration | empty result, never another association's rows |
| Effect of decisions | the old session after a disabling or a transfer | integration | the first request after the decision is already refused |
| Integrity | the request-serving database role cannot change or delete extracts, checks, filings, decisions and state changes | integration | each attempt is refused by the database |
| Integrity | the downloaded extract matches the stored fingerprint | integration | equal on every download |
| File size and kind | the extract upload boundary | unit + integration + contract | exactly 5 MB accepted; one byte more refused; a file that is neither a PDF nor a signed container refused; a separate signature file refused; nothing stored on refusal |

## Edge cases / error paths

- No session on any operation → refused as not signed in; nothing read.
- The sign-in provider returns an unverified address → sign-in refused; no account created.
- A body that cannot be decoded → refused as unreadable, apart from field validation.
- Another person's заявка or its extract, by id → answered as not existing, the same as a missing id.
- The extract is downloaded only by its applicant and the оператор; a член правління reaches it only through the association card of their own association.
- A refile when the code was connected meanwhile → refused as already connected; the rejected заявка stays відхилена.
- A member whose term ended yesterday → denied; a member whose term ends today → still served.
- A competing запит with an earlier term start approved later → the previous entry ends no earlier than its first day (SAD §8 Dates, §11).
- The roster step of an approval fails → the decision, the association and the roster change are all undone; the заявка stays очікує.
- The storage provider refuses the consent or is unavailable → nothing connected; the member is told to try again.
- Returning from the storage consent with a foreign or replayed confirmation → refused; nothing connected.
- A filing or a decision with several problems → every broken field named in one answer.
- The оператор's address also on a roster → the queue and that one association, nothing wider.

## Test data

- Seed strategy: factories per entity of [data-model.md](./data-model.md) — associations (connected, disabled, set up by hand without a code), roster entries with positions and terms (current, ended yesterday, ending today), accounts (on no roster, on one, on two, the оператор on one roster), applications of both kinds in each state with their filings, checks, decisions and competing pairs, state changes, a storage connection. Extract fixtures: a small signed PDF, a file of exactly 5 MB, one of 5 MB plus one byte, a non-PDF. Placeholder data only (`<name>@example.test`, «Тестовий Голова», codes like 12345678).
- Clock: an injected date in Europe/Kyiv for every term, filing-date and extract-age case.
- Integration dependency: a fresh migrated ephemeral database per test, connected as the request-serving role so row-level security is exercised; never a mocked store. Sign-in and storage providers are in-process fakes behind their ports, with one smoke test per provider against a test account before release.
- Cleanup boundary: per test for integration (its own database); per suite for e2e-through-UI (each flow files under its own code so flows do not collide; the database is torn down after the run).
- Visual baselines: stored with the repo, produced in one fixed environment, updated only in a commit that says so.

## NFR validation (load)

The load tool already in the repo, or a load-testing tool appropriate for the stack. Seed: 20 associations, 200 applications with extracts.

- Read p95 ≤ 200 ms and throughput ≥ 50 req/s per instance → sustain 50 requests/s for 5 min over a mix of the application page, the оператор's queue and the association card; assert p95 ≤ 200 ms and no errors.
- Write p95 ≤ 500 ms (without a file) → 5 requests/s of approvals, rejections, disablings and enablings for 5 min over pre-seeded pending applications; assert p95 ≤ 500 ms; refusals of an already decided заявка are expected and not counted as errors.
- Filing p95 ≤ 5 s for a 5 MB file → 20 sequential filings with the 5 MB fixture; assert p95 ≤ 5 s.
- Not load tests, verified elsewhere: isolation, effect of decisions and integrity (integration above), availability (uptime check), RPO/RTO (quarterly restore drill), retention (storage policy and the privilege test).

## CI placement

- On every PR: unit, integration (the isolation check included — it blocks the release), contract, component, visual-regression, e2e-through-UI (all with fakes; the repo already runs its browser suite on PRs).
- Pre-release / on schedule: the load scenarios and the smoke tests against the real sign-in and storage providers.
- Order of arrival follows PRD §9: the isolation check and the roster positions first, then filing and review, then transfers, then disabling and the storage connection.
