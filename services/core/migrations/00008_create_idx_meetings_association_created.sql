-- +goose NO TRANSACTION
-- +goose Up
-- Serves ListMeetings: keyset page newest first within one association.
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_meetings_association_created
    ON meetings (association_id, created_at DESC, id DESC);

-- +goose Down
DROP INDEX CONCURRENTLY IF EXISTS idx_meetings_association_created;
