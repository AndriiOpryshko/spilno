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

**Technical** (versions from `mise.toml`, `go.mod`, `package.json`, `infra/compose.yaml`):
- <Go x.y, chi, oapi-codegen strict server, sqlc + pgx, goose>
- <Postgres x>
- <React x, TanStack Router, shadcn/ui, generated TS client>

**Repo ADRs that bind this feature:**
- <0001 stack> · <0002 bounded contexts> · <0003 spec-first codegen> · <0004 migrations and queries>

**Conventions** (`CLAUDE.md`): one OpenAPI contract, one tag per context, paths `/<context>/`;
contexts import each other only through `module.go`; problem+json errors; `r.Group`, never
`r.Mount`; generated files are never edited.

**Organisational:** <effort, deadline, team — from PRD or TBD with a §11 row>

**Regulatory / legal:** <personal data of co-owners, record retention, signature law — from PRD §6.1 and brief §18>

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
