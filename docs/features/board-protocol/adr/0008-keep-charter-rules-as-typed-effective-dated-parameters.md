---
status: Deferred
owner: "Andrii Opryshko"
reviewers: []
updated_at: "2026-10-25"
feature_size: L
stage: "04-05"
ticket: ""
---

# 0008 — Keep правила статуту as a typed catalog of effective-dated parameters

- **Status:** Deferred to Stage 2
- **Date:** 2026-10-25
- **Deciders:** Andrii Opryshko (Architect), with Claude during the architecture-design pass

> **Deferred to Stage 2 (2026-10-25).** Stage 1 computes no quorum or majority and stores only the current association settings (meeting form, signers, start number — PRD AC-42); effective-dated rules come with the Stage 2 checks (PRD §3). Re-validate this ADR when Stage 2 is specified.

## Context

Quorum, majority, the board's competence limit and whether a non-voting member counts toward a кворум пункту come from the charter and change over time; each check must use the rule in force on the meeting date (AC-01), quorum may be a count or a share of the roster in force (AC-02c), and every change carries a charter or decision reference, author and time and is shown on every protocol that applied it (AC-02b). Rules are part of every locked snapshot (ADR-0007). Charter rules are kept as data so neighbouring associations can follow (PRD §1).

## Decision drivers

- QG-2 (PRD §6 "Quorum and rule accuracy — 100% agreement with hand calculation on the association's past protocols").
- PRD §6.1 abuse case "Relaxing a rule to clear a block": every change referenced, authored and visible.
- Rule entry is done by a volunteer член правління, not a lawyer or developer.

## Considered options

1. **Typed catalog** — a fixed set of rule kinds with typed values, versioned by effective date, evaluated by pure domain functions.
2. **Expression language (CEL)** — each rule stored as an expression string such as `present >= ceil(roster * 2 / 3)`, also effective-dated.

## Decision outcome

**Chosen:** Option 1. Each kind has a fixed meaning, a validated value form and tests, so the regression suite over the historical protocols covers every rule the system can hold; an expression language would put legally significant logic into strings a volunteer types, which tests cannot see in advance.

Kinds (first release): `quorum_size` — `{count: n}` or `{share: "2/3", of: "roster_in_force"}`, rounded up; `majority` — threshold (`more_than_half`, a fraction) and basis (`present_eligible_for_item`, `roster_in_force`); `competence_limit` — amount in kopiyky; `non_voting_counts_for_item_quorum` — boolean (default `false`, PRD §8). Each change is an append-only version: kind, value, `effective_from`, reference (charter clause or decision), author, recorded at. The rule in force for a meeting is the latest version with `effective_from` ≤ meeting date. `governance/domain` computes meeting quorum, кворум пункту and majority as pure functions of (rules in force, roster in force, attendance, неучасть у голосуванні, votes); the regression suite encodes the 30+ historical protocols as fixtures and runs in `mise run test`. Шаблони формулювання follow the same effective-dated versioning (AC-23b).

## Consequences

**Positive**
- Every rule the system can hold is covered by tests; the input form is simple for a non-lawyer.
- Rule versions appear verbatim in snapshots and exports.

**Negative**
- A charter rule outside the catalog (another association's unusual quorum formula) needs a new kind in code and a release.

**Neutral**
- A new kind is additive: older versions and snapshots keep their meaning.

## Links

- PRD: [PRD.md](../PRD.md) §1, §6, §6.1, §8, AC-01, AC-02b, AC-02c, AC-05, AC-05b, AC-06, AC-11, AC-23b
- SAD: [sad.md](../sad.md) §4, §5, §10
- Related ADR: [ADR-0007](./0007-lock-protocols-as-fingerprinted-immutable-snapshots.md)
