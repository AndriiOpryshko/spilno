# Backfill — `meetings.association_id`

Staged migrations 07 → 08 → 09 (expand → backfill → contract), ADR-0015.

| Phase | File | Effect |
|---|---|---|
| 1 expand | `20261027120007_add_association_and_chair_to_meetings` | `association_id UUID NULL`, `chair_board_member_id UUID NULL` FK → `board_members` |
| 2 backfill | `20261027120008_backfill_meetings_association` | every `NULL` → bootstrap association `00000000-0000-7000-8000-000000000001` |
| 3 contract | `20261027120009_set_meetings_association_not_null` | `SET NOT NULL` |

**App code (dual-write, not a trigger).** From phase 1 on, `CreateMeeting` in `internal/governance/adapters/postgres/queries/meetings.sql` must insert `association_id` (from `ActorFrom(ctx)`), and `GetMeeting` / `ListMeetings` must filter on it. Phase 3 fails every `CreateMeeting` that does not — that query ships in the same release as phase 1 at the latest.

**Volume and ETA.** `meetings` holds a few dozen rows (15-20 board meetings a year since the association was founded). One `UPDATE` in one transaction, well under a second; no batching needed. The `WHERE association_id IS NULL` filter makes a re-run a no-op, so it is resumable.

**One release or three?** goose applies every pending migration on API start and spilno runs one API replica (SAD §7), so phases 1–3 can go out in one release **only if** the `CreateMeeting` change is in that release. Otherwise promote 07–08 first and 09 in the next release. `implement-tasks` decides when it promotes.

**Rollback.** Each phase has a `.down.sql`: 09 drops `NOT NULL`, 08 sets the backfilled ids back to `NULL`, 07 drops both columns.

**Verify after phase 2:** `SELECT count(*) FROM meetings WHERE association_id IS NULL;` → `0`.
