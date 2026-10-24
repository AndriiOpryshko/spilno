---
status: Accepted
owner: "Andrii Opryshko"
reviewers: []
updated_at: "2026-10-25"
feature_size: L
stage: "04-05"
ticket: ""
---

# 0012 — Integrate contexts through synchronous module.go facades wired in the composition root

- **Status:** Accepted
- **Date:** 2026-10-25
- **Deciders:** Andrii Opryshko (Architect), with Claude during the architecture-design pass

## Context

With five contexts (ADR-0002), several need each other's data: identity asks governance whether a person is a current член правління; governance reads document versions from cases (AC-09), related contracts by джерело коштів from finance (AC-25); finance asks governance whether an approving decision is still чинне; assistant reads a case's documents and hands proposals to governance. Repo ADR 0002 allows importing only a context's `module.go`. Governance and finance read each other, which in Go would be an import cycle if their packages imported one another.

## Decision drivers

- QG-2, AC-09: a slot must stop counting as confirmed the moment its document is replaced — no window where a stale view lets a draft through.
- Repo ADR 0002: no imports of another context's sub-packages; depguard rules per context.
- One developer: the least machinery that keeps boundaries.

## Considered options

1. **Synchronous in-process facades** — the consumer declares a port in its own types; the provider's `module.go` exports query and command functions; the composition root adapts one to the other.
2. **Domain events through an outbox** — contexts publish events in the writing transaction (carried as River jobs); consumers keep local copies of what they need.

## Decision outcome

**Chosen:** Option 1. Reads are always current, tests replace a port with a fake, and there is no second copy of data to drift.

Rules:
- The consumer owns the port (`governance/ports.CaseDocuments`, `governance/ports.FundingContracts`, `finance/ports.DecisionValidity`, `identity/ports.Membership`, `assistant/ports.Drafts`) in its own types.
- The provider exposes plain-data functions on its `module.go` and never returns domain types.
- Contexts never import each other. `cmd/api` and `cmd/worker` (the composition roots) build thin adapters from provider functions to consumer ports when wiring modules, so there are no import cycles and the depguard rules stay unchanged.
- Cross-context references are ids only (UUID); no foreign keys between contexts' tables.
- "Document replaced" is detected on read: a slot keeps the document version it was confirmed against and compares it with the latest version (ADR-0010).
- Asynchronous hand-offs exist only as River jobs (ADR-0006).

## Consequences

**Positive**
- No eventual consistency anywhere a check depends on.
- Boundaries stay lint-enforced; no shared tables.

**Negative**
- A draft read makes several in-process calls across contexts; the p95 ≤ 1 s budget must hold (§10).
- Moving a context into its own service later means turning these calls into network calls.
- The composition roots grow an adapter layer that needs its own small tests.

**Neutral**
- Events can be added later for a consumer that tolerates lag, without changing the facades.

## Links

- PRD: [[../PRD.md]] AC-09, AC-16, AC-19, AC-25, AC-25b, AC-26
- SAD: [[../sad.md]] §5, §8
- Related ADR: [[0002-extend-governance-and-add-cases-assistant-identity-contexts]], [[0010-evaluate-checks-on-read-and-bind-overrides-to-input-fingerprints]]
