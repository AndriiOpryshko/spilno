---
status: Draft
skill: ao-architecture
owner: "Andrii Opryshko"
reviewers: ["Andrii Opryshko"]
updated_at: "2026-11-05"
feature_size: "M"
stage: "04-05"
depth: "easy"
ticket: ""
target_surfaces: [backend-service, web-frontend]
contexts: [identity, governance, finance, platform]
---

# Software Architecture Document — association-onboarding

> **Inputs:** [PRD](./PRD.md) (Approved 2026-11-05, ao-prd, amended at this pass: manual extract check) · [CONTEXT](./CONTEXT.md) · [board-protocol CONTEXT](../board-protocol/CONTEXT.md) · repo ADRs [`docs/adr/`](../../adr/) · board-protocol [SAD](../board-protocol/sad.md) and [ADRs](../board-protocol/adr/) · reference contexts `services/core/internal/identity`, `services/core/internal/governance`

## 1. Introduction and goals

**Intent.** A голова правління connects their association to Спільно, or takes it over after an election, without the developer: they sign in, file a заявка на підключення or a запит на передачу with a КЕП-signed витяг з ЄДР, and the оператор checks the extract outside Спільно and decides with a reason. The оператор can disable and enable an association; the board connects its own document storage. With a second association the data of each association becomes visible only to the people on its roster, and one person may sit on several rosters (PRD §1, §2).

**Top-3 quality goals** (full scenarios in §10):

1. **Isolation** — 0 records of association B readable by a person not on B's roster, enforced by the database and by a release-blocking check (PRD §6 Isolation).
2. **Immediate effect of access decisions** — a disabling or a transfer is in force from the next request (PRD §6).
3. **Traceability of office** — every голова in Спільно links to the stored extract and the оператор's recorded check and decision, kept ≥ 5 years (PRD §6 Retention, AC-29).

**Stakeholders.**

| Role (glossary) | Interest | Sign-off? |
|---|---|---|
| голова правління | files, follows and refiles a заявка or a запит; becomes the first член правління | No |
| оператор | checks extracts, decides, disables and enables associations | Yes (product owner) |
| член правління | works in one or several associations; connects the storage; reads the registry of a вимкнене ОСББ | No |
| ревізійна комісія | later reads the chain from Спільно to ЄДР (no access in this feature) | No |
| Tech lead | SAD approval | Yes |

Product-owner decisions at this pass (2026-11-05): the extract and the КЕП are checked by hand outside Спільно («вся перевірка — ручна і за кадром», PRD amended); the association id is part of the path («хоч рут різний з айді осбб»); row-level security now.

<!-- Decision overrides from the critic: «- Decision override: <headline> — rationale: <reason>» -->

## 2. Constraints

**Technical** (versions from `mise.toml`, `infra/compose.yaml`, board-protocol SAD §2):
- Go 1.27.3, chi v5, oapi-codegen strict server over one OpenAPI 3.1 contract `services/core/api/openapi.yaml`; every operation has one context tag and a `/<context>/` path (repo [ADR 0001](../../adr/0001-stack.md), [ADR 0003](../../adr/0003-spec-first-codegen.md)).
- Postgres 18.6 (`postgres:18.6-alpine`), pgx/v5 + sqlc, goose migrations in `services/core/migrations/` applied on start (repo [ADR 0004](../../adr/0004-migrations-and-queries.md)); next migration `00028`. Today one database role owns the schema and serves the app.
- Web: React 19.3, TypeScript 6.0, Vite 8, TanStack Router + Query, orval hooks per tag, shadcn/ui, Storybook 10; routes today `index`, `cases.index`, `cases.$caseId.protocol`.
- Existing contexts: `identity` (stub handler: session, sign-in, callback, sign-out; tables `associations` with one seeded row, `accounts` with `association_id NOT NULL`, `sessions`, `google_connections`), `governance` (roster `board_members` + `board_member_revisions` with term and no position; ~30 operations under `/governance/...`; `association_id` + explicit filter per board-protocol ADR-0015), `finance` (`contracts` **without** `association_id`), `system`, `platform` (`httpapi`, `testdb`).
- Background work is River (board-protocol ADR-0006); this feature adds no job.

