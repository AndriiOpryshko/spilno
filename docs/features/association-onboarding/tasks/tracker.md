# Tracker — association-onboarding

> Status of every task in the epic. `implement-tasks` updates `done` as it commits each task.
> States: `todo` · `in_progress` · `blocked` · `review` · `done`. Estimates: S = 2 h, M = half a day, L = a day.

| # | Task | Layer | Owner | Estimate | Blocked by | Status |
|---|---|---|---|---|---|---|
| [T01](./T01-move-paths-under-association.md) | Move the governance and finance operations under `/<context>/associations/{association_id}/` | wiring | Andrii Opryshko | L | — | done |
| [T02](./T02-merge-contract-delta.md) | Merge the onboarding contract delta and stub the new identity operations | wiring | Andrii Opryshko | L | T01 | done |
| [T03](./T03-identity-problem-codes.md) | Register the identity problem codes of PRD §5.1 and governance.drive_not_connected | infra | Andrii Opryshko | S | — | done |
| [T04](./T04-in-association-transaction.md) | Add platform/db.InAssociation: a transaction that carries the association and the actor | infra | Andrii Opryshko | M | — | done |
| [T05](./T05-promote-expand-migrations.md) | Promote the expand migrations: application tables, association and roster columns, the app role | migration | Andrii Opryshko | M | — | done |
| [T06](./T06-promote-drop-accounts-association.md) | Promote the drop of accounts.association_id ahead of the manifest order | migration | Andrii Opryshko | S | T05 | done |
| [T07](./T07-roster-position-domain.md) | Give roster entries a position and a decision reference, with the date rules of a transfer | domain | Andrii Opryshko | M | — | done |
| [T08](./T08-filing-validation-domain.md) | Validate a filing: every field at once, the term rules and the extract file | domain | Andrii Opryshko | M | T03 | done |
| [T09](./T09-application-lifecycle-domain.md) | Model the application lifecycle and the guards of a decision | domain | Andrii Opryshko | M | T03 | done |
| [T10](./T10-name-match-extract-age.md) | Compare the three names after levelling and flag an extract older than 30 days | domain | Andrii Opryshko | S | — | done |
| [T11](./T11-association-state-access-matrix.md) | Model the association state and the access matrix of PRD §10 | domain | Andrii Opryshko | M | T03 | done |
| [T12](./T12-two-database-roles.md) | Run goose as the owner and serve as spilno_app; give tests an app-role pool with two associations | infra | Andrii Opryshko | M | T05 | done |
| [T13](./T13-finance-contracts-association.md) | Scope finance contracts to the association in the path | app | Andrii Opryshko | S | T01, T04, T05 | done |
| [T14](./T14-roster-read-facade.md) | Serve the roster facade: the current member, a person's memberships and the members' sign-in addresses | infra | Andrii Opryshko | L | T04, T05, T07 | done |
| [T15](./T15-seat-head-end-term.md) | Seat a голова and end a term through the governance facade inside the caller's transaction | app | Andrii Opryshko | M | T14 | done |
| [T16](./T16-promote-backfills.md) | Promote the backfills: association state, roster positions and the contracts' association | migration | Andrii Opryshko | S | T06, T13, T14 | done |
| [T17](./T17-promote-not-null-contracts.md) | Promote the NOT NULL steps and the contracts index | migration | Andrii Opryshko | S | T16 | done |
| [T18](./T18-sign-in-any-verified-person.md) | Sign in any verified person, keep the session in Postgres and tell the оператор | app | Andrii Opryshko | L | T02, T03, T06 | done |
| [T19](./T19-association-and-operator-guards.md) | Guard association paths with the session, the roster and a scoped transaction, and operator paths with the оператор flag | wiring | Andrii Opryshko | L | T04, T11, T14, T18 | done |
| [T20](./T20-promote-row-level-security.md) | Promote row-level security for the governance, finance and identity tables | migration | Andrii Opryshko | M | T12, T13, T17, T19 | done |
| [T21](./T21-isolation-release-check.md) | Add the release-blocking isolation test over every association-scoped operation of the contract | tests | Andrii Opryshko | M | T02, T20 | done |
| [T22](./T22-file-connection-application.md) | File a заявка на підключення and keep the signed extract unchanged | ports | Andrii Opryshko | L | T05, T08, T12, T18 | done |
| [T23](./T23-follow-and-refile-application.md) | Let the applicant follow and refile their application | ports | Andrii Opryshko | M | T09, T22 | done |
| [T24](./T24-operator-queue-review-check.md) | Serve the оператор's queue and review page, and record the check of the extract | ports | Andrii Opryshko | L | T10, T19, T22 | done |
| [T25](./T25-reject-application.md) | Reject an application with a reason | ports | Andrii Opryshko | S | T09, T24 | done |
| [T26](./T26-approve-connection-application.md) | Approve a заявка на підключення: create the association and seat its first голова in one transaction | ports | Andrii Opryshko | L | T09, T15, T20, T24 | done |
| [T27](./T27-file-transfer-request.md) | File a запит на передачу for a connected association | ports | Andrii Opryshko | M | T14, T22 | done |
| [T28](./T28-review-transfer-request.md) | Show competing запити and the storage account in the review of a запит | ports | Andrii Opryshko | S | T24, T27 | done |
| [T29](./T29-approve-transfer-request.md) | Approve a запит на передачу: seat the new голова and end the previous term in one transaction | ports | Andrii Opryshko | L | T26, T28 | done |
| [T30](./T30-disable-enable-association.md) | Disable and enable an association as the оператор | ports | Andrii Opryshko | M | T11, T19 | done |
| [T31](./T31-disabled-association-read-only.md) | Make a disabled association read-only in the middleware | wiring | Andrii Opryshko | M | T19, T30 | done |
| [T32](./T32-memberships-and-association-card.md) | Serve my associations and the association card with the confirmation chain | ports | Andrii Opryshko | M | T19, T26 | done |
| [T33](./T33-connect-storage-in-app.md) | Connect the association's document storage from the web app | ports | Andrii Opryshko | L | T14, T19, T20 | done |
| [T34](./T34-refuse-documents-without-storage.md) | Refuse case documents and assistant turns while the storage is not connected | ports | Andrii Opryshko | S | T02, T03, T19 | done |
| [T35](./T35-ui-shell-sign-in-picker.md) | Build the signed-in shell: sign-in page, association picker, routes under /o/$associationId and the оператор's entry | ui | Andrii Opryshko | L | T02 | done |
| [T36](./T36-ui-filing-forms.md) | Build the filing forms for a заявка на підключення and a запит на передачу | ui | Andrii Opryshko | L | T35 | done |
| [T37](./T37-ui-my-applications.md) | Show my applications: status, the оператор's contact, the reason and refiling | ui | Andrii Opryshko | M | T36 | done |
| [T38](./T38-ui-operator-queue-review.md) | Build the оператор's queue and the review page with sources, warnings and the check form | ui | Andrii Opryshko | L | T36 | done |
| [T39](./T39-ui-operator-decisions.md) | Add the оператор's decisions: approve, reject and their refusals | ui | Andrii Opryshko | M | T38 | done |
| [T40](./T40-ui-operator-associations.md) | Build the оператор's associations list with disabling and enabling | ui | Andrii Opryshko | M | T35 | done |
| [T41](./T41-ui-association-card.md) | Build the association card with the confirmation chain and the roster marks | ui | Andrii Opryshko | M | T35 | done |
| [T42](./T42-ui-storage-and-suspended-notice.md) | Show the storage connection, the connect flow and the «Доступ призупинено» notice | ui | Andrii Opryshko | M | T41 | done |
| [T43](./T43-visual-baselines.md) | Add visual-regression checks for the filing form, the review page and the association card | tests | Andrii Opryshko | S | T36, T38, T41 | done |
| [T44](./T44-e2e-connection-flows.md) | Prove filing, following and deciding a заявка end to end through the UI (US-01, US-02, US-03) | tests | Andrii Opryshko | L | T23, T25, T26, T32, T37, T39 | done |
| [T45](./T45-e2e-transfer-and-card.md) | Prove a transfer, the takeover and the association card end to end (US-04, US-05, US-09) | tests | Andrii Opryshko | M | T29, T41, T44 | done |
| [T46](./T46-e2e-disable-storage-access.md) | Prove disabling, the read-only association, the storage connection and every access denial end to end (US-06, US-07, US-08) | tests | Andrii Opryshko | L | T21, T31, T33, T34, T40, T42, T44 | done |
| [T47](./T47-load-scenarios.md) | Add the load scenarios for the latency and throughput targets of PRD §6 | tests | Andrii Opryshko | M | T26, T30, T32 | done |
| [T48](./T48-docs-rules-and-amendments.md) | Update the repo rules and environment docs, amend the board-protocol documents and write the оператор's runbook | docs | Andrii Opryshko | M | T20, T33 | done |
| [T49](./T49-security-review-record.md) | Run and record the security review that PRD §6.1 requires | docs | Andrii Opryshko | M | T45, T46 | done |
| [T50](./T50-kyiv-calendar-day.md) | Take «today» for every roster date from the Europe/Kyiv calendar | infra | Andrii Opryshko | M | — | done |
| [T51](./T51-carry-term-end-end-previous-head.md) | Carry the filed term end into the roster and end the previous голова whatever their term end | app | Andrii Opryshko | M | T50 | done |
| [T52](./T52-refuse-stale-filing-and-foreign-code.md) | Refuse a decision on a stale filing and a card transfer that claims another association's code | ports | Andrii Opryshko | M | — | done |
| [T53](./T53-limit-request-bodies.md) | Limit request bodies and read time before the request validator runs | infra | Andrii Opryshko | M | — | done |
| [T54](./T54-commit-before-response.md) | Commit the association transaction before the response leaves, and keep outside calls out of it | wiring | Andrii Opryshko | L | T50 | done |
| [T55](./T55-bound-sign-in-and-consent-state.md) | Bound the sign-in and storage-consent state | app | Andrii Opryshko | M | T54 | done |
| [T56](./T56-extract-download-header-and-type-check.md) | Send a valid Content-Disposition on the review download and tighten the extract type check | ports | Andrii Opryshko | S | — | done |
| [T57](./T57-refuse-rls-bypassing-role.md) | Refuse to start when the serving role bypasses row-level security | wiring | Andrii Opryshko | S | — | done |
| [T58](./T58-storage-check-injection-and-pending-gauge.md) | Pass the storage check into governance from cmd/api and count the pending age from the latest filing | wiring | Andrii Opryshko | S | T57 | done |
| [T59](./T59-index-lower-sign-in-email.md) | Make the memberships lookup use an index on the lowered sign-in address | migration | Andrii Opryshko | S | — | done |
| [T60](./T60-contract-disabled-and-refile-codes.md) | Advertise the two missing problem codes in the contract and reconcile the sync report | wiring | Andrii Opryshko | M | — | done |
| [T61](./T61-record-single-release-migrations.md) | Record that the expand, backfill and contract migrations shipped in one release | docs | Andrii Opryshko | S | T59 | done |
| [T62](./T62-serve-roster-endpoints.md) | Serve the roster: list, add and change члени правління with their position and decision | ports | Andrii Opryshko | L | T51, T58, T59 | done |
| [T63](./T63-serve-registry-read.md) | Serve the реєстр протоколів: search and the signed protocol | ports | Andrii Opryshko | L | T62 | done |
| [T64](./T64-serve-file-download-read-only-check.md) | Serve file downloads, and only signed copies while the association is read-only | ports | Andrii Opryshko | M | T58, T63 | done |
| [T65](./T65-serve-cases-and-document-upload.md) | Serve справи and document upload behind the storage guard | ports | Andrii Opryshko | L | T64 | done |
| [T66](./T66-ui-sign-in-link-generated-urls.md) | Point «Увійти» at the API and build every API link from the generated URL helpers | ui | Andrii Opryshko | S | — | done |
| [T67](./T67-ui-operator-layout-entry-guard.md) | Give the оператор's section an index, its own navigation and one access guard | ui | Andrii Opryshko | M | T66 | done |
| [T68](./T68-ui-association-shell-card-notice.md) | Put the association card in the navigation and the suspended notice in the association layout | ui | Andrii Opryshko | M | T67 | done |
| [T69](./T69-ui-review-latest-filing-extract-download.md) | Review the latest filing with its latest check, and let the оператор download the extract | ui | Andrii Opryshko | M | T67 | done |
| [T70](./T70-ui-review-source-labels-competing.md) | Label the source of every value on the review page and show competing заявки side by side | ui | Andrii Opryshko | M | T60, T69 | done |
| [T71](./T71-ui-ukrainian-warnings-field-errors.md) | Write the warnings and the field errors in Ukrainian from structured data | ui | Andrii Opryshko | M | T70 | done |
| [T72](./T72-card-ended-mark-only-when-a-term-ended.md) | Show «повноваження завершено» in the card history only when the decision ended a term | ports | Andrii Opryshko | M | T51, T60 | done |
| [T73](./T73-ui-operator-contact-address.md) | Show the оператор's contact address in every environment | ui | Andrii Opryshko | S | — | done |
| [T74](./T74-ui-shared-problem-fallback.md) | Give every page a loading state and one fallback for a failed request | ui | Andrii Opryshko | M | T68 | done |
| [T75](./T75-ui-mutation-errors-and-focus.md) | Report every failed operator action and keep focus after a decision | ui | Andrii Opryshko | M | T67 | done |
| [T76](./T76-ui-one-date-formatter.md) | Format every date with one Europe/Kyiv formatter | ui | Andrii Opryshko | S | — | done |
| [T77](./T77-ui-reuse-inline-alert-hooks-links.md) | Reuse InlineAlert, the generated hooks and router links across the identity screens | ui | Andrii Opryshko | M | T71, T74, T75 | done |
| [T78](./T78-ui-roster-screen.md) | Build the roster screen: list, add and change члени правління, with the decision mark on each голова | ui | Andrii Opryshko | L | T62, T68 | done |
| [T79](./T79-ui-registry-screens.md) | Build the реєстр протоколів: search and the signed protocol with its download | ui | Andrii Opryshko | M | T63, T64, T68 | done |
| [T80](./T80-ui-cases-on-api-with-upload.md) | Move the справи list and the case screen to the API, with document upload and the storage message | ui | Andrii Opryshko | L | T65, T68 | done |
| [T81](./T81-e2e-review-fixes-through-clicks.md) | Prove the fixed flows end to end by clicking, not by typing addresses | tests | Andrii Opryshko | M | T52, T68, T71, T72, T73 | done |
| [T82](./T82-e2e-roster-registry-documents.md) | Prove the roster, the registry and the document refusal end to end | tests | Andrii Opryshko | L | T78, T79, T80 | done |
| [T83](./T83-concurrency-and-failure-tests.md) | Test the decisions and the filings under concurrency | tests | Andrii Opryshko | M | T52, T54 | done |
| [T84](./T84-refresh-visual-baselines.md) | Refresh the visual baselines of the review page and the association card | tests | Andrii Opryshko | S | T71, T72, T77 | done |
| [T85](./T85-sync-test-plan-tracker-board-protocol.md) | Bring the test plan, the trackers and the board-protocol tasks in line with what was built | docs | Andrii Opryshko | M | T61, T81, T82, T83, T84 | done |
| [T86](./T86-decision-entries-by-first-revision.md) | Find the голова a decision seated and the entry it ended by the first revision that carries the decision | infra | Andrii Opryshko | M | — | done |
| [T87](./T87-roster-decision-effect-and-no-reopen.md) | Tell each roster entry what its decision did to it, and refuse to reopen an entry an approval ended | ports | Andrii Opryshko | M | — | done |
| [T88](./T88-ui-roster-marks-by-effect-and-kind.md) | Mark roster entries by the decision's effect and word the mark by the kind of the decision | ui | Andrii Opryshko | M | T87 | done |
| [T89](./T89-ui-entry-to-connection-form.md) | Lead a signed-in person to the заявка на підключення form | ui | Andrii Opryshko | S | — | done |
| [T90](./T90-one-definition-per-problem-code.md) | Advertise the disabled refusal of a file download and define every problem code once | ports | Andrii Opryshko | M | — | done |
| [T91](./T91-governance-actor-lookup-injected.md) | Pass the read-only lookup into governance from cmd/api and refuse a request without an actor | wiring | Andrii Opryshko | S | T90 | done |
| [T92](./T92-files-per-association.md) | Resolve the file storage per association | ports | Andrii Opryshko | M | — | done |
| [T93](./T93-guard-check-then-handler-for-outside-io.md) | Run operations with outside I/O after the guard's check transaction, not inside it | wiring | Andrii Opryshko | L | — | done |
| [T94](./T94-upload-outside-transaction.md) | Upload a case document before the write transaction and remove the file when the write fails | app | Andrii Opryshko | M | T92, T93 | done |
| [T95](./T95-download-outside-transaction.md) | Read a download from the storage outside the transaction and answer 503 when the read fails | app | Andrii Opryshko | M | T91, T92, T93 | done |
| [T96](./T96-code-exchange-outside-transaction.md) | Exchange the storage consent code with Google outside the transaction | app | Andrii Opryshko | M | T93 | done |
| [T97](./T97-wire-drive-storage-per-association.md) | Build each association's Drive storage from its Google connection in cmd/api | wiring | Andrii Opryshko | L | T94, T95 | done |
| [T98](./T98-oversize-body-413-everywhere.md) | Answer 413 for every oversized body and limit the bodies of finance and system | infra | Andrii Opryshko | M | — | done |
| [T99](./T99-sign-in-state-in-sealed-cookie.md) | Keep the pending sign-in in a sealed cookie of the browser that started it | app | Andrii Opryshko | L | — | done |
| [T100](./T100-competing-within-one-kind.md) | Mark as competing only pending applications of the same kind for the same code | infra | Andrii Opryshko | S | — | done |
| [T101](./T101-race-tests-and-term-dates-that-last.md) | Report race-test failures on the test goroutine and stop the term-end tests from expiring | tests | Andrii Opryshko | S | T86 | done |
| [T102](./T102-ui-operator-fallback-and-focus.md) | Show the shared fallback on the оператор's review and associations pages and keep focus after an approval | ui | Andrii Opryshko | M | — | done |
| [T103](./T103-ui-governance-generated-hooks-and-fallback.md) | Use the generated hooks, their query keys and the shared fallback in the governance screens | ui | Andrii Opryshko | M | — | done |
| [T104](./T104-ui-page-roster-and-memberships.md) | Read every page of the roster and of a person's memberships | ui | Andrii Opryshko | M | T88 | done |
| [T105](./T105-ui-page-operator-lists.md) | Read every page of the оператор's queue and list of associations | ui | Andrii Opryshko | M | T102 | done |
| [T106](./T106-ui-extract-file-check-before-upload.md) | Check the extract's size and type before the upload and name the field on a 413 | ui | Andrii Opryshko | S | — | done |
| [T107](./T107-load-smoke-in-ci.md) | Run the load smoke of PRD §6 in CI | tests | Andrii Opryshko | M | T99 | done |
| [T108](./T108-e2e-second-review-fixes.md) | Prove the second review's fixes end to end | tests | Andrii Opryshko | M | T88, T89, T99 | done |
| [T109](./T109-refresh-visual-baselines-second-review.md) | Refresh the visual baselines the second review's fixes changed | tests | Andrii Opryshko | S | T88, T89, T102, T104, T105, T106 | done |
| [T110](./T110-contract-texts-and-reconcile.md) | Correct the contract's descriptions and reconcile the sync report | wiring | Andrii Opryshko | M | T87, T90, T100 | done |
| [T111](./T111-sync-docs-after-second-review.md) | Bring the test plan, the trackers and the review record in line with the second follow-up | docs | Andrii Opryshko | M | T96, T97, T98, T101, T103, T107, T108, T109, T110 | done |
| [T112](./T112-wire-consent-drive-address-and-test-api-wiring.md) | Give the production storage consent its Drive address and test the wiring cmd/api runs | wiring | Andrii Opryshko | M | — | done |
| [T113](./T113-drive-resolver-one-read-cached-token-deadline.md) | Resolve an association's Drive storage with one read, a reused token and a deadline | app | Andrii Opryshko | M | T112 | todo |
| [T114](./T114-refuse-another-storage-account-while-files-exist.md) | Refuse to connect another storage account while the association has files | ports | Andrii Opryshko | L | — | todo |
| [T115](./T115-ui-storage-account-refusal.md) | Tell the член правління why another storage account was refused | ui | Andrii Opryshko | S | T114 | todo |
| [T116](./T116-ui-lasting-links-to-applications.md) | Give every signed-in person a lasting way to their заявки and to the заявка на підключення form | ui | Andrii Opryshko | M | — | done |
| [T117](./T117-ui-transfer-request-link-for-every-member.md) | Offer the запит на передачу on the association card to every member | ui | Andrii Opryshko | S | — | todo |
| [T118](./T118-ui-extract-refusal-with-every-other-field.md) | Name a refused extract together with every other broken field | ui | Andrii Opryshko | S | — | todo |
| [T119](./T119-ui-one-reader-for-memberships.md) | Read memberships through one hook everywhere and stop the page loop on a repeated cursor | ui | Andrii Opryshko | M | — | todo |
| [T120](./T120-ui-protocol-screen-one-cache-key.md) | Read and write the protocol draft under one cache key and tighten the reuse test | ui | Andrii Opryshko | S | T119 | todo |
| [T121](./T121-ui-operator-loading-states-and-way-back.md) | Show a loading state in the оператор's queue and a way back from a missing protocol or справа | ui | Andrii Opryshko | S | — | todo |
| [T122](./T122-bound-sign-in-replay-set.md) | Bound the set of used sign-in states | app | Andrii Opryshko | S | — | todo |
| [T123](./T123-sign-in-cookie-path-from-callback-address.md) | Scope the pending sign-in cookie by the path of the callback address | app | Andrii Opryshko | S | T122 | todo |
| [T124](./T124-multipart-operations-outside-guard-transaction.md) | Run every multipart operation outside the guard's transaction and fail the build when one is missed | wiring | Andrii Opryshko | M | — | todo |
| [T125](./T125-storage-race-test-reports-on-test-goroutine.md) | Report the storage race test's failures on the test goroutine | tests | Andrii Opryshko | S | — | todo |
| [T126](./T126-decision-effect-after-two-approvals.md) | Prove each roster entry's decision effect after two transfers in a row | tests | Andrii Opryshko | S | — | todo |
| [T127](./T127-dev-storage-unique-file-ids.md) | Give the development storage file ids that are unique across associations | infra | Andrii Opryshko | S | — | todo |
| [T128](./T128-load-job-gates-on-throughput-and-errors.md) | Fail the CI load job on throughput and errors and report its latency | tests | Andrii Opryshko | S | — | todo |
| [T129](./T129-guard-refusal-codes-from-identity.md) | Take the guard's refusal codes from identity through cmd/api and drop the mirror | wiring | Andrii Opryshko | M | T112 | todo |
| [T130](./T130-contract-cookie-413-term-end-and-decision-effect.md) | Bring the contract in line with the sign-in cookie, the 413 answers, the term-end refusal and the decision effect | wiring | Andrii Opryshko | M | T114 | todo |
| [T131](./T131-reconcile-sync-report-and-design-docs.md) | Reconcile the sync report and write the roster rules into ADR-0006, the SAD and the data model | docs | Andrii Opryshko | M | T124, T129, T130 | todo |
| [T132](./T132-record-open-gaps-in-prd.md) | Record every gap left open in PRD §8 with an owner and a due date | docs | Andrii Opryshko | S | — | todo |
| [T133](./T133-e2e-third-review-fixes.md) | Prove the third review's fixes end to end by clicking | tests | Andrii Opryshko | M | T116, T117, T127 | todo |
| [T134](./T134-refresh-visual-baselines-third-review.md) | Refresh the visual baseline of the association card | tests | Andrii Opryshko | S | T116, T117, T118, T121 | todo |
| [T135](./T135-sync-docs-after-third-review.md) | Bring the test plan, the trackers and the review record in line with the third follow-up | docs | Andrii Opryshko | M | T113, T115, T119, T120, T123, T125, T126, T128, T131, T132, T133, T134 | todo |

