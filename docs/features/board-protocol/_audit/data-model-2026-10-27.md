# Audit — generate-data-model board-protocol (--mode brownfield), 2026-10-27

## Generated files

- [`data-model.md`](../data-model.md) — ER (31 entities), entity tables, indexes with the query each serves, staged-migration list, test fixtures.
- 23 staged migration pairs (`.up.sql` + `.down.sql`) under [`migrations/`](../migrations/), `20261027120000_create_associations` … `20261027120022_create_integrity_runs`, plus [`migrations/backfill-meetings-association-id.md`](../migrations/backfill-meetings-association-id.md).
- `.claude/rules/migrations.md` — **bootstrapped** from the course baseline (it was missing), with a "spilno specifics" section: goose format, `NO TRANSACTION` for `CONCURRENTLY`, the ADR-0007 trigger exception, legacy tables left as they are. Edit it if the team disagrees with any default.

**Migrations are staged — not yet in the live `services/core/migrations/` tree; `implement-tasks` promotes them (re-stamping the version at promote-time).** At promote time each pair is merged into one goose file: `-- +goose Up` + the `.up.sql`, `-- +goose Down` + the `.down.sql`; files 10 and 11 also get `-- +goose NO TRANSACTION`. File 21 already carries `-- +goose StatementBegin/End` around its function.

No architecture map (`docs/architecture-map.md` is absent): the migration location and tool were taken from repo ADR 0004, `services/core/sqlc.yaml` and SAD §2 instead of a re-scan.

## Verification run

On a throwaway `postgres:18.6-alpine` container (removed afterwards): existing `00001`–`00003` Up sections + one seeded meeting → all 23 `.up.sql` in order → all 23 again (idempotency, no errors) → backfill left 0 NULLs. Triggers: UPDATE, DELETE, TRUNCATE on `protocol_snapshots` and DELETE, TRUNCATE on `protocols` rejected; UPDATE on `protocols` allowed; a second non-withdrawn protocol for the same meeting rejected by `uq_protocols_meeting_not_withdrawn`. On a fresh database: all ups, then all downs in reverse → only `contracts` and `meetings` remain, `meetings` with its original columns, 2 indexes (the PKs), no leftover function.

## Self-check (4/4)

| Check | Result |
|---|---|
| Naming | Pass — plural snake_case tables (`meeting_attendance` renamed to `meeting_attendees` during the check); `<timestamp>_<verb>_<entity>` filenames; every `.up.sql` has its `.down.sql` |
| down.sql reversibility | Pass — every CREATE TABLE / INDEX / TRIGGER and ADD COLUMN has its DROP; proven by the down run above |
| FK indexes | Pass — every `REFERENCES` column has an index; 13 are covered by the leading column of a UNIQUE or composite PK (`agenda_items.protocol_id`, `poll_rounds.protocol_id`, `print_copies.protocol_id`, `meeting_facts.protocol_id`, `protocol_snapshots.protocol_id`, …) |
| Forbidden features | Pass with **one documented exception** — no `CHECK`, no business `DEFAULT`, no sequence PK, no `updated_at` / `deleted_at`; `CREATE OR REPLACE TRIGGER` ×4 in `20261027120021_create_protocol_immutability_triggers.up.sql` lines 14, 18, 22, 26 — ADR-0007, kept by product-owner decision 2026-10-27 |

## Convention divergences