**Repo ADRs that bind this feature:** 0001 stack · 0002 bounded contexts · 0003 spec-first codegen · 0004 migrations and queries.

**Feature ADRs of board-protocol it changes:** [ADR-0003](../board-protocol/adr/0003-sign-in-with-oidc-and-postgres-sessions.md) (sign-in refuses a person off the roster → superseded in part by ADR-0004), [ADR-0005](../board-protocol/adr/0005-store-files-in-association-owned-google-drive-folder.md) (storage authorized once by a developer command → ADR-0008), [ADR-0015](../board-protocol/adr/0015-scope-owned-rows-to-one-association-with-explicit-query-filters.md) (RLS deferred until a second association → ADR-0003), tasks T08 and T12 (replaced).

**Conventions** (`CLAUDE.md`, `.claude/rules/openapi.md`, `.claude/rules/migrations.md`): one OpenAPI contract, one tag per context, paths `/<context>/`; contexts import each other only through `module.go`; RFC 9457 problem+json with `code`; `r.Group`, never `r.Mount`; generated files never edited; new tables without business `CHECK`, triggers or stored procedures; zero-downtime 3-step for a new NOT NULL column and for a dropped column.

**Organisational:** one developer with coding agents; effort ~3 person-weeks, Dec 2026 – Jan 2027, after board-protocol Stage 1 (brief §12). The path change of ADR-0002 touches operations that board-protocol Stage 1 tasks are building (§11).

**Regulatory / legal:** Law «On personal data protection» — the applicant's name, address, position and term; the РНОКПП inside the КЕП certificate of the stored extract (shown to nobody, kept only in the file); classification Confidential (PRD §6.1). The оператор decides access to Спільно, never who heads an association (brief §18, AC-12). Security review required before release (PRD §6.1).

## 3. Context and scope

Спільно is one web application in which each association works in its own space. People reach it with their sign-in account; the trust boundary is the session plus the association's roster: nothing of an association is served to a person who is not on its roster in force today. The КЕП and the ЄДР extract are checked by the оператор with public services outside Спільно; Спільно only stores the file and what the оператор recorded.

| Actor or system | Type | Interaction |
|---|---|---|
| голова правління | Person | signs in, files a заявка на підключення or a запит на передачу with the signed extract, follows it |
| член правління | Person | works in each association whose roster they are on; connects the storage |
| оператор | Person | checks extracts outside Спільно, records the check, decides, disables and enables |
| Google sign-in | System (external) | proves the person's sign-in address (OIDC) |
| Google Drive of the association | System (external) | the association's document storage, connected by the board |
| Public signature verification service (ЦЗО) | System (external) | used by the оператор by hand to check the КЕП; Спільно does not call it |

```mermaid
C4Context
    title association-onboarding — System Context

    Person(head, "голова правління", "files a заявка or a запит with a signed extract")
    Person(member, "член правління", "works in the associations whose roster they are on")
    Person(operator, "оператор", "checks extracts and decides access")
    System(spilno, "Спільно", "board records of each association, isolated per association")
    System_Ext(google, "Google sign-in", "proves the sign-in address")
    System_Ext(drive, "Google Drive of the association", "document storage connected by the board")
    System_Ext(czo, "Signature verification service", "used by the оператор by hand")

    Rel(head, spilno, "files and follows applications", "HTTPS")
    Rel(member, spilno, "works in an association", "HTTPS")
    Rel(operator, spilno, "records checks and decisions", "HTTPS")
    Rel(spilno, google, "signs people in", "OIDC")
    Rel(spilno, drive, "stores case files", "Drive API")
    Rel(operator, czo, "checks the КЕП of the extract", "HTTPS")
```

## 4. Solution strategy

**Target surface(s):** `[backend-service, web-frontend]` — the core API and the existing SPA; no worker, CLI or new process (the extract check is manual, ADR-0005).

**Strategic pillars:**

