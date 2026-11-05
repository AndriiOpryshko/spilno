---
status: accepted
date: 2026-11-05
feature: association-onboarding
sad_section: "§4"
deciders: ["Andrii Opryshko", "Claude (ao-architecture, easy depth, assumptions ledger accepted)"]
---

# 0001 Serve the operator back office as an operator-only section of the web app

## Context and problem statement

The оператор reviews заявки and запити на передачу, records the check of the extract, decides, and disables or enables associations (PRD US-03, US-05, US-06). PRD §8 left open where this back office lives. The Stage 1 screens spec drew it as «Proposed — outside PRD Stage 1» (BR-21, section `51:6511`). It must never show an association's records to the оператор (AC-16b).

## Decision drivers

- PRD AC-16, AC-16b: only the оператор reaches the queue; the оператор reaches no association's records.
- PRD §3: one оператор, no operator roles.
- SAD §2: one developer; one SPA and one contract already exist (repo ADR 0001, 0003).

## Blast radius

- Irreversible: yes — a separate surface would mean a second build, deploy and auth path; merging later is days of work.
- Multi-context: yes — web and core (`identity` operator operations).
- Legitimate alternative: yes — a separate admin app, or a CLI.
- Touches the database: no.

## Considered options

1. **An `/operator/` section of the existing SPA, backed by `identity` operations under `/identity/operator/...`, gated by the session address equal to `SPILNO_OPERATOR_EMAIL`.**
2. **A separate admin web app** on its own origin with its own sign-in.
3. **A CLI for the developer** (T08/T12 style).

## Decision outcome

**Chosen:** Option 1. One build, one contract, one sign-in. The operator check is a single comparison in the identity middleware, and the оператор gets no association path at all.

### Consequences

- Good: reuses the shadcn components, the Figma back-office frames and the generated client; e2e covers it with the rest.
- Bad: the operator code ships to every browser (hidden routes, no data without the server check).
- Neutral: a second оператор later means a table of operators instead of one environment variable.

### Migration and rollback

N/A — no stored data changes.

## Links

- PRD: [PRD.md](../PRD.md) — US-03, US-05, US-06, AC-16, AC-16b, §8
- SAD: [sad.md](../sad.md) §4
- Related ADR: [0004](./0004-let-any-verified-person-sign-in-and-derive-memberships-from-rosters.md); repo [0003 Spec-first codegen](../../../adr/0003-spec-first-codegen.md)
