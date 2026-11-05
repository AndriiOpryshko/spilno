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
target_surfaces: []
contexts: []
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

<2-3 sentences: what the system does for the association and where its trust boundary is>

| Actor or system | Type | Interaction |
|---|---|---|
| <glossary role> | Person | <...> |
| <text generation service> | System (external) | <wording only> |

```mermaid
C4Context
    title <slug> — System Context

    Person(role1, "<glossary role>", "<intent>")
    System(sys, "<product>", "<one sentence>")
    System_Ext(ext, "<external>", "<one sentence>")

    Rel(role1, sys, "<interaction>", "HTTPS")
    Rel(sys, ext, "<interaction>", "<protocol>")
```

## 4. Solution strategy

**Target surface(s):** `<[backend-service, web-frontend]>`

**Strategic pillars** (each one seeds ADRs; cite a §2 constraint or a past incident/limitation):

1. **<e.g. New `protocols` context, governance stays the owner of meetings>** — <2-3 sentences; constraint: repo ADR 0002>.
2. **<e.g. Signed records are immutable; corrections are new records>** — <2-3 sentences; PRD §10>.
3. **<e.g. Deterministic checks in Go, AI writes wording only>** — <2-3 sentences; brief §16>.
4. **<UI architecture: SPA route under apps/web reusing shadcn components>** — <per UI surface>.

## 5. Building block view

<one paragraph: hexagonal per context, why the feature extends or adds a context>

**Context ownership** (fills frontmatter `contexts`):

| Context | New? | Owns (aggregates, tables) | Provides to others via module.go | Uses from others |
|---|---|---|---|---|
| <protocols> | new | <protocol, agenda item, vote> | <...> | <governance: meetings, roster> |
| <governance> | existing | <meeting, board term> | <meeting by id, roster on date> | — |

**Layout of the new context:**

```
services/core/internal/<bc>/
├── domain/              <entities, sentinel errors; stdlib only>
├── app/                 <use cases; authorization against the board term on the action date>
├── ports/               <repositories, other contexts' ports>
├── adapters/http/       <StrictServerInterface; problem+json>
├── adapters/postgres/   <queries/ → sqlcgen/>
└── module.go            <wiring>
```

New-context wiring (CLAUDE.md): spec tag · `adapters/http/gen/oapi-codegen.yaml` · `sqlc.yaml` entry ·
`<bc>-private` depguard rule · `cmd/api/main.go`.

```mermaid
C4Container
    title <slug> — Containers

    Person(member, "<glossary role>")

    System_Boundary(sys, "<product>") {
        Container(web, "Web client", "React", "<screens>")
        Container_Boundary(core, "Core API — Go modular monolith") {
            Container(ctx_new, "<new context>", "Go bounded context", "<owns>")
            Container(ctx_ref, "<existing context>", "Go bounded context", "<provides>")
        }
        ContainerDb(db, "Database", "Postgres", "<tables per context>")
        ContainerDb(files, "File storage", "<per ADR>", "<signed scans>")
    }

    System_Ext(otel, "Observability", "OpenTelemetry → Grafana LGTM")

    Rel(member, web, "uses", "HTTPS")
    Rel(web, ctx_new, "calls", "OpenAPI, JSON")
    Rel(ctx_new, ctx_ref, "reads", "through module.go")
    Rel(ctx_new, db, "reads/writes own tables", "SQL")
    Rel(ctx_new, files, "stores files")
    Rel(ctx_new, otel, "traces, metrics", "OTLP")
```

## 6. Runtime view

<!-- Seed the primary flows here; complete-sequence-diagrams adds the rest. Participants are §5
     container aliases. Messages are semantic — no HTTP verbs, paths or status codes.
     With an ao-prd PRD: one flow per critical §10 transition (e.g. mark ready, sign and lock,
     refused edit after lock) and name the PRD AC each flow covers. -->

**Flow 1: <happy path> — covers AC-<NN>**

```mermaid
sequenceDiagram
    actor M as <glossary role>
    participant W as web
    participant P as <new context>
    participant G as <existing context>
    participant D as db
    M->>W: <action>
    W->>P: <command>
    P->>G: <read through module.go>
    P->>D: <write in one transaction>
    P-->>W: <result>
```

**Flow 2: <refused transition / failure mode> — covers AC-<NN>**

## 7. Deployment view

<!-- Usually N/A for a feature inside the core process: <!-- N/A: reuses the core deployment unit -->
     plus one line of scaling numbers. Fill when the feature adds a process, a queue or storage. -->

<topology; storage for files; backup and restore numbers from PRD §6>

## 8. Crosscutting concepts

| Concept | Convention for this feature | Where defined |
|---|---|---|
| Errors | problem+json; codes from PRD §5.1 (`<context>.<error_name>`) | `internal/platform/httpapi`, PRD §5.1 |
| Authorization | <roles × actions from PRD §10, checked in `app/`> | PRD §10 |
| Validation | OpenAPI validator (422) + domain `Validate()` | CLAUDE.md |
| IDs | UUID | repo convention |
| Migrations | goose; <what this feature migrates, link ADR> | ADR 0004 |
| Audit / immutability | <how locked records are protected> | ADR <NNNN> |
| Files | <where scans live, size limit from PRD §6> | ADR <NNNN> |
| Observability | OTel spans + request-duration histogram | repo convention |
| i18n | UI in Ukrainian | — |

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
