---
status: Draft
skill: ao-architecture
owner: "Andrii Opryshko"
reviewers: ["Andrii Opryshko"]
updated_at: "2026-10-25"
feature_size: "L"
prd: "PRD.ao-prd.md"
depth: "medium"
stage: "04-05"
ticket: ""
target_surfaces: []  # filled by the first §4 decision: backend-service | web-frontend | worker | cli | ...
contexts: []         # bounded contexts this feature touches, new ones marked (new) — filled in §5
---

# Software Architecture Document — board-protocol

<!-- ao-architecture: Arc42 12 sections, C4 L1 in §3, C4 L2 in §5. N/A sections carry
     <!-- N/A: <reason> -->. Numbers in §10 come verbatim from PRD §6.
     Inputs: PRD.md (§5.1 / §9 / §10 when written by ao-prd), CONTEXT.md, docs/adr/, CLAUDE.md,
     references/repo-practice.md. -->

> **Inputs:** [PRD](./PRD.ao-prd.md) · [CONTEXT](./CONTEXT.md) · repo ADRs [`docs/adr/`](../../adr/) · reference contexts `services/core/internal/governance`, `services/core/internal/finance`

## 1. Introduction and goals

**Intent.** board-protocol lets a член правління turn a кошторис робіт into a printable протокол правління in one sitting. The protocol is assembled from typed слоти рішення: an amount, counterparty or джерело коштів reaches print only with a document of the справа and a person's confirmation, and кворум, кворум пункту, majority, the competence limit and the allowed meeting form are computed from the board in force and the правила статуту in force on the meeting date. The помічник протоколу writes wording and попередження and proposes amounts, but is never the source of a number. Попереднє голосування is advisory; printing is gated on блокуючі перевірки, some of which a member may lift with an обхід блокування and a reason. Uploading the signed scan locks the protocol in the реєстр протоколів with its справа, votes and overrides; a draft that is not signed is closed as знятий з розгляду and kept (PRD §1-§2, §10).

**Top-3 quality goals** (full scenarios in §10):

1. **Integrity of signed protocols** — a підписаний protocol changes only through a new protocol that references it; 0 mismatches between printed numbers and confirmed slots (PRD §6 Integrity, AC-30, AC-30b).
2. **Traceability and rule accuracy** — every number shows its справа document and the person who confirmed it; кворум, majority and the competence limit are computed from the rule in force on the meeting date (PRD §2, AC-02, AC-09, AC-19).
3. **Confidentiality within one association** — only a член правління with a term in force acts on or reads the association's records; to anyone from another association a protocol is "not found" (PRD §6.1, AC-32, AC-33).

Latency, durability (RPO ≤ 24 h, RTO ≤ 4 h) and availability (99.5 %) from PRD §6 are additional scenarios in §10; they yield to the three goals above when they conflict.

**Stakeholders.**

| Role (glossary) | Interest | Sign-off? |
|---|---|---|
| член правління | Keeps the board and правила статуту, drafts with the помічник протоколу, confirms slots, casts own попереднє голосування, overrides, marks ready, uploads the signed copy, withdraws, searches the реєстр протоколів | No |
| головуючий | Chairs the meeting and signs the paper protocol; a slot of the protocol, not a separate permission (PRD §1 ¶4) | No |
| голова правління | Answers the charter open questions (meeting forms, competence limit, кворум vs majority) | No |
| ревізійна комісія | Checks spending against protocols at the annual meeting; no access in this feature (PRD §3) | No |
| управляюча компанія | Issues the кошторис робіт and акт виконаних робіт held in a справа; no access | No |
| співвласник | Receives printed protocols as today; no access | No |
| Tech lead | SAD and ADR approval; security review (PRD §6.1) | Yes |

## 2. Constraints

**Technical** (versions from `mise.toml`, `services/core/go.mod`, `apps/web/package.json`, `infra/compose.yaml`):
- Go 1.27.3, chi v5, oapi-codegen v2.8 strict server, sqlc + pgx/v5, otelpgx, goose (repo ADR 0001, 0003, 0004).
- Postgres 18.6 (`postgres:18.6-alpine`), one database for the whole `core` process.
- React 19.3, TypeScript 6.0.3, Vite 8, Tailwind 4.3, shadcn/ui, TanStack Router 1.170 + Query 5.104, orval-generated hooks.
- Observability: OpenTelemetry → `grafana/otel-lgtm` 0.37 (traces, metrics, logs via `otelslog`).