1. **The association is explicit in every association-scoped path** — `/<context>/associations/{association_id}/...` in the API and `/o/{association_id}/...` in the web app; the identity middleware checks the person against that association's roster before any context runs ([ADR-0002](./adr/0002-put-the-association-id-in-every-association-scoped-path.md)). Constraint: repo ADR 0003 (`/<context>/` prefix kept), AC-04d.
2. **The database enforces isolation** — row-level security on every association-owned table, the app connects as a role without `BYPASSRLS`, each association-scoped request runs in a transaction with the association set; the explicit filters of board-protocol ADR-0015 stay as a second line, and a release-blocking test seeds two associations and a person on both ([ADR-0003](./adr/0003-enforce-association-isolation-with-postgres-row-level-security.md)). Constraint: PRD §6 Isolation; brief §10 top risk.
3. **People are not bound to one association** — any verified person holds a session; which associations they may open comes from the rosters in force today; an applicant has a session and no association ([ADR-0004](./adr/0004-let-any-verified-person-sign-in-and-derive-memberships-from-rosters.md)). Constraint: PRD §1 ¶4 (several associations per person), board-protocol ADR-0003.
4. **Proof of office is a document a person checked** — Спільно keeps the uploaded extract unchanged and the оператор's recorded check; approval is refused without it; Спільно parses or verifies nothing ([ADR-0005](./adr/0005-check-the-signed-extract-manually-and-keep-the-uploaded-file-in-identity.md)). Constraint: PRD AC-03, AC-29; product owner 2026-11-05.
5. **UI architecture** — the same SPA: association routes under `/o/{association_id}/`, an association picker after sign-in, the applicant's pages under `/applications`, and the оператор's back office under `/operator/` shown only to the оператор; reuses the shadcn components and the Stage 1 Figma frames (BR-21) ([ADR-0001](./adr/0001-serve-the-operator-back-office-as-an-operator-only-section-of-the-web-app.md)). Constraint: PRD §8 back-office question, `docs/design/stage1-screens-spec.md` §14.

## 5. Building block view

Hexagonal per context, as the repo already does. The feature extends `identity` (it owns associations, accounts, sessions and the storage connection — board-protocol ADR-0014) with applications, the оператор's queue and the association state; `governance` gains positions on the roster and two facade functions; `finance` gains the association column; `platform` gains the transaction-with-association helper. No new context.

**Context ownership** (fills frontmatter `contexts`):

| Context | New? | Owns (aggregates, tables) | Provides to others via module.go | Uses from others |
|---|---|---|---|---|
| identity | existing, extended | association (+ ЄДРПОУ, address, state, state changes); account (no association); session; заявка на підключення / запит на передачу with filings, stored extract, оператор's check and decision; storage connection | `ActorFrom(ctx)` (account, email, association, member, read-only flag); storage token source | governance: memberships by email, current member in an association, seat a голова, end a term, members' sign-in addresses |
| governance | existing, extended | roster entry + revisions (+ position голова правління / член правління, + decision reference); everything of board-protocol | `CurrentMember(ctx, association, email, date)`, `MembershipsOf(ctx, email, date)`, `SeatHead(ctx, association, person, term, decision)`, `EndTerm(ctx, association, member, date, decision)`, `MemberSignInAddresses(ctx, association)` | identity: actor, storage token source |
| finance | existing | contracts (+ `association_id`) | — | identity: actor |
| platform | existing | `platform/db`: `InAssociation(ctx, id, fn)` transaction helper; `testdb` seeds two associations | helpers only | — |

**Layout inside identity** (no new context, so no CLAUDE.md wiring checklist; new operations stay under the `identity` tag):

```
services/core/internal/identity/
├── domain/              association state, application and filing, extract check, decision; sentinels for PRD §5.1 identity.* codes
├── app/                 file, refile, record check, approve, reject, disable, enable, connect storage; operator check
├── ports/               Membership (governance), Applications, Associations, Extracts repositories
├── adapters/http/       StrictServerInterface; session + association middleware
├── adapters/postgres/   queries/ → sqlcgen/
└── module.go            wiring; ActorFrom; token source
```

