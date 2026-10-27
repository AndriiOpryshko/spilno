-- Existing table: CONCURRENTLY, one statement, outside a transaction (goose: NO TRANSACTION).
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_meetings_association_created
    ON meetings (association_id, created_at DESC, id DESC);
