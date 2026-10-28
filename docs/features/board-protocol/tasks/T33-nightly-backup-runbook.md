---
id: T33
title: "Back up Postgres nightly off the VM and write the restore-drill runbook"
layer: "infra"
deps: []
acs: []
nfr: ["PRD §6 Durability (RPO ≤ 24 h, RTO ≤ 4 h)", "QG-8", "ADR-0016"]
files_hint:
  - "infra/backup/"
  - "docs/runbooks/backup-restore.md"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T33 — Back up Postgres nightly off the VM and write the restore-drill runbook

## Why

PRD §6 Durability (RPO ≤ 24 h, RTO ≤ 4 h, quarterly drill); [ADR-0016](../adr/0016-back-up-postgres-nightly-off-the-vm-with-quarterly-restore-drills.md); [sad §6 Flow 11](../sad.md).

**Satisfies:** [PRD §5](../PRD.md) — (enabling / NFR task) · **Blocked by:** —

## What

- Backup script + schedule (compose service or cron on the VM); alert on failure to two members.
- Runbook: restore steps and a drill record table.

## Definition of Done

- [ ] Acceptance: A script dumps the database with `pg_dump`, encrypts it with `age` and pushes it off-site; a local run against the dev database restores into a fresh container and the runbook records that drill.
- [ ] local drill: dump → encrypt → restore → row counts match
- [ ] runbook committed with the first drill record
- [ ] `mise run lint` and `mise run test` green

## Notes

NFR task, no PRD §5 AC. Independent of the app code — parallel branch from day one.