```mermaid
C4Container
    title association-onboarding — Containers

    Person(head, "голова правління")
    Person(member, "член правління")
    Person(operator, "оператор")

    System_Boundary(sys, "Спільно") {
        Container(web, "Web client", "React, TanStack Router", "association space /o/id, applications, operator back office")
        Container_Boundary(core, "Core API — Go modular monolith") {
            Container(identity, "identity", "Go bounded context", "sessions, associations, applications, extracts, operator decisions, storage connection")
            Container(governance, "governance", "Go bounded context", "roster with positions, protocols, registry")
            Container(finance, "finance", "Go bounded context", "contracts per association")
        }
        ContainerDb(db, "Database", "Postgres 18.6, goose + sqlc, row-level security", "tables per context, stored extracts")
    }

    System_Ext(google, "Google sign-in", "OIDC")
    System_Ext(drive, "Google Drive of the association", "case files")
    System_Ext(otel, "Observability", "OpenTelemetry to Grafana LGTM")

    Rel(head, web, "uses", "HTTPS")
    Rel(member, web, "uses", "HTTPS")
    Rel(operator, web, "uses", "HTTPS")
    Rel(web, identity, "calls", "OpenAPI, JSON")
    Rel(web, governance, "calls", "OpenAPI, JSON")
    Rel(web, finance, "calls", "OpenAPI, JSON")
    Rel(identity, governance, "memberships, seat head, end term", "through module.go")
    Rel(governance, identity, "actor, storage token", "through module.go")
    Rel(identity, db, "reads/writes own tables", "SQL")
    Rel(governance, db, "reads/writes own tables", "SQL")
    Rel(finance, db, "reads/writes own tables", "SQL")
    Rel(identity, google, "signs in, storage consent", "OIDC, OAuth")
    Rel(governance, drive, "stores files", "Drive API")
    Rel(identity, otel, "traces, metrics", "OTLP")
```

