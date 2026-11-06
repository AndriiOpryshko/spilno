---
status: accepted
date: 2026-11-05
feature: association-onboarding
sad_section: "§4"
deciders: ["Andrii Opryshko", "Claude (ao-architecture)"]
---

# 0005 Check the signed extract manually and keep the uploaded file in identity

## Context and problem statement

Proof of office is a витяг з ЄДР signed with the голова's КЕП (PRD US-01, AC-03). Ukrainian КЕП uses DSTU 4145 in CAdES/ASiC, with no Go library; automatic checking needs a sidecar or a third-party service. The product owner decided at this pass: «вся перевірка — ручна і за кадром». The extract must survive a change of the association's storage, and the association has no storage at filing time (PRD §9).

## Decision drivers

- PRD AC-03, AC-03b, AC-10, AC-29: approval needs a recorded check; every value shows who entered it.
- PRD §6: extract ≤ 5 MB; retention ≥ 5 years, never hard-deleted.
- `.claude/rules/migrations.md`: no triggers outside board-protocol ADR-0007.
- ~6 filings a year.

## Blast radius

- Irreversible: yes — where legal proof is kept for years.
- Multi-context: no — identity only.
- Legitimate alternative: yes — automatic checking through a sidecar or an external API; storage in a Drive of the оператор.
- Touches the database: yes — new identity tables, a file column, retention.

## Considered options

1. **Manual check by the оператор outside Спільно; the file stored as uploaded in an identity table (`bytea` + sha256, size, media type); the check stored as its own row (signer, signing date, керівник, the extract's ЄДРПОУ and date, оператор, time); the app role has `SELECT, INSERT` only on these tables.**
2. **A `qes-verifier` sidecar with a DSTU-capable library**, the file in Postgres.
3. **A third-party verification API**, the file in Postgres.
4. **Manual check, file in the оператор's Google Drive**, metadata in Postgres.

## Decision outcome

**Chosen:** Option 1. No cryptography or extra process to run; the file and its check commit in one transaction and are covered by the nightly backup. Insert-only by privilege avoids a trigger.

### Consequences

- Good: the chain «голова → decision → check → stored file» is complete and immutable; nothing personal leaves Спільно.
- Bad: correctness rests on the оператор; the decision time includes a manual step (PRD §7 KPI).
- Neutral: automatic checking can be added later behind an `identity/ports` interface that pre-fills the same check row.

### Migration and rollback

- Migration: new identity tables for applications, filings, extracts, checks and decisions; grants `SELECT, INSERT` to `spilno_app` on the insert-only ones.
- Rollback: down migration drops the tables only before the first real filing; after that, rollback keeps the tables (decisions and extracts are never deleted).

## Links

- PRD: [PRD.md](../PRD.md) — US-01, AC-03, AC-03b, AC-10, AC-29, §9
- SAD: [sad.md](../sad.md) §4, §8
- Related ADR: [0003](./0003-enforce-association-isolation-with-postgres-row-level-security.md); board-protocol [ADR-0016](../../board-protocol/adr/0016-back-up-postgres-nightly-off-the-vm-with-quarterly-restore-drills.md)
