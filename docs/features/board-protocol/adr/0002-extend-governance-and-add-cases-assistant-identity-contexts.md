---
status: Accepted
owner: "Andrii Opryshko"
reviewers: []
updated_at: "2026-10-25"
feature_size: L
stage: "04-05"
ticket: ""
---

# 0002 — Extend governance with board decisions and add cases, assistant and identity contexts

- **Status:** Accepted
- **Date:** 2026-10-25
- **Deciders:** Andrii Opryshko (Architect), with Claude during the architecture-design pass

## Context

The core service has three bounded contexts: `system`, `finance` (contracts) and `governance` (board meetings only). board-protocol adds the board roster, правила статуту, шаблони формулювання, protocol drafts with слоти рішення, attendance, неучасть у голосуванні, попереднє голосування, блокуючі перевірки with overrides, the реєстр протоколів, справи with document versions, the помічник протоколу backed by a text-generation service, sign-in, and a джерело коштів plus approving decision on finance contracts (PRD §4, §8). Context boundaries are enforced by depguard (repo ADR 0002); moving tables and code between contexts after data accumulates is weeks of work.

## Decision drivers

- QG-1 / QG-2: locking must capture roster, rule versions, attendance, votes and computed results atomically; quorum is computed from all of them together.
- PRD §3 non-goal "The autonomous part of approach B" and AC-28: the text-generation service only proposes — best enforced by a dependency rule, not by discipline.
- AC-09, AC-17b, AC-25: protocols and finance contracts both cite case documents; finance must not depend on governance to reach a document.
- §2: each new context costs a spec tag, oapi-codegen config, sqlc entry, depguard rule and a `main.go` line.

## Considered options

1. **Extend `governance` + new `cases`, `assistant`, `identity`; extend `finance`.**
2. **Everything in `governance` + new `identity`** — cases, documents and the assistant live inside governance.
3. **Finer split** — `board` (roster, rules, templates) and `protocols` (drafts, votes, registry) instead of one `governance`, plus `cases`, `assistant`, `identity`.

## Decision outcome

**Chosen:** Option 1. Everything a board decision is computed from stays in one context, so a lock is one local transaction; documents get a context both protocols and finance can cite; and the assistant's separate context lets depguard forbid `governance` from importing the model client at all. Option 3 would push every quorum computation and every lock across a context boundary.

Ownership:

| Context | Owns |
|---|---|
| `governance` (extended) | meetings (existing), board roster with terms, правила статуту, шаблони формулювання, protocol drafts, agenda items, слоти рішення, attendance, неучасть у голосуванні, wording versions, попереднє голосування, блокуючі перевірки and обходи блокування, protocol numbers, locked snapshots, the реєстр протоколів, decision links (revokes / replaces / refers to) |
| `cases` (new) | справи, case documents and their versions, file metadata (Drive ids, sha256), text read from documents |
| `assistant` (new) | conversations of the помічник протоколу, proposals (slot values with quoted fragments, connecting text, scan readings), the text-generation port and adapter |
| `identity` (new) | accounts linked to roster persons, sessions, the sign-in flow |
| `finance` (extended) | джерела коштів; each contract's джерело коштів and the id of the decision that approved it |

## Consequences

**Positive**
- The "AI only proposes" boundary is a lint failure if broken: only `assistant` may import the model client; `governance` accepts proposals through its `module.go`.
- A lock and its checks run in one transaction inside `governance`.
- `finance` links a contract to a document version in `cases` without depending on `governance`.

**Negative**
- `governance` becomes the largest context; its `app/` needs internal sub-packages to stay readable.
- Three new contexts × the five wiring steps of CLAUDE.md, three new depguard rules and three new spec tags.
- Cross-context reads (governance → cases for document versions, governance → finance for related contracts, governance ↔ identity for the current member) need explicit `module.go` facades.

**Neutral**
- `board` and `protocols` can still be split out of `governance` later if it grows; the snapshot format (ADR-0007) does not depend on the split.

## Links

- PRD: [[../PRD.md]] §4, §8
- SAD: [[../sad.md]] §4, §5
- Related ADR: [[0004-use-gemini-api-for-document-reading-and-drafting]], [[0007-lock-protocols-as-fingerprinted-immutable-snapshots]]