Dropped recommended containers: `files` (case files stay in the association's Drive, board-protocol ADR-0005; extracts live in `db`, ADR-0005), `ai` and `qes` (no AI; the КЕП is checked by hand).

**Roster change on approval** → [ADR-0006](./adr/0006-apply-approvals-to-the-roster-through-a-governance-facade-in-one-transaction.md): identity's approval use case calls `governance.SeatHead` / `EndTerm` inside one transaction carried in `ctx`.

## 6. Runtime view

Primary flows per critical PRD §10 transition; `complete-sequence-diagrams` adds the rest.

**Flow 1: filing a заявка на підключення (— → Очікує) — covers AC-01, AC-02, AC-04, AC-04b, AC-04c, AC-30**

```mermaid
sequenceDiagram
    actor H as голова правління
    participant W as web
    participant I as identity
    participant D as db
    H->>W: fill the form and attach the signed extract
    W->>I: file application with the extract file
    I->>I: validate every field at once and the term end
    alt invalid fields or term ended
        I-->>W: refused with every broken field named
    else valid
        I->>D: is the code connected, or pending by this person
        alt code connected
            I-->>W: refused, already connected, offer a transfer request
        else free or pending by someone else
            I->>D: insert application, filing and extract in one transaction
            I-->>W: filed, waiting for the operator
        end
    end
```

**Flow 2: the оператор records the check and approves a заявка (Очікує → Схвалена, — → Підключене) — covers AC-03, AC-03b, AC-08, AC-09, AC-12, AC-04c**

```mermaid
sequenceDiagram
    actor O as оператор
    participant W as web
    participant I as identity
    participant G as governance
    participant D as db
    O->>O: check the КЕП and the extract outside Спільно
    O->>W: record signer, керівник, code and date of the extract
    W->>I: record check
    I->>D: insert the check, compute name mismatch and age warnings
    O->>W: approve, naming a document if a mismatch is open
    W->>I: approve
    I->>I: check recorded, codes equal, document named if needed
    I->>D: begin, code still free, insert the association
    I->>G: seat the applicant as голова in the new association
    G->>D: insert roster entry with position and decision reference
    I->>D: record the decision and commit
    I-->>W: approved, association connected
```

**Flow 3: approving a запит на передачу (Очікує → Схвалена, roster change) — covers AC-19, AC-19b, AC-19c, AC-24b**

```mermaid
sequenceDiagram
    actor O as оператор
    participant W as web
    participant I as identity
    participant G as governance
    participant D as db
    O->>W: approve the transfer request
    W->>I: approve
    I->>I: check recorded, codes equal, document named if needed
    I->>D: begin with the association set
    I->>G: end the term of the recorded голова on the day before the new term start
    G->>D: insert a roster revision with term end and decision reference
    I->>G: seat the new голова from the typed term start
    G->>D: insert roster entry with position and decision reference
    I->>D: record the decision and commit
    Note over I,G: the ended term is before today, so the next membership check refuses the previous голова
    I-->>W: approved, competing requests stay waiting
```

**Flow 4: an association-scoped request through the middleware (access, disabled, not a member) — covers AC-07, AC-16b, AC-21, AC-25, AC-26**

```mermaid
sequenceDiagram
    actor M as член правління
    participant W as web
    participant I as identity
    participant G as governance
    participant D as db
    M->>W: open a page of the association
    W->>I: request with the association id in the path
    I->>D: resolve the session
    I->>D: begin, set the association for row-level security
    I->>G: is this person a current member today
    alt not a member, the operator, or unknown association
        I-->>W: denied, same answer as a missing record
    else member and association disabled
        alt registry read or signed copy download
            I->>G: serve read only
            G-->>W: registry or signed copy
        else any change
            I-->>W: access suspended with the reason
        end
    else member and association connected
        I->>G: handle the request inside the same transaction
        G-->>W: result
    end
```

## 7. Deployment view

Same VM and compose topology as board-protocol SAD §7 (`caddy`, `core-api`, `worker`, `postgres`, `backup`, `otel-lgtm`); no new process. Changes:

- **Two database roles** (ADR-0003): `spilno_owner` owns the schema and runs goose on start (`MIGRATIONS_DATABASE_URL`); `spilno_app` (no `BYPASSRLS`, not the owner) serves requests and River jobs (`DATABASE_URL`); `spilno_app` has no `UPDATE` or `DELETE` on the extract table (ADR-0005). Local `infra/compose.yaml` creates both roles in an init script; `testdb.New(t)` does the same.
- **Configuration:** `SPILNO_OPERATOR_EMAIL` (the one оператор, ADR-0001); the OAuth redirect URI for the storage connection registered with Google (ADR-0008).
- **Backups:** extracts are inside Postgres, covered by the nightly dump (board-protocol ADR-0016); at ~6 filings a year of ≤ 5 MB the dump grows by ≤ 30 MB a year.

**Monitoring additions:**

| Signal | Source | Alert |
|---|---|---|
| Request duration p95 per route | `http.server.request.duration` | read > 200 ms, write > 500 ms, filing > 5 s for 15 min |
| Isolation denials | `identity.access.denied` by reason (not_member, disabled_write, operator_on_association) | sudden rise > 10× weekly mean (probing) |
| Applications waiting | `identity.applications.pending_age_days` (oldest) | oldest > 5 working days |
| Row-level-security misses | `platform.db.association_unset` (a query on a scoped table without the association set returns nothing) | any |

**Scaling:** ~6 filings a year; tens of associations at most before brief §14's revisit trigger; one API replica.

## 8. Crosscutting concepts

| Concept | Convention for this feature | Where defined |
|---|---|---|
| Errors | problem+json; codes from PRD §5.1: `identity.application_invalid`, `identity.signature_not_checked`, `identity.extract_code_mismatch`, `identity.association_already_connected`, `identity.application_already_open`, `identity.mismatch_needs_document`, `identity.reason_required`, `identity.application_not_pending`, `identity.not_operator`, `identity.association_not_connected`, `identity.association_state_unchanged`, `identity.association_disabled`, `identity.term_ended`, `identity.storage_account_personal`, `governance.drive_not_connected` (planned); `identity.not_current_member` (existing) also answers another association's object and another person's application, so existence is not revealed | `internal/platform/httpapi`, PRD §5.1 |
| Authentication | OIDC sign-in for any verified address; server-side session; no session → `identity.session_required` | board-protocol ADR-0003, ADR-0004 |
| Authorization | PRD §10 matrix. Association-scoped paths: middleware → session → association state → `governance.CurrentMember(association, email, today Europe/Kyiv)` → actor. Applicant paths (`/identity/applications`): own records by account. Operator paths (`/identity/operator/...`): session address equals `SPILNO_OPERATOR_EMAIL`, else `identity.not_operator`; the оператор is refused on every association path (AC-16b) | ADR-0001, ADR-0002, ADR-0004 |
| Association scope | path parameter `association_id`, never the body; RLS on every association-owned table; explicit filters kept | ADR-0002, ADR-0003 |
| Disabled association | actor carries `read_only`; only the registry read and signed-copy download operations pass; everything else `identity.association_disabled` with the reason | ADR-0007 |
| Validation | OpenAPI validator (422) + `domain.Validate()` naming every broken field (ЄДРПОУ 8 digits, name, address, term start, extract attached) | CLAUDE.md |
| IDs | UUID v7 generated app-side for new tables; cross-context references by id only | `.claude/rules/migrations.md`, repo ADR 0002 |
| Migrations | goose; new identity tables; `associations` + ЄДРПОУ (unique, nullable), address, state; `board_member_revisions` + position, decision reference; `finance.contracts` + `association_id` in 3 steps; `accounts.association_id` dropped in 3 steps; RLS policies and roles | ADR-0003, ADR-0004, ADR-0005, ADR-0006 |
| Audit / immutability | filings, checks, decisions and state changes are insert-only rows with author and time; the extract table is insert-only by privilege (no trigger); a decided application changes only by a new filing or decision | ADR-0005, PRD §6 Integrity |
| Files | extract: PDF or signed container ≤ 5 MB, content-sniffed, stored as uploaded with sha256, downloaded only by the оператор and the applicant as attachment; case files unchanged (board-protocol ADR-0005) | ADR-0005, PRD §6 |
| Storage connection | in-app OAuth consent by a current член правління; refused when the Google address is a member's sign-in address | ADR-0008 |
| Dates | terms are calendar dates in Europe/Kyiv; previous голова's term ends the day before the new term start, no later than the day before approval (AC-19) | PRD AC-19 |
| Observability | OTel spans; metrics in §7; logs never carry the extract, the РНОКПП or the reason text | repo convention |
| i18n | UI in Ukrainian; messages of PRD §5.1 verbatim | CONTEXT.md |
| Rate limiting | <!-- N/A: ~6 filings a year; probing is watched through identity.access.denied --> N/A | — |

## 9. Architecture decisions

| # | Decision | Status | Section | Blast radius (I / C / A / D) |
|---|---|---|---|---|
| <NNNN> | <decision as a sentence> | accepted | §<N> | <✓ / – / ✓ / ✓> |

ADR files: `docs/features/<slug>/adr/NNNN-<decision>.md`. Repo-level ADRs stay in `docs/adr/`.

## 10. Quality requirements

**QG-1. <quality>**
- **When:** <trigger>
- **Then:** <PRD §6 number verbatim>
- **How verify:** <testdb integration test / Playwright e2e / k6 / OTel metric name>

**QG-2. <quality>** — <...>

**QG-3. <quality>** — <...>

## 11. Risks and technical debt

| Risk / debt | Severity | Mitigation | Owner |
|---|---|---|---|
| <migration of existing records (PRD §9)> | <Low/Medium/High> | <...> | <...> |
| Open architectural decision: <headline> | Open question | Resolve before <stage or date>; <reason> | <owner> |

**Accepted debt:** <...>

## 12. Glossary

<!-- Terms from CONTEXT.md used in this SAD + technical terms of this feature. New domain terms go
     to CONTEXT.md first. -->

| Term | Meaning |
|---|---|
| <term> | <meaning> |