| # | Course default vs repo / SAD | What was done | Who decides |
|---|---|---|---|
| 1 | Repo uses sequential goose versions (`00001`–`00003`); SAD §2 says "next migration number `00004`". Course default is timestamps. | Staged with timestamps. goose orders by integer version, so `20261027…` sorts after `00003`, but the tree would mix two schemes. | `implement-tasks` at promote: keep timestamps (`goose create` default) or re-stamp to `00004`… |
| 2 | Course pairs `.up.sql` / `.down.sql`; goose reads one file with `-- +goose Up/Down`. | Pairs staged; merged at promote (see above). | — |
| 3 | Course forbids `TRIGGER`; ADR-0007 requires two. | Kept, plus TRUNCATE guards (row triggers do not fire on TRUNCATE, so ADR-0007's "no delete" would otherwise have a hole). Recorded in `.claude/rules/migrations.md`. | Decided 2026-10-27 |
| 4 | Existing `contracts` / `meetings` use `CHECK`, `updated_at` + `set_updated_at()` trigger, `DEFAULT gen_random_uuid()`. | Left as they are (brownfield); new tables follow the course defaults — UUID v7 generated in Go (`uuid.NewV7`), no DB default. | — |
| 5 | ADR-0017 puts the counter in `association_settings.next_protocol_number`. | Settings became append-only revisions (decision 2026-10-27), so the counter moved to its own row `protocol_registers`. The locking query is the same, on another table. | **ADR-0017 needs an amendment note**; SAD §5 table list too (see below) |
| 6 | SAD §5 "board roster with terms (`board_members`)" as one mutable table. | `board_members` (one person, one term; stable id) + append-only `board_member_revisions` (decision 2026-10-27). | SAD §5 update |
| 7 | SAD §8 / QG-1 write `governance.protocol_snapshots`, `identity.google_connections`. | Read as context ownership: tables stay in `public` like `meetings`; no Postgres schemas introduced. | — |
| 8 | Course: PK is UUID v7. | Three pure join tables (`assistant_turn_documents`, `meeting_attendees`, `meeting_item_votes`) use a composite PK — no identity of their own. | — |
| 9 | ADR-0003: session "random id". | The cookie token is random; only its SHA-256 (`sessions.token_hash`) is stored, so a database leak yields no usable cookie. | — |

## Drift findings

**Schema vs source** (domain layer read: `internal/governance/domain/meeting.go`, `adapters/postgres/queries/meetings.sql`):

- `meetings.association_id`, `meetings.chair_board_member_id` — column-without-field once phase 1 lands. Expected: `implement-tasks` adds them to `domain.Meeting` and the sqlc queries. No `_drift/*.sql` proposed — the source must follow the schema, not the other way round.
- **Breaking for existing code:** `CreateMeeting` inserts without `association_id`; after phase 3 (`SET NOT NULL`) it fails. `GetMeeting` / `ListMeetings` do not filter by association (ADR-0015). Covered in the backfill note.
- No field-without-column, type or nullability mismatch for `meetings`; all other tables are new and have no domain code yet.

**Model vs spec** (PRD §4–§5 Stage 1, SAD §5 ownership table, §6 flows):

- Every SAD §5 table has a table: `meetings`, `association_settings`, `board_members`, `cases`, `case_documents`, `protocols`, `assistant_turns`, `poll_rounds`, `poll_votes`, `print_copies`, `meeting_facts`, `protocol_snapshots`, integrity runs (`integrity_runs`), `associations`, `accounts`, `sessions`, `google_connections`.
- Tables added without a name in SAD §5, each with an origin: `protocol_registers` (ADR-0017 + decision 2026-10-27), `board_member_revisions` (PRD §6.1 author of each change + decision 2026-10-27), `files` (ADR-0005 per-file metadata), `protocol_transitions` (SAD §8 audit trail, PRD §6.1), `agenda_items` / `agenda_item_versions` (AC-03, AC-29, ADR-0011), `decision_slots` / `slot_values` / `slot_confirmations` (AC-21, AC-22, AC-28b, AC-09; SAD §5 "слоти рішення and confirmations"), `assistant_turn_documents` (AC-26), `meeting_attendees` / `meeting_item_votes` (AC-14, AC-34, ADR-0018), `registry_entries` (AC-18; ADR-0007 "search indexes derived from snapshots"), `integrity_findings` (Flow 10).
- `meetings` gains `chair_board_member_id`: AC-28 and ADR-0011 render the головуючий "from the meeting record", which had no such column.
- **SAD has no §6.4 ER stub** — relationships were derived from §5, §6 and the ADRs. The ER in `data-model.md` can be linked from SAD §5.
- Stage 2 entities (правила статуту with effective dates, checks, overrides, шаблони формулювання, export) are intentionally absent (PRD §3).
- Flow 12 hint "board_members looked up by (association, sign-in email) and by term covering a date" → `idx_board_member_revisions_association_email` + `idx_board_member_revisions_member_created`; Flow 13 hint "protocol_snapshots by counterparty, amount, subject and meeting date" → served by the `registry_entries` projection (date index only; text and amount filters unindexed at this volume — see data-model.md "Deliberately not indexed").

## Breaking changes decomposed

- `meetings.association_id` NOT NULL on an existing table → 3 steps: 07 expand (nullable) → 08 backfill → 09 contract, with [`backfill-meetings-association-id.md`](../migrations/backfill-meetings-association-id.md). Both new indexes on `meetings` are `CONCURRENTLY`, one statement per file (10, 11).
- No rename, drop or retype.

## TBDs

| Where | What |
|---|---|
| `data-model.md:364` | `associations.name`: no PRD AC or flow edits it — api-forge decides whether the settings carry it |
| `data-model.md:388` | Expired-session cleanup has no flow — add a periodic delete and an `expires_at` index only if specified |
| `data-model.md:424` | How the first член правління is created before anyone can sign in (ops command; kept out of migrations by the PII guard) — break-tasks |
| `data-model.md:431` | `association_settings.signers` free text vs a list of roles — follows PRD §8 «who signs the protocol» |
| `data-model.md:738` | River schema migration generated at promote time from the pinned River version (ADR-0006) |

## Follow-ups for the upstream artifacts

- ADR-0017: amendment note — the counter is `protocol_registers.next_number`, not `association_settings.next_protocol_number`.
- SAD §5 ownership table: add `protocol_registers`, `board_member_revisions`, `files`, `protocol_transitions`, `registry_entries`; link `data-model.md` as the ER (§6.4 is absent).
- SAD §2: "Next migration number `00004`" depends on the naming choice at promote (divergence 1).
