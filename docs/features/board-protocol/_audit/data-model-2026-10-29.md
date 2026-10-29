# Audit — generate-data-model board-protocol (drift pass + review finding 3), 2026-10-29

## Generated files

- [`data-model.md`](../data-model.md): `assistant_turns.idempotency_key` (ER, entity table, index table, staged-migration rows 23–24).
- 2 staged migration pairs under [`migrations/`](../migrations/):
  - `20261029120000_add_idempotency_key_to_assistant_turns.{up,down}.sql`
  - `20261029120001_create_uq_assistant_turns_idempotency_key.{up,down}.sql`: `CREATE UNIQUE INDEX CONCURRENTLY`, one statement; the goose file needs `-- +goose NO TRANSACTION`.

**These migrations are staged and not yet in the live `services/core/migrations/` tree.** T25 promotes them (`/promote-migration board-protocol`). They land after the live `00027`, as a delta on the table that `00019` already created.

Everything else was already in place. `data-model.md` and the staged pairs 00–22 exist, and all of them are promoted (`00004`–`00027`, T01–T05). So the full generation pass was skipped.

## Why this delta

Review [finding 3](../_review/review-2026-10-29.md) (AC-26): `createAssistantTurn` requires `Idempotency-Key` (24 h), and nothing stored it. [`api-sync-report.md`](../contracts/api-sync-report.md) assigned the column to this skill.

| Decision | Reason |
|---|---|
| `UUID`, not the `VARCHAR(64)` the sync report suggested | The contract header is `format: uuid` (`services/core/api/openapi.yaml` `IdempotencyKey`) |
| Nullable for good, no 3-step | Assistant turns and older rows have no key. The column never becomes NOT NULL |
| UNIQUE `(association_id, idempotency_key)`, no TTL cleanup | Two concurrent retries hit the index instead of racing. The key is kept longer than the promised 24 h, which is harmless for random client keys. Read with the replay lookup in T25 |
| `CONCURRENTLY` | `assistant_turns` is live (`00019`) |

**Verification:** run on a throwaway `postgres:18.6-alpine` against a minimal `assistant_turns`. Results: up ×2 (idempotent) → several NULL keys coexist → a duplicate key is rejected by `uq_assistant_turns_association_idempotency_key` → down ×1 restores the table without the column or index. The container was removed.

## Self-check (new files, 4/4)

| Check | Result |
|---|---|
| Naming | Pass: `<timestamp>_<verb>_<entity>`, each `.up` has its `.down` |
| down reversibility | Pass: DROP INDEX CONCURRENTLY / DROP COLUMN, proven above |
| FK indexes | n/a: no `REFERENCES` added |
| Forbidden features | Pass: no `CHECK`, `TRIGGER`, business `DEFAULT` |

## Drift findings

**Schema vs source.** The domain layer was read: `internal/governance/domain/{meeting,board_member,settings,protocol}.go` and `adapters/postgres/queries/meetings.sql`. The schema is the source of truth, so no `_drift/*.sql` is proposed. The code must follow the schema.

| Table.column | Domain | Kind | Owner |
|---|---|---|---|
| `meetings.chair_board_member_id` | no field in `domain.Meeting`; CreateMeeting drops it | column-without-field | T22 (finding 2) |
| `meetings.association_id` | no field; the adapter writes the constant `stage1Association`; `GetMeeting` does not filter by association (ADR-0015) | column-without-field | T11 (association from the session) |
| `protocol_transitions` row with `from_state` NULL (creation) | `NewProtocol` ignores actor and time and records no transition | missing write | T15 follow-up (finding 8) |
| `protocol_transitions.actor_member_id` UUID NULL (NULL = worker) | `ProtocolTransition.ActorID string` | type + nullability mismatch | T15 follow-up / T20: use `uuid.NullUUID` or map `""` ↔ NULL in the repository |
| `protocols.number` INTEGER NULL | `int`, 0 = none | nullability (mapping) | T20: 0 ↔ NULL in the repository |
| `board_member_revisions.sign_in_email` VARCHAR(320) NULL, lowercased in app | `string`, `""` = none, not lowercased | nullability (mapping) + missing normalisation | T10 / finding 12 |
| `association_settings.register_start_number` NOT NULL | `0` means missing, `Validate` accepts 0 | value-range mismatch | T09 follow-up (finding 4: `< 1` invalid) |
| `id`, `association_id`, `created_by_member_id`, `created_at` on roster, settings and protocols | not in the domain structs | expected: the repositories are not written yet | T10, T20 |

Lengths match: `full_name` 200, `sign_in_email` 320, `meeting_form` / `signers` 500, meeting `title` / `location` 200. Protocol states (`draft`, `ready_for_vote`, `ready_for_signing`, `signed`, `withdrawn`) fit `VARCHAR(32)` and match the partial index `state <> 'withdrawn'`.

**Model vs spec.** PRD and SAD are unchanged since the 2026-10-27 pass (no commits after `0b3bb45`), so that report still holds. One gap is closed: the `Idempotency-Key` storage. Still open: review finding 16, a `sessions` exemption to record in data-model.md, which this pass did not change.

## Breaking changes decomposed

None. The new column is nullable for good, and the index is `CONCURRENTLY`.

## TBDs

Unchanged from the [2026-10-27 report](./data-model-2026-10-27.md). The `board_members` TBD (first член правління) now has an owner: T08, review finding 1.

## Task plan updated (review findings 1, 2, 3, 6, 7)

`tasks.json`, the task files, the `_epic.md` rows and `tracker.md` were updated for T08, T11, T22, T25, T37, T41 and T43. T08 gains deps T02 and T09. The DAG is checked acyclic (44 tasks).

## Follow-ups

- `/sdlc:api-forge board-protocol --reconcile` moves the `Idempotency-Key` origin from `medium` to `high` (data-model column exists).
