-- governance: Idempotency-Key of createAssistantTurn (AC-26, review finding 3). Nullable for good:
-- assistant turns and turns written before this migration have no key, so no backfill is needed.
ALTER TABLE assistant_turns ADD COLUMN IF NOT EXISTS idempotency_key UUID;
