# Epic — association-onboarding

> **PRD:** [PRD.md](../PRD.md) · **Design:** [sad.md](../sad.md) · **Data model:** [data-model.md](../data-model.md) · **API:** [openapi.yaml](../contracts/openapi.yaml) · [api-sync-report](../contracts/api-sync-report.md) · **ADRs:** [adr/](../adr/) · **Tests:** [test-plan.md](../test-plan.md) · **Migrations:** [MANIFEST](../migrations/MANIFEST.md)

## Goal

A голова правління connects their association to Спільно, or takes it over after an election, without the developer: they sign in, file a заявка на підключення or a запит на передачу with a КЕП-signed витяг з ЄДР, and the оператор records the check made outside Спільно and decides with a reason (PRD §2). Every association and every change of голова then carries a checkable chain from the extract to the decision, and with a second association nothing of one association is readable by a person off its roster — enforced by the database and by a test that blocks a release.

## Scope

- **In:** surfaces `[backend-service, web-frontend]` (sad frontmatter); contexts `identity` (applications, the оператор's queue, association state, storage connection, sign-in and the guards), `governance` (positions on the roster, the roster facade), `finance` (the association on contracts), `platform` (`db.InAssociation`, two database roles, test fixtures); the 20 staged migrations; the contract delta and the path move of ADR-0002 in `services/core/api/openapi.yaml`; the SPA shell, the applicant's pages, the `/operator/` section, the association card and the storage connection.
- **Out:** everything in PRD §3 (no automatic check of the КЕП or the extract, no periodic re-check of ЄДР, no notifications, one оператор, no co-owners' accounts); editing the association's name and address by the board (named in PRD §9, no AC — api-sync-report §D); a River job (none in this feature); the board-protocol screens and handlers themselves (see «Board-protocol tasks» below), apart from the roster, the registry read, the file download and справи, which the review follow-up brings in (see «Review follow-up»).

## Task map

Same DAG as [tasks.json](./tasks.json). Colours = layer. Day-one parallel starts: T01, T03, T04, T05, T07, T10. Longest chain: 10 phases.

```mermaid
flowchart LR
    T01["T01 wiring: move-paths-under-association"]:::wir
    T02["T02 wiring: merge-contract-delta"]:::wir
    T03["T03 infra: identity-problem-codes"]:::inf
    T04["T04 infra: in-association-transaction"]:::inf
    T05["T05 migration: promote-expand-migrations"]:::mig
    T06["T06 migration: promote-drop-accounts-association"]:::mig
    T07["T07 domain: roster-position-domain"]:::dom
    T08["T08 domain: filing-validation-domain"]:::dom
    T09["T09 domain: application-lifecycle-domain"]:::dom
    T10["T10 domain: name-match-extract-age"]:::dom
    T11["T11 domain: association-state-access-matrix"]:::dom
    T12["T12 infra: two-database-roles"]:::inf
    T13["T13 app: finance-contracts-association"]:::app
    T14["T14 infra: roster-read-facade"]:::inf
    T15["T15 app: seat-head-end-term"]:::app
    T16["T16 migration: promote-backfills"]:::mig
    T17["T17 migration: promote-not-null-contracts"]:::mig
    T18["T18 app: sign-in-any-verified-person"]:::app
    T19["T19 wiring: association-and-operator-guards"]:::wir
    T20["T20 migration: promote-row-level-security"]:::mig
    T21["T21 tests: isolation-release-check"]:::tst
    T22["T22 ports: file-connection-application"]:::prt
    T23["T23 ports: follow-and-refile-application"]:::prt
    T24["T24 ports: operator-queue-review-check"]:::prt
    T25["T25 ports: reject-application"]:::prt
    T26["T26 ports: approve-connection-application"]:::prt
    T27["T27 ports: file-transfer-request"]:::prt
    T28["T28 ports: review-transfer-request"]:::prt
    T29["T29 ports: approve-transfer-request"]:::prt
    T30["T30 ports: disable-enable-association"]:::prt
    T31["T31 wiring: disabled-association-read-only"]:::wir
    T32["T32 ports: memberships-and-association-card"]:::prt
    T33["T33 ports: connect-storage-in-app"]:::prt
    T34["T34 ports: refuse-documents-without-storage"]:::prt
    T35["T35 ui: ui-shell-sign-in-picker"]:::ui
    T36["T36 ui: ui-filing-forms"]:::ui
    T37["T37 ui: ui-my-applications"]:::ui
    T38["T38 ui: ui-operator-queue-review"]:::ui
    T39["T39 ui: ui-operator-decisions"]:::ui
    T40["T40 ui: ui-operator-associations"]:::ui
    T41["T41 ui: ui-association-card"]:::ui
    T42["T42 ui: ui-storage-and-suspended-notice"]:::ui
    T43["T43 tests: visual-baselines"]:::tst
    T44["T44 tests: e2e-connection-flows"]:::tst
    T45["T45 tests: e2e-transfer-and-card"]:::tst
    T46["T46 tests: e2e-disable-storage-access"]:::tst
    T47["T47 tests: load-scenarios"]:::tst
    T48["T48 docs: docs-rules-and-amendments"]:::doc
    T49["T49 docs: security-review-record"]:::doc
    T01 --> T02
    T05 --> T06
    T03 --> T08
    T03 --> T09
    T03 --> T11
    T05 --> T12
    T01 --> T13
    T04 --> T13
    T05 --> T13
    T04 --> T14
    T05 --> T14
    T07 --> T14
    T14 --> T15
    T06 --> T16
    T13 --> T16
    T14 --> T16
    T16 --> T17
    T02 --> T18
    T03 --> T18
    T06 --> T18
    T04 --> T19
    T11 --> T19
    T14 --> T19
    T18 --> T19
    T12 --> T20
    T13 --> T20
    T17 --> T20
    T19 --> T20
    T02 --> T21
    T20 --> T21
    T05 --> T22
    T08 --> T22
    T12 --> T22
    T18 --> T22
    T09 --> T23
    T22 --> T23
    T10 --> T24
    T19 --> T24
    T22 --> T24
    T09 --> T25
    T24 --> T25
    T09 --> T26
    T15 --> T26
    T20 --> T26
    T24 --> T26
    T14 --> T27
    T22 --> T27
    T24 --> T28
    T27 --> T28
    T26 --> T29
    T28 --> T29
    T11 --> T30
    T19 --> T30
    T19 --> T31
    T30 --> T31
    T19 --> T32
    T26 --> T32
    T14 --> T33
    T19 --> T33
    T20 --> T33
    T02 --> T34
    T03 --> T34
    T19 --> T34
    T02 --> T35
    T35 --> T36
    T36 --> T37
    T36 --> T38
    T38 --> T39
    T35 --> T40
    T35 --> T41
    T41 --> T42
    T36 --> T43
    T38 --> T43
    T41 --> T43
    T23 --> T44
    T25 --> T44
    T26 --> T44
    T32 --> T44
    T37 --> T44
    T39 --> T44
    T29 --> T45
    T41 --> T45
    T44 --> T45
    T21 --> T46
    T31 --> T46
    T33 --> T46
    T34 --> T46
    T40 --> T46
    T42 --> T46
    T44 --> T46
    T26 --> T47
    T30 --> T47
    T32 --> T47
    T20 --> T48
    T33 --> T48
    T45 --> T49
    T46 --> T49
    classDef mig fill:#fde68a,stroke:#92400e,color:#111
    classDef dom fill:#bbf7d0,stroke:#166534,color:#111
    classDef inf fill:#bfdbfe,stroke:#1e40af,color:#111
    classDef app fill:#ddd6fe,stroke:#5b21b6,color:#111
    classDef prt fill:#fbcfe8,stroke:#9d174d,color:#111
    classDef ui fill:#fed7aa,stroke:#9a3412,color:#111
    classDef wir fill:#e5e7eb,stroke:#374151,color:#111
    classDef tst fill:#fecaca,stroke:#991b1b,color:#111
    classDef doc fill:#f5f5f4,stroke:#57534e,color:#111
```

**Serialized lanes** (`implement-tasks` derives them from `layer` and overlapping `files_hint`): migrations T05 → T06 → T16 → T17 → T20; the live contract T01 → T02; the identity adapters T18 → T19 → T22 … T33 (shared `adapters/postgres` and its generated `sqlcgen`); the e2e specs T44 → T45 → T46. The domain tasks (T07–T11), the platform tasks (T04, T12) and the ui tasks after T35 (T36 / T40 / T41) run next to them.

## Rollout order (PRD §9)

| PRD §9 step | Tasks | Ships |
|---|---|---|
| 1. Isolation guard, positions, backfill | T01–T07, T12–T21 | association in every path, sign-in and the guards, row-level security, the release-blocking check; members keep working as before |
| 2. Заявки на підключення and the оператор's review | T08–T11, T22–T26, T32, T35–T39, T44 | a new association connects without the developer |
| 3. Запити на передачу | T27–T29, T41, T45 | a new голова takes over; the hand-made association can get its extract (AC-19c) |
| 4. Disable / enable, the storage connected by the board | T30, T31, T33, T34, T40, T42, T46 | replaces the developer's storage command |
| Across | T43, T47, T48, T49 | visual baselines, load scenarios, docs and amendments, the security review |

## Tasks

See [tracker.md](./tracker.md) for status. Total 49 tasks, ~29.5 person-days (S = 2 h, M = half a day, L = a day).

| # | Task | Layer | Blocked by | ACs | Est. |
|---|---|---|---|---|---|
| [T01](./T01-move-paths-under-association.md) | Move the governance and finance operations under `/<context>/associations/{association_id}/` | wiring | — | AC-04d, AC-07 | L |
| [T02](./T02-merge-contract-delta.md) | Merge the onboarding contract delta and stub the new identity operations | wiring | T01 | AC-01, AC-10, AC-17, AC-21, AC-27, AC-29 | L |
| [T03](./T03-identity-problem-codes.md) | Register the identity problem codes of PRD §5.1 and governance.drive_not_connected | infra | — | AC-02, AC-03, AC-03b, AC-04, AC-04b, AC-12, AC-13, AC-15, AC-16, AC-16c, AC-18, AC-22, AC-24, AC-25, AC-28, AC-28b, AC-30 | S |
| [T04](./T04-in-association-transaction.md) | Add platform/db.InAssociation: a transaction that carries the association and the actor | infra | — | AC-04d, AC-07 | M |
| [T05](./T05-promote-expand-migrations.md) | Promote the expand migrations: application tables, association and roster columns, the app role | migration | — | AC-01, AC-06, AC-08, AC-21, AC-29 | M |
| [T06](./T06-promote-drop-accounts-association.md) | Promote the drop of accounts.association_id ahead of the manifest order | migration | T05 | AC-04d, AC-05 | S |
| [T07](./T07-roster-position-domain.md) | Give roster entries a position and a decision reference, with the date rules of a transfer | domain | — | AC-08, AC-19, AC-19b, AC-19c | M |
| [T08](./T08-filing-validation-domain.md) | Validate a filing: every field at once, the term rules and the extract file | domain | T03 | AC-01, AC-02, AC-30 | M |
| [T09](./T09-application-lifecycle-domain.md) | Model the application lifecycle and the guards of a decision | domain | T03 | AC-03, AC-03b, AC-06, AC-12, AC-13, AC-15, AC-16c, AC-18b | M |
| [T10](./T10-name-match-extract-age.md) | Compare the three names after levelling and flag an extract older than 30 days | domain | — | AC-11, AC-11b | S |
| [T11](./T11-association-state-access-matrix.md) | Model the association state and the access matrix of PRD §10 | domain | T03 | AC-07, AC-16, AC-16b, AC-21, AC-22, AC-23, AC-24, AC-25, AC-26 | M |
| [T12](./T12-two-database-roles.md) | Run goose as the owner and serve as spilno_app; give tests an app-role pool with two associations | infra | T05 | AC-06, AC-15 | M |
| [T13](./T13-finance-contracts-association.md) | Scope finance contracts to the association in the path | app | T01, T04, T05 | AC-04d, AC-07 | S |
| [T14](./T14-roster-read-facade.md) | Serve the roster facade: the current member, a person's memberships and the members' sign-in addresses | infra | T04, T05, T07 | AC-04d, AC-07, AC-09, AC-28b | L |
| [T15](./T15-seat-head-end-term.md) | Seat a голова and end a term through the governance facade inside the caller's transaction | app | T14 | AC-08, AC-19, AC-19c, AC-20 | M |
| [T16](./T16-promote-backfills.md) | Promote the backfills: association state, roster positions and the contracts' association | migration | T06, T13, T14 | AC-19c, AC-29 | S |
| [T17](./T17-promote-not-null-contracts.md) | Promote the NOT NULL steps and the contracts index | migration | T16 | AC-08, AC-21 | S |
| [T18](./T18-sign-in-any-verified-person.md) | Sign in any verified person, keep the session in Postgres and tell the оператор | app | T02, T03, T06 | AC-01, AC-05, AC-16 | L |
| [T19](./T19-association-and-operator-guards.md) | Guard association paths with the session, the roster and a scoped transaction, and operator paths with the оператор flag | wiring | T04, T11, T14, T18 | AC-04d, AC-07, AC-16, AC-16b | L |
| [T20](./T20-promote-row-level-security.md) | Promote row-level security for the governance, finance and identity tables | migration | T12, T13, T17, T19 | AC-04d, AC-07, AC-16b | M |
| [T21](./T21-isolation-release-check.md) | Add the release-blocking isolation test over every association-scoped operation of the contract | tests | T02, T20 | AC-04d, AC-07, AC-16b | M |
| [T22](./T22-file-connection-application.md) | File a заявка на підключення and keep the signed extract unchanged | ports | T05, T08, T12, T18 | AC-01, AC-02, AC-04, AC-04b, AC-04c, AC-30 | L |
| [T23](./T23-follow-and-refile-application.md) | Let the applicant follow and refile their application | ports | T09, T22 | AC-05, AC-06, AC-07, AC-14 | M |
| [T24](./T24-operator-queue-review-check.md) | Serve the оператор's queue and review page, and record the check of the extract | ports | T10, T19, T22 | AC-04c, AC-10, AC-11, AC-11b, AC-16 | L |
| [T25](./T25-reject-application.md) | Reject an application with a reason | ports | T09, T24 | AC-13, AC-14, AC-15, AC-16c | S |
| [T26](./T26-approve-connection-application.md) | Approve a заявка на підключення: create the association and seat its first голова in one transaction | ports | T09, T15, T20, T24 | AC-03, AC-03b, AC-04c, AC-08, AC-09, AC-12, AC-15, AC-16c | L |
| [T27](./T27-file-transfer-request.md) | File a запит на передачу for a connected association | ports | T14, T22 | AC-02, AC-17, AC-18, AC-19c, AC-24b, AC-30 | M |
| [T28](./T28-review-transfer-request.md) | Show competing запити and the storage account in the review of a запит | ports | T24, T27 | AC-19b, AC-28b | S |
| [T29](./T29-approve-transfer-request.md) | Approve a запит на передачу: seat the new голова and end the previous term in one transaction | ports | T26, T28 | AC-16c, AC-18b, AC-19, AC-19b, AC-19c, AC-20, AC-24b | L |
| [T30](./T30-disable-enable-association.md) | Disable and enable an association as the оператор | ports | T11, T19 | AC-16, AC-21, AC-22, AC-23, AC-24 | M |
| [T31](./T31-disabled-association-read-only.md) | Make a disabled association read-only in the middleware | wiring | T19, T30 | AC-21, AC-23, AC-25, AC-26 | M |
| [T32](./T32-memberships-and-association-card.md) | Serve my associations and the association card with the confirmation chain | ports | T19, T26 | AC-04d, AC-19c, AC-20, AC-29 | M |
| [T33](./T33-connect-storage-in-app.md) | Connect the association's document storage from the web app | ports | T14, T19, T20 | AC-27, AC-28b | L |
| [T34](./T34-refuse-documents-without-storage.md) | Refuse case documents and assistant turns while the storage is not connected | ports | T02, T03, T19 | AC-28 | S |
| [T35](./T35-ui-shell-sign-in-picker.md) | Build the signed-in shell: sign-in page, association picker, routes under /o/$associationId and the оператор's entry | ui | T02 | AC-04d, AC-07, AC-16, AC-16b, AC-19 | L |
| [T36](./T36-ui-filing-forms.md) | Build the filing forms for a заявка на підключення and a запит на передачу | ui | T35 | AC-01, AC-02, AC-04, AC-04b, AC-17, AC-18, AC-30 | L |
| [T37](./T37-ui-my-applications.md) | Show my applications: status, the оператор's contact, the reason and refiling | ui | T36 | AC-05, AC-06, AC-14 | M |
| [T38](./T38-ui-operator-queue-review.md) | Build the оператор's queue and the review page with sources, warnings and the check form | ui | T36 | AC-04c, AC-10, AC-11, AC-11b, AC-16, AC-19b, AC-28b | L |
| [T39](./T39-ui-operator-decisions.md) | Add the оператор's decisions: approve, reject and their refusals | ui | T38 | AC-03, AC-03b, AC-08, AC-12, AC-13, AC-14, AC-15, AC-16c, AC-18b | M |
| [T40](./T40-ui-operator-associations.md) | Build the оператор's associations list with disabling and enabling | ui | T35 | AC-21, AC-22, AC-23, AC-24 | M |
| [T41](./T41-ui-association-card.md) | Build the association card with the confirmation chain and the roster marks | ui | T35 | AC-19c, AC-20, AC-29 | M |
| [T42](./T42-ui-storage-and-suspended-notice.md) | Show the storage connection, the connect flow and the «Доступ призупинено» notice | ui | T41 | AC-09, AC-21, AC-25, AC-27, AC-28, AC-28b | M |
| [T43](./T43-visual-baselines.md) | Add visual-regression checks for the filing form, the review page and the association card | tests | T36, T38, T41 | AC-01, AC-10, AC-29 | S |
| [T44](./T44-e2e-connection-flows.md) | Prove filing, following and deciding a заявка end to end through the UI (US-01, US-02, US-03) | tests | T23, T25, T26, T32, T37, T39 | AC-01, AC-02, AC-03, AC-03b, AC-04, AC-04b, AC-04c, AC-05, AC-06, AC-08, AC-09, AC-10, AC-11, AC-11b, AC-12, AC-13, AC-14, AC-15, AC-16c, AC-30 | L |
| [T45](./T45-e2e-transfer-and-card.md) | Prove a transfer, the takeover and the association card end to end (US-04, US-05, US-09) | tests | T29, T41, T44 | AC-17, AC-18, AC-18b, AC-19, AC-20, AC-29 | M |
| [T46](./T46-e2e-disable-storage-access.md) | Prove disabling, the read-only association, the storage connection and every access denial end to end (US-06, US-07, US-08) | tests | T21, T31, T33, T34, T40, T42, T44 | AC-04d, AC-07, AC-09, AC-16, AC-16b, AC-21, AC-22, AC-23, AC-24, AC-25, AC-26, AC-27, AC-28, AC-28b | L |
| [T47](./T47-load-scenarios.md) | Add the load scenarios for the latency and throughput targets of PRD §6 | tests | T26, T30, T32 | AC-01, AC-05, AC-08, AC-29 | M |
| [T48](./T48-docs-rules-and-amendments.md) | Update the repo rules and environment docs, amend the board-protocol documents and write the оператор's runbook | docs | T20, T33 | AC-03, AC-05, AC-12 | M |
| [T49](./T49-security-review-record.md) | Run and record the security review that PRD §6.1 requires | docs | T45, T46 | AC-04c, AC-07, AC-16b, AC-16c, AC-28b | M |

**AC coverage:** all 42 acceptance criteria of PRD §5 are in at least one task's `acs` (checked against the AC headings of PRD §5 when these files were written). The NFRs of PRD §6 without an AC are carried by T12 (integrity of the insert-only tables), T20 and T21 (isolation), T29 and T31 (effect of a decision on the next request), T22 (file size, filing latency) and T47 (latency and throughput).

## Review follow-up (2026-11-10)

The [feature review](../_review/review-2026-11-10.md) ran after PR #48 was merged and asked for changes: 22 stage-1 findings (F01–F22) and 15 stage-2 findings (B1–B9, W1–W6), all resolved as «fix now». T50–T85 carry them on a new branch off `main`; T01–T49 stay as built. F22 (no `ui` task claimed AC-19) is closed by this breakdown itself: T35 now lists AC-19.

Same DAG as [tasks.json](./tasks.json), follow-up tasks only. Parallel starts: T50, T52, T53, T56, T57, T59, T60, T66, T73, T76. Longest chain: 9 phases. ~19.25 person-days.

```mermaid
flowchart LR
    T50["T50 infra: kyiv-calendar-day"]:::inf
    T51["T51 app: carry-term-end-end-previous-head"]:::app
    T52["T52 ports: refuse-stale-filing-and-foreign-code"]:::prt
    T53["T53 infra: limit-request-bodies"]:::inf
    T54["T54 wiring: commit-before-response"]:::wir
    T55["T55 app: bound-sign-in-and-consent-state"]:::app
    T56["T56 ports: extract-download-header-and-type-check"]:::prt
    T57["T57 wiring: refuse-rls-bypassing-role"]:::wir
    T58["T58 wiring: storage-check-injection-and-pending-gauge"]:::wir
    T59["T59 migration: index-lower-sign-in-email"]:::mig
    T60["T60 wiring: contract-disabled-and-refile-codes"]:::wir
    T61["T61 docs: record-single-release-migrations"]:::doc
    T62["T62 ports: serve-roster-endpoints"]:::prt
    T63["T63 ports: serve-registry-read"]:::prt
    T64["T64 ports: serve-file-download-read-only-check"]:::prt
    T65["T65 ports: serve-cases-and-document-upload"]:::prt
    T66["T66 ui: ui-sign-in-link-generated-urls"]:::ui
    T67["T67 ui: ui-operator-layout-entry-guard"]:::ui
    T68["T68 ui: ui-association-shell-card-notice"]:::ui
    T69["T69 ui: ui-review-latest-filing-extract-download"]:::ui
    T70["T70 ui: ui-review-source-labels-competing"]:::ui
    T71["T71 ui: ui-ukrainian-warnings-field-errors"]:::ui
    T72["T72 ports: card-ended-mark-only-when-a-term-ended"]:::prt
    T73["T73 ui: ui-operator-contact-address"]:::ui
    T74["T74 ui: ui-shared-problem-fallback"]:::ui
    T75["T75 ui: ui-mutation-errors-and-focus"]:::ui
    T76["T76 ui: ui-one-date-formatter"]:::ui
    T77["T77 ui: ui-reuse-inline-alert-hooks-links"]:::ui
    T78["T78 ui: ui-roster-screen"]:::ui
    T79["T79 ui: ui-registry-screens"]:::ui
    T80["T80 ui: ui-cases-on-api-with-upload"]:::ui
    T81["T81 tests: e2e-review-fixes-through-clicks"]:::tst
    T82["T82 tests: e2e-roster-registry-documents"]:::tst
    T83["T83 tests: concurrency-and-failure-tests"]:::tst
    T84["T84 tests: refresh-visual-baselines"]:::tst
    T85["T85 docs: sync-test-plan-tracker-board-protocol"]:::doc
    T50 --> T51
    T50 --> T54
    T54 --> T55
    T57 --> T58
    T59 --> T61
    T51 --> T62
    T58 --> T62
    T59 --> T62
    T62 --> T63
    T58 --> T64
    T63 --> T64
    T64 --> T65
    T66 --> T67
    T67 --> T68
    T67 --> T69
    T60 --> T70
    T69 --> T70
    T70 --> T71
    T51 --> T72
    T60 --> T72
    T68 --> T74
    T67 --> T75
    T71 --> T77
    T74 --> T77
    T75 --> T77
    T62 --> T78
    T68 --> T78
    T63 --> T79
    T64 --> T79
    T68 --> T79
    T65 --> T80
    T68 --> T80
    T52 --> T81
    T68 --> T81
    T71 --> T81
    T72 --> T81
    T73 --> T81
    T78 --> T82
    T79 --> T82
    T80 --> T82
    T52 --> T83
    T54 --> T83
    T71 --> T84
    T72 --> T84
    T77 --> T84
    T61 --> T85
    T81 --> T85
    T82 --> T85
    T83 --> T85
    T84 --> T85
    classDef mig fill:#fde68a,stroke:#92400e,color:#111
    classDef dom fill:#bbf7d0,stroke:#166534,color:#111
    classDef inf fill:#bfdbfe,stroke:#1e40af,color:#111
    classDef app fill:#ddd6fe,stroke:#5b21b6,color:#111
    classDef prt fill:#fbcfe8,stroke:#9d174d,color:#111
    classDef ui fill:#fed7aa,stroke:#9a3412,color:#111
    classDef wir fill:#e5e7eb,stroke:#374151,color:#111
    classDef tst fill:#fecaca,stroke:#991b1b,color:#111
    classDef doc fill:#f5f5f4,stroke:#57534e,color:#111
```

| Task | Title | Layer | Blocked by | ACs | Findings | Est. |
|---|---|---|---|---|---|---|
| [T50](./T50-kyiv-calendar-day.md) | Take «today» for every roster date from the Europe/Kyiv calendar | infra | — | AC-09, AC-19, AC-29 | F07 | M |
| [T51](./T51-carry-term-end-end-previous-head.md) | Carry the filed term end into the roster and end the previous голова whatever their term end | app | T50 | AC-08, AC-19, AC-19c, AC-30 | F05, F06 | M |
| [T52](./T52-refuse-stale-filing-and-foreign-code.md) | Refuse a decision on a stale filing and a card transfer that claims another association's code | ports | — | AC-15, AC-19c | F01 (backend half), B3 | M |
| [T53](./T53-limit-request-bodies.md) | Limit request bodies and read time before the request validator runs | infra | — | AC-02 | B1 | M |
| [T54](./T54-commit-before-response.md) | Commit the association transaction before the response leaves, and keep outside calls out of it | wiring | T50 | AC-25, AC-27 | B2 | L |
| [T55](./T55-bound-sign-in-and-consent-state.md) | Bound the sign-in and storage-consent state | app | T54 | AC-05, AC-27 | B4 | M |
| [T56](./T56-extract-download-header-and-type-check.md) | Send a valid Content-Disposition on the review download and tighten the extract type check | ports | — | AC-02, AC-03 | B5, B6 | S |
| [T57](./T57-refuse-rls-bypassing-role.md) | Refuse to start when the serving role bypasses row-level security | wiring | — | AC-04d, AC-07 | F16 | S |
| [T58](./T58-storage-check-injection-and-pending-gauge.md) | Pass the storage check into governance from cmd/api and count the pending age from the latest filing | wiring | T57 | AC-05, AC-28 | B7, B8 | S |
| [T59](./T59-index-lower-sign-in-email.md) | Make the memberships lookup use an index on the lowered sign-in address | migration | — | AC-04d | F20 | S |
| [T60](./T60-contract-disabled-and-refile-codes.md) | Advertise the two missing problem codes in the contract and reconcile the sync report | wiring | — | AC-06, AC-25 | F17, F18, F21 | M |
| [T61](./T61-record-single-release-migrations.md) | Record that the expand, backfill and contract migrations shipped in one release | docs | T59 | AC-04d | F19 | S |
| [T62](./T62-serve-roster-endpoints.md) | Serve the roster: list, add and change члени правління with their position and decision | ports | T51, T58, T59 | AC-09, AC-16b, AC-20, AC-25 | F15 | L |
| [T63](./T63-serve-registry-read.md) | Serve the реєстр протоколів: search and the signed protocol | ports | T62 | AC-09, AC-26 | F15 | L |
| [T64](./T64-serve-file-download-read-only-check.md) | Serve file downloads, and only signed copies while the association is read-only | ports | T58, T63 | AC-25, AC-26 | F15 | M |
| [T65](./T65-serve-cases-and-document-upload.md) | Serve справи and document upload behind the storage guard | ports | T64 | AC-25, AC-28 | F15 | L |
| [T66](./T66-ui-sign-in-link-generated-urls.md) | Point «Увійти» at the API and build every API link from the generated URL helpers | ui | — | AC-04d, AC-05 | F02, W5 (hand-built URLs) | S |
| [T67](./T67-ui-operator-layout-entry-guard.md) | Give the оператор's section an index, its own navigation and one access guard | ui | T66 | AC-16, AC-16b, AC-21 | F03, W1 | M |
| [T68](./T68-ui-association-shell-card-notice.md) | Put the association card in the navigation and the suspended notice in the association layout | ui | T67 | AC-04d, AC-21, AC-25, AC-27, AC-29 | F08, F14 | M |
| [T69](./T69-ui-review-latest-filing-extract-download.md) | Review the latest filing with its latest check, and let the оператор download the extract | ui | T67 | AC-03, AC-06, AC-08, AC-15 | F01 (web half), F04 | M |
| [T70](./T70-ui-review-source-labels-competing.md) | Label the source of every value on the review page and show competing заявки side by side | ui | T60, T69 | AC-04c, AC-10, AC-19b | F10, F11 | M |
| [T71](./T71-ui-ukrainian-warnings-field-errors.md) | Write the warnings and the field errors in Ukrainian from structured data | ui | T70 | AC-02, AC-11, AC-11b | F12 | M |
| [T72](./T72-card-ended-mark-only-when-a-term-ended.md) | Show «повноваження завершено» in the card history only when the decision ended a term | ports | T51, T60 | AC-19c, AC-20 | F09 | M |
| [T73](./T73-ui-operator-contact-address.md) | Show the оператор's contact address in every environment | ui | — | AC-05 | F13 | S |
| [T74](./T74-ui-shared-problem-fallback.md) | Give every page a loading state and one fallback for a failed request | ui | T68 | AC-05, AC-07 | W2 | M |
| [T75](./T75-ui-mutation-errors-and-focus.md) | Report every failed operator action and keep focus after a decision | ui | T67 | AC-13, AC-15, AC-22 | W3, W6 | M |
| [T76](./T76-ui-one-date-formatter.md) | Format every date with one Europe/Kyiv formatter | ui | — | AC-05, AC-29 | W4 | S |
| [T77](./T77-ui-reuse-inline-alert-hooks-links.md) | Reuse InlineAlert, the generated hooks and router links across the identity screens | ui | T71, T74, T75 | AC-13, AC-21 | W5 | M |
| [T78](./T78-ui-roster-screen.md) | Build the roster screen: list, add and change члени правління, with the decision mark on each голова | ui | T62, T68 | AC-09, AC-20 | F15 | L |
| [T79](./T79-ui-registry-screens.md) | Build the реєстр протоколів: search and the signed protocol with its download | ui | T63, T64, T68 | AC-09, AC-26 | F15 | M |
| [T80](./T80-ui-cases-on-api-with-upload.md) | Move the справи list and the case screen to the API, with document upload and the storage message | ui | T65, T68 | AC-25, AC-28 | F15, F08 (the mock) | L |
| [T81](./T81-e2e-review-fixes-through-clicks.md) | Prove the fixed flows end to end by clicking, not by typing addresses | tests | T52, T68, T71, T72, T73 | AC-04c, AC-05, AC-06, AC-08, AC-11, AC-21, AC-29 | F01, F02, F03, F04, F08, F11, F12 (e2e halves) | M |
| [T82](./T82-e2e-roster-registry-documents.md) | Prove the roster, the registry and the document refusal end to end | tests | T78, T79, T80 | AC-09, AC-16b, AC-20, AC-26, AC-28 | F15 | L |
| [T83](./T83-concurrency-and-failure-tests.md) | Test the decisions and the filings under concurrency | tests | T52, T54 | AC-04c, AC-15 | B9 | M |
| [T84](./T84-refresh-visual-baselines.md) | Refresh the visual baselines of the review page and the association card | tests | T71, T72, T77 | AC-10, AC-29 | follow-up of F09, F10, F11, F12 | S |
| [T85](./T85-sync-test-plan-tracker-board-protocol.md) | Bring the test plan, the trackers and the board-protocol tasks in line with what was built | docs | T61, T81, T82, T83, T84 | AC-09, AC-20, AC-26, AC-28 | F15 (bookkeeping), F21 | M |

**Finding coverage:** every finding of the review record is named by at least one task above, F22 excepted (closed here). F15 is carried by T62–T65 (handlers), T78–T80 (screens) and T82 (e2e).

### Board-protocol work carried here (F15)

The product owner chose to deliver AC-09, AC-20, AC-26, AC-28 and the on-roster part of AC-16b in this feature. That takes work out of six board-protocol tasks; each is built once, here, and T85 updates the board-protocol tracker.

| board-protocol task | Carried here by | Left in board-protocol |
|---|---|---|
| [T10](../../board-protocol/tasks/T10-settings-roster-endpoints.md) settings and roster endpoints | T62: the three board-member operations | the settings operations, `governance.register_start_locked` |
| [T31](../../board-protocol/tasks/T31-registry-search.md) registry search | T63: whole | nothing |
| [T21](../../board-protocol/tasks/T21-cases-endpoints.md) cases and files | T64: `downloadFile` with the ADR-0007 check; T65: the four case operations | the domain event for the needs-review hook, the upload-duration metric |
| [T35](../../board-protocol/tasks/T35-ui-settings-roster.md) settings and roster screens | T78: the roster screen | the settings screen |
| [T43](../../board-protocol/tasks/T43-ui-registry.md) registry screens | T79: whole | nothing |
| [T36](../../board-protocol/tasks/T36-ui-cases-documents.md) cases and documents screens | T80: whole | nothing |

Not carried: the lock with a signed copy (board-protocol T29) and everything before it. Until it exists no protocol can be signed in the app, so the registry is proven on seeded signed protocols (T63, T82), and the Drive adapter (board-protocol T13) stays there: files are served through the in-memory `Files` fake under `SPILNO_DEV_STORAGE=true` and answer 503 without it (T64).

### Risks of the follow-up

- **T54 changes the path of every association-scoped request** (commit before the response). It is the one L task with a wide blast radius; the review did not verify B2, so the task starts by reproducing it.
- **T51 fixes two defects that hide each other.** Storing the filed term end without the change to `GetEntryOfDecision` would stop transfers from ending the previous голова.
- **Three findings rest on the reviewer's read alone** (B3 in T52, F18 in T60, F20 in T59): each task starts with a failing test or an `EXPLAIN`, and records in the review record if the finding does not reproduce.
- **T65 and T80 are the slice to drop first.** They exist only for the UI half of AC-28; without them that half needs an explicit de-scope in PRD §8.
- **Contract lane.** T60, T70 and T72 edit `services/core/api/openapi.yaml`; their shared `files_hint` serializes them, and each lands with its generated code.

## Board-protocol tasks this epic replaces or needs

`tasks.json` holds only this epic's ids, so these are not `deps`; they are named in the Notes of the task on the right.

| board-protocol task | Status there | Relation | Here |
|---|---|---|---|
| [T08](../../board-protocol/tasks/T08-identity-oidc-sessions.md) sign-in and sessions | todo | **replaced**: any verified person signs in (ADR-0004); `bootstrap-member` is not built | T18 |
| [T11](../../board-protocol/tasks/T11-require-session-current-member.md) session and current member on governance and finance routes | todo | **superseded**: the association is in the path and the check runs inside the scoped transaction | T19, T31 |
| [T12](../../board-protocol/tasks/T12-google-connection-token-source.md) / [T12d](../../board-protocol/tasks/T12d-google-config.md) Google connection | in progress | `cmd/google-connect` **replaced** by the in-app consent (ADR-0008); the cipher, the store and the consent exchange (T12a–c) are reused; T12d's configuration is **needed** | T33 |
| [T34](../../board-protocol/tasks/T34-ui-sign-in-session.md) sign-in page and session guard | todo | **superseded**: the principal changed and the shell gains the picker | T35 |
| [T10](../../board-protocol/tasks/T10-settings-roster-endpoints.md) settings and roster endpoints | todo | **shared**: the roster repository and `CurrentMember`; its handlers must return `position` and `decision_id`; AC-09's «add a член правління» needs it | T14, T44 |
| [T21](../../board-protocol/tasks/T21-cases-endpoints.md), [T25](../../board-protocol/tasks/T25-assistant-propose-draft.md) case documents, assistant turns | todo | keep the storage guard as their first step; `downloadFile` checks the signed snapshot when the actor is read-only | T34, T31 |
| [T29](../../board-protocol/tasks/T29-lock-with-signed-copy.md), [T31](../../board-protocol/tasks/T31-registry-search.md), [T43](../../board-protocol/tasks/T43-ui-registry.md) lock, registry, registry screen | todo | **needed** for the e2e step «download a signed protocol while disabled» | T46 |
| [T35](../../board-protocol/tasks/T35-ui-settings-roster.md), [T36](../../board-protocol/tasks/T36-ui-cases-documents.md) roster and case-document screens | todo | mount `RosterDecisionMark`; trigger the «connect the storage first» message | T41, T42 |
| [T14](../../board-protocol/tasks/T14-river-worker-jobs.md) River worker | todo | every job opens `InAssociation` from its payload once row-level security is on | T20 |
| every board-protocol handler and ui task not started | todo | builds on the moved paths and on routes under `/o/$associationId/` | T01, T35 |

## Risks / Hard rules

- **Order against board-protocol (sad §11, High).** The SAD asks to agree the order before this breakdown; none was recorded, so the default of its mitigation is taken: T01 moves the paths first, in one PR, before any further board-protocol handler or ui task starts. If board-protocol goes first instead, T01 grows with every handler it has to move.
- **Migration order deviates from MANIFEST.md (T06).** The drop of `accounts.association_id` is promoted second, not last: nothing reads the column, an applicant cannot get an account while it is `NOT NULL`, and the staged order would loop (T20 needs T19, T19 needs T18, T18 needs the drop). T06 records the change in MANIFEST.md and data-model.md.
- **«PR 3» of the manifest is split.** Files 13–16 (T17) need only the backfills; files 17–19 (T20) wait for the app-role connection (T12) and the guards (T19), because row-level security enabled earlier hides every row from the app.
- **Carried open items** (api-sync-report §D asked for them before this stage): no code for a bad storage `state` or a refused code — T33 answers 422 without a code; editing the association's name and address has no AC and no task; PRD AC-19 lacks the «never before the entry's first day» rule that SAD §8 and ADR-0006 state — T07 implements the SAD; the open questions of PRD §8 (retention of rejected заявки, replacing a connected storage) keep their defaults.
- **Size.** 49 tasks and ~29.5 person-days against the SAD's ~22.5 (4.5 person-weeks): the test plan chose e2e-through-UI for every refusal and visual baselines (T43–T46), the epic absorbs three board-protocol tasks (T18, T19, T35) and adds the load scenarios and the security review (T47, T49). The skill's 8–20 band is exceeded on purpose, as in board-protocol.
- **Contract first, handlers later (T02).** The delta is merged once with 501 stubs and each ports task removes its own — the accepted deviation of board-protocol T07, so the ui lane starts early.
- **Isolation has three lines and no bypass** (ADR-0003, ADR-0009): row-level security, the explicit `association_id` filters of board-protocol ADR-0015, the release-blocking test (T21). Every repository of an association-owned table takes its runner from `db.From(ctx)`. The оператор role opens nothing on an association path; an operator use case reads one association-owned value only, the storage account address (T28).
- **Proof of office is a document a person checked** (ADR-0005): no task parses the extract or verifies the КЕП. Extracts, filings, checks, decisions and state changes are insert-only by privilege, never by trigger, and are never deleted.
- **Contexts meet only through `module.go`** (repo ADR 0002): identity reaches the roster through its `Membership` port, which governance implements and `cmd/api` wires — identity does not import governance. No foreign key crosses a context.
- **Generated code is never edited** (`*.gen.go`, `sqlcgen/`, `apps/web/src/api/gen/`, `routeTree.gen.ts`); `mise run gen` after the contract, a query or a migration changes. No `r.Mount` for a context.
- **Messages are PRD §5.1 verbatim, in Ukrainian;** no refusal names a голова, an applicant or another заявка (PRD §6.1 «probing»).
- **Logs never carry** the extract, the РНОКПП or a reason text (sad §8). **Fixtures carry no real person:** `<name>@example.test`, «Тестовий Голова», codes like 12345678.
