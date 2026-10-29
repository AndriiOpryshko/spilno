# Tracker — board-protocol

> Status of every task in the epic. `implement-tasks` updates `done` as it commits each task.
> States: `todo` · `in_progress` · `blocked` · `review` · `done`. Estimates: S = 2 h, M = half a day, L = a day.

| # | Task | Layer | Owner | Estimate | Blocked by | Status |
|---|---|---|---|---|---|---|
| [T01](./T01-stage-identity-migrations.md) | Promote the identity migrations: accounts, sessions, google_connections | migration | Andrii Opryshko | S | — | done |
| [T02](./T02-stage-roster-settings-migrations.md) | Promote the roster, settings and meetings-chair migrations | migration | Andrii Opryshko | S | T01 | todo |
| [T03](./T03-stage-case-protocol-migrations.md) | Promote the case, protocol, agenda-item and decision-slot migrations | migration | Andrii Opryshko | S | T02 | todo |
| [T04](./T04-stage-assistant-poll-river-migrations.md) | Promote the assistant-turn and poll migrations and add the River schema | migration | Andrii Opryshko | M | T03 | todo |
| [T05](./T05-stage-lock-integrity-migrations.md) | Promote the print-copy, meeting-facts, snapshot, trigger and integrity-run migrations | migration | Andrii Opryshko | M | T04 | todo |
| [T06](./T06-problem-codes-registry.md) | Add problem `code` and `urn:problem:` types to httpapi with a governance error registry | infra | Andrii Opryshko | M | — | todo |
| [T07](./T07-merge-contract-identity-skeleton.md) | Merge the contract delta, add the identity context skeleton and regenerate server and client | wiring | Andrii Opryshko | L | — | todo |
| [T08](./T08-identity-oidc-sessions.md) | Sign in with Google OIDC and keep sessions in Postgres | app | Andrii Opryshko | L | T01, T06, T07 | todo |
| [T09](./T09-settings-roster-domain.md) | Model association settings and roster entries with all-fields validation | domain | Andrii Opryshko | M | — | todo |
| [T10](./T10-settings-roster-endpoints.md) | Serve settings and roster: use cases, Postgres repositories, handlers and CurrentMember | ports | Andrii Opryshko | L | T02, T06, T07, T09 | todo |
| [T11](./T11-require-session-current-member.md) | Require a session and a current член правління on the governance and finance routes | wiring | Andrii Opryshko | M | T08, T10 | todo |
| [T12](./T12-google-connection-token-source.md) | Store the association's Google connection and expose GoogleTokenSource | infra | Andrii Opryshko | M | T01, T07 | todo |
| [T13](./T13-drive-files-adapter.md) | Implement the Drive Files adapter with sha256 checks and an in-memory fake | infra | Andrii Opryshko | M | T12 | todo |
| [T14](./T14-river-worker-jobs.md) | Wire the River client, the cmd/worker binary and the job status endpoints | wiring | Andrii Opryshko | L | T04, T07, T11 | todo |
| [T15](./T15-protocol-lifecycle-domain.md) | Model the protocol lifecycle: states, revision, transitions and refusals | domain | Andrii Opryshko | L | — | todo |
| [T16](./T16-agenda-slots-domain.md) | Model agenda items and слоти рішення: confirm, correct, needs review and readiness | domain | Andrii Opryshko | M | T15 | todo |
| [T17](./T17-text-fragments-proposal-rules.md) | Add text fragments, the fact validator and the amount recommendation rule | domain | Andrii Opryshko | L | T16 | todo |
| [T18](./T18-poll-domain.md) | Model poll rounds and the own-vote rule | domain | Andrii Opryshko | M | T15 | todo |
| [T19](./T19-meeting-facts-snapshot-domain.md) | Model meeting facts and the canonical fingerprinted snapshot | domain | Andrii Opryshko | M | T16, T18 | todo |
| [T20](./T20-protocol-repository.md) | Persist the protocol aggregate with a revision guard | infra | Andrii Opryshko | L | T03, T16 | todo |
| [T21](./T21-cases-endpoints.md) | Serve cases: create, list, open, add document versions and download files | ports | Andrii Opryshko | L | T03, T11, T13 | todo |
| [T22](./T22-open-draft-endpoints.md) | Open a draft for a meeting from a case, get and list protocols | ports | Andrii Opryshko | L | T10, T20, T21 | todo |
| [T23](./T23-wording-slot-endpoints.md) | Edit wording and слоти рішення: save wording, add, confirm and correct slots | ports | Andrii Opryshko | L | T17, T22 | todo |
| [T24](./T24-gemini-textmodel-adapter.md) | Implement the Gemini TextModel adapter as the only importer of the SDK | infra | Andrii Opryshko | M | T07, T17 | todo |
| [T25](./T25-assistant-propose-draft.md) | Take conversation turns and propose a draft in the propose_draft job | app | Andrii Opryshko | L | T04, T14, T23, T24 | todo |
| [T26](./T26-poll-endpoints.md) | Start the попереднє голосування, cast own votes and return to draft | ports | Andrii Opryshko | L | T04, T18, T22 | todo |
| [T27](./T27-gdocs-renderer-adapter.md) | Render the print copy from a pinned Google Docs template, with a fake Renderer | infra | Andrii Opryshko | M | T12 | todo |
| [T28](./T28-ready-for-signing-print.md) | Mark ready for signing: reserve the number and render the print copy in a job | ports | Andrii Opryshko | L | T05, T14, T23, T26, T27 | todo |
| [T29](./T29-lock-with-signed-copy.md) | Lock with the signed scan and meeting facts, and apply Drive restrictions | ports | Andrii Opryshko | L | T19, T28 | todo |
| [T30](./T30-withdraw-and-copy.md) | Withdraw a protocol with a reason and copy a withdrawn draft | ports | Andrii Opryshko | M | T29 | todo |
| [T31](./T31-registry-search.md) | Search the реєстр протоколів and open a signed protocol | ports | Andrii Opryshko | L | T29 | todo |
| [T32](./T32-nightly-integrity-check.md) | Run the nightly integrity check of snapshots and clean up orphan files | app | Andrii Opryshko | M | T14, T29, T30 | todo |
| [T33](./T33-nightly-backup-runbook.md) | Back up Postgres nightly off the VM and write the restore-drill runbook | infra | Andrii Opryshko | M | — | todo |
| [T34](./T34-ui-sign-in-session.md) | Add the sign-in page, the session guard and 401/403 handling | ui | Andrii Opryshko | M | T07 | todo |
| [T35](./T35-ui-settings-roster.md) | Build the association settings and roster screens | ui | Andrii Opryshko | L | T07 | todo |
| [T36](./T36-ui-cases-documents.md) | Move the cases list and case screen to generated hooks with document upload | ui | Andrii Opryshko | M | T07 | todo |
| [T37](./T37-ui-protocol-on-hooks.md) | Move the protocol screen to generated hooks: agenda, slots and revision conflicts | ui | Andrii Opryshko | L | T07 | todo |
| [T38](./T38-ui-text-editor-slot-chips.md) | Build the decision text editor with slot chips and highlighted typed facts | ui | Andrii Opryshko | L | T37 | todo |
| [T39](./T39-ui-assistant-panel.md) | Add the помічник протоколу conversation panel with job polling | ui | Andrii Opryshko | M | T38 | todo |
| [T40](./T40-ui-live-poll.md) | Make the попереднє голосування live: start, own vote, tally, informal label, return to draft | ui | Andrii Opryshko | M | T39 | todo |
| [T41](./T41-ui-print-withdraw.md) | Wire printing and withdrawal: ready for signing, job polling, void warning, withdraw and copy | ui | Andrii Opryshko | L | T40 | todo |
| [T42](./T42-ui-signed-copy-form.md) | Build the signed-scan upload with the meeting facts form | ui | Andrii Opryshko | M | T41 | todo |
| [T43](./T43-ui-registry.md) | Build the реєстр протоколів search and the signed protocol view | ui | Andrii Opryshko | M | T07 | todo |
| [T44](./T44-e2e-draft-to-registry.md) | Prove the flow end to end: draft to registry through the UI with fake Google and Gemini | tests | Andrii Opryshko | L | T25, T28, T29, T30, T31, T32, T34, T35, T36, T42, T43 | todo |

**Total:** 44 tasks, ~31.25 person-days, 12 dependency phases.
