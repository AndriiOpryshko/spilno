-- assistant_turns is live: build the index without a write lock. One statement per file; the goose
-- file gets `-- +goose NO TRANSACTION`. NULL keys (assistant turns) never collide.
CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS uq_assistant_turns_association_idempotency_key
    ON assistant_turns (association_id, idempotency_key);
