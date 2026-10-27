---
status: accepted
date: 2026-10-25
feature: board-protocol
sad_section: "§7"
deciders: ["Andrii Opryshko"]
---

# 0014 Run core, worker and Postgres on one VM with docker compose

## Context and problem statement

The repo has only a local `infra/compose.yaml`; board-protocol is the first feature that real associations will use, and it adds a second process (ADR 0001). PRD §6 sets availability 99.5 % monthly, RPO ≤ 24 h and RTO ≤ 4 h. The hosting choice applies to every feature, so this ADR is a candidate to be promoted to `docs/adr/` as repo ADR 0005.

## Decision drivers

- PRD §6: 99.5 % monthly, RPO ≤ 24 h (daily backup), RTO ≤ 4 h (quarterly restore drill).
- One developer; ~13 associations, ~20 protocols each a year.
- SAD §2: the stack already runs under docker compose locally.

## Blast radius

- Irreversible: no — the same images move to another host in less than three days.
- Multi-context: yes — every context, web, `core`, `worker` and the database.
- Legitimate alternative: yes — Google Cloud Run with Cloud SQL.
- Touches the database: yes — where it runs, how it is backed up and restored.

## Considered options

1. **One VM with docker compose** — caddy, core, worker, postgres, otel-lgtm; nightly encrypted dump off the VM.
2. **Google Cloud Run + Cloud SQL** — managed containers and database with automatic backups.
3. **Leave hosting open** as a repo-level open question.

## Decision outcome

**Chosen:** option 1. The compose file reuses the local service definitions; `core` and `worker` are one image with two commands; caddy terminates TLS and serves the web bundle. Postgres is dumped nightly with `pg_dump`, encrypted, to storage outside the VM, and a restore onto a fresh VM is drilled quarterly. Files on Google Drive are not copied out (the user chose to rely on Google; accepted debt in SAD §11).

### Consequences

- Good: cheapest option; production equals local compose; no cloud-specific code.
- Bad: one host is a single point of failure; a long host outage eats the 99.5 % budget; OS and Postgres updates are manual.
- Bad: backups and restores are scripts the developer owns and must test.
- Neutral: Google is still a dependency through sign-in, Drive and Gemini.

### Migration and rollback

- Migration: none for data; `cmd/api` applies pending goose migrations on start (repo ADR 0004), and the worker starts after the API is healthy so River's schema exists (ADR 0015).
- Rollback: restore the latest dump on a new host; locked protocols are never deleted.

## Links

- PRD: [PRD.ao-prd.md](../PRD.ao-prd.md) — US-09, AC-36 (the реєстр протоколів must be reachable and restorable); §6 Availability, Durability
- SAD: [sad.ao-architecture.md](../sad.ao-architecture.md) §7
- Related ADR: [0001 Surfaces](./0001-deliver-board-protocol-as-core-endpoints-spa-routes-and-a-worker-process.md) · repo [0001 Stack](../../../adr/0001-stack.md)
