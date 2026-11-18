---
status: Draft
owner: "Andrii Opryshko"
reviewers: ["Andrii Opryshko"]
updated_at: "2026-11-18"
feature_size: "M"
---

# Test plan — association-onboarding

> **Inputs:** [PRD](./PRD.md) §5 (43 acceptance criteria), §5.1 error catalogue, §6 NFRs, §10 roles and states ·
> [sad.md](./sad.md) §6 flows 1-12 and the coverage table, §10 quality scenarios · [data-model.md](./data-model.md) ·
> [contracts/openapi.yaml](./contracts/openapi.yaml) · `target_surfaces: [backend-service, web-frontend]`, so the UI tiers apply.

A голова правління connects an association, or takes it over after an election, by filing a заявка на
підключення or a запит на передачу with a signed витяг з ЄДР; the оператор records the check and decides with
a reason, can disable and enable an association, and the board connects its own document storage. With a
second association, nothing of one association is readable by a person off its roster. Every criterion below
has at least one test that `implement-tasks` writes red first.

Updated 2026-11-17 after the follow-up of the [review of 2026-11-10](./_review/review-2026-11-10.md) (tasks
T50–T85): the rows below name the tests as they are now, and «Where the tests are» lists the browser specs
by criterion. No row waits for a skipped test.

Updated again 2026-11-17 after the second follow-up ([review of 2026-11-17](./_review/review-2026-11-17.md),
tasks T86–T111): the rows it added or changed are in place below, and «Tests of the second follow-up» under
«Where the tests are» names each new test function, component test and spec by criterion.

Updated a third time 2026-11-18 after the third follow-up ([second re-review of 2026-11-17](./_review/review-2026-11-17-2.md),
tasks T112–T134): AC-27b is mapped, every criterion names tests that exist (each name was grepped), the tests
the second follow-up named under their old titles are renamed, and «Tests of the third follow-up» under «Where
the tests are» names each new test by criterion. The criteria that have no browser flow are said so in their rows.

## Level decisions (product owner, 2026-11-10)

