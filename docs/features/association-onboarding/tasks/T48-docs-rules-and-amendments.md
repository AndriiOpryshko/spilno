---
id: T48
title: "Update the repo rules and environment docs, amend the board-protocol documents and write the оператор's runbook"
layer: "docs"
deps: ["T20", "T33"]
acs: ["AC-03", "AC-05", "AC-12"]
files_hint:
  - "CLAUDE.md"
  - ".claude/rules/openapi.md"
  - "README.md"
  - "docs/features/board-protocol/PRD.md"
  - "docs/features/board-protocol/adr/"
  - "docs/features/board-protocol/tasks/tracker.md"
  - "docs/features/association-onboarding/data-model.md"
  - "docs/runbooks/operator-extract-check.md"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T48 — Update the repo rules and environment docs, amend the board-protocol documents and write the оператор's runbook

## Why

[PRD](../PRD.md) §1 «Traceability and deviations» and §9: this feature amends board-protocol PRD §3 and AC-19 and replaces its tasks T08 and T12; [sad](../sad.md) §2 lists the board-protocol ADRs it changes; [ADR-0005](../adr/0005-check-the-signed-extract-manually-and-keep-the-uploaded-file-in-identity.md): the check of the extract is a manual procedure and needs to be written down.

**Satisfies:** [PRD §5](../PRD.md) AC-03, AC-05, AC-12 · **Blocked by:** [T20](./T20-promote-row-level-security.md), [T33](./T33-connect-storage-in-app.md)

## What

- `CLAUDE.md`: association-scoped paths, the two database roles, `InAssociation`, the e2e configuration, the baseline update command. `.claude/rules/openapi.md`: paths `/<context>/associations/{association_id}/…` and the `MemberOrDisabled` 403.
- `README.md`: `MIGRATIONS_DATABASE_URL`, `SPILNO_OPERATOR_EMAIL`, the storage redirect address, `VITE_OPERATOR_CONTACT_EMAIL`.
- board-protocol: the amendment notes on PRD §3 and AC-19; «superseded in part» lines on ADR-0003, ADR-0005 and ADR-0015 with links here; the tracker rows of T08, T11, T12 and T34 marked replaced.
- `docs/runbooks/operator-extract-check.md`: where to check a КЕП, what to record, when a document must be named (AC-12), what to do with an extract of another code. `data-model.md`: the missing `identity.operator_own_application` row (api-sync-report §D).

## Definition of Done

- [ ] Acceptance: CLAUDE.md, the OpenAPI rule, the README's environment section and the amended board-protocol documents describe the system as built, the оператор's runbook exists, and every link in the edited files resolves.
- [ ] each listed file is changed as described; no code file is touched
- [ ] a link check over the edited files and `docs/features/association-onboarding/` finds no broken link
- [ ] `/ao-sdlc-audit association-onboarding` reports no stale artifact
- [ ] `mise run lint` green

## Notes

The board-protocol tracker has uncommitted edits on `feat/core-google-config` at the time of this breakdown — edit it after that branch is merged.
