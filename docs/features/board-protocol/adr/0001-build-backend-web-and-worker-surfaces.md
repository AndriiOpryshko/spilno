---
status: Accepted
owner: "Andrii Opryshko"
reviewers: []
updated_at: "2026-10-25"
feature_size: L
stage: "04-05"
ticket: ""
---

# 0001 — Build board-protocol as backend service, web frontend and a separate worker

- **Status:** Accepted
- **Date:** 2026-10-25
- **Deciders:** Andrii Opryshko (Architect), with Claude during the architecture-design pass

## Context

PRD §1-§4 describe члени правління working in a browser against a server that keeps cases, drafts and the реєстр протоколів. Several operations are slow by nature: a proposed protocol from a conversation (PRD §6: p95 ≤ 60 s for up to 10 attached documents), reading the protocol number and date from a signed scan (p95 ≤ 20 s including that read), copying files in Google Drive, and the periodic integrity check of every locked protocol. The repo today runs one process, `services/core/cmd/api`, plus the `apps/web` SPA. The surface choice decides which C4 containers §5 draws and what api-forge, break-tasks and plan-tests generate.

## Decision drivers

- PRD §6 NFR: open a draft or a registry decision p95 ≤ 1 s — request handling must not be held up by 60-second model calls.
- PRD §6 NFR: "Integrity of locked protocols — periodic fingerprint comparison of every locked protocol, alert on mismatch" — needs a scheduled job.
- PRD §6 NFR: availability 99% monthly. (Durability RPO ≤ 24 h is deliberately not met in the first release — no database backup; SAD §1 decision override, §7, §11.)
- §2: modular monolith, one Go module (repo ADR 0002); SPA web app (repo ADR 0001).

## Considered options

1. **backend-service + web-frontend, background work in the API process** — goroutine workers inside `cmd/api`.
2. **backend-service + web-frontend + worker** — a second process, `cmd/worker`, built from the same module, runs all slow and scheduled work.
3. **Add mobile-app** — a native app for preliminary voting; the PRD does not ask for it and a responsive SPA covers it.

## Decision outcome

**Chosen:** Option 2 — `target_surfaces: [backend-service, web-frontend, worker]`. Slow model calls, Drive transfers and nightly jobs run in their own process, so a stuck or memory-heavy job (50 MB scans) never degrades the API's ≤ 1 s paths, and the worker can be restarted or scaled independently. Both processes share the contexts and the database; the worker reaches contexts only through their `module.go`, like the API.

## Consequences

**Positive**
- API latency is isolated from model latency and from retries.
- Scheduled jobs (the nightly integrity check; a database backup once one is added) have a natural home with their own logs and metrics.
- break-tasks gets `ui` tasks for the SPA and async tasks for the worker; plan-tests adds component / visual-regression / e2e-through-UI tiers and worker integration tests.

**Negative**
- Two deployable processes to run, monitor and keep at the same version; `mise run dev` (mprocs) gains a worker entry.
- api-forge must also describe the job contracts (`contracts/events.md`) for the worker surface.

**Neutral**
- The worker could later fold back into the API process by starting the same job runner there; the job contracts do not change.

## Links

- PRD: [[../PRD.md]] §1, §6
- SAD: [[../sad.md]] §4, §5, §7
- Related ADR: [[0006-run-background-work-in-worker-via-river-queue]]
