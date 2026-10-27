---
status: accepted
date: 2026-10-25
feature: board-protocol
sad_section: "§4"
deciders: ["Andrii Opryshko"]
---

# 0001 Deliver board-protocol as core endpoints, SPA routes in apps/web and a separate worker process

## Context and problem statement

board-protocol needs a browser UI for members of the board, the protocol rules and records in the core API, and slow external work: assistant drafts (≤ 30 s p95), rendering the print copy, Drive read-only restrictions after lock and a nightly integrity check (PRD §6, US-02, US-06, US-07; reading the signed scan joins in Stage 2). SAD §4 decides which runnable containers the feature owns; every downstream stage reads this as `target_surfaces`.

## Decision drivers

- PRD §6: draft from the помічник протоколу ≤ 30 s p95; signed-scan upload ≤ 3 s p95 for 20 MB; availability 99.5 % monthly.
- SAD §2: one Go module, `core` is the only process today; no broker in the stack (repo ADR 0001).
- The web client already exists (`apps/web`) with a governance prototype (`src/features/governance/`).

## Blast radius

- Irreversible: yes — a second deploy unit and the web routes shape tasks, tests and deployment for every later stage.
- Multi-context: yes — `apps/web` and `services/core` (`core` and `worker` binaries).
- Legitimate alternative: yes — run background work as goroutines inside `core`.
- Touches the database: no — the job queue is decided in ADR 0015 (which superseded ADR 0008).

## Considered options

1. **Core endpoints + SPA routes + separate `cmd/worker` process** — slow and external work runs out of the API process.
2. **Core endpoints + SPA routes, background goroutines inside `core`** — one deploy unit; slow AI calls share the API process.

## Decision outcome

**Chosen:** option 1, `target_surfaces: [backend-service, web-frontend, worker]`. The user chose to keep AI, Drive and PDF work out of the API process so that a slow or failing external call never holds API capacity, and so the worker can be restarted or scaled on its own.

### Consequences

- Good: the API stays within the read/write latency targets while Gemini calls take tens of seconds.
- Good: the web reuses the prototype components and shadcn; no new UI stack.
- Bad: a second binary to deploy, monitor and keep in version step with `core` (§7).
- Neutral: the worker imports contexts only through their `module.go`, like `cmd/api`.

### Migration and rollback

N/A — no stored data changes. Removing the worker leaves queued jobs in River's table (ADR 0015).

## Links

- PRD: [PRD.ao-prd.md](../PRD.ao-prd.md) — US-02, US-06, US-07; §6 latency rows
- SAD: [sad.ao-architecture.md](../sad.ao-architecture.md) §4
- Related ADR: [0015 River jobs](./0015-run-background-jobs-in-cmd-worker-with-river-on-postgres.md) · repo [0001 Stack](../../../adr/0001-stack.md)
