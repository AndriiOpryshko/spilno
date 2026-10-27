---
status: Accepted
owner: "Andrii Opryshko"
reviewers: []
updated_at: "2026-10-25"
feature_size: L
stage: "04-05"
ticket: ""
---

# 0014 — Keep board-protocol in governance and add only an identity context

- **Status:** Accepted
- **Date:** 2026-10-25
- **Deciders:** Andrii Opryshko (Architect), with Claude during the Stage 1 rescope
- **Supersedes:** [ADR-0002](./0002-extend-governance-and-add-cases-assistant-identity-contexts.md), [ADR-0012](./0012-integrate-contexts-through-synchronous-module-facades.md)

## Context

The core service has three bounded contexts: `system`, `finance` (contracts) and `governance` (board meetings only). ADR-0002 added `cases`, `assistant` and `identity` and extended `finance` with джерела коштів; ADR-0012 wired five contexts through facades. The Stage 1 rescope of PRD.md (2026-10-25) removes what justified the split: no finance cross-checks (AC-25 is Stage 2), no document versions cited by finance contracts, and the помічник протоколу only proposes into one protocol draft. Context boundaries are enforced by depguard (repo ADR 0002); each new context costs a spec tag, oapi-codegen config, sqlc entry, depguard rule and a `main.go` line, and moving tables between contexts later is expensive.

## Decision drivers

- QG-1, QG-2: a transition or a lock must write text, slots, documents, poll rounds and meeting facts in one local transaction.
- PRD §3 non-goal "The autonomous part of approach B" and AC-28: the text-generation service only proposes — best enforced by a dependency rule, not by discipline.
- Sign-in is a prerequisite with its own data (accounts, sessions, the association's Google connection) that is not board-specific.
- One developer: the fewest contexts that keep a real boundary.

## Considered options

1. **Extend `governance` with everything board-protocol owns + new `identity`; `finance` untouched.**
2. **Keep ADR-0002** — `governance` + new `cases`, `assistant`, `identity`; extended `finance`.
3. **Everything in `governance`, including sign-in and sessions.**

## Decision outcome

**Chosen:** Option 1.

Ownership:

| Context | Owns |
|---|---|
| `governance` (extended) | meetings (existing, + `association_id`), association settings (meeting form, signers, register start number), board roster with terms and sign-in email, справи with document versions and file metadata, protocols (state, revision, reserved number), agenda items, слоти рішення and confirmations, text fragments, conversations and proposals of the помічник протоколу, poll rounds and votes, print copies, meeting facts, locked snapshots, integrity runs |
| `identity` (new) | the association record, accounts (Google subject, verified email), sessions, the association's Google connection (encrypted refresh token) |
| `finance` | unchanged in Stage 1 (its route group is put behind sign-in in `cmd/api`) |

Integration rules (restated from ADR-0012, narrowed to two contexts):
- `identity` → `governance`: `CurrentMember(ctx, email, date)` — is this email a current член правління on that date, and which roster person.
- `governance` → `identity`: `ActorFrom(ctx)` (account, person, association) and `GoogleTokenSource(ctx, association)`.
- The consumer owns the port in its own types; the provider exposes plain-data functions on its `module.go`; `cmd/api` and `cmd/worker` adapt one to the other, so the two contexts never import each other and there is no import cycle.
- Cross-context references are ids only; no foreign keys between the contexts' tables.

Dependency rules in `services/core/.golangci.yml`: a new `identity-private` rule (and the `_test.go` exclusion pattern widened to `(system|finance|governance|identity)-private`), and a `gemini-sdk-in-adapter` rule that denies the Gemini SDK for `$all` except `**/internal/governance/adapters/gemini/**` — only that adapter may call the model ([ADR-0004](./0004-use-gemini-api-for-document-reading-and-drafting.md)).

## Consequences

**Positive**
- Every Stage 1 transition and the lock run in one transaction inside `governance`.
- "The AI only proposes" is a lint failure if broken, without a separate context.
- One new context instead of three; the CLAUDE.md wiring checklist runs once.

**Negative**
- `governance` becomes the largest context; its `app/` needs sub-packages (settings, roster, cases, drafts, assistant, poll, printing, signing, withdrawal, registry) to stay readable.
- If a later feature needs справи outside board decisions (e.g. finance contracts citing documents), they must be extracted into their own context then.

**Neutral**
- Stage 2 finance cross-checks (AC-25) add a governance → finance read facade under the same rules.

## Links

- PRD: [PRD.md](../PRD.md) §1, §3, §6.1, AC-16, AC-19, AC-26, AC-28, AC-29
- SAD: [sad.md](../sad.md) §4, §5, §8
- Related ADR: supersedes [ADR-0002](./0002-extend-governance-and-add-cases-assistant-identity-contexts.md) and [ADR-0012](./0012-integrate-contexts-through-synchronous-module-facades.md); [ADR-0003](./0003-sign-in-with-oidc-and-postgres-sessions.md), [ADR-0004](./0004-use-gemini-api-for-document-reading-and-drafting.md), [ADR-0015](./0015-scope-owned-rows-to-one-association-with-explicit-query-filters.md)
