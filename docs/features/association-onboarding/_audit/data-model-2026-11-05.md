---
skill: ao-data-model
feature: association-onboarding
date: 2026-11-05
---

# Data-model audit — association-onboarding

## Staged paths

20 goose files in `docs/features/association-onboarding/migrations/` (`NNNNN_` prefix), order and PRs in
`MANIFEST.md`, three backfill notes. **Staged, not live**: nothing was written to `services/core/migrations/`.

## Inputs and scan

- Gate: `PRD.md` (ao-prd, Approved, amended 58e6280) and `sad.md` (ao-architecture, Draft) — the only variants.
- Repo scan done inline (no Explore subagent): live migrations `00001`–`00027`, `sqlc.yaml` (4 contexts,
  shared schema), queries per context, the 28 tables that carry `association_id`, `finance.contracts`
  without it, `postgres:18.6-alpine`.

## Rules file vs skill defaults

`.claude/rules/migrations.md` exists, so it wins over the skill's template (step 2). Applied from the rules
file and live practice of `00009`–`00027`, against the skill's defaults:

| Skill default | Applied instead | Source |
|---|---|---|
| `uuid DEFAULT gen_random_uuid()` | `UUID PRIMARY KEY`, UUID v7 from Go | rules «Defaults», SAD §8 |
| `text` + `CHECK (char_length)` | `VARCHAR(N)` | rules «Strings» |
| `updated_at` + `set_updated_at()` | `created_at` only; changing columns like `protocols.state` | rules «Audit columns» |
| closed-set `CHECK (state IN …)` | no CHECK; closed sets as Go constants | rules «No CHECK on business invariants» |
| no `IF NOT EXISTS` | `IF NOT EXISTS` on `CREATE TABLE` / `CREATE INDEX` / `ADD COLUMN` | rules «Required constraints» |
| `<table>_<cols>_idx` / `_key` | `idx_<table>_<cols>` / `uq_<table>_<cols>` | live migrations |

Self-check items 4 (`VARCHAR(` → FAIL) and 2/6 were read against the rules file, not the skill text.

## Divergences and debt found in the live schema

- `associations` (`00004`) predates the rules: `name text CHECK`, no `IF NOT EXISTS` — kept.
- `contracts`, `meetings` (`00002`, `00003`) keep `DEFAULT gen_random_uuid()`, CHECKs and `updated_at`
  triggers (documented exception in the rules file); `meetings.kind DEFAULT 'REGULAR'` is business-default debt.
- `finance.contracts` had no `association_id` — a cross-association leak once a second association joins;
  fixed by files 8, 12, 15, 16, 18.
- SAD §7 says «a goose migration creates `spilno_app`»: file 9 creates it as a **NOLOGIN group role**; the
  login role per environment (compose, testdb, production) is granted membership outside migrations, so no
  password lives in git. The SAD wording stands; this detail is for the implementation task.
- ADR-0003 `FORCE ROW LEVEL SECURITY` means the migration owner needs `BYPASSRLS` (the compose superuser
  has it) so later backfills still see every row; production must give the owner role that attribute.

## Expand / contract sequences

| Table.column | Expand | Backfill | Contract |
|---|---|---|---|
| `associations.state` | file 2 (PR 1) | file 10 (PR 2) | file 13 (PR 3) |
| `board_member_revisions.position` | file 6 (PR 1) | file 11 (PR 2) | file 14 (PR 3) |
| `contracts.association_id` | file 8 (PR 1) | file 12 (PR 2) | file 15 + index 16 (PR 3) |
| `accounts.association_id` (drop) | — (code stops reading in the release before) | — | file 20 (PR 4) |

Row-level security (17–19) ships in PR 3 together with the app's switch to a `spilno_app` login and
`platform/db.InAssociation`; earlier, it would hide every row from the app.

## Self-check

```
self-check: 10/10 PASS
  1 naming PASS · 2 Up/Down symmetry PASS · 3 FK indexes PASS · 4 banned features PASS (rules file)
  5 goose markers PASS (4 CONCURRENTLY files with NO TRANSACTION, one statement each)
  6 Down restores PASS with one note (below) · 7 query coverage PASS · 8 no cross-context FK PASS
  9 association scope PASS (governance/finance tables; identity directory tables scoped by account or оператор, ADR-0003)
  10 sqlc PASS (compile in a scratch copy)
  roundtrip: 27 -> up (47) -> down-to 27 -> up, on a throwaway postgres:18.6-alpine container;
             diff(27, down) differs only in the column position of accounts.association_id
             (re-added last by file 20's Down); diff(up, up2) empty
  rls smoke: as a member of spilno_app — no association set: 0 rows; association A: only A;
             actor email set: own roster entries in both associations; insert into B with A set: refused;
             UPDATE on application_extracts: permission denied
```

## TBDs and open points

- File 20 Down (`NNNNN_drop_accounts_association_id.sql`, Down section) fills the re-added column with the
  seeded association: exact only while that is the only association; after a second association a rollback
  past file 20 needs a Go script that picks the association per account from the roster.
- PRD §8: retention of rejected заявки and their extracts (default ≥ 5 years; no column depends on it).
- `CurrentMember` compares `sign_in_email` exactly while the RLS own-entry clause uses `lower()`;
  normalise addresses to lower case on write in governance (implementation task).