**Total:** 135 tasks, ~71 person-days: T01–T49 (~29.5, done, merged in PR #48), the review follow-up T50–T85 (~19.25, done, PRs #52–#55, [review of 2026-11-10](../_review/review-2026-11-10.md) with the commit that closed each finding), the second follow-up T86–T111 (~13, done, PR #56, [review of 2026-11-17](../_review/review-2026-11-17.md) with the commit that closed each finding) and the third follow-up T112–T135 (~9.25, todo, [second re-review of 2026-11-17](../_review/review-2026-11-17-2.md)).

## Known gaps (found during implementation)

- **No adapter for the real sign-in provider.** T18 built the provider port, an in-process fake and the
  development shim (`SPILNO_DEV_SIGN_IN=true`); without the shim the four sign-in operations answer 501.
  The Google OIDC adapter (ID-token verification, client configuration from the environment) is a
  separate piece of work after this feature (decided 2026-11-10).
- **Sign-in state is sealed with a per-process key** (`identity/app/signin.go`: state, PKCE verifier, nonce and
  expiry travel in a cookie of the browser that started the sign-in; since T99 the server keeps nothing). One API
  instance is enough; more than one needs a shared key. The storage consent keeps one server-side entry per account.
- **Security review findings are open** ([record](../_review/security-2026-11-10.md)): it counts no High, six
  Medium (SEC-01 to SEC-06, due 2026-11-24) and eight Low, and has not been run again after either follow-up.
  T53 answers SEC-03 (body limits), T57 SEC-04 (a role that bypasses row-level security) and T99 SEC-02 and
  SEC-12 (the sign-in `state` bound to the browser, no map of pending sign-ins). SEC-01 with the first gap means
  the feature must not be deployed where anyone but the developer can reach it until the real provider adapter
  exists.
- **The e2e suite needs a fresh database per run.** The оператор's queue and list of associations page with
  «Показати ще» since T105, but the specs look for a заявка on the first page of the queue (20 rows, oldest
  first) and do not press it, and the AC-16c spec files as the оператор from `/`, which offers the form only
  while the оператор is on no roster or on several (the AC-16b spec seats them on one). The two runs on one
  database that T105 asked for were made on 2026-11-17: the first passes, the second fails AC-16c for that
  reason and AC-26 (`us07-read-only-association.spec.ts:63`, the link of a just attached document did not
  appear; cause not looked into).
- **A person on exactly one roster has no link to the заявка на підключення form.** `/` opens their
  association and no screen inside it leads to the form; the list of заявки and the picker carry the link
  (T89). A член правління who connects a second association (AC-04d) has to type the address. Found by T108.
- **No screen creates a справа.** The справи list and the case screen read the API (T80); a справа is created
  only through `createCase` until board-protocol builds the draft flow. The e2e specs create it through the API.
- **The Drive storage has not run against real Google.** T97 builds each association's storage from its Google
  connection and T112 gives the consent its Drive address; both are proven on a fake Google through HTTP
  (`cmd/api/drive_consent_test.go` runs the consent `wireGoogle` builds and the resolver `driveFiles` builds;
  `drive_wiring_test.go` calls the same `governance.NewDriveResolver`). What remains is a run with a test Google
  account: board-protocol T13.
- **An uploaded file whose record failed is removed in the request, best effort** (T94). A removal that fails
  is logged with the storage id and the file stays in the storage until board-protocol builds the
  `remove_orphan_file` job (T14, T32).
- **`identity.association_disabled` and `identity.not_current_member` are defined twice:** in the identity
  registry and as constants of `governance/domain/errors.go`, which may not import identity (T91).
  `TestGuardCodesMirrorTheIdentityModule` keeps the two definitions equal.
- **The load smoke in CI proves the throughput and the error rate, not production latency** (T107): 30 s of
  reads and writes and 5 filings on a shared runner. The full run of PRD §6 stays `mise run load` before a release.
- **Board-protocol work still open behind this feature's ACs:** the association settings (board-protocol T10,
  T35), the needs-review event and the upload metric of document versions (T21), and the lock that writes a
  signed protocol (T29; the реєстр is tested on seeded protocols). Each task file there says what this feature
  delivered.
- **The correction note of a signed protocol** is one section under the agenda, not beside the corrected item
  (board-protocol AC-15, noted in board-protocol T43).
- **SAD §6 Flow 5 has no branch for 409 `identity.application_already_open`** on a refile; the contract lists
  it since T60 and [api-sync-report.md](../contracts/api-sync-report.md) carries the deferral.