| Group | Levels chosen | Default that was proposed |
|---|---|---|
| Refusals by a rule (AC-02, 03, 03b, 04, 04b, 04c, 11, 11b, 12, 13, 15, 16c, 18, 18b, 22, 24, 27b, 28, 28b, 30) | unit + integration + contract + e2e-through-UI, each refusal | unit + integration; the message checked at the component level |
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
| AC-01 | голова files a заявка through the form (Flow US-01) | e2e-through-UI | the form is opened by the link on the person's list of заявки, not by a typed address; «Заявку подано, очікує розгляду оператором» |
| AC-01 | the list of заявки leads to the form | component | «Подати заявку на підключення» next to the empty message and above a list that has a заявка |
| AC-01, AC-05, AC-06, AC-14 | a person on one roster reaches their заявки and the form from inside the association | component (association sidebar, picker) | «Мої заявки» and «Подати заявку на підключення» in the sidebar of the association layout, for a person on one roster and on several; «Мої заявки» in the picker too |
| AC-01, AC-05 | a member of one association files a second заявка from inside it and follows it | e2e-through-UI (the AC-04d spec) | the second заявка is filed by the sidebar link (the person's first is approved, so `/` opens it) and followed by «Мої заявки» from the sidebar; no address typed |
| AC-01 | the position is shown as «голова правління» and cannot be chosen | component + visual-regression (filing form) | no position control; term end optional |
| AC-02 error | every missing or wrong field of a заявка and of a запит is named at once | unit (table: code not of 8 digits, empty name, empty address, empty full name, another position, no term start, term start after the filing date, no file; the запит's own fields) | all broken fields listed together |
| AC-02 | an incomplete filing records nothing | integration + contract | refused; no application, filing or extract stored; the answer names each broken field in the agreed shape |
| AC-02 | the form marks every broken field after one submit | e2e-through-UI | each field carries its message; nothing filed |
| AC-02 | the form checks the extract before it is sent | component | a file over 5 MB or of another type is refused under the field and the refused file is never sent; a 413 answer shows the same field message |
| AC-02 | a refused extract is named together with every other broken field | component | the other fields go to the server without the file and the server's answer marks every other broken field at once, beside «Витяг»; with all other fields valid only the extract is named and nothing is recorded |
| AC-02 | a body over the limit sent without its length is refused as too large | unit + integration (a filing, a governance JSON body and a document upload, a finance contract) | answered 413 in the problem shape, not as a field error or a server error; nothing stored |
| AC-04 domain invariant | a заявка for a connected code is refused without naming who heads it | unit + integration + contract | refused as already connected; nothing recorded; the answer carries no person |
| AC-04 | the refusal offers the запит на передачу form | e2e-through-UI | «Це ОСББ вже підключене…» with a link that opens the запит form |
| AC-04b domain invariant | a second заявка by the same person for the same code is refused and the open one shown | unit + integration + contract + e2e-through-UI | «У вас уже є заявка на це ОСББ»; the open заявка is shown; nothing recorded |
| AC-04d cross-context | a person on two rosters works in each association separately | unit (access matrix) + integration + contract | the list of own associations has both; records, roster and settings of each are returned only under that association and only to a person on its roster |
| AC-04d | two windows hold two associations after sign-in | e2e-through-UI | the person's second заявка is filed from inside the first association (see AC-01 above); the windows are two pages of one browser, so one session: the person picks an association in each and opens its card from the sidebar; each window keeps its own after reloads; every page names its association |
| AC-04d | a person reaches every association they are in, and the form for another one | component | every screen reads memberships through one hook, so the twenty-first association opens on its protocol page and from the storage callback as well as in the layout; a server that repeats a cursor ends in the shared fallback after a bounded number of requests; the picker carries «Подати заявку на підключення» under the associations |
| AC-30 legal competence | a filing by a голова whose term already ended is refused | unit (table: end before the filing date refused; end on the filing date accepted; no end accepted; заявка and запит) | refused only when the end is earlier than the filing date |
| AC-30 | the ended term is refused at filing and nothing is recorded | integration + contract + e2e-through-UI | «Строк повноважень голови вже завершився — подайте витяг після переобрання» |

### US-02 — follow, fix and refile

| AC (PRD §5) | Test name (intent-based) | Level | Expected outcome |
|---|---|---|---|
| AC-05 happy path | the applicant sees their pending заявка and nothing of the association | integration | own заявка with status, filing date and the оператор's contact address; no association data in any answer |
| AC-05 | the applicant signs in and reads the status (Flow US-02) | e2e-through-UI + component (status line) | «очікує розгляду з 10.11.2026» and the contact address |
| AC-05 | a person signs in with the «Увійти» button (Flow US-03) | e2e-through-UI | the button leads through the provider back into the app, signed in as the оператор |
| AC-05 | the pending sign-in belongs to the browser that started it | integration | it travels in a sealed cookie scoped to the path of the provider's redirect address, with or without the web app's `/api` prefix, not readable by scripts; a callback with no cookie, another browser's, a tampered one or a replay is refused; any number of other starts leaves a sign-in in progress able to finish; callbacks with a genuine cookie and a bogus code leave the set of used states bounded and an earlier sign-in able to finish |
| AC-06 happy path | only the same applicant may refile a rejected заявка | unit | Відхилена → Очікує allowed for the applicant, refused for anyone else |
| AC-06 | a rejected заявка is refiled with a new extract | integration | the same заявка is очікує with the new filing date; the rejected version, its extract and reason stay in the history; the earlier check is not carried over |
| AC-06 | a rejection that names a filing that is no longer the latest is refused | integration | answered as a conflict, not a server error; nothing is recorded |
| AC-06 | the applicant reads the reason, fixes and refiles; the оператор decides the refiled заявка (Flow US-02) | e2e-through-UI + component (review page) | the new status is shown; the history shows the rejection; the review page shows the new filing with the earlier one below, the extract the оператор downloads is the refiled file byte for byte, and the approval goes through |
| AC-07 authorization | a person off the roster is denied the registry, a справа and the roster | unit (access matrix) + integration (an applicant with a pending заявка, a member of another association, a person with nothing; an existing and a missing id) | denied in every combination; nothing read |
| AC-07 | the denial is the same for an existing and a missing record | contract | identical body in both cases |
| AC-07 | an applicant opening an association address sees no access | e2e-through-UI | a no-access page; no association data rendered |

### US-03 — decide an application

| AC (PRD §5) | Test name (intent-based) | Level | Expected outcome |
|---|---|---|---|
| AC-03 domain invariant | approval is refused until the check of the extract is recorded | unit + integration + contract + e2e-through-UI | «Спершу перевірте підпис КЕП…»; the заявка stays очікує; no association created |
| AC-03b domain invariant | approval is refused when the extract names another code | unit + integration + contract + e2e-through-UI | «Витяг стосується іншого ОСББ…»; the заявка stays очікує |
| AC-04c domain invariant | a competing заявка is accepted silently and shown side by side | integration + contract | the second applicant gets the same answer as a lone applicant; the оператор's view marks both as competing; a заявка and a запит for one code do not mark each other |
| AC-04c | approving the second of two competing заявки is refused once the first is approved | unit + integration + contract | refused as for a connected code; rejecting it with a reason still works; two approvals racing create exactly one association |
| AC-04c | the оператор sees «конкуруюча заявка» and is refused the second approval | e2e-through-UI + component (competing marker) | both marked in the queue; each review page lists the other and opens it; after the first approval the second is no longer marked and its approval is refused with the message |
| AC-04c | the queue shows every pending заявка | component | «Показати ще» while there is a next page; the next page is added under the first; the page skeleton while the first page is read |
| AC-08 happy path | approving a checked заявка creates the association and its first член правління | integration | association with the code, name and address of the заявка; decision with date and оператор; the applicant on the roster as голова правління with the term of the заявка; all or nothing |
| AC-08 | the term end typed in the заявка is kept on the roster entry | integration | the голова's entry ends on the filed date; without one it stays open |
| AC-08 | the оператор opens the заявка from the queue, records the check and approves (Flow US-03) | e2e-through-UI | the заявка shows схвалена with the decision date |
| AC-09 cross-context | the approved applicant works in the new association | integration | a current member from the term start; can open the registry and add a член правління; the storage shows as not connected |
| AC-09 | the roster is listed, added to and changed | integration + contract + component (roster screen) | entries with position, term and decision; every broken field named at once; a seat made by a decision keeps its position and decision |
| AC-09 | the registry and a справа with its documents open for a member | integration + contract + component (registry, case screen) | the registry lists signed decisions by filter and page; a справа shows its documents and versions |
| AC-09 | the approved applicant signs in and adds a board member (Flow US-03) | e2e-through-UI | the association opens; «сховище не підключене» visible; the member is added on the roster screen, signs in and lands in the association with the same roster |
| AC-10 traceability | every value on the review page names its source | integration + contract | typed values marked as entered by the applicant; checked values marked with the check date; «не перевірено» before the check |
| AC-10 | source labels on the review page | component + visual-regression (review page) + e2e-through-UI (Flow US-03) | each of the six typed and five checked values carries its label; «не перевірено» before the check and gone after it |
| AC-10 | the review page never renders empty | component | a skeleton while the read is out, and while a competing заявка loads it sits inside a list item; a mistyped id, a server error and a network error show the shared fallback; no session leads to sign-in |
| AC-11 traceability | three names are compared after levelling their spelling | unit (table: case, extra spaces, apostrophe forms, word order are the same person; another surname, a missing по батькові, Latin for Cyrillic are a розбіжність) | same person or a розбіжність naming which values differ |
| AC-11 | the розбіжність імен warning appears on recording and on reopening and blocks nothing by itself | integration + contract | warning with the differing values in the review answer; rejection works; approval works once a document is named |
| AC-11 | the warning is shown to the оператор | component + e2e-through-UI | «Розбіжність імен: голова за витягом «…»», in Ukrainian, naming the value that differs from the other two |
| AC-11b traceability | an extract older than 30 days raises an advisory | unit (30 days none, 31 days warning) + integration + contract | «Витяг сформовано <дата> — старший за 30 днів»; approval needs nothing more |
| AC-11b | the age warning is shown to the оператор | component + e2e-through-UI | «Витяг сформовано <дата> — старший за 30 днів» with the extract date as the screens print dates; approve stays available |
| AC-12 legal competence | approving over a розбіжність імен needs the document relied on | unit + integration + contract | refused without a reference; approved with a reference of up to 500 characters, which is stored with the decision; a longer one is refused as a broken field |
| AC-12 | the оператор is asked for the document | e2e-through-UI | «Схвалення з розбіжністю потребує документа…»; approval succeeds after typing it |
| AC-13 error | rejecting with an empty reason is refused | unit + integration + contract + e2e-through-UI | «Вкажіть причину — її побачить заявник»; stays очікує |
| AC-14 happy path | rejecting with a reason records it and creates nothing | integration | відхилена with the reason, date and оператор; no association; the applicant's own view carries the reason |
| AC-14 | the оператор rejects and the applicant reads the reason (Flow US-02) | e2e-through-UI | the reason is shown to the applicant after sign-in |
| AC-15 domain invariant | a decided заявка or запит cannot be decided again | unit (every refused transition of PRD §10) + integration + contract + e2e-through-UI | «Рішення за цією заявкою вже ухвалено»; the recorded decision is unchanged |
| AC-15 | the outcome of a decision is announced where the оператор is | component (decision panel) | after an approval, as after a rejection, focus moves to the status line that names the outcome |
| AC-16 authorization | a person who is not the оператор gets nothing of the queue | unit (access matrix) + integration (a член правління, an applicant, a person with nothing; list, approve, reject, disable, enable) + contract | denied each time; no заявка listed; nothing changed |
| AC-16 | the operator section is absent and its address shows no access | e2e-through-UI | no menu entry; the direct address gives a no-access page |
| AC-16b authorization | the оператор role opens no association's records | unit (access matrix) + integration + contract | off the roster: denied with the same body as any outsider; on the roster of A: a member's rights in A and nothing in B |
| AC-16b | the оператор sees only their own association besides the queue | e2e-through-UI | off every roster: the queue, and no access to an association's address; seated on a roster by its голова: the roster read and a record changed there, another association refused, the queue still theirs |
| AC-16c domain invariant | the оператор cannot decide their own filing | unit + integration (approve and reject; заявка and запит) + contract + e2e-through-UI | «Це ваша власна заявка…»; stays очікує |

### US-04 — request a transfer

| AC (PRD §5) | Test name (intent-based) | Level | Expected outcome |
|---|---|---|---|
| AC-17 happy path | a запит на передачу is filed for a connected association by a person on or off its roster | integration + contract | recorded as очікує with the extract; the answer reveals nobody of the current roster |
| AC-17 | the new голова files a запит (Flow US-04) | e2e-through-UI | «Запит на передачу подано, очікує розгляду оператором» |
| AC-17 | a newly elected голова who is already on the roster files a запит from the card | e2e-through-UI | the roster entry is made first; `/` opens their association, the link on the card opens the запит form and the filing succeeds with nothing of the current roster revealed |
| AC-17 | the card offers the запит to every member | component + visual-regression (association card) | the link is on the card of an association connected by a заявка as well as of one connected by hand; the «підключено вручну» badge only on the hand-made one |
| AC-18 domain invariant | a запит for a code that is not connected is refused | unit + integration + contract + e2e-through-UI | «Це ОСББ ще не підключене…» with a link to the заявка form; nothing recorded |

### US-05 — decide a transfer

| AC (PRD §5) | Test name (intent-based) | Level | Expected outcome |
|---|---|---|---|
| AC-18b domain invariant | a запит is not approved without a recorded check or with an extract of another code | unit + integration + contract + e2e-through-UI | refused the same way as a заявка; stays очікує; the roster unchanged |
| AC-19 happy path | the previous голова's entry ends the day before the new term start | unit (table: start 20.11 → end 19.11; never before the first day of the previous entry; an applicant already on the roster is re-seated, no second entry) | the end date and the re-seat follow the rule |
| AC-19 | approving a запит changes the roster in one step | integration | the new голова is current from the typed term start; the previous голова's whole entry is ended; other entries untouched; both changes link to the extract and the decision; all or nothing |
| AC-19 | the previous голова loses access at the moment of approval | integration | the next request of their open session is denied; also when their entry already had a later term end; at Kyiv midnight the term start day is already today |
| AC-19 | an entry an approval ended cannot be reopened | integration + contract | a cleared or later term end on that entry is refused, a corrected name is saved; a previous голова added again is a second entry, a член правління without a decision, and one current голова remains |
| AC-19 | the association's screens read the memberships once | component (справи, registry, signed protocol, roster, protocol, storage callback) + reuse test | one memberships request for a visit, through `useAllMemberships` and the layout's cache that every navigation re-reads; the reuse test forbids the generated hook elsewhere; an expired session leads to sign-in |
| AC-19 | the оператор approves and the new голова takes over (Flow US-05) | e2e-through-UI | the new голова opens the association; the previous голова gets no access |
| AC-19b happy path | two запити for one association are shown as competing and decided independently | integration | approving one leaves the other очікує; rejecting it needs a reason; approving it later ends the term of the голова confirmed first, also after the previous голова's entry was edited; a заявка for the same code is not a competitor of a запит |
| AC-19b | the «конкуруючий запит» marker | component | both запити listed together with the marker |
| AC-19c cross-context | a запит for the association set up by hand attaches the extract and the code | integration | the code is recorded from the запит; the голова's entry gets the position; no term ends; while that голова is the оператор the запит cannot be decided |
| AC-19c | the card replaces «підключено вручну, без витягу» with the extract | component | the mark before approval; the extract, date and signer after |
| AC-20 traceability | roster entries of the new and the previous голова name the запит and the decision | integration + contract | appointed / ended notes with the запит date, the decision date and a link to the extract; each entry says whether its decision seated or ended it, and an edit of the ended entry changes neither answer |
| AC-20 | the notes are shown on the roster (Flow US-09) | component (roster screen, card) + e2e-through-UI | «призначено / повноваження завершено за запитом на передачу від …, схвалено оператором …» on the card and on the two roster entries the decision touched, each with a link that downloads the extract; «повноваження завершено» only where a term ended |
| AC-19b, AC-20 | each entry reports what its decision did after two transfers in a row | integration | a connection and two transfers approved through the API: the голова of the first decision is `ended` and the second `seated`, the new голова `seated`; an applicant re-seated on an existing entry has one entry, reported `seated` |
| AC-20 | the note follows what the decision did and what kind it was | component (roster screen, mark) + e2e-through-UI (Flows US-03, US-05) | two голови with one full name are told apart; the first голова of a new association reads «призначено за заявкою на підключення від …», never «за запитом на передачу»; after the previous голова's name is corrected the same two entries keep their notes; an ended entry shows its term as text with «Щоб повернути людину до правління, додайте запис знову» |
| AC-20 | the roster shows every entry | component (roster screen) | «Показати ще» while there is a next page; a later entry appears with its note |

### US-06 — disable and enable

| AC (PRD §5) | Test name (intent-based) | Level | Expected outcome |
|---|---|---|---|
| AC-21 happy path | disabling records the reason and takes effect on the member's next action | integration | вимкнене with reason, date and оператор; the next change of a member's open session is refused with the reason |
| AC-21 | the оператор disables an association (Flow US-06) | e2e-through-UI | the оператор reaches the list from the section's navigation; on the member's next click in the sidebar they see «Доступ призупинено: спір щодо складу правління» |
| AC-21 | the list of associations shows every association and never renders empty | component | «Показати ще» while there is a next page; a skeleton while the read is out and the shared fallback on a failed read |
| AC-22 error | disabling with an empty reason is refused | unit + integration + contract + e2e-through-UI | «Вкажіть причину — її побачать члени правління»; stays підключене |
| AC-23 happy path | enabling restores the association as it was | integration | підключене; every record equal to its state before disabling; the enabling recorded with date and оператор |
| AC-23 | the оператор enables it again (Flow US-06) | e2e-through-UI | members change records again |
| AC-24 domain invariant | disabling a disabled and enabling an enabled association is refused | unit + integration + contract + e2e-through-UI | «ОСББ уже в цьому стані»; no state change recorded |
| AC-24b cross-context | a transfer is approved for a disabled association | integration | the roster changes as in AC-19; the association stays вимкнене |

### US-07 — work in a disabled association

| AC (PRD §5) | Test name (intent-based) | Level | Expected outcome |
|---|---|---|---|
| AC-25 authorization | a member of a disabled association can change nothing | unit (access matrix) + integration (edit a draft, upload a document, change the roster, change the settings, connect the storage) + contract | each refused with the reason; nothing changes; the contract lists the disabled refusal on the file download too |
| AC-25 | the operations that talk to the storage are checked first and run afterwards | integration | for a document upload, a file download and the storage consent no database connection is held while the body is read or the outside call runs; the roster and read-only refusals are unchanged |
| AC-25 | every multipart operation is on that list | unit (a walk over the contract) | an association-scoped operation with a multipart body that the guard does not run outside its transaction fails the build, `createAssistantTurn` and `lockProtocolWithSignedCopy` among them |
| AC-25 | the protocol screen shows the new draft after a write | component (protocol screen) | a saved slot is no longer dirty and a transition changes the state and the buttons, with no reload: the draft is read and written under one cache key |
| AC-25 | the change controls answer «Доступ призупинено» | e2e-through-UI + component (notice, roster form, case screen) | the notice with the reason on every page of the association; no add, change or upload control; no change saved |
| AC-26 happy path | the registry and signed copies stay readable in a disabled association | unit (access matrix) + integration + contract | the registry is listed; the signed copy is served unchanged; a file no signed protocol pins is refused with the reason; a request that reaches the download without a checked person is refused as a person off the roster |
| AC-26 | a file is read from the storage of its own association | integration (in-memory storage; a fake storage provider over HTTP) | another association's storage does not hold the bytes; with the real adapter wired the download returns the uploaded bytes from the folder of the association's connection; the bytes are read with no transaction open, and a read that fails answers as storage unavailable |
| AC-26 | a file is read through a resolver that asks once and gives up | integration (a fake Google over HTTP) | two requests of one association ask the token endpoint once and a reconnect is used by the next request; a Drive that never answers ends in 503 within the deadline and an unreadable token answers 503 `governance.storage_unavailable` on upload and on download, not 500 |
| AC-26 | the development storage never repeats a file id across associations | unit + integration | the first uploads of two associations, and two instances of the fake, get distinct ids; the browser spec of AC-26 passes a second time on one database |
| AC-26 | the registry and the signed protocol render under the notice | component (registry screens) | the list and the protocol with its download links, the notice above; a protocol that is not found offers a link back to the registry beside the retry |
| AC-26 | a member downloads a signed protocol while disabled (Flow US-07) | e2e-through-UI | the member opens the реєстр and the protocol from the sidebar under the notice; the signed scan downloads byte for byte; the search answers; a document outside a signed protocol stays refused |

### US-08 — connect the document storage

| AC (PRD §5) | Test name (intent-based) | Level | Expected outcome |
|---|---|---|---|
| AC-27 happy path | a член правління connects the association's own storage account | integration (fake storage provider) | the connection is recorded with the account address and the member; every член правління reads it |
| AC-27 | the member connects the storage (Flow US-08) | e2e-through-UI + component (storage line) | «Сховище документів: <адреса акаунта>» |
| AC-27 | the consent `cmd/api` wires completes and the documents go through the resolver `cmd/api` runs | integration (`cmd/api`, a fake Google) | the consent built by `wireGoogle`, with its Drive address, creates the connection with a folder; a document uploaded and downloaded goes to that folder through `governance.NewDriveResolver`; `wireGoogle` refuses a consent without the address (`TestWireGoogleRefusesAConfigurationWithoutTheDriveAddress`) |
| AC-27 | the consent is exchanged outside the database transaction and one account holds one pending consent | integration (fake storage provider) | while one completion waits for the provider another request is served; a later start replaces the earlier one of the same account; a flood of starts by one member leaves another member's consent able to finish |
| AC-27b domain invariant | another account is refused while the association has files | unit + integration (fake storage provider) + contract | 409 `identity.storage_has_files`; the connection and the file stay as they were; the same account (letter case ignored) connects again and an association without files connects any account; the code is registered once as a conflict, is in the PRD §5.1 catalogue and in both contracts; there is no browser flow: no spec seeds a file before the consent |
| AC-27b | the member is told why and which account holds the files | component (storage callback) | «У сховищі ОСББ уже є файли — підключіть той самий акаунт, з яким вони були збережені», the connected account named, the link back to the card |
| AC-28 cross-context | attaching a document before the storage is connected is refused | integration + contract + component (case screen) + e2e-through-UI | «Спершу підключіть сховище документів ОСББ» on the case screen; says any член правління can connect it and links to the card; nothing attached |
| AC-28 | a document goes to the association's storage before it is recorded | integration (in-memory storage; a fake storage provider over HTTP) | the bytes are stored with no transaction open, then the file and the document are recorded together; a failed record removes the stored file, and a removal that fails is logged with the storage id only; a checksum mismatch answers as storage unavailable; two associations use two folders, each with its own token; an association without a connection is refused as not connected, one without a folder as storage unavailable; a resolver that fails answers as storage unavailable |
| AC-28b domain invariant | a member's own sign-in account is refused as the storage account | unit + integration + contract + e2e-through-UI | «Це особистий акаунт члена правління — підключіть акаунт ОСББ»; nothing connected |
| AC-28b | the оператор sees which account holds the files on a later запит | integration | the review answer of a запит carries the storage account address |

### US-09 — how the голова was confirmed

| AC (PRD §5) | Test name (intent-based) | Level | Expected outcome |
|---|---|---|---|
| AC-29 traceability | the association card shows the confirmation chain | integration + contract | the current голова, the extract with its date and signer, the decision date, every later запит and disabling with its reason; the hand-made association is marked as connected by hand |
| AC-29 | the card renders the chain | component + visual-regression (association card) | all parts present in both variants (with an extract, connected by hand) |
| AC-29 | a member opens the card and the extract (Flow US-09) | e2e-through-UI | the card opens from the sidebar; the extract downloads; after a transfer the chain has both decisions and both extracts |
| AC-29 | the card keeps the chain after a roster edit (Flow US-05) | integration + e2e-through-UI | after the previous голова's name is corrected the card names them as ended under the corrected name, once, and the new голова as the голова правління |

### Quality checks behind the criteria (PRD §6)

| Requirement | Test name (intent-based) | Level | Expected outcome |
|---|---|---|---|
| Isolation (release-blocking) | every association-scoped operation of the contract is called with another association's id | integration (two seeded associations, a person on A, a person on both, the оператор on A's roster only) | zero records of B returned to a person off B's roster; a new scoped operation is picked up without editing the test |
| Isolation | a query on a scoped table without an association set reads nothing | integration | empty result, never another association's rows |
| Effect of decisions | the old session after a disabling or a transfer | integration | the first request after the decision is already refused |
| Integrity | the request-serving database role cannot change or delete extracts, checks, filings, decisions and state changes | integration | each attempt is refused by the database |
| Integrity | the downloaded extract matches the stored fingerprint | integration | equal on every download |
| Effect of decisions | one заявка decided twice at once; two filings racing into one open заявка | integration | exactly one decision and one association; the loser is refused, never a server error |
| Effect of decisions | the change is committed before the answer leaves | integration | a failed commit answers a server error with no success body; a refused request leaves nothing |
| Isolation | the API serves only as a role that row-level security binds | integration | start-up is refused for a role that bypasses the policies |
| File size and kind | a request body over the limit is cut off before it is read | integration | refused before validation and before any session check; nothing stored |
| File size and kind | the extract upload boundary | unit + integration + contract | exactly 5 MB accepted; one byte more refused; a file that is neither a PDF nor a signed container refused; a separate signature file refused; nothing stored on refusal |
| File size and kind | a body over the limit sent in chunks, without its length | unit + integration (identity, governance, finance) | 413 in the problem shape from every context; nothing stored |
| Throughput | the load smoke in CI | load | on every PR, after the core job: 50 requests/s of reads kept with no failed request; the job fails on lost throughput or a failed request and prints each p95 beside its PRD §6 limit; without the report-only flag the tool enforces the p95 limits too |
| Error catalogue | every governance problem code has one constant | unit | no governance source outside the registry spells a code as a literal or spells an identity code at all (the guard's two refusal codes come from identity through `cmd/api`); the governance adapter imports nothing from identity |
| Contract fidelity | the contract says what the code does | unit + contract | the delta equals the live contract; the three upload operations advertise the shared 413 and an oversize filing answers it; the sign-in operations describe the sealed cookie; the 422 of `updateBoardMember` names `/term_end`; `decision_effect` is a string enum with no `nil` constant in the generated code |
| Concurrency tests | a failure inside a racing call is reported | integration | the racing calls return their error and the test fails on its own goroutine, in the approval tests and in the storage consent test |

## Where the tests are

The browser specs by criterion (`apps/web/e2e/`); a typed address is used only where there is nothing to click.
No spec types the address of a filing form: the заявка form is opened by its link from `/`, from the list of
заявки or from the sidebar of the association (`openConnectionForm` in `support.ts`), the запит form by the link
on the refusal or on the card (`openTransferFormFromCard` in `transfer-support.ts`).

| Spec | Criteria |
|---|---|
| `us01-file-application.spec.ts` | AC-01 |
| `us02-follow-and-refile.spec.ts` | AC-05, AC-06, AC-08, AC-14 |
| `us03-decide-application.spec.ts` | AC-05 (sign-in by the button), AC-08, AC-09, AC-10, AC-20 (the first голова's note) |
| `us04-file-transfer.spec.ts` | AC-17 (as a person off the roster and as a member), AC-18 |
| `us05-decide-transfer.spec.ts` | AC-18b, AC-19, AC-20 and AC-29 (after a roster edit) |
| `us06-disable-enable.spec.ts` | AC-21, AC-22, AC-23, AC-24 |
| `us07-read-only-association.spec.ts` | AC-25, AC-26, AC-28 |
| `us08-connect-storage.spec.ts` | AC-27, AC-28, AC-28b |
| `us09-association-card.spec.ts` | AC-20, AC-29 |
| `refusals.spec.ts` | AC-02, AC-03, AC-03b, AC-04, AC-04b, AC-04c, AC-11, AC-11b, AC-12, AC-13, AC-15, AC-30 |
| `access.spec.ts` | AC-01 and AC-05 (the second заявка from the sidebar), AC-04d, AC-07, AC-16, AC-16b, AC-16c |

AC-19b, AC-19c, AC-24b and AC-27b have no browser flow of their own, as the rows above say: integration and component.

The other tiers: unit and integration tests sit beside the code in `services/core/internal/identity/`,
`services/core/internal/governance/` (roster, registry, files, справи) and `services/core/internal/platform/`
(isolation, row-level security, the request-serving role); component tests and stories beside the components in
`apps/web/src/features/identity/` and `apps/web/src/features/governance/`; the visual baselines in
`apps/web/src/features/identity/__screenshots__/`; the load scenarios in `tests/load/`, run by the `load` job
of `.github/workflows/ci.yml`.

### Tests of the second follow-up

What tasks T86–T110 added or changed, by criterion. Go tests are functions under `services/core/internal/`,
component tests are `describe` or `it` titles under `apps/web/src/features/`, specs are under `apps/web/e2e/`.

| AC | Go tests | Component tests and specs |
|---|---|---|
| AC-01 | — | `identity/shell/page-states.test.tsx` «the link to the заявка на підключення form (AC-01)»; `us01-file-application.spec.ts` «US-01: a голова files a заявка and finds it pending in their list (AC-01)», now through `openConnectionForm` |
| AC-02 | `identity/applications_file_test.go` `TestFileApplicationRefusesAnOversizeChunkedBody`; `governance/cases_endpoints_test.go` `TestOversizeChunkedBodiesAnswer413`; `finance/module_test.go` `TestCreateContractRefusesAnOversizeChunkedBody`; `platform/httpapi/problem_test.go` `TestStrictRequestError_MaxBytesIs413` | `identity/applications/filing-form.test.tsx` «refuses an extract over 5 MB under the field, sending no refused file (AC-02)», «refuses an extract of another type under the field, sending no refused file (AC-02)» (renamed by T118), «shows a 413 answer as the same extract field message (AC-02)» |
| AC-04c | `identity/review_queue_competing_test.go` `TestQueue_AConnectionAndATransferOfOneCodeDoNotMarkEachOther`, `TestQueue_TwoApplicationsStillMarkEachOtherBesideATransfer` | `identity/operator/operator-section.test.tsx` «every page of the оператор's queue (AC-04c)» |
| AC-04d | — | `identity/shell/shell.test.tsx` «AssociationPicker: filing for another association (AC-04d)»; `identity/shell/association-shell.test.tsx` «a person on more than one page of memberships (AC-04d)»; `access.spec.ts` «AC-04d: two windows of one person hold two associations and keep them after reloads», now two pages of one browser context |
| AC-05 | `identity/signin_test.go` `TestSignInStateCookieFlags`, `TestSignInCallbackNeedsTheCookieOfItsState`; `identity/state_bounded_test.go` `TestSignInFloodLeavesAnEarlierSignInAbleToFinish` | — |
| AC-08, AC-15, AC-19c, AC-30 | no new test: `raceAll` in `identity/decision_race_test.go` reports the racing calls' errors on the test goroutine, and `approvalTermEnd` in `identity/approval_test.go` takes the term end from the tests' clock | `identity/operator/decision-panel.test.tsx` «announces an approval as a status (AC-08)», now with the focus on the status |
| AC-10, AC-21 | — | `identity/shell/page-states.test.tsx`: the pages «the оператор's review page», «the оператор's competing item» and «the оператор's associations list» in the table of pages; `identity/operator/associations.test.tsx` «every page of the associations list (AC-21)» |
| AC-16c | — | `access.spec.ts` «AC-16c: the оператор decides nothing they filed themselves», moved from `refusals.spec.ts` |
| AC-19 | `governance/roster_endpoints_test.go` `TestListBoardMembersSaysWhatTheDecisionDid`, `TestUpdateBoardMemberRefusesToReopenAnEndedEntry`, `TestReAddedPreviousHeadIsAMemberEntry` | `governance/cases.test.tsx` «the справи screen reads through the layout cache and the shared fallback (AC-19, AC-28, T103)»; `governance/registry/registry.test.tsx` «the registry screens read through the layout cache and the shared fallback (AC-19, AC-26, T103)» |
| AC-19b | `identity/approval_transfer_test.go` `TestApprovalTransfer_AnEditOfThePreviousHeadDoesNotTurnTheNextTransferAround`; `identity/review_queue_competing_test.go` `TestQueue_TwoTransfersStillMarkEachOtherBesideAnApplication` | — |
| AC-20, AC-29 | `identity/card_test.go` `TestCard_AnEditOfThePreviousHeadKeepsItAsTheEndedEntry` | `governance/roster/roster-screen.test.tsx` «RosterScreen: the mark follows the decision effect and kind (AC-20, AC-29)», «RosterScreen: every page of the roster (AC-04d, AC-20)»; `identity/association/roster-decision-mark.test.tsx` «words the mark of a connection by the заявка, not by a запит на передачу», «words the mark of a transfer by the запит на передачу, for both effects»; `us03-decide-application.spec.ts` «US-03: the approved голова adds a член правління, who then opens the association (AC-09, AC-20)»; `us05-decide-transfer.spec.ts` «US-05: a fix of the previous голова's name leaves the marks of the transfer on the same entries (AC-20, AC-29)» |
| AC-25 | `governance/download_contract_test.go` `TestDownloadFileContractAdvertisesBothForbiddenCodes`; `identity/middleware_outside_io_test.go` `TestGuardOutsideIOHoldsNoConnection` | — |
| AC-26 | `governance/file_download_actor_test.go` `TestDownloadFileRefusesARequestWithoutAnActor`, `TestGovernanceAdaptersDoNotDependOnIdentity`; `governance/file_download_test.go` `TestDownloadFileReadsTheStorageOutsideTheTransaction`; `governance/files_per_association_test.go` `TestDevFilesKeepOneStoragePerAssociation`, `TestDownloadFileResolvesTheStoragePerAssociation`; `governance/drive_wiring_test.go` `TestDocumentsLiveInTheDriveFolderOfTheAssociationsConnection`, `TestAnAssociationWithoutAConnectionOrAFolderGetsNoDrive`; `platform/config/google_test.go` `TestLoadGoogle_T97_DriveAddresses` | `governance/registry/registry.test.tsx`, the same `describe` as AC-19 |
| AC-27 | `identity/storage_test.go` `TestStorageCodeExchangeHoldsNoConnection`; `identity/state_bounded_test.go` `TestStorageConsentFloodLeavesAnotherMembersConsentAbleToFinish`, `TestStorageConsentHoldsOnePendingEntryPerAccount` | — |
| AC-28 | `governance/cases_endpoints_test.go` `TestAddCaseDocumentUploadsBeforeTheWriteTransaction`, `TestAddCaseDocumentRemovesTheUploadedFileWhenTheWriteFails`, `TestAddCaseDocumentLogsAFailedRemovalWithTheStorageIDOnly`, `TestAddCaseDocumentAnswersStorageUnavailableOnAChecksumMismatch`; the `Files` contract of `governance/adapters/filestest/contract.go`, subtest «remove deletes a file, is idempotent and the removed file is not found», run for the fake and for the Drive adapter; the two tests of `governance/drive_wiring_test.go` named under AC-26 | `governance/cases.test.tsx`, the same `describe` as AC-19 |
| Error catalogue (Q08, Q09) | `governance/download_contract_test.go` `TestGovernanceHasNoProblemCodeLiteralsOutsideTheRegistry`; `governance/domain/errors_test.go` `TestErrorRegistry_SearchInvalidHasAConstant` | `identity/reuse.test.tsx` «the screens under src/features follow the reuse conventions (W5)» |
| Throughput (PRD §6) | `tests/load/` through `mise run load` | the `load` job of `.github/workflows/ci.yml` |

No visual baseline changed: the three baselined stories (the filing form, the review page, the association
card) render as before, and the new notes, links and «Показати ще» are in stories without a baseline (T109).

### Tests of the third follow-up

What tasks T112–T134 added or changed, by criterion, in the same notation. Every name below was grepped in the
tree. The test of the second follow-up that was replaced is `TestGuardCodesMirrorTheIdentityModule`
(`governance/guard_codes_mirror_test.go`, gone with T129).

| AC | Go tests | Component tests and specs |
|---|---|---|
| AC-01, AC-05, AC-06, AC-14 | — | `identity/shell/association-shell.test.tsx` «links «Мої заявки» and «Подати заявку на підключення» in the sidebar for a person on %s (AC-01, AC-05, AC-06, AC-14)» (one roster, several rosters); `identity/shell/shell.test.tsx` «AssociationPicker: the lasting link to the list of заявки (AC-05, AC-14)»; `access.spec.ts` «AC-04d: two windows of one person hold two associations and keep them after reloads», which now files the second заявка from the sidebar and follows it from «Мої заявки» |
| AC-02 | `platform/contract_test.go` `TestFileOperationsAdvertisePayloadTooLarge`; `identity/applications_413_contract_test.go` `TestOversizeFilingAnswersTheAdvertised413` | `identity/applications/filing-form.test.tsx` «names a refused extract together with every other broken field after one submit (AC-02)», «names only the extract and records nothing when a refused extract sits next to valid fields (AC-02)»; the two refusal tests of the second follow-up were renamed «… sending no refused file (AC-02)» (see above) |
| AC-04c, AC-21 | — | `identity/shell/page-states.test.tsx` the page «the оператор's queue» in the table of pages; `identity/operator/review-page.test.tsx` «keeps the list markup valid while a competing заявка loads: only <li> children, the skeleton inside one» |
| AC-04d, AC-19 | — | `identity/shell/association-shell.test.tsx` «stays in the twenty-first association on the protocol page (T119, R06)», «ends in the shared fallback after a bounded number of requests when the server repeats a cursor (T119, S12)»; `identity/storage/storage-line.test.tsx` «StorageCallbackLost without an association id (T119, R06)»; `identity/reuse.test.tsx` «reads memberships only through useAllMemberships (T119, R06)» |
| AC-05 | `identity/signin_test.go` `TestSignInCookiePathFollowsTheRedirectAddress`; `platform/contract_test.go` `TestSignInOperationsDescribeThePendingCookie`; `identity/state_bounded_test.go` `TestBogusSignInCallbacksLeaveServerMemoryBounded`, `TestBogusCallbackFloodKeepsEarlierSignInAndReplayRefusal` | the AC-04d spec above, for «Мої заявки» from the sidebar |
| AC-16c | — | `access.spec.ts` «AC-16c: the оператор decides nothing they filed themselves», which no longer depends on the оператор's rosters and passes twice on one database |
| AC-17 | — | `identity/association/association-card.test.tsx` «offers the запит link on the card of an association connected by a заявка (AC-17)», «shows the by-hand badge only on the hand-made association»; `us04-file-transfer.spec.ts` «US-04: a newly elected голова who is already on the roster files a запит from the card (AC-17)»; the baselines `association-card-chromium-linux-arm64.png` and `-linux-x64.png` |
| AC-19, AC-20 | `platform/contract_test.go` `TestUpdateBoardMemberNamesTheTermEndRefusal`, `TestDecisionEffectIsAStringEnumByRef`, `TestGeneratedGovernanceCodeHasNoNilConstant`, `TestDeltaEqualsTheLiveContract` | — |
| AC-19b, AC-20 | `identity/approval_transfer_test.go` `TestApprovalTransfer_TheRosterReportsEachEntrysDecisionEffectAfterTwoTransfers`, `TestApprovalTransfer_AReseatedApplicantHasOneEntryReportedSeated` | — |
| AC-25 | `identity/middleware_outside_io_spec_test.go` `TestEveryAssociationMultipartOperationIsOutsideIO` | `governance/protocol-screen.test.tsx` the `describe` «the protocol screen shows the new draft after a write, without a reload» (`a saved slot is no longer dirty`, `a transition changes the state and the buttons`); `governance/cases.test.tsx` «a справа that is not found (AC-25)» |
| AC-25, AC-26 | `governance/download_contract_test.go` `TestGovernanceSourcesSpellNoIdentityCode`; `cmd/api/drive_consent_test.go` and `identity/middleware_test.go` take the two codes from identity | — |
| AC-26, AC-28 | `governance/drive_resolver_test.go` `TestTwoRequestsOfOneAssociationAskTheTokenEndpointOnce`, `TestAReconnectMakesTheNextRequestUseTheNewConnection`, `TestADriveThatNeverAnswersEndsTheRequestWith503WithinTheDeadline`, `TestAnUnreadableTokenAnswers503OnUploadAndOnDownload`; `governance/adapters/filesfake/fake_test.go` `TestFake_IDsDifferAcrossInstances`; `governance/files_per_association_test.go` `TestDevFilesGiveFirstUploadsOfTwoAssociationsDistinctIDs` | `governance/registry/registry.test.tsx` «a protocol that is not found (AC-26)»; `governance/cases.test.tsx` the same `describe` as AC-25; `us07-read-only-association.spec.ts` «AC-26: a member of a вимкнене ОСББ opens the реєстр and downloads the signed copy, byte for byte», which passes a second time on one database |
| AC-26, AC-27, AC-28, AC-28b | `cmd/api/drive_consent_test.go` `TestConsentBuiltByWireGoogleCompletesAndDocumentsGoThroughTheDriveResolver` | — |
| AC-27 | `identity/storage_test.go` `TestStorageCodeExchangeHoldsNoConnection`, now reporting on the test goroutine | — |
| AC-27b | `identity/storage_test.go` `TestStorageConnectionRefusesAnotherAccountWhileFilesExist`, `TestStorageConnectionAcceptsTheSameAccountOrAnotherWithoutFiles`; `identity/domain/storage_has_files_test.go` `TestStorageHasFilesCodeIsRegisteredOnceAsAConflict`, `TestStorageHasFilesCodeIsInThePRDCatalogueAndBothContracts` | `identity/storage/storage-line.test.tsx` «tells why another account was refused while the files exist: names the connected account, says the files stay, links back to the card (AC-27b)» |
| Throughput (PRD §6) | `tests/load/stats_test.go` `TestJudgeReportOnlyDoesNotFailOnP95`, `TestJudgeReportOnlyStillFailsOnThroughputAndErrors`, `TestJudgeEnforcingFailsOnP95` | the `load` job of `.github/workflows/ci.yml` |

One visual baseline changed: the association card's, which shows the запит link on every card (T117 for
linux-arm64, T134 for linux-x64, regenerated in the pinned Playwright image). The filing form and the review
page hold on both platforms after T118 and T121.

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
- Signed protocol for the browser flow of AC-26: nothing in the app can lock a protocol yet, so the suite writes one over the owner connection of its database. The files it names are uploaded through the case screen first, so the download serves real bytes. The seed refuses any database that is not marked as the e2e one, because a seeded protocol cannot be deleted. A справа is created through the API: no screen creates one yet.
- Clock: an injected date in Europe/Kyiv for every term, filing-date and extract-age case. A test that files on the real clock takes its term end from that clock, never from a fixed calendar date. The browser suite compares dates as the Europe/Kyiv day, whatever the time zone of the runner.
- Integration dependency: a fresh migrated ephemeral database per test, connected as the request-serving role so row-level security is exercised; never a mocked store. Sign-in and storage providers are in-process fakes behind their ports, with one smoke test per provider against a test account before release.
- Cleanup boundary: per test for integration (its own database); per suite for e2e-through-UI (each flow files under its own code so flows do not collide; the database is torn down after the run). A run needs a fresh database: the specs look for a заявка on the first page of the оператор's queue (oldest first) without asking for more, so заявки left pending by an earlier run push new ones off it. AC-16c and AC-26 no longer need it and pass a second time on one database (T127: unique development file ids; T133: the AC-16c spec reaches the form by the lasting link, whatever the оператор's rosters).
- Visual baselines: stored with the repo, produced in one fixed environment, updated only in a commit that says so.

## NFR validation (load)

The load tool already in the repo, or a load-testing tool appropriate for the stack. Seed: 20 associations, 200 applications with extracts.

- Read p95 ≤ 200 ms and throughput ≥ 50 req/s per instance → sustain 50 requests/s for 5 min over a mix of the application page, the оператор's queue and the association card; assert p95 ≤ 200 ms and no errors.
- Write p95 ≤ 500 ms (without a file) → 5 requests/s of approvals, rejections, disablings and enablings for 5 min over pre-seeded pending applications; assert p95 ≤ 500 ms; refusals of an already decided заявка are expected and not counted as errors.
- Filing p95 ≤ 5 s for a 5 MB file → 20 sequential filings with the 5 MB fixture; assert p95 ≤ 5 s.
- In CI the same scenarios run as a smoke after the core job: 30 s of reads, 30 s of writes and 5 filings with the limits above as thresholds. On a shared runner this proves the throughput and the absence of failed requests: the CI job fails only on those and prints each p95 beside its limit (T128), and `mise run load` without the report-only flag enforces the limits before a release.
- Not load tests, verified elsewhere: isolation, effect of decisions and integrity (integration above), availability (uptime check), RPO/RTO (quarterly restore drill), retention (storage policy and the privilege test).

## CI placement

- On every PR: unit, integration (the isolation check included — it blocks the release), contract, component, visual-regression, e2e-through-UI (all with fakes; the repo already runs its browser suite on PRs), and the short load smoke.
- Pre-release / on schedule: the load scenarios at their full length and the smoke tests against the real sign-in and storage providers. Neither provider has run yet: the sign-in adapter is not built and the storage adapter is proven on a fake provider only.
- Order of arrival follows PRD §9: the isolation check and the roster positions first, then filing and review, then transfers, then disabling and the storage connection.
