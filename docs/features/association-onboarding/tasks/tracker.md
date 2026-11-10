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
| [T19](./T19-association-and-operator-guards.md) | Guard association paths with the session, the roster and a scoped transaction, and operator paths with the оператор flag | wiring | Andrii Opryshko | L | T04, T11, T14, T18 | todo |
| [T20](./T20-promote-row-level-security.md) | Promote row-level security for the governance, finance and identity tables | migration | Andrii Opryshko | M | T12, T13, T17, T19 | todo |
| [T21](./T21-isolation-release-check.md) | Add the release-blocking isolation test over every association-scoped operation of the contract | tests | Andrii Opryshko | M | T02, T20 | todo |
| [T22](./T22-file-connection-application.md) | File a заявка на підключення and keep the signed extract unchanged | ports | Andrii Opryshko | L | T05, T08, T12, T18 | todo |
| [T23](./T23-follow-and-refile-application.md) | Let the applicant follow and refile their application | ports | Andrii Opryshko | M | T09, T22 | todo |
| [T24](./T24-operator-queue-review-check.md) | Serve the оператор's queue and review page, and record the check of the extract | ports | Andrii Opryshko | L | T10, T19, T22 | todo |
| [T25](./T25-reject-application.md) | Reject an application with a reason | ports | Andrii Opryshko | S | T09, T24 | todo |
| [T26](./T26-approve-connection-application.md) | Approve a заявка на підключення: create the association and seat its first голова in one transaction | ports | Andrii Opryshko | L | T09, T15, T20, T24 | todo |
| [T27](./T27-file-transfer-request.md) | File a запит на передачу for a connected association | ports | Andrii Opryshko | M | T14, T22 | todo |
| [T28](./T28-review-transfer-request.md) | Show competing запити and the storage account in the review of a запит | ports | Andrii Opryshko | S | T24, T27 | todo |
| [T29](./T29-approve-transfer-request.md) | Approve a запит на передачу: seat the new голова and end the previous term in one transaction | ports | Andrii Opryshko | L | T26, T28 | todo |
| [T30](./T30-disable-enable-association.md) | Disable and enable an association as the оператор | ports | Andrii Opryshko | M | T11, T19 | todo |
| [T31](./T31-disabled-association-read-only.md) | Make a disabled association read-only in the middleware | wiring | Andrii Opryshko | M | T19, T30 | todo |
| [T32](./T32-memberships-and-association-card.md) | Serve my associations and the association card with the confirmation chain | ports | Andrii Opryshko | M | T19, T26 | todo |
| [T33](./T33-connect-storage-in-app.md) | Connect the association's document storage from the web app | ports | Andrii Opryshko | L | T14, T19, T20 | todo |
| [T34](./T34-refuse-documents-without-storage.md) | Refuse case documents and assistant turns while the storage is not connected | ports | Andrii Opryshko | S | T02, T03, T19 | todo |
| [T35](./T35-ui-shell-sign-in-picker.md) | Build the signed-in shell: sign-in page, association picker, routes under /o/$associationId and the оператор's entry | ui | Andrii Opryshko | L | T02 | done |
| [T36](./T36-ui-filing-forms.md) | Build the filing forms for a заявка на підключення and a запит на передачу | ui | Andrii Opryshko | L | T35 | todo |
| [T37](./T37-ui-my-applications.md) | Show my applications: status, the оператор's contact, the reason and refiling | ui | Andrii Opryshko | M | T36 | todo |
| [T38](./T38-ui-operator-queue-review.md) | Build the оператор's queue and the review page with sources, warnings and the check form | ui | Andrii Opryshko | L | T36 | todo |
| [T39](./T39-ui-operator-decisions.md) | Add the оператор's decisions: approve, reject and their refusals | ui | Andrii Opryshko | M | T38 | todo |
| [T40](./T40-ui-operator-associations.md) | Build the оператор's associations list with disabling and enabling | ui | Andrii Opryshko | M | T35 | todo |
| [T41](./T41-ui-association-card.md) | Build the association card with the confirmation chain and the roster marks | ui | Andrii Opryshko | M | T35 | todo |
| [T42](./T42-ui-storage-and-suspended-notice.md) | Show the storage connection, the connect flow and the «Доступ призупинено» notice | ui | Andrii Opryshko | M | T41 | todo |
| [T43](./T43-visual-baselines.md) | Add visual-regression checks for the filing form, the review page and the association card | tests | Andrii Opryshko | S | T36, T38, T41 | todo |
| [T44](./T44-e2e-connection-flows.md) | Prove filing, following and deciding a заявка end to end through the UI (US-01, US-02, US-03) | tests | Andrii Opryshko | L | T23, T25, T26, T32, T37, T39 | todo |
| [T45](./T45-e2e-transfer-and-card.md) | Prove a transfer, the takeover and the association card end to end (US-04, US-05, US-09) | tests | Andrii Opryshko | M | T29, T41, T44 | todo |
| [T46](./T46-e2e-disable-storage-access.md) | Prove disabling, the read-only association, the storage connection and every access denial end to end (US-06, US-07, US-08) | tests | Andrii Opryshko | L | T21, T31, T33, T34, T40, T42, T44 | todo |
| [T47](./T47-load-scenarios.md) | Add the load scenarios for the latency and throughput targets of PRD §6 | tests | Andrii Opryshko | M | T26, T30, T32 | todo |
| [T48](./T48-docs-rules-and-amendments.md) | Update the repo rules and environment docs, amend the board-protocol documents and write the оператор's runbook | docs | Andrii Opryshko | M | T20, T33 | todo |
| [T49](./T49-security-review-record.md) | Run and record the security review that PRD §6.1 requires | docs | Andrii Opryshko | M | T45, T46 | todo |

**Total:** 49 tasks, ~29.5 person-days.
