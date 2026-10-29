---
status: Draft
owner: "Andrii Opryshko"
reviewers: ["Andrii Opryshko"]
updated_at: "2026-10-29"
feature_size: "L"
---

# Test plan — board-protocol

> **Inputs:** [PRD](./PRD.md) §5 (37 Stage 1 acceptance criteria + 24 deferred to Stage 2), §6 NFRs ·
> [sad.md](./sad.md) §6 flows 1-13 and the coverage table · [data-model.md](./data-model.md) entities ·
> `target_surfaces: [backend-service, web-frontend, worker]`, so the UI tiers apply.

Stage 1 of board-protocol turns a conversation and case documents into a protocol draft whose proposed
numbers wait for a person's confirmation, runs an informal poll, prints with a reserved number, and locks
the protocol when the signed scan and the meeting facts are uploaded; a draft can be withdrawn. Every
Stage 1 criterion below has at least one test that `implement-tasks` writes red first.

## Levels

| Level | Scope | Strategy (generic — no tool names) |
|---|---|---|
| Unit | Pure governance domain logic: the lifecycle state machine, readiness rules, the final-amount choice rule, the fact validator, poll rounds and tallies, field validators. | In-memory, no I/O; table-driven cases per rule. |
| Integration | A use case through its module against the real database it owns, plus the worker jobs. | A fresh migrated ephemeral database per test (the repo's existing per-test database helper); external services (text generation, file storage, document rendering, sign-in provider) replaced by in-process fakes behind their ports. |
| Contract | The HTTP surface against the feature's OpenAPI delta (shapes, problem bodies, the same body for a missing and a forbidden id) and the worker job payloads against `contracts/events.md`. | Validate real responses and enqueued payloads against the agreed documents; no hand-rolled stubs. |
| E2E | A full lifecycle through the public API with a signed-in member. | Real core service and ephemeral database; fakes for the external services. |
| Load | Only the numeric NFRs of PRD §6. | The load tool already in the repo, or a load-testing tool appropriate for the stack. |
| Component | A UI piece in isolation: poll label and tally, slot origin/status, readiness card, registry entry. | Render in the repo's component harness with its accessibility checks; assert output + interaction. |
| Visual-regression | <!-- N/A: the repo has no approved visual baseline yet; the component tier covers states --> | — |
| E2E-through-UI | The user-story flows driven through the rendered web app. | Real web app + core service + ephemeral database; external services faked. |

## AC coverage

Expected outcomes are in plain words. «Fake generator» = the in-process fake of the text-generation service.

### US-01 — roster and settings

| AC (PRD §5) | Test name (intent-based) | Level | Expected outcome |
|---|---|---|---|
| AC-42 domain invariant | missing settings are named all at once | unit | the meeting form, signers and register start that are missing are listed together |
| AC-42 | first draft is refused until settings are saved, then the print copy carries the meeting form and signers | integration | refused with a link to the settings; after saving, the print copy shows them and nothing is compared against them |
| AC-42 | settings gap sends the member to the settings page | e2e-through-UI | the refusal lists the gaps and its link opens the settings |
| AC-02 error | roster entry without a term start or with the end before the start names every broken field | unit + contract | rejected; each broken field is named in one answer |

### US-12 / US-09 — conversational drafting and pre-fill

| AC (PRD §5) | Test name (intent-based) | Level | Expected outcome |
|---|---|---|---|
| AC-26 happy path | conversation about the roof waterproofing contract creates a draft with recommended unconfirmed slots | integration (fake generator) | draft for that meeting with that item; every slot unconfirmed with its source fragment and document link; request, turns and documents kept with the draft |
| AC-26 | member drafts a protocol from a conversation | e2e-through-UI | the proposed draft opens with the slots marked unconfirmed |
| AC-21 domain invariant | amount and counterparty are pre-filled as «from document, unconfirmed» with the quoted fragment | integration (fake generator) | both slots carry origin «document», status unconfirmed, fragment and link |
| AC-21 | ready-to-sign is refused while any proposed slot is unconfirmed | unit | readiness lists every unconfirmed slot |
| AC-21 | slot origin and status are shown on the slot | component | «from document, unconfirmed» badge with a link; confirming flips it to confirmed |
| AC-22 happy path | final-amount choice rule picks the full billed amount | unit (table: with/without VAT, superseded estimate, advance vs full, non-VAT contractor, no invoice yet) | the latest-dated full-cost document wins; the other values are listed with their documents |
| AC-28 domain invariant | a fact in generated text outside the slots becomes a new unconfirmed slot | unit | amounts, dates, document references, names and counts outside the slots are never kept in the text |
| AC-28 | generator text with an extra amount is turned into a slot | integration (fake generator returning a stray amount) | the stored text has no stray fact; a new unconfirmed slot holds it; members present and vote counts stay blank |
| AC-28b domain invariant | a conversation change never overwrites a confirmed slot | unit + integration | the new value sits beside the confirmed one; the slot needs review until a member picks one |

### US-02 — the draft

| AC (PRD §5) | Test name (intent-based) | Level | Expected outcome |
|---|---|---|---|
| AC-03 traceability | opened draft shows the meeting date, the roster in force on that date, the case documents and an item per case | integration + component | all four parts present; a member whose term does not cover the date is not in the roster |
| AC-03c happy path | a second case for the same meeting becomes a new agenda item of the existing draft | integration | one draft for the meeting, one more item; no lifecycle rule blocks it |
| AC-29 happy path | another member rewords the draft of a meeting chaired by someone else | integration | the wording is kept with its author and time; every slot keeps its value and source |
| AC-04 error | ready-to-sign is refused for a draft without items or with an item without a subject | unit + integration | every missing part named at once; the draft stays in чернетка |

### US-03 — informal poll

| AC (PRD §5) | Test name (intent-based) | Level | Expected outcome |
|---|---|---|---|
| AC-30 happy path | starting the poll freezes the wording as round 1, and as the next round after a return | unit + integration | state готовий до голосування; round numbers 1 then 3; unconfirmed slots shown as such and do not block |
| AC-31 happy path | round-1 tally of 3 for, 1 against, 1 abstain shows each member and time | unit + component | tally per member with time, labelled informal; no quorum, majority or adoption wording |
| AC-41 legal competence | poll label is shown on the poll, the tally and the registry history | component | exact label «Попереднє внутрішнє опитування правління — не є голосуванням засідання і не замінює його» |
| AC-41 | poll tallies never reach the print copy or the meeting facts | integration | the print copy has no poll result; meeting-facts entry starts blank; registry shows the poll only as draft history |
| AC-05d domain invariant | an edit or a return during round 2 closes the round and the next start opens round 3 empty | unit + integration | round 2 votes kept for history and not counted; round 3 has no votes |
| AC-05e authorization | a member cannot cast or change another member's vote | integration + contract | refused; the other member's vote is unchanged |
| AC-32 domain invariant | a vote is refused outside готовий до голосування | unit (each of чернетка, готовий до підпису, підписаний, знятий з розгляду) | refused with «poll is open only while the draft is готовий до голосування» |

### US-04 — numbers linked to documents

| AC (PRD §5) | Test name (intent-based) | Level | Expected outcome |
|---|---|---|---|
| AC-09 domain invariant | a newer version of a cited document marks the citing slots as needing review | integration | confirmed value shown beside the current document's value; slot no longer confirmed; a ready draft returns to чернетка |
| AC-09 | a document changed in file storage is caught by the nightly check | integration (worker, fake file storage returning a different hash) | the citing slots are marked as needing review |

### US-06 — print and lock

| AC (PRD §5) | Test name (intent-based) | Level | Expected outcome |
|---|---|---|---|
| AC-14c happy path | ready-to-sign freezes the text, renders the print copy and reserves the next number | integration (fake renderer) | frozen text; print copy with meeting date, form, signers, blank attendance and vote fields; next register number printed |
| AC-14c | two drafts marked ready at the same moment get different numbers | integration (concurrent) | one sequence, no duplicates |
| AC-33 happy path | ready-to-sign from an open poll closes the round | integration | готовий до підпису with copy and number; round 2 closed and kept; no poll result on the copy |
| AC-14b domain invariant | an edit or a return from готовий до підпису voids the copy and keeps No 14 | unit + integration | back to чернетка; copy void and kept; warning not to sign earlier copies; No 14 kept and never given to another protocol |
| AC-14 happy path | signed scan with the meeting facts locks the protocol | integration | підписаний; registry entry with scan, facts, uploader and time; text no longer editable |
| AC-14 | full lifecycle: draft → poll → print → upload signed scan → registry | e2e + e2e-through-UI | the protocol appears in the registry locked with its facts |
| AC-34 error | upload without members present or a vote count per item names every missing fact | unit + integration | not locked; every missing fact named at once; stays готовий до підпису |
| AC-15 domain invariant | reported handwritten corrections require a note that is shown beside the frozen text | unit + integration + component | refused without a note; with it, locked, frozen text unchanged, note shown in the registry |
| AC-40 domain invariant | a signed copy is refused before готовий до підпису | unit | refused in чернетка and готовий до голосування with «mark готовий до підпису and print first» |
| AC-16 authorization | a former member cannot upload the signed copy of a draft they worked on | integration + contract | refused; nothing locked; the attempt changes no state; actions of current members are recorded with author |
| AC-17 domain invariant | text, meeting facts and signed copy of a locked protocol cannot change | integration | every change refused with «changes only through a new board decision» |
| AC-17 | stored locked rows reject update and delete at the database | integration | the database itself refuses (append-only rule of ADR-0007) |

### US-07 — registry

| AC (PRD §5) | Test name (intent-based) | Level | Expected outcome |
|---|---|---|---|
| AC-18 happy path | registry search by counterparty, exact amount, amount range, subject and period over 30 locked protocols | integration | exactly the matching decisions, each opening with signed copy, facts, cited documents and poll history labelled informal |
| AC-18 | member finds and opens a past decision | e2e-through-UI | search → result → decision page with its basis |
| AC-19 authorization | a former or never member is denied the registry and cases without revealing existence | integration + contract | denied; the same answer for an existing and a missing id |

### US-13 — withdrawal

| AC (PRD §5) | Test name (intent-based) | Level | Expected outcome |
|---|---|---|---|
| AC-35 happy path | a чернетка is withdrawn with a reason and takes no number | unit + integration | знятий з розгляду; cases and versions kept; no number |
| AC-36 happy path | a draft готовий до голосування is withdrawn with its poll rounds | integration | знятий з розгляду; rounds kept; never printed; no number |
| AC-37 happy path | a printed unsigned No 14 is withdrawn and keeps its number | integration | знятий з розгляду; frozen text and copy kept as never signed; No 14 never reused |
| AC-38 domain invariant | a locked protocol cannot be withdrawn | unit | refused; a new revoking or correcting decision is offered |
| AC-39 domain invariant | a withdrawn draft cannot be edited or polled | unit | refused; a new draft copied from it is offered |

### Deferred to Stage 2

The 24 deferred criteria (AC-01, AC-02b, AC-02c, AC-05, AC-05b, AC-05c, AC-06, AC-07, AC-08, AC-10, AC-11,
AC-12, AC-12b, AC-13, AC-14d, AC-17b, AC-17c, AC-20, AC-23, AC-23b, AC-24, AC-25, AC-25b, AC-27) get their
rows when Stage 2 is specified and re-validated (PRD §5 «Deferred to Stage 2»). AC-03b is dropped
(replaced by AC-29) and needs no test.

## Edge cases / error paths

- No session → refused as not signed in; nothing read.
- A stale draft revision on any change → refused as a conflict; the newer state is kept (ADR-0009).
- Text-generation service or document renderer fails or exceeds the job timeout → the job is retried with backoff and, after the attempt limit, marked failed; the member sees the failure with a retry, can fill the slots by hand, and the committed state is never undone (SAD §6, background jobs) — integration on the worker with failing fakes.
- Text-generation service answers with an unusable structure → nothing it proposed is stored; the job is marked failed with a retry offered.
- A case document instructing the assistant (text meant to steer it) → every proposed number still arrives unconfirmed; the stored text holds no fact outside the slots (PRD §6.1).
- Upload over 20 MB, or not PDF / JPEG / PNG → rejected and counted; nothing stored.
- Any uploaded file is served for download only, never rendered as active content.
- File storage unavailable at a case-document or signed-scan upload → nothing is committed, the state does not change, the member is told storage is unavailable and to try again; a file stored before a failed transaction is removed by the orphan-file job.
- Settings or roster with several broken fields → every field named in one answer.
- Integrity check finds a changed or trashed file of a locked protocol → a mismatch is recorded and alerted; a trashed file is restored.

## Test data

- Seed strategy: factories per entity of [data-model.md](./data-model.md) — association with settings and register, board members with terms (current, ended, future), meetings, cases with versioned documents, protocols in each lifecycle state, poll rounds with votes, print copies, meeting facts. Placeholder data only (`<name>@example.test`, «Тестовий Член Правління»).
- Integration dependency: a fresh migrated ephemeral database per test, never a mocked store. External services are in-process fakes behind their ports, with one smoke test per service against a test account before release.
- Cleanup boundary: per test for integration (its own database); per suite for e2e and e2e-through-UI (one seeded association, torn down after the run).

## NFR validation (load)

The load tool already in the repo, or a load-testing tool appropriate for the stack. Scale: one association, a board of 5-9 members, so rates are small; the thresholds are what matters.

- Read p95 ≤ 200 ms (open a draft or a registry decision, registry search) → 10 requests/s for 5 min over a registry seeded with 300 locked protocols; assert p95 ≤ 200 ms and no errors.
- Write p95 ≤ 500 ms (without a file) → 5 requests/s of slot confirmations and rewordings for 5 min; assert p95 ≤ 500 ms; revision conflicts are expected and not counted as errors.
- Upload p95 ≤ 3 s for 20 MB → 20 sequential 20 MB uploads; assert p95 ≤ 3 s.
- Print copy p95 ≤ 5 s → 20 ready-to-sign transitions against the real renderer of the test account (pre-release only); assert p95 ≤ 5 s.
- Proposal p95 ≤ 30 s → 10 conversational drafts against the real text-generation service of the test account (pre-release only); assert p95 ≤ 30 s from request to filled draft.
- Not load tests, verified elsewhere: integrity (the nightly check, integration above + production alert), RPO/RTO (quarterly restore drill, ADR-0016), availability (uptime check), retention (storage policy).

## CI placement

- On every PR: unit, integration, contract, component, e2e and e2e-through-UI (all with fakes; the repo already runs its browser suite on PRs).
- Pre-release / on schedule: load scenarios and the smoke tests against the real external services.
