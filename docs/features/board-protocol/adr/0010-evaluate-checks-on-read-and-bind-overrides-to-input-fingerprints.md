---
status: Deferred
owner: "Andrii Opryshko"
reviewers: []
updated_at: "2026-10-25"
feature_size: L
stage: "04-05"
ticket: ""
---

# 0010 — Evaluate checks on read and bind each override to the check's input fingerprint

- **Status:** Deferred to Stage 2
- **Date:** 2026-10-25
- **Deciders:** Andrii Opryshko (Architect), with Claude during the architecture-design pass

> **Deferred to Stage 2 (2026-10-25).** Blocking checks and overrides moved to Stage 2 (PRD §3). The only Stage 1 gates are validation, not блокуючі перевірки: an incomplete agenda (AC-04) and an unconfirmed proposed value (AC-21) block printing. Re-validate this ADR when Stage 2 is specified.

## Context

Deterministic блокуючі перевірки stop a draft from being marked готовий до підпису; попередження only advise (PRD §1 ¶4, CONTEXT). A член правління may lift a блокуюча перевірка with an обхід блокування and a written reason (AC-12, AC-13, AC-16), but the override covers only that check and the values it was recorded for — a changed amount or a new check needs a new override (AC-12b). "Number without a confirmed source" is never lifted by an override, only by a source or a person's confirmation (CONTEXT; AC-10, AC-27). Inputs to checks change from many directions: slot edits, attendance, back-dated правила статуту, a replaced case document (AC-09).

## Decision drivers

- QG-2: checks must always reflect the current rules, roster and documents.
- AC-12b: override validity tied to the exact check and values.
- PRD §7 KPI "Override rate"; PRD §6.1 abuse case "Routine overrides" — every override visible with reason.

## Considered options

1. **Evaluate on read; overrides keyed by check key + input fingerprint.**
2. **Persist check results** — recompute and store rows after each change; overrides reference those rows.

## Decision outcome

**Chosen:** Option 1. Checks are a pure function `Evaluate(draft, rules in force, roster in force, document versions) → []Check` in `governance/domain`, run on every read of a draft and inside the lock transaction, so a stale stored check can never let a draft through after a back-dated rule or a replaced document.

Shape:
- Each check has a stable **key** (kind + agenda item id + slot id where relevant) and an **input fingerprint** (SHA-256 over the canonical inputs it used: e.g. confirmed amount and the `competence_limit` rule version id; present members and the `quorum_size` version id).
- Catalog, first release — blocking: `no_meeting_quorum` (AC-05c), `amount_over_competence_limit` (AC-11), `number_without_confirmed_source` (AC-10, AC-27, not overridable), `incomplete_agenda` (AC-04, not overridable). Advisory: `item_without_majority` (AC-05b), `item_without_quorum` (AC-06), `possible_personal_interest` (AC-24), `related_contracts_same_funding_source` (AC-25), `related_valid_decision` (AC-25b), `scan_number_or_date_mismatch` (AC-14d). Each names the rule version and charter clause or the source it links to.
- An override row stores draft id, check key, input fingerprint, reason (non-empty), author (current член правління), time. It applies only while a current check has the same key and fingerprint; otherwise it stays in the draft's history and the check blocks again (AC-12b). Applied overrides go into the lock snapshot (ADR-0007).

## Consequences

**Positive**
- No synchronization of stored checks; correctness follows from the inputs.
- A new check kind is one function plus tests; the override mechanism is generic.

**Negative**
- The fingerprint's canonical input form is a contract: changing it voids existing overrides on open drafts (locked protocols keep theirs in the snapshot).
- Every draft read evaluates all checks, including cross-context reads (finance, related decisions) — must fit the p95 ≤ 1 s budget at ≤ 500 locked protocols.

**Neutral**
- A history of check results is not stored for drafts; the snapshot keeps the final set.

## Links

- PRD: [PRD.md](../PRD.md) §1, §6.1, §7, AC-04, AC-05b, AC-05c, AC-06, AC-10, AC-11, AC-12, AC-12b, AC-13, AC-14d, AC-16, AC-24, AC-25, AC-25b, AC-27
- SAD: [sad.md](../sad.md) §5, §8
- Related ADR: [ADR-0008](./0008-keep-charter-rules-as-typed-effective-dated-parameters.md), [ADR-0012](./0012-integrate-contexts-through-synchronous-module-facades.md)