**Repo ADRs that bind this feature:** [0001 Stack](../../adr/0001-stack.md) · [0002 Bounded contexts](../../adr/0002-bounded-contexts.md) · [0003 Spec-first codegen](../../adr/0003-spec-first-codegen.md) · [0004 Migrations and queries](../../adr/0004-migrations-and-queries.md). None is superseded by this feature.

**Starting point of the repo** (scan 2026-10-25): contexts `system`, `finance` (contracts) and `governance` (meetings: kind, held date, optional end and location, `quorum_required`); no sign-in, no association scope, no file storage, no outbound HTTP client, no background jobs. Migrations live in one directory `services/core/migrations/` (`0000N_*.sql`), not per context; tables use CHECK constraints and the shared `set_updated_at` trigger.

**Conventions** (`CLAUDE.md`): one OpenAPI contract, one tag per context, paths `/<context>/`; contexts import each other only through `module.go` (depguard `<bc>-private`); `domain/` imports only stdlib and uuid; problem+json errors; `r.Group`, never `r.Mount`; generated files are never edited; integration tests on `testdb.New(t)`.

**Organisational:** one developer; rollout Stage 0 (sign-in, association scope) → Stage 1 (US-01-US-09) → Stage 2 (US-10-US-12) (PRD §9). There is no fixed project deadline: the date of an association's annual general meeting, which the KPI "100 % of spending decisions locked before the annual meeting" is measured against, is the association's own setting, kept with the association record (Stage 0).

**Regulatory / legal** (PRD §6.1, §6, §3):
- Data class Confidential: names and terms of board members, signatures on scans, co-owners' names and apartment numbers in case documents, reasons for неучасть у голосуванні (personal data under the Law of Ukraine "On Personal Data Protection").
- Signed and withdrawn protocols kept ≥ 5 years and never hard-deleted.
- The binding signature is on paper; no electronic signing, no binding electronic vote.
- Case documents go to a text-generation service only under an agreement that excludes training on the data.
- Security review required before Stage 1 release (first feature with authorization, personal data and an external AI service).

## 3. Context and scope

spilno is the association's own record of board governance: a член правління prepares, checks and locks протоколи правління against the board roster, the правила статуту and the documents of each справа, and finds past decisions in the реєстр протоколів. The trust boundary is the association: everything inside spilno is visible only to the association's members of the board with a term in force; the only data that leaves it is case-document content sent to the text-generation service for wording and reading, and printed paper protocols that people carry out by hand. How a член правління signs in is decided in §4; if it adds an external system, §5 shows it.

| Actor or system | Type | Interaction |
|---|---|---|
| член правління | Person | Uses spilno for every protocol action; also acts as головуючий of a meeting (a slot, not a role in the system) |
| управляюча компанія | Person (external) | Hands over кошторис робіт, invoices and акти виконаних робіт; a член правління attaches them to a справа. No access |
| ревізійна комісія | Person (external) | Checks signed paper protocols at the annual meeting. No access in this feature (PRD §3) |
| співвласник | Person (external) | Receives printed protocols as today. No access |
| Text generation service | System (external) | Receives case documents and the draft; returns wording, попередження, proposed amounts and the number/date read from a signed scan. Never the source of a confirmed number |

```mermaid
C4Context
    title board-protocol — System Context

    Person(member, "член правління", "Drafts, confirms, votes, marks ready, uploads the signed copy, searches the реєстр протоколів")
    Person_Ext(uk, "управляюча компанія", "Issues кошторис робіт and акти виконаних робіт")
    Person_Ext(audit, "ревізійна комісія", "Checks signed paper protocols; no access")
    System(spilno, "spilno", "HOA management app: meetings, contracts, board protocols")
    System_Ext(ai, "Text generation service", "Wording, попередження, proposed amounts, reads number and date from a scan")

    Rel(uk, member, "Hands over estimates and acts", "paper, email")
    Rel(member, spilno, "Prepares, checks and locks protocols", "HTTPS")
    Rel(spilno, ai, "Sends case documents, receives wording and proposals", "HTTPS")
    Rel(member, audit, "Presents signed protocols", "paper")
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
